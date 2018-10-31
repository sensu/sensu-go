# Sensu 2.0

TravisCI: [![TravisCI Build Status](https://travis-ci.org/sensu/sensu-go.svg?branch=master)](https://travis-ci.org/sensu/sensu-go)

CircleCI: [![CircleCI Build Status](https://circleci.com/gh/sensu/sensu-go/tree/master.svg?style=svg)](https://circleci.com/gh/sensu/sensu-go/tree/master)

Sensu is an open source monitoring tool for ephemeral infrastructure
and distributed applications. It is an agent based monitoring system
with built-in auto-discovery, making it very well-suited for cloud
environments. Sensu uses service checks to monitor service health and
collect telemetry data. It also has a number of well defined APIs for
configuration, external data input, and to provide access to Sensu's
data. Sensu is extremely extensible and is commonly referred to as
"the monitoring router".

To learn more about Sensu, [please visit the
website](https://sensu.io/).

## What is Sensu 2.0?

Sensu 2.0 is a complete rewrite of Sensu in Go, with new capabilities
and reduced operational overhead. It eliminates several sources of
friction for new and experienced Sensu users.

## Installation

Sensu 2.0 installer packages are available for a number of computing
platforms (e.g. Debian/Ubuntu, RHEL/Centos, etc), but the easiest way
to get started is with the official Docker image, sensu/sensu.

Please note the following installation steps to get Sensu up and
running on your local workstation with Docker.

_NOTE: the following instructions are based on Docker Community
Edition (CE), though they may be easily adapted for other container
platforms. Please download and install Docker CE before proceeding._

1. Start the Sensu 2.0 Backend process

   ```
   $ docker run -d --name sensu-backend \
   -p 2380:2380 -p 3000:3000 -p 8080:8080 -p 8081:8081 \
   sensu/sensu:2.0.0-beta.3 sensu-backend start
   ```

2. Start the Sensu 2.0 Agent process

   ```
   $ docker run -d --name sensu-agent --link sensu-backend \
   sensu/sensu:2.0.0-beta.3 sensu-agent start \
   --backend-url ws://sensu-backend:8081 \
   --subscriptions workstation,docker
   ```

3. Download and install the Sensu 2.0 CLI tool

   **On macOS**

   ```
   $ latest=$(curl -s https://storage.googleapis.com/sensu-binaries/latest.txt)

   $ curl -LO https://storage.googleapis.com/sensu-binaries/$latest/darwin/amd64/sensuctl

   $ chmod +x sensuctl

   $ sudo mv sensuctl /usr/local/bin/
   ```

   **On Debian/Ubuntu Linux**

   ```
   $ curl -s \
   https://packagecloud.io/install/repositories/sensu/nightly/script.deb.sh \
   | sudo bash

   $ sudo apt-get install sensu-cli
   ```

   **On RHEL/CentOS Linux**

   ```
   $ curl -s \
   https://packagecloud.io/install/repositories/sensu/nightly/script.rpm.sh \
   | sudo bash

   $ sudo yum install sensu-cli
   ```

4. Configure the Sensu 2.0 CLI tool

   ```
   $ sensuctl configure
   ? Sensu Backend URL: http://127.0.0.1:8080
   ? Username: admin
   ? Password: P@ssw0rd!
   ? Namespace: default
   ? Preferred output format: tabular
   ```

5. List Sensu 2.0 Entities

   ```
   $ sensuctl entity list
   ```

Congratulations! You now have a local Sensu 2.0 deployment!

To learn more about Sensu 2.0 and what you can do with it, please
check out the [official project documentation](https://docs.sensu.io/sensu-core/2.0/).

## Getting Started

Now that you have a [local Sensu 2.0 deployment up and
running](#installation), it's time to configure your first Sensu
monitoring check! Sensu checks are commands (or scripts) that allow
you to monitor server resources, services, and application health, as
well as collect & analyze metrics.

1. Register a Sensu 2.0 Asset for the check executable

   ```
   $ sensuctl asset create check-plugins \
   --url https://github.com/portertech/sensu-plugins-go/releases/download/0.0.1/sensu-check-plugins.tar.gz \
   --sha512 4e6f621ebe652d3b0ba5d4dead8ddb2901ea03f846a1cb2e39ddb71b8d0daa83b54742671f179913ed6c350fc32446a22501339f60b8d4e0cdb6ade5ee77af16
   ```

2. Create a check to monitor Google via ICMP from your workstation

   ```
   $ sensuctl check create google \
   --runtime-assets check-plugins \
   --command "check-ping -h google.ca -P 80" \
   --subscriptions workstation --interval 10 --timeout 5
   ```

## Contributing

To make a good faith effort to ensure the criteria of the MIT License
are met, Sensu Inc. requires the Developer Certificate of Origin (DCO)
process to be followed.

For guidelines on how to contribute to this project and more
information on the DCO, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Development

Sensu is written in Go, and targets the latest stable release of the Go
compiler. To work on Sensu, you will need the latest release of Go.

[Go installation instructions](https://golang.org/doc/install)

### Protobuf

#### Overview

We are using the version **proto3** of the protocol buffers language. Here are some useful resources:

[To learn more about protocol buffers](https://developers.google.com/protocol-buffers/docs/overview)

[The proto3 language guide](https://developers.google.com/protocol-buffers/docs/proto3)

#### Installation

Install the protobuf compiler since we don't use the one that golang uses.
```
brew install protobuf
```
Otherwise, see the **for non-C++ users** [instructions here.](https://github.com/google/protobuf#protocol-compiler-installation)

#### Quick Start

Once you make a change to any `*.proto` file within the **types** package, you will need to regenerate the associated `*.pb.go` file. To do so, simply run `go generate` on the package.

### Dependencies

Sensu uses [golang/dep](https://github.com/golang/dep) for managing its
dependencies. You will need to install the latest stable version of dep in
order to install Sensu's dependencies.

[Dep releases](https://github.com/golang/dep/releases)

#### Usage

Running the following will pull all required dependencies, including static
analysis and linter tools.

```shell
./build.sh deps
```

Later, if you would like to add a dependency, run:

```shell
dep ensure -add https://my-repo.com/my/dep
```

If you would like to update a dependency, run:

```shell
dep ensure -update https://my-repo.com/my/dep
```

When you would like to remove a dependency, remove the it from `Gopkg.toml` and
then run:

```shell
dep prune
```

#### Further Reading

- [The Saga of Go Dependency Management](https://blog.gopheracademy.com/advent-2016/saga-go-dependency-management/)
- [`dep` Usage](https://github.com/golang/dep#usage)

## Building

### Docker

The simplest way to the build Sensu is with the `sensu-go-build` image. The
image contains all the required tools to build the agent, backend and sensuctl.

```sh
docker pull sensu/sensu-go-build
docker run -it -e GOOS=darwin -v `pwd`:/go/src/github.com/sensu/sensu-go --entrypoint='/go/src/github.com/sensu/sensu-go/build.sh' sensu/sensu-go-build
```

If you would like to build for different platforms and architectures use GOOS
and GOARCH env variables. See [Optional environment variables](https://golang.org/doc/install/source#environment) for more.

When complete your binaries will be present in the `target` directory.

### Manually

First ensure that you have the required tools installed to build the programs.

* Ensure that you have the Go tools installed and your environment configured.
  If not follow the official
  [Install the Go tools](https://golang.org/doc/install#install) guide.
* When building the Sensu backend you will need NodeJS and Yarn installed so
  that the web UI may be included in the binary. Follow
  [Installing Node.js](https://nodejs.org/en/download/package-manager/) and
  [Yarn Installation](https://yarnpkg.com/en/docs/install) for installation
  instructions for your platform.

Once all the tools are installed you are now ready to use the build script. To
build the Sensu backend, agent and sensuctl, run:

```sh
./build.sh build
```

Each product can built separately, with one of the following:

```sh
./build.sh build_agent
./build.sh build_backend
./build.sh build_cli
```

By default the web UI is built along side and bundled into the backend, as this
can be a time intensive process, we provide an escape hatch. Use the `dev` tag
to avoid building the web UI.

```sh
./build.sh build_backend -tags dev
```

## Testing

Install dependencies:

```shell
./build.sh deps
```

Run test suites:

```shell
./build.sh ci
```

Run end-to-end tests:

```shell
./build.sh e2e

To run a specific test:

./build.sh e2e -run TestRBAC

To prevent tests from running in parallel:

./build.sh e2e -parallel 1
```
