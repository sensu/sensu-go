#!/usr/bin/env bash
set -o pipefail
set -e

# On CircleCI, we want to see what commands are being run, at least for now
test -n "$CIRCLECI" && set -x

REPO_PATH="github.com/sensu/sensu-go"
DASHBOARD_PATH="dashboard"

eval $(go env)

cmd=${1:-"all"}

RACE=""

VERSION_CMD="go run ./version/cmd/version/version.go"

set_race_flag() {
    if [ "$GOARCH" == "amd64" ] && [ "$CGO_ENABLED" == "1" ]; then
        RACE="-race"
    fi
}

case "$GOOS" in
    darwin)
        set_race_flag
        ;;
    freebsd)
        set_race_flag
        ;;
    linux)
        set_race_flag
        ;;
    windows)
        set_race_flag
        ;;
esac

cmd_name_map() {
    local cmd=$1

    case "$cmd" in
        backend)
            echo "sensu-backend"
            ;;
        agent)
            echo "sensu-agent"
            ;;
        cli)
            echo "sensuctl"
            ;;
    esac
}

build_tool_binary () {
    local goos=$1
    local goarch=$2
    local cmd=$3
    local subdir=$4

    local outfile="target/${goos}-${goarch}/${subdir}/${cmd}"

    GOOS=$goos GOARCH=$goarch go build -o $outfile ${REPO_PATH}/${subdir}/${cmd}/...

    echo $outfile
}

build_binary () {
    # Unset GOOS and GOARCH so that the version command builds on native OS/arch
    unset GOOS
    unset GOARCH

    local version=$($VERSION_CMD -v)
    local prerelease=$($VERSION_CMD -p)

    local goos=$1
    local goarch=$2
    local cmd=$3
    local cmd_name=$4
    local ext="${@:5}"

    local outfile="target/${goos}-${goarch}/${cmd_name}"

    local build_date=$(date +"%Y-%m-%dT%H:%M:%S%z")
    local build_sha=$(git rev-parse HEAD)

    local version_pkg="github.com/sensu/sensu-go/version"
    local ldflags=" -X $version_pkg.Version=${version}"
    local ldflags+=" -X $version_pkg.PreReleaseIdentifier=${prerelease}"
    local ldflags+=" -X $version_pkg.BuildDate=${build_date}"
    local ldflags+=" -X $version_pkg.BuildSHA=${build_sha}"
    local main_pkg="cmd/${cmd_name}"

    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags "${ldflags}" $ext -o $outfile ${REPO_PATH}/${main_pkg}

    echo $outfile
}

build_tool () {
    local cmd=$1
    local subdir=$2

    if [ ! -d bin/${subdir} ]; then
        mkdir -p bin/${subdir}
    fi

    echo "Building $subdir/$cmd for ${GOOS}-${GOARCH}"
    out=$(build_tool_binary $GOOS $GOARCH $cmd $subdir)
    rm -f bin/$(basename $out)
    cp ${out} bin/${subdir}
}

build_commands () {
    echo "Build all commands..."

    build_agent $@
    build_backend $@
    build_cli $@
}

build_agent() {
    build_command agent $@
}

build_backend() {
    build_dashboard $@
    build_command backend $@
}

build_cli() {
    build_command cli $@
}

build_dashboard() {
    echo "Building web UI"
    GOOS=$HOST_GOOS GOARCH=$HOST_GOARCH go generate $@ ./dashboard
}

build_command () {
    local cmd=$1
    local cmd_name=$(cmd_name_map $cmd)
    local ext="${@:2}"

    if [ ! -d bin/ ]; then
        mkdir -p bin/
    fi

    echo "Building $cmd for ${GOOS}-${GOARCH}"
    out=$(build_binary $GOOS $GOARCH $cmd $cmd_name $ext)
    rm -f bin/$(basename $out)
    cp ${out} bin
}

linter_commands () {
    echo "Running linter..."

    go get honnef.co/go/tools/cmd/megacheck
    go get gopkg.in/alecthomas/gometalinter.v2
    go get github.com/gordonklaus/ineffassign
    go get github.com/jgautheron/goconst/cmd/goconst

    megacheck $(go list ./... | grep -v dashboardd | grep -v agent/assetmanager | grep -v scripts)
    if [ $? -ne 0 ]; then
        echo "Linting failed..."
        exit 1
    fi

    gometalinter.v2 --vendor --disable-all --enable=vet --enable=ineffassign --enable=goconst --tests ./...
    if [ $? -ne 0 ]; then
        echo "Linting failed..."
        exit 1
    fi

    # Make sure every package has a LICENSE
    go list ./... | sed -n '1!p' | sed -e 's_github.com/sensu/sensu-go/__g' | xargs -I '{}' stat '{}'/LICENSE > /dev/null
}

unit_test_commands () {
    echo "Running unit tests..."

    go test -timeout=60s $RACE $(go list ./... | egrep -v '(testing|vendor|scripts)')
    if [ $? -ne 0 ]; then
        echo "Unit testing failed..."
        exit 1
    fi
}

integration_test_commands () {
    echo "Running integration tests..."

    go test -timeout=180s -tags=integration $(go list ./... | egrep -v '(testing|vendor|scripts)')
    if [ $? -ne 0 ]; then
        echo "Integration testing failed..."
        exit 1
    fi
}

e2e_commands () {
    echo "Running e2e tests..."
    go test ${REPO_PATH}/testing/e2e $@
}

docker_commands () {
    local cmd=$1

    if [ "$cmd" == "build" ]; then
        docker_build ${@:2}
    elif [ "$cmd" == "push" ]; then
        docker_push ${@:2}
    elif [ "$cmd" == "deploy" ]; then
        docker_build
        docker_push ${@:2}
    else
        docker_commands build $@
    fi
}

docker_build() {
    local build_sha=$(git rev-parse HEAD)
    local ext=$@

    # When publishing image, ensure that we can bundle the web UI.
    build_dashboard $ext

    for cmd in agent backend cli; do
        echo "Building $cmd for linux-amd64"
        local cmd_name=$(cmd_name_map $cmd)
        build_binary linux amd64 $cmd $cmd_name $ext
    done

    # build the docker image with master tag
    docker build --label build.sha=${build_sha} -t sensu/sensu:master .
}

docker_push() {
    local release=$1
    local version=$(echo sensu/sensu:$($VERSION_CMD -v))
    local version_iteration=$(echo sensu/sensu:$($VERSION_CMD -v).$($VERSION_CMD -i))

    # ensure we are authenticated
    docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

    # push master - tags and pushes latest master docker build only
    if [ "$release" == "master" ]; then
        docker push sensu/sensu:master
        exit 0
    fi

    if [ "$release" == "nightly" ]; then
        docker tag sensu/sensu:master sensu/sensu:nightly
        docker push sensu/sensu:nightly
        exit 0
    fi

    # if versioned release push to 'latest' tag
    if [ "$release" == "versioned" ]; then
        docker tag sensu/sensu:master sensu/sensu:latest
        docker push sensu/sensu:latest
    fi

    # push current revision
    docker tag sensu/sensu:master $version_iteration
    docker push $version_iteration
    docker tag $version_iteration $version
    docker push $version
}

test_dashboard() {
    pushd "${DASHBOARD_PATH}"
    yarn install
    yarn test
    popd
}

prompt_confirm() {
    read -r -n 1 -p "${1} [y/N] " response
    case "$response" in
        [yY][eE][sS]|[yY])
            true
            ;;
        *)
            false
            ;;
    esac
}

check_deploy() {
    echo "Checking..."

    # Prompt for confirmation if deploying outside Travis
    if [[ -z "${TRAVIS}" || "${TRAVIS}" != true ]]; then
        prompt_confirm "You are trying to deploy outside of Travis. Are you sure?" || exit 0
    fi
}

deploy() {
    local release=$1

    echo "Deploying..."

    # Authenticate to Google Cloud and deploy binaries
    if [[ "${release}" != "nightly" ]]; then
        openssl aes-256-cbc -K $encrypted_abd14401c428_key -iv $encrypted_abd14401c428_iv -in gcs-service-account.json.enc -out gcs-service-account.json -d
        gcloud auth activate-service-account --key-file=gcs-service-account.json
        ./build-gcs-release.sh
    fi

    # Deploy system packages to PackageCloud
    make clean
    docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
    docker pull sensu/sensu-go-build
    docker run -it -e SENSU_BUILD_ITERATION=$SENSU_BUILD_ITERATION -e CI=$CI -v `pwd`:/go/src/github.com/sensu/sensu-go sensu/sensu-go-build
    docker run -it -v `pwd`:/go/src/github.com/sensu/sensu-go -e PACKAGECLOUD_TOKEN="$PACKAGECLOUD_TOKEN" -e SENSU_BUILD_ITERATION=$SENSU_BUILD_ITERATION -e CI=$CI sensu/sensu-go-build publish_travis
    docker run -it -v `pwd`:/go/src/github.com/sensu/sensu-go sensu/sensu-go-build clean

    # Deploy Docker images to the Docker Hub
    docker_build
    docker_push $release
}

case "$cmd" in
    "build")
        build_commands "${@:2}"
        ;;
    "build_agent")
        build_agent "${@:2}"
        ;;
    "build_backend")
        build_backend "${@:2}"
        ;;
    "build_cli")
        build_cli "${@:2}"
        ;;
    "dashboard")
        test_dashboard
        ;;
    "dashboard-ci")
        test_dashboard
        ./codecov.sh -t $CODECOV_TOKEN -cF javascript -s dashboard
        ;;
    "deploy")
        check_deploy
        deploy "${@:2}"
        ;;
    "docker")
        docker_commands "${@:2}"
        ;;
    "e2e")
        # Accepts specific test name. E.g.: ./build.sh e2e -run TestAgentKeepalives
        build_command agent
        build_command backend
        build_command cli
        e2e_commands "${@:2}"
        ;;
    "lint")
        linter_commands
        ;;
    "none")
        echo "noop"
        ;;
    "quality")
        unit_test_commands
        ;;
    "unit")
        unit_test_commands
        ;;
    "integration")
        integration_test_commands
        ;;
    *)
        unit_test_commands
        integration_test_commands
        build_commands
        e2e_commands
        ;;
esac
