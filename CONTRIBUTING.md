# Contributing

We, the maintainers, love pull requests from everyone, but often find
we must say "no" despite how reasonable the proposal may seem.

For this reason, we ask that you open an issue to discuss proposed
changes prior to submitting a pull request for the implementation.
This helps us to provide direction as to implementation details, which
branch to base your changes on, and so on.

1. Open an issue to describe your proposed improvement or feature
1. Fork https://github.com/sensu/sensu-go
1. Clone the fork to $GOPATH/src/github.com/sensu/sensu-go. (This exact path must be used. If GOPATH is undefined, use /home/go.)
1. Create your feature branch (`git checkout -b my-new-feature`)
1. If applicable, add a [CHANGELOG.md entry](#changelog) describing your change.
1. Commit your changes with a [DCO Signed-off-by statement](#dco) (`git commit --signoff`)
1. Push your feature branch (`git push origin my-new-feature`)
1. Create a Pull Request as appropriate based on the issue discussion

## DCO

To make a good faith effort to ensure the criteria of the MIT License
are met, Sensu Inc. requires the Developer Certificate of Origin (DCO)
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

Git makes this easy with `git commit --signoff`!

The DCO text can either be manually added to your commit body, or you
can add either `-s` or `--signoff` to your usual git commit commands.
If you forget to add the sign-off you can also amend a previous commit
with the sign-off by running `git commit --amend -s`. If you've pushed
your changes to Github already you'll need to force push your branch
after this with `git push -f`. -- Thanks Chef!

## Changelog

The Sensu [Changelog](CHANGELOG.md) is based on the Sensu Community 
[Changelog guidelines](https://github.com/sensu-plugins/community/blob/master/HOW_WE_CHANGELOG.md).

All new changes go underneath the _Unreleased_ heading at the top of the Changelog.
Beyond that, here are some additional guidelines that should make it more clear where your
change goes in the Changelog.

### Added

Any _new_ functionality goes here. This may be a new field on a data type or a new data
type altogether; a new API endpoint; or possibly a whole new feature. In general, these
are sentences that start with the word "added."

Examples:

- `begin` field to silences that initiates silencing at a given timestamp
- /healthz endpoint that reports health of the sensu-agent process

### Changed

Changes to any existing component or functionality of the system that does not cause
breaking changes to users or developers go here. _Changed_ is distinguishable from
_Fixed_ in that it is an intentional change to existing functionality.

Examples:

- `sensu-agent` exits gracefully instead of crashing upon disconnect
- Refactored the API to use reusable controller logic

### Fixed

Fixed bugs go here.

Examples:

- `sensu-agent` no longer ignores keepalive configuration
- Don't delete auth tokens at startup

### Deprecated

Deprecated should include any soon-to-be removed functionality. An entry here that
is user facing will likely yield entries in _Removed_ or _Breaking_ eventually.

Examples:

- The /health API endpoint is being replaced by /healthz on the backend
- The /stash API endpoint is being removed in a future release

### Removed

Removed is for the removal of functionality that does not directly impact users,
these entries most likely only impact developers of Sensu.  If user facing
functionality is removed, an entry should be added to the _Breaking Changes_
section instead.

Examples:

- Removed references to `encoding/json` in favor of `json-iter`.
- Removed unused `Store` interface for `BlobStore`.

### Security

Any fixes to address security exploits should be added to this section. If
available, include an associated CVE entry.

Examples:

- Upgraded build to use Go 1.9.1 to address [CVE-2017-15041](https://www.cvedetails.com/cve/CVE-2017-15041/)
- Fixed issue where users could view entities without permission

### Breaking Changes

Whenever you have to make a change that will cause users to be unable to
upgrade versions of Sensu without intervention by an operator, your change
goes here. Try to avoid these. If they're required, we should have documented
justification in a GitHub issue and preferably a proposal. We should also bump
minor versions at this time.

Examples:

- Refactored how Checks are stored in Etcd, `sensu-backend migrate` is required to upgrade
