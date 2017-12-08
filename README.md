# Sensu 2.0

[![Build Status](https://travis-ci.com/sensu/sensu-go.svg?token=bQ4K7jzHALx4myyBoqcu&branch=master)](https://travis-ci.com/sensu/sensu-go)

[Engineering Wiki](https://github.com/sensu/engineering/wiki)

## Contributing/Development

To make a good faith effort to ensure the criteria of the MIT License
are met, Sensu Inc requires the Developer Certificate of Origin (DCO)
process to be followed.

The DCO is an attestation attached to every contribution made by every
developer. In the commit message of the contribution, the developer
simply adds a Signed-off-by statement and thereby agrees to the DCO,
which you can find below or at http://developercertificate.org/.

```
Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

The following is an example DCO Signed-off-by statement.

```
 Author: Sean Porter <sean@sensu.io>

 Committer: Greg Poirier <greg@sensu.io>

   Let's name it WizardFormat.

   Calling it the Sensu Metric Format was a mistake.

   Signed-off-by: Sean Porter <sean@sensu.io>
   Signed-off-by: Grep Poirier <greg@sensu.io>
```

The general development process is:

1. Fork this repo and clone it to your workstation.
2. Create a feature branch for your change.
3. Write code and tests.
4. Commit your changes with a DCO Signed-off-by statement.
5. Push your feature branch to github and open a pull request against master.

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
```
