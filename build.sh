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
    build_command backend $@
}

build_cli() {
    build_command cli $@
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

docker_commands () {
    local cmd=$1

    if [ "$cmd" == "build" ]; then
        docker_build ${@:2}
    else
        docker_commands build $@
    fi
}

docker_build() {
    # install sensuctl to make sure it's current with the docker build
    go install ./cmd/sensuctl

    local build_sha=$(git rev-parse HEAD)
    local ext=$@

    for cmd in agent backend cli; do
        echo "Building $cmd for linux-amd64"
        local cmd_name=$(cmd_name_map $cmd)
        build_binary linux amd64 $cmd $cmd_name $ext
    done

    # build the docker image with master tag
    docker build --label build.sha=${build_sha} -t sensu/sensu:master .
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
        ;;
    "docker")
        docker_commands "${@:2}"
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
        ;;
esac
