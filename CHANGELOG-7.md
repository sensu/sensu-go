# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking
- Embedded etcd is no longer supported, all related configuration has been
removed.
- The prefix that keepalives are stored under has now changed. This could lead
to dangling references when using an older Sensu database with 7.x, if the
software is upgraded when there are active keepalive failures.
- Etcd client configuration options have changed.
- Entity configuration can now be stored in PostgreSQL. Existing entity
configuration will not be migrated from Etcd.
- Namespaces can now be stored in PostgreSQL. Existing namespaces will not be
migrated from Etcd.
- PostgreSQL >= 9.6 is now required.

### Added
- Developer mode can now be enabled with the --dev flag.
- Added sensu-backend configuration for postgresql.
- Added configuration store selector to sensu-backend.
- Added postgresql state store.
- GlobalResource interface in core/v3 allows core/v3 resources to
be marked as global resources.

### Fixed
- Fixed an issue where multi-expression exclusive "Deny" filters were not
  evaluated as described in the documentation.

### Changed
- Changed parameters for `sensuctl cluster-role create` to be plural
- Deregistration events are now silenced if a silenced entry exists matching the
entity subscriptions and/or a check named `deregistration`.
- Upgraded Go version from 1.17.1 to 1.18.1.
- Changed sensu-backend etcd configuration options.
- Upgraded etcd version from 3.5.2 to 3.5.4.

### Removed
- Removed sensu-backend upgrade command. May make an appearance again in later versions.
