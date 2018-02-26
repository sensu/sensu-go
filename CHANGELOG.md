# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- Add CLI support for adhoc check requests.
- Check scheduler now handles adhoc check requests.
- Added `set-FIELD` and `remove-FIELD` commands for all updatable fields
of a check. This allows updating single fields and completely clearing out
non-required fields.
- Add built-in only_check_output mutator to pipelined.
- Allow publish, cron, ttl, timeout, low flap threshold and more fields to be
set when importing legacy settings.
- Add CPU architecture in system information of entities.
- The `sensuctl user change-password` subcommand now accepts flag parameters.
- Configured and enabled etcd autocompaction.
- Add event metrics type, implementing the Sensu Metrics Format.
- Agents now try to reconnect to the backend if the connection is lost.
- Added non-functional selections for resolving and silencing to web ui
- Add LastOk to check type. This will be updated to reflect the last timestamp
of a successful check.

### Changed
- Refactor Check data structure to not depend on CheckConfig. This is a breaking
change that will cause existing Sensu alpha installations to break if upgraded.
This change was made before beta release so that further breaking changes could
be avoided.
- Make indentation in protocol buffers files consistent.
- Refactor Hook data structure. This is similar to what was done to Check,
except that HookConfig is now embedded in Hook.
- Refactor CheckExecutor and AdhocRequestExecutor into an Executor interface.
- Changed the sensu-backend etcd flag constants to match the etcd flag names.
- Upgraded to Etcd v3.3.1
- Removed 3DES from the list of allowed ciphers in the backend and agent.
- Password input fields are now aligned in  `sensuctl user change-password`
subcommand.
- Agent backend URLs without a port specified will now default to port 8081.
- Travis encrypted variables have been updated to work with travis-ci.org
- Upgraded all builds to use Go 1.10.
- Use megacheck instead of errcheck.
- Cleaned agent configuration.

### Fixed
- Fixed a bug in time.InWindow that in some cases would cause subdued checks to
be executed.
- Fixed a bug in the HTTP API where resource names could not contain special
characters.
- Resolved a bug in the keepalive monitor timer which was causing it to
erroneously expire.
- Resolved a bug in how an executor processes checks. If a check contains proxy
requests, the check should not duplicately execute after the proxy requests.
- Removed an erroneous validation statement in check handler.

## [2.0.0-alpha.17] - 2018-02-13
### Added
- Add .gitattributes file with merge strategy for the Changelog.
- Context switcher added for dashboard.
- Add API support for adhoc check requests.
- Check scheduler now supports round-robin scheduling.
- Added better error checking for CLI commands and support for mutually
exclusive fields.
- Added `--interactive` flag to CLI which is required to run interactive mode.
- Added CLI role rule-add Organization and Environment interactive prompts.
- Added events page list and simple buttons to filter

### Changed
- Silenced `begin` supports human readable time (Format: Jan 02 2006 3:04PM MST)
in `sensuctl` with optional timezone. Stores the field as unix epoch time.
- Increased the timeout in the store's watchers tests.
- Incremental retry mechanism when waiting for agent and backend in e2e tests.
- Renamed CLI asset create interactive prompt "Org" to "Organization".

### Fixed
- Fixed required flags in `sensuctl` so requirements are enforced.
- Add support for embedded fields to dynamic.Marshal.

## [2.0.0-alpha.16] - 2018-02-07
### Added
- Add an e2e test for proxy check requests.
- Add integration tests to our CI.
- Context switcher added for dashboard
- Add api support for adhoc check requests.

### Fixed
- Tracks in-progress checks with a map and mutex rather than an array to
increase time efficiency and synchronize goroutines reading from and writing
to that map.
- Fixed a bug where we were attempting to kill processes that had already
finished before its allotted execution timeout.
- Fixed a bug where an event could erroneously be shown as silenced.
- Properly log errors whenever a check request can't be published.
- Fixed some build tags for tests using etcd stores.
- Keepalive monitors now get updated with changes to a keepalive timeout.
- Prevent tests timeout in queue package
- Prevent tests timeout in ring package
- Fixed a bug in the queue package where timestamps were not parsed correctly.
- Fixed Ring's Next method hanging in cases where watch events are not propagated.

### Changed
- Queues are now durable.
- Refactoring of the check scheduling integration tests.
- CLI resource delete confirmation is now `(y/N)`.

### Removed
- Dependency github.com/chzyer/readline

## [2.0.0-alpha.15] - 2018-01-30
### Added
- Add function for matching entities to a proxy check request.
- Added functions for publishing proxy check requests.
- Added proxy request validation.
- CLI functionality for proxy check requests (add set-proxy-requests command).
- Entities have been added to the state manager and synchronizer.
- Added package leader, for facilitating execution by a single backend.
- Proxy check requests are now published to all entities described in
`ProxyRequests` and `EntityAttributes`.
- Add quick navigation component for dashboard

### Changed
- Govaluate logic is now wrapped in the `util/eval` package.
- Cron and Interval scheduling are now mutually exclusive.

### Fixed
- Fixed a bug where retrieving check hooks were only from the check's
organization, rather than the check's environment, too.

## [2.0.0-alpha.14] - 2018-01-23
### Added
- Add `Timeout` field to CheckConfig.
- CLI functionality for check `Timeout` field.
- Add timeout support for check execution.
- Add timeout support for check hook execution.
- Token substitution is now available for check hooks
- Add an e2e test for logging redaction
- Support for `When` field in `Filter` which enables filtering based on days
and times of the week.
- New gRPC inspired GraphQL implementation. See
[graphql/README](backend/apid/graphql/README.md) for usage.
- Support for TTLs in check configs to monitor stale check results.

### Changed
- Moved monitor code out of keepalived and into its own package.
- Moved KeyBuilder from etcd package to store package.

## [2.0.0-alpha.13] - 2018-01-16
### Added
- Logging redaction for entities

### Changed
- Removed the Visual Studio 2017 image in AppVeyor to prevent random failures

### Fixed
- Fixed e2e test for token substitution on Windows
- Fixed check subdue unit test for token substitution on Windows
- Consider the first and last seconds of a time window when comparing the
current time
- Fixed Travis deploy stage by removing caching for $GOPATH
- Parse for [traditional cron](https://en.wikipedia.org/wiki/Cron) strings, rather than [GoDoc cron](https://godoc.org/github.com/robfig/cron) strings.

### Changed
- Removed the Visual Studio 2017 image in AppVeyor to prevent random failures
- Made some slight quality-of-life adjustments to build-gcs-release.sh.

### Fixed
- Fixed e2e test for token substitution on Windows
- Fixed check subdue unit test for token substitution on Windows
- Consider the first and last seconds of a time window when comparing the
current time
- Fixed Travis deploy stage by removing caching for $GOPATH
- Parse for [traditional cron](https://en.wikipedia.org/wiki/Cron) strings, rather than [GoDoc cron](https://godoc.org/github.com/robfig/cron) strings.

## [2.0.0-alpha.12] - 2018-01-09
### Added
- Add check subdue mechanism. Checks can now be subdued for specified time
windows.
- Silenced entries now include a `begin` timestamp for scheduled maintenance.
- Store clients can now use [watchers](https://github.com/sensu/sensu-go/pull/792) to be notified of changes to objects in the store.
- Add check `Cron` field. Checks can now be scheduled according to the cron
string stored in this field.
- Add a distributed queue package for use in the backend.
- Token substitution is now available for checks.
- CLI functionality for check `Cron` field.
- Add an e2e test for cron scheduling.
- Add an e2e test for check hook execution.

## [2.0.0-alpha.11] - 2017-12-19
### Breaking Changes
- The `Source` field on a check has been renamed to `ProxyEntityID`. Any checks
using the Source field will have to be recreated.

### Added
- Silenced entries with ExpireOnResolve set to true will now be deleted when an
event which has previously failing was resolved
- TCP/UDP sockets now accept 1.x backward compatible payloads. 1.x Check Result gets translated to a 2.x Event.
- Custom attributes can be added to the agent at start.
- New and improved Check Hooks are implemented (see whats new about hooks here: [Hooks](https://github.com/sensu/sensu-alpha-documentation/blob/master/08-hooks.md))
- Add check subdue CLI support.

### Changed
- Avoid using reflection in time.InWindows function.
- Use multiple parallel jobs in CI tools to speed up the tests
- Pulled in latest [github.com/coreos/etcd](https://github.com/coreos/etcd).
- Includes fix for panic that occurred on shutdown.
- Refer to their
[changelog](https://github.com/gyuho/etcd/blob/f444abaae344e562fc69323c75e1cf772c436543/CHANGELOG.md)
for more.
- Switch to using [github.com/golang/dep](https://github.com/golang/dep) for
managing dependencies; `vendor/` directory has been removed.
- See [README](README.md) for usage.

## [2.0.0-alpha.10] - 2017-12-12
### Added
- End-to-end test for the silencing functionality
- Silenced events are now identified in sensuctl

### Changed
- Events that transitioned from incidents to a healthy state are no longer
filtered by the pipeline
- Errcheck was added to the build script, and the project was given a once-over
to clean up existing errcheck lint.
- Creating a silenced entry via sensuctl no longer requires an expiry value

### Fixed
- Entities can now be silenced using their entity subscription
- Fixed a bug in the agent where it was ignoring keepalive interval and timeout
settings on start
- Keepalives now alert when entities go away!
- Fixed a bug in package dynamic that could lead to an error in json.Marshal
in certain cases.
- Fixed an issue in keepalived to handle cases of nil entities in keepalive
messages

## [2.0.0-alpha.9] - 2017-12-5
### Added
- Proxy entities are now dynamically created through the "Source" attribute of a
check configuration
- Flag to sensuctl configure allowing it to be configured non-interactively
(usage: --non-interactive or -n)
- New function SetField in package dynamic, for setting fields on types
supporting extended attributes.
- Automatically append entity:entityID subscription for agent entities
- Add silenced command to sensuctl for silencing checks and subscriptions.
- Add healthz endpoint to agent api for checking agent liveness.
- Add ability to pass JSON event data to check command STDIN.
- Add POST /events endpoint to manually create, update, and resolve events.
- Add "event resolve" command to sensuctl to manually resolve events.
- Add the time.InWindow & time.InWindows functions to support time windows, used
in filters and check subdue

### Fixed
- Fixed a bug in how silenced entries were deleted. Only one silenced entry will
be deleted at a time, regardless of wildcard presence for subscription or check.

## [2.0.0-alpha.8] - 2017-11-28
### Added
- New "event delete" subcommand in sensuctl
- The "Store" interface is now properly documented
- The incoming request body size is now limited to 512 KB
- Silenced entries in the store now have a TTL so they automatically expire
- Initial support for custom attributes in various Sensu objects
- Add "Error" type for capturing pipeline errors
- Add registration events for new agents
- Add a migration tool for the store directly within sensu-backend

### Changed
- Refactoring of the sensu-backend API
- Modified the description for the API URL when configuring sensuctl
- A docker image with the master tag is built for every commit on master branch
- The "latest" docker tag is only pushed once a new release is created

### Fixed
- Fix the "asset update" subcommand in sensuctl
- Fix Go linting in build script
- Fix querying across organizations and environments with sensuctl
- Set a standard redirect policy to sensuctl HTTP client

### Removed
- Removed extraneous GetEnv & GetOrg getter methods
