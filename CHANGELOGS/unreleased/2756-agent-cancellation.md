### Fixed
- Fixed a bug where agents would sometimes refuse to terminate on SIGTERM and
SIGINT.
- Fixed a bug where agents would always try to reconnect to the same backend,
even when multiple backends were specified. Agents will now try to connect to
other backends, in pseudorandom fashion.
