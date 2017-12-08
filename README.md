# Sensu 2.0

[![Build Status](https://travis-ci.com/sensu/sensu-go.svg?token=bQ4K7jzHALx4myyBoqcu&branch=master)](https://travis-ci.com/sensu/sensu-go)

## Contributing/Development

To make a good faith effort to ensure the criteria of the MIT License
are met, Sensu Inc. requires the Developer Certificate of Origin (DCO)
process to be followed.

For guidelines on how to contribute to this project and more
information on the DCO, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Protobuf

### Overview

We are using the version **proto3** of the protocol buffers language. Here's some useful ressources:

[To learn more about protocol buffers](https://developers.google.com/protocol-buffers/docs/overview)

[The proto3 language guide](https://developers.google.com/protocol-buffers/docs/proto3)


### Installation

Install the protobuf compiler since we don't use the one that golang uses.
```
brew install protobuf
```
Otherwise, see the **for non-C++ users** [instructions here.](https://github.com/google/protobuf#protocol-compiler-installation)

### Quick Start

Once you make a change to any `*.proto` file within the **types** package, you will need regenerate the associated `*.pb.go` file. To do so, simply run the [genproto.sh](https://github.com/sensu/sensu-go/blob/master/scripts/genproto.sh) script, which will install all required dependencies and launch the code generation.

## Testing

Run test suites:

```
./build.sh ci
```

Run end-to-end tests:

```
./build.sh e2e

# To run a specific test:

./build.sh e2e -run TestRBAC

# To prevent tests from running in parallel:

./build.sh e2e -parallel 1
```
