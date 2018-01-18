# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- Add `Timeout` field to CheckConfig.
- CLI functionality for check `Timeout` field.
- Add timeout support for check execution.
- Add timeout support for check hook execution.

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
