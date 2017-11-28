#!/usr/bin/env bash
set -o pipefail
set -e

REPO_PATH="github.com/sensu/sensu-go"
DASHBOARD_PATH="dashboard"

eval $(go env)

cmd=${1:-"all"}

RACE=""

set_race_flag() {
	if [ "$GOARCH" == "amd64" ]; then
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

install_deps () {
	go get github.com/axw/gocov/gocov
	go get gopkg.in/alecthomas/gometalinter.v1
	go get github.com/gordonklaus/ineffassign
	go get github.com/jgautheron/goconst/cmd/goconst
	go get -u github.com/golang/lint/golint
}

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

	GOOS=$goos GOARCH=$goarch go build -i -o $outfile ${REPO_PATH}/${subdir}/${cmd}/...

	echo $outfile
}

build_binary () {
	local goos=$1
	local goarch=$2
	local cmd=$3
	local cmd_name=$4

	local outfile="target/${goos}-${goarch}/${cmd_name}"

	local version=$(cat version/version.txt)
	local prerelease=$(cat version/prerelease.txt)
	local build_date=$(date +"%Y-%m-%dT%H:%M:%S%z")
	local build_sha=$(git rev-parse HEAD)

	local version_pkg="github.com/sensu/sensu-go/version"
	local ldflags=" -X $version_pkg.Version=${version}"
	local ldflags+=" -X $version_pkg.PreReleaseIdentifier=${prerelease}"
	local ldflags+=" -X $version_pkg.BuildDate=${build_date}"
	local ldflags+=" -X $version_pkg.BuildSHA=${build_sha}"

	CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -ldflags "${ldflags}" -i -o $outfile ${REPO_PATH}/${cmd}/cmd/...

	echo $outfile
}

build_tools () {
	echo "Running tool & plugin builds..."

	for cmd in cat false sleep true; do
		build_tool $cmd "tools"
	done

	for cmd in slack; do
		build_tool $cmd "handlers"
	done
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
	echo "Running build..."

	for cmd in agent backend cli; do
		build_command $cmd
	done
}

build_command () {
	local cmd=$1
	local cmd_name=$(cmd_name_map $cmd)

	if [ ! -d bin/ ]; then
		mkdir -p bin/
	fi

	echo "Building $cmd for ${GOOS}-${GOARCH}"
	out=$(build_binary $GOOS $GOARCH $cmd $cmd_name)
	rm -f bin/$(basename $out)
	cp ${out} bin
}

linter_commands () {
	echo "Running linter..."

	gometalinter.v1 --vendor --disable-all --enable=vet --enable=golint --enable=ineffassign --enable=goconst --tests ./...
	if [ $? -ne 0 ]; then
		echo "Linting failed..."
		exit 1
	fi
}

unit_test_commands () {
	echo "Running tests..."

	echo "" > coverage.txt
	for pkg in $(go list ./... | egrep -v '(testing|vendor)'); do
		go test -timeout=60s -v $RACE -coverprofile=profile.out -covermode=atomic $pkg
		if [ $? -ne 0 ]; then
		  echo "Tests failed..."
		  exit 1
		fi

		if [ -f profile.out ]; then
			cat profile.out >> coverage.txt
			rm profile.out
		fi
	done
}

e2e_commands () {
	echo "Running e2e tests..."
	go test -v ${REPO_PATH}/testing/e2e $@
}

docker_commands () {
  # make this one var (push or release - master or versioned)
	local push=$1
	local release=$2
	local build_sha=$(git rev-parse HEAD)

	for cmd in cat false sleep true; do
		echo "Building tools/$cmd for linux-amd64"
		build_tool_binary linux amd64 $cmd "tools"
	done

	for cmd in slack; do
		echo "Building handlers/$cmd for linux-amd64"
		build_tool_binary linux amd64 $cmd "handlers"
	done

	# install_dashboard_deps
	# build_dashboard
	# bundle_static_assets

	for cmd in agent backend cli; do
		echo "Building $cmd for linux-amd64"
		local cmd_name=$(cmd_name_map $cmd)
		build_binary linux amd64 $cmd $cmd_name
	done

	docker build --label build.sha=${build_sha} -t sensuapp/sensu-go:master .

	# push master - tags and pushes latest master docker build only
	if [ "$push" == "push" ] && [ "$release" == "master" ]; then
		docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
		docker push sensuapp/sensu-go:master
	# push versioned - tags and pushes with version pulled from
	# version/prerelease/iteration files
	elif [ "$push" == "push" ] && [ "$release" == "versioned" ]; then
		docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
		local version_alpha=$(echo sensuapp/sensu-go:$(cat version/version.txt)-alpha)
		local version_alpha_iteration=$(echo sensuapp/sensu-go:$(cat version/version.txt)-$(cat version/prerelease.txt).$(cat version/iteration.txt))
        docker tag sensuapp/sensu-go:master sensuapp/sensu-go:latest
        docker push sensuapp/sensu-go:latest
		docker tag sensuapp/sensu-go:master $version_alpha_iteration
		docker push $version_alpha_iteration
		docker tag $version_alpha_iteration $version_alpha
		docker push $version_alpha
	fi
}

check_for_presence_of_yarn() {
  if hash yarn 2>/dev/null; then
    echo "Yarn is installed, continuing."
  else
    echo "Please install yarn to build dashboard."
    exit 1
  fi
}

install_dashboard_deps() {
  go get -u github.com/UnnoTed/fileb0x
  check_for_presence_of_yarn
  pushd "${DASHBOARD_PATH}"
  yarn install
  yarn precompile
  popd
}

test_dashboard() {
  pushd "${DASHBOARD_PATH}"
  yarn lint
  yarn test --coverage
  popd
}

build_dashboard() {
  pushd "${DASHBOARD_PATH}"
  yarn install
  yarn precompile
  yarn build
  popd
}

bundle_static_assets() {
	fileb0x ./.b0x.yaml
}

if [ "$cmd" == "deps" ]; then
	install_deps
elif [ "$cmd" == "quality" ]; then
	linter_commands
	unit_test_commands
elif [ "$cmd" == "lint" ]; then
	linter_commands
elif [ "$cmd" == "unit" ]; then
	unit_test_commands
elif [ "$cmd" == "build_tools" ]; then
	build_tools
elif [ "$cmd" == "e2e" ]; then
	# Accepts specific test name. E.g.: ./build.sh e2e -run TestAgentKeepalives
	build_commands
	e2e_commands "${@:2}"
elif [ "$cmd" == "ci" ]; then
  subcmd=${2:-"go"}
  if [[ "$subcmd" == "dashboard" ]]; then
    install_dashboard_deps
    test_dashboard
  elif [[ "$subcmd" == "none" ]]; then
    echo "noop"
  else
    linter_commands
    unit_test_commands
    build_commands
    e2e_commands
  fi
elif [ "$cmd" == "coverage" ]; then
  subcmd=${2:-"go"}
  if [ "$subcmd" == "dashboard" ]; then
    ./codecov.sh -t $CODECOV_TOKEN -cF javascript -s dashboard
  elif [ "$subcmd" == "none" ]; then
    echo "noop"
  else
    ./codecov.sh -t $CODECOV_TOKEN -cF go
  fi
elif [ "$cmd" == "build" ]; then
	build_commands
elif [ "$cmd" == "docker" ]; then
	docker_commands "${@:2}"
elif [ "$cmd" == "build_agent" ]; then
	build_command agent
elif [ "$cmd" == "build_backend" ]; then
	build_command backend
elif [ "$cmd" == "build_cli" ]; then
	build_command cli
elif [ "$cmd" == "build_dashboard" ]; then
	install_dashboard_deps
	build_dashboard
	bundle_static_assets
else
	install_deps
	linter_commands
	build_tools
	unit_test_commands
	build_commands
	e2e_commands
fi
