# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Breaking
- Embedded etcd is no longer supported, all related configuration has been
removed.
- The prefix that keepalives are stored under has now changed. This could lead
to dangling references when using an older Sensu database with 7.x, if the
software is upgraded when there are active keepalive failures.

### Added
- Developer mode can now be enabled with the --dev flag.
