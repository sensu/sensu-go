## Unreleased
### Added
- Proxy entities are now dynamically created through the "Source" attribute of a
check configuration
- Flag to sensuctl configure allowing it to be configured non-interactively (usage:
--non-interactive or -n)

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
