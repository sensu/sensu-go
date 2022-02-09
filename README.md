# Sensu Go

CircleCI: [![CircleCI Build Status](https://circleci.com/gh/sensu/sensu-go/tree/main.svg?style=svg)](https://circleci.com/gh/sensu/sensu-go/tree/main)

Sensu is an open source monitoring tool for ephemeral infrastructure
and distributed applications. It is an agent based monitoring system
with built-in auto-discovery, making it very well-suited for cloud
environments. Sensu uses service checks to monitor service health and
collect telemetry data. It also has a number of well defined APIs for
configuration, external data input, and to provide access to Sensu's
data. Sensu is extremely extensible and is commonly referred to as
"the monitoring router".

To learn more about Sensu, [please visit the
website](https://sensu.io/) and [read the documentation](https://docs.sensu.io/sensu-go/latest/).

## What is Sensu Go?

Sensu Go is a complete rewrite of Sensu in Go, with new capabilities
and reduced operational overhead. It eliminates several sources of
friction for new and experienced Sensu users.

The original Sensu required external services like Redis or RabbitMQ.
Sensu Go can rely on an embedded etcd datastore for persistence, making
the product easier to get started with. External etcd services can also be
used, in the event that you already have them deployed.

Sensu Go replaces Ruby expressions with JavaScript filter expressions, by
embedding a JavaScript interpreter.

Unlike the original Sensu, Sensu Go events are always handled, unless
explicitly filtered.

## Installation

Sensu Go installer packages are available for a number of computing
platforms (e.g. Debian/Ubuntu, RHEL/Centos, etc), but the easiest way
to get started is with the official Docker image, sensu/sensu.

See the [installation documentation](https://docs.sensu.io/sensu-go/latest/operations/deploy-sensu/install-sensu/)
to get started with one of Sensu's [supported packages](https://docs.sensu.io/sensu-go/latest/platforms/).

**NOTE**: Starting with Sensu Go version 6.0, for instances built from source, the web UI is now a [standalone product](https://github.com/sensu/web) — it is no longer included with the Sensu backend. To build the web UI from source, use the [installation instructions](https://github.com/sensu/web/blob/master/INSTALL.md) in the Sensu Go Web repository.

### Building from source

The various components of Sensu Go can be manually built from this repository.
You will first need [Go >= 1.16](https://golang.org/doc/install#install)
installed. Then, you should clone this repository **outside** of the GOPATH
since Sensu Go uses [Go Modules](https://github.com/golang/go/wiki/Modules):
```
$ git clone https://github.com/sensu/sensu-go.git
$ cd sensu-go
```

To compile and then run Sensu Go within a single step:
```
$ go run ./cmd/sensu-agent
$ go run ./cmd/sensu-backend
$ go run ./cmd/sensuctl
```

To build Sensu Go binaries and save them into the `bin/` directory:
```
$ go build -o bin/sensu-agent ./cmd/sensu-agent
$ go build -o bin/sensu-backend ./cmd/sensu-backend
$ go build -o bin/sensuctl ./cmd/sensuctl
```

To build Sensu Go binaries with the version information:
```
# sensu-agent
$ go build -ldflags '-X "github.com/sensu/sensu-go/version.Version=5.14.0" -X "github.com/sensu/sensu-go/version.BuildDate=2019-10-08" -X "github.com/sensu/sensu-go/version.BuildSHA='`git rev-parse HEAD`'"' -o bin/sensu-agent ./cmd/sensu-agent

# sensu-backend
$ go build -ldflags '-X "github.com/sensu/sensu-go/version.Version=5.14.0" -X "github.com/sensu/sensu-go/version.BuildDate=2019-10-08" -X "github.com/sensu/sensu-go/version.BuildSHA='`git rev-parse HEAD`'"' -o bin/sensu-backend ./cmd/sensu-backend

# sensuctl
$ go build -ldflags '-X "github.com/sensu/sensu-go/version.Version=5.14.0" -X "github.com/sensu/sensu-go/version.BuildDate=2019-10-08" -X "github.com/sensu/sensu-go/version.BuildSHA='`git rev-parse HEAD`'"' -o bin/sensuctl ./cmd/sensuctl
```

## Contributing

For guidelines on how to contribute to this project, how to hack on Sensu, and
information about what we require from project contributors, please see
[CONTRIBUTING.md](CONTRIBUTING.md).

Sensu is and always will be open source, and we continue to highly
value community contribution. The packages we’re releasing for new
versions are from our Enterprise repo; Sensu Go is the upstream for
Sensu Enterprise (as they’d say in the Go community: Sensu Go is
vendored into the Sensu Enterprise Go repo). We encourage you to
download new versions, as the functionality will be identical to what
you find in the public repo, and access to the enterprise-only
features can be unlocked with a license key. Because these releases
are in our Enterprise repo, there may be times that you don’t see the
actual work being done on an issue you open, but that doesn’t mean
we’re not working on it! Our team is committed to updating progress on
open issues in the sensu-go repo, even if that work is being done in
our Enterprise repo.
