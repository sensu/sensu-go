# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Breaking
- Etcd is no longer supported. All persistent data is now stored in postgresql.
- The enterprise postgresql event store is no longer supported. Events are now
  stored in a different format in postgresql, and postgresql is configured with
  the command line or configuration file (--pg-dsn). Additional environment
  variables are supported, see the libpq documentation for details.
- core/v2.Namespace is no longer supported. Users must use core/v3.Namespace,
which now allows for the addition of labels and annotations.
- API keys can no longer be retrieved from the database.
- The REST APIs now use the wrapped resource format.
- When multiple asset builds are filtered the first asset build with the highest
number of filters is returned. Previously the first filtered asset build was
returned regardless of filter count.
- Removes the deprecated agent socket APIs
- Disables round robin scheduling. Checks configured with round robin will
  no longer be scheduled.

### Added
- Added sensu-backend configuration for postgresql.
- Added configuration store selectors to sensu-backend, previously an enterprise
  feature.
- Added postgresql support for storing all state and configuration.
- GlobalResource interface in core/v3 allows core/v3 resources to
  be marked as global resources.
- The authentication module now logs successful (INFO) and unsuccessful (ERROR)
  login attempts.
- Added a redesigned keepalive system which does not rely on asynchronous
  notification. Unlike the old system, it is much more resilient in the face of
  mass outages. This system also underpins the redesigned round robin scheduling
  system.
- Added a flag to sensu-agent to prevent the collection of network host
  information. This information can be quite lengthy and can reduce the overall
  system performance if agent entities grow to be too large.
- API keys can now be created with sensuctl create.

### Fixed
- Fixed an issue where multi-expression exclusive "Deny" filters were not
  evaluated as described in the documentation.
- API keys are now securely stored in the database.

### Changed
- Changed parameters for `sensuctl cluster-role create` to be plural
- Deregistration events are now silenced if a silenced entry exists matching the
entity subscriptions and/or a check named `deregistration`.
- Upgraded Go version to 1.19.5. Old Go versions are not supported.
- The sensuctl api-key grant command now returns additional information.
- Handler errors now logged at the error level instead of info level

### Removed
- Removed sensu-backend upgrade command. May make an appearance again in later versions.
- Removed support for older, unsupported postgresql versions.
