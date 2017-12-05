# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- Proxy entities are now dynamically created through the "Source" attribute of a
check configuration
- Flag to sensuctl configure allowing it to be configured non-interactively (usage:
--non-interactive or -n)
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

### Changed
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
