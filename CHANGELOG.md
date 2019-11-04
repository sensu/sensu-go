# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed
- Added better error logging for mutator execution.

### Added
- Added the `APIKey` resource and HTTP API support for POST, GET, and DELETE.
- Added sensuctl commands to manage the `APIKey` resource.
- Added support for api keys to be used in api authentication.
- Added support for sensu-backend service environment variables.

## [5.14.2] - 2019-11-04

### Changed
- Upgraded etcd to 3.3.17

### Fixed
- As a result of upgrading etcd, TLS etcd clients that lose their connection will
successfully reconnect when using --no-embed-etcd.
- Check TTL switches are now correctly buried when associated events and entities
are deleted.
- Keepalive switches are now correctly buried when the keepalive event is deleted.
- Sensu now uses far fewer leases for keepalives and check TTLs, resulting in a
stability improvement for most deployments.
- Fixed a minor UX issue in interactive filter commands in sensuctl.

## [5.14.1] - 2019-10-16

### Added
- Added prometheus gauges for check schedulers.

### Fixed
- Opening an already open Bolt database should not cause sensu-agent to hang
indefinitely.
- [CLI] Dump multiple types as YAML to a file would print separator STDOUT
instead of specified file
- Fixed a bug where Sensu would crash with a panic due to a send on a closed channel.

## [5.14.0] - 2019-10-08

### Added
- [Web] Added an additional option to the error dialog allowing users to
completely wipe the application's persisted state; in lieu of them having to
manually wipe their local / session storage. This may help in the rare cases
where something in said state is leading to an uncaught exception.
- [Web] For operating systems with support for selecting a preferred light /dark
theme, the application now respects the system preference by default.
- sensuctl dump can now list the types of supported resources with --types.
- Added the `sensu_agent_version` field to the `Entity` resource, which reflects
the Sensu semver version of the agent entity.
- Added the `--etcd-heartbeat-interval` and `--etcd-election-timeout` flags to
`sensu-backend`

### Changed
- [Web] Github is not always the best place for feature requests and discussion,
as such we've changed CTA for feedback to point to Discourse instead of the
web repository's issues page.
- [Web] When a user lands on a page inside a namespace that no longer exists or
they do not have access to, the drawer is now opened so that namespace switcher
is front and center. Hopefully this should reduce any confusion around next
steps.
- Support agent TLS authentication, usable with a licensed sensu-backend.
- Updated Go version from 1.12.3 to 1.13.1.
- [GraphQL] `putWrapped` mutation now accepts wrapped JSON with empty
outer objectmeta.

### Fixed
- [Web] Fixed issue where a user with an appropriate role may have been unable
to resolve events, queue checks, and create silenced entries.
- Splayed proxy checks are now executed every interval, instead of every
`interval + interval * splay_coverage`.
- [GraphQL] Ensures that proxy entity label & annotations are redacted.
- Fixed a bug in the ring where round robin schedules would not recover
after quorum loss.
- [Web] Unauthorized errors emitted while creating silences or resolving events
are now caught and a toast is presented to communicate what occurred.
- [Web] Internal errors are now avoided when a user attempts to queue an ad-hoc
check for a keepalive.
- Do not separate asset builds into several assets unless the the tabular format
is used in `sensuctl asset list`.
- Fix the 'flag accessed but not defined' error in `sensuctl asset outdated`
- Fix generic API client's `SetTypeMeta` method. The APIGroup is now correctly
configured and by virtue unintended authorization denied errs are avoided.
- Fixed a bug where checks would stop executing after a network error.
- Fixed a bug where sensuctl create with stdin was not working.

## [5.13.2] - 2019-09-19

### Fixed
- Enterprise bug fix.

## [5.13.1] - 2019-09-10

### Fixed
- Multi-build asset definitions with no matching filters will no longer cause a panic.

## [5.13.0] - 2019-09-09

### Added
- Added the `sensuctl env` command.
- sensuctl asset add (fetches & adds assets from Bonsai).
- sensuctl asset outdated (checks for newer versions of assets from Bonsai).
- Add HTTP and directory support to `sensuctl create`
- Only validate check interval/cron when publish true

### Fixed
- sensuctl dump no longer silently discards errors.
- Interactive check create and update modes now have 'none' as the first
highlighted option, instead of nagios-perfdata.
- Fixed a bug where silences would not expire on event resolution.

## [5.12.0] - 2019-08-22

### Added
- Added functionality for the agent `--allow-list` configuration, which
whitelists check and check hook executables.
- Added the `runtime_assets` field to `HookConfig`. Assets are enabled
for check hook execution.
- Added backwards compatible content negotiation to the websocket connection.
Protobuf will be used for serialization/deserialization unless indicated by the
backend to use JSON.
- Added delete functionality for assets in the API and sensuctl.
- Added `sensuctl dump` to dump resources to a file or STDOUT.
- Added `event.check.name` as a supported field selector.
- [Web] Added timeline chart to event details view.
- Added `entity.system.arm_version` to record the value of `GOARM` at compile time.
- Added `ProviderType` field to `AuthProviderClaims`
- Added `builds` field to the `Asset` type to allow assets to specify different
URLs for each platform/architecture/arch_version.

### Changed
- The project now uses Go modules instead of dep for dependency management.
- The internal reverse proxy relied on by the dashboard has been eliminated.
- The generic etcd watcher now keeps track of revisions.
- The resource caches can now rebuild themselves in case of failures.
- Event and Entity resources can now be created without an explicit namespace;
the system will refer to the namespace in the URL.
- Events and Entities can now be created with the POST verb.
- [Web] Changed styling of namespace labels.
- Log token substitution failures more clearly.

### Fixed
- Fixed the tabular output of `sensuctl filter list` so inclusive filter expressions
are joined with `&&` and exclusive filter expressions are joined with `||`.
- The REST API now correctly only returns events for the specific entity
queried in the `GET /events/:entity` endpoint (#3141)
- Prevent a segmentation fault when running `sensuctl config view` without
configuration.
- Added entity name to the interactive sensuctl survey.
- Check hooks with `stdin: true` now receive actual event data on STDIN instead
  of an empty event.
- Prevent a segmentation fault on the agent when a command execution returns an
error.
- [Web] Fixed issue where a bad or revoked access token could crash the app.

### Removed
- Removed encoded protobuf payloads from log messages (when decoded, they can reveal
redacted secrets).

## [5.11.1] - 2019-07-18

### Fixed
- The agent now sends heartbeats to the backend in order to detect network
failures and reconnect faster.
- The default handshake timeout for the WebSocket connection negotiation has
been lowered from 45 to 15 seconds and is now configurable.

## [5.11.0] - 2019-07-10

### Added
- Silenced entries are now retrieved from the cache when determining if an event
is silenced.
- Added --disable-assets flag to sensu-agent.
- Added ability to query mutators to the GraphQL service
- Added ability to query event filters to the GraphQL service
- Added prometheus metrics for topics in wizard bus and agent sessions.
- The buffer size and worker count of keepalived, eventd & pipelined can now be
configured on sensu-backend.
- Added a `headers` field to the `Asset` struct. Headers is a map of key/value
string pairs used as HTTP headers for asset retrieval.
- Added the current user to the output of `sensuctl config view`.
- [Web] Adds list and details views for mutators
- [Web] Adds list and details views for event filters
- Added sensuctl delete command

### Changed
- [Web] Updated embedded web assets from `46cd0ee` ... `8f50155`
- The REST API now returns the `201 Created` success status response code for
POST & PUT requests instead of `204 No Content`.

### Fixed
- The REST API now returns an error when trying to delete an entity that does
not exist.
- Fixed a bug where basic authorization was not being performed on the agent websocket connection.
- Fixed an aliasing regression where event timestamps from the /events API
were not getting properly populated.
- Fixed a bug where multiple nested set handlers could be incorrectly flagged as
deeply nested.
- Fixed a bug where round robin proxy checks could fail to execute.
- Fixed a bug where watchers could enter a tight loop, causing very high CPU
usage until sensu-backend was restarted.

## [5.10.1] - 2019-06-25

### Fixed
- Fixed the entity_attributes in proxy_requests so all attributes must match
instead of only one of them.
- Fixed a bug where events were not deleted when their corresponding entity was.

## [5.10.0] - 2019-06-18

### Added
- Added POST `/api/core/v2/tessen/metrics`.
- Added the ability in TessenD to listen for metric points on the message bus,
populate, and send them to the Tessen service.
- [Web] Adds ability to delete entities
- [GraphQL] Adds simple auto-suggestion feature.
- Added a tag to all Tessen metrics to differentiate internal builds.
- Added a unique sensu cluster id, accessible by GET `/api/core/v2/cluster/id`.
- Added `sensuctl cluster id` which exposes the unique sensu cluster id.

### Changed
- [Web] Updated embedded web assets from `275386a` ... `46cd0ee`
- Refactoring of the REST API.
- Changed the identifying cluster id in TessenD from the etcd cluster id to
the sensu cluster id.
- [GraphQL] Updates `PutResource` mutation to accept an `upsert` boolean flag parameter. The `upsert` param defaults to `true`, but if set to `false` the mutation will return an error when attempting to create a duplicate resource.
- Eventd has been refactored. Users should not perceive any changes, but a
substantial amount of business logic has been moved into other packages.
- The `sensuctl create` command now accepts resources without a declared
namespace. If the namespace is omitted, the resource will be created in the
current namespace, or overridden by the `--namespace` flag.
- Eventd now uses a constant number of requests to etcd when working with
silenced entries, instead of a number that is proportional to the number of
subscriptions in a check.

### Fixed
- The check state and check total_state_change properties are now more correct.
- Scheduling proxy checks now consumes far fewer CPU resources.
- [Web] Unless required- scrollbars on code blocks are hidden.
- [Web] Ensure that we redirect user to a valid namespace when first signing in.
- [Web] Correctly display timeout value for handlers.
- [Web] Avoid exception when parsing non-standard cron statements. (Eg.
`@every 1h` or `@weekly`)
- The resources metadata are now validated with the request URI.

## [5.9.0] - 2019-05-29

### Added
- [GraphQL] Added field to retrieve REST API representation of a resource to
  each core type
- [Web] Add views for handlers

### Changed
- [Web] Updated embedded web assets from `9d91d7f` ... `275386a`
- [Web] Implements simpler & more efficient filtering.
- [GraphQL] fields that previously accepted a JS filter have been deprecated and
  replaced with a simpler syntax.

### Fixed
- Fixed the behaviors for check `Occurrences` and `OccurrencesWatermark`.
- Fixed a panic that could occur when seeding initial data.
- [Web] Compress dashboard assets
- [Web] Fixed regression where dashboard assets were no longer compressed.
- Fixed listing of silenced entries by check or subscription.
- The docker-compose.yaml file now refers to the sensu/sensu:latest image.

## [5.8.0] - 2019-05-22

### Added
- Added per resource counts to tessen data collection.
- Added event processing counts to tessen data collection.
- Added ability to query for `Handlers` (individual and collections) from the GraphQL query endpoint.
- Added `/version` to retrieve the current etcd server/cluster version and the sensu-backend version.
- --etcd-cipher-suites option is now available for sensu-backend.
- Added the `--chunk-size` flag to `sensuctl * list` sub-commands

### Changed
- eventd and keepalived now use 1000 handlers for events.
- etcd database size and request size are now configurable.
- Most resources now use protobuf serialization in etcd.

### Fixed
- Only bury switchsets of checks that no longer have a TTL, in order to reduce
the number of write operations made to etcd.
- Fixed keepalives switchsets for entities with deregistration.
- Fixed continue token generation in namespace and user pagination.

## [5.7.0] - 2019-05-09

### Added
- Added a Windows service wrapper for sensu-agent. See
"sensu-agent service --help" for more information.

### Fixed
- Fixed `sensuctl` color output on Windows.
- Fixed a regression in `sensuctl cluster` json/wrapped-json output.
- Fixed a regression that caused listing objects for a given namespace to also
  include results from namespaces sharing a similar prefix.

## [5.6.0] - 2019-04-30

### Added
- Added filtering support to `sensuctl`. This feature only works against a
  `sensu-backend` with a valid enterprise license.
- Added fields getter functions for resources available via the REST API.
- Added the message bus to Tessend in order to track Tessen configuration changes from the API.
- Added a performance optimizing `Count()` function to the generic store.
- Added a hexadecimal Cluster ID title to the `sensuctl cluster health` and
`sensuctl cluster member-list` commands in tabular format.
- Added a `Header` field to the `HealthResponse` type returned by `/health`.

### Fixed
- Fixed the agent `--annotations` and `--labels` flags.

## [5.5.1] - 2019-04-15

### Changed
- Added parsing annoatations to sensu-agent, both from agent.yml and command line arguments
- Updated Go version from 1.11.4 to 1.12.3 for CI builds.
- Changed the 1.x `client` field to `source` in the 1.x compatible agent socket. The `client` field is now deprecated.
- Deprecated the agent TCP/UDP sockets in favor of the agent rest api.
- [GraphQL] Added mutation to create / update using wrapped resources.
- [GraphQL] Added field returning wrapped resource given ID.
- apid uses a new generic router for listing resources.
- The store uses the generic List function for listing resources.

### Fixed
- Fixed an issue where etcd watchers were used incorrectly. This was causing
100% CPU usage in some components, as they would loop endlessly trying to get
results from watchers that broke, due to their stream terminating. Other
components would simply stop updating. Watchers now get reinstated when the
client regains connectivity.
- Fixed the `/events/:entity` route in the REST API.
- Fixed a bug where the --labels arg was not working as expected in sensu-agent.

## [5.5.0] - 2019-04-03

### Added
- Added the TessenD daemon.
- Added an etcd watcher for tessen configuration.
- Added ring support for TessenD so that the service is invoked in a
round-robin fashion within a cluster.
- Added `tessen opt-in` command to `sensuctl`.
- Added `tessen opt-out` command to `sensuctl`.
- Added `tessen info` command to `sensuctl`.
- Added more verbose logging to indicate when a proxy request matches an entity according to its entity attributes.

### Removed
- Removed the unused etcd watcher for hook configurations.

### Fixed
- [Web] Ensure user chip is never rendered when creator is not present.

## [5.4.0] - 2019-03-27

### Added
- Add support for pagination to the API
- Add two new flags for `backend` daemon to optionally allow for separate TLS
  cert/key for dashboard. the flags are: `--dashboard-cert-file` and
  `--dashboard-key-file`. The dashboard will use the same TLS config of the API
  unless these new flags are specified.
- Added notion of asset collections to dashboard daemon
- Added a store for Tessen opt-in/opt-out configuration.
- Added /tessen GET and PUT endpoints to the API.
- Added queueing to the agent /events API

### Changed
- [Web] Updated dependencies that had warnings
- [Web] Updated dependency babel to ^7.4
- [Web] Updated UI library to ^3.8

### Fixed
- Fixed a bug in `sensuctl` where global/persistent flags, such as `--namespace`
  and `--config-dir`, would get ignored if they were passed after a sub-command
  local flag, such as `--format`.
- Fixed a bug in `sensuctl` where handlers and filters would only be deleted
  from the default namespace, unless a `--namespace` flag was specified.
- Fixed a bug where events could be stored without a timestamp.
- Fixed a bug where metrics could be persisted to etcd in some cases.
- Fixed a bug where agents would sometimes refuse to terminate on SIGTERM and
  SIGINT.
- Fixed a bug where agents would always try to reconnect to the same backend,
  even when multiple backends were specified. Agents will now try to connect to
  other backends, in pseudorandom fashion.
- [Web] Avoids crash when the creator of a check is inaccessible.
- [Api] Respond with 404 from the users endpoint when user for given name cannot
  be found.
- Commands wrap on the event details page and will display "-" if there is no
  command (keepalives)

## [5.3.0] - 2019-03-11

### Added
- Added additional check config and entity information to event details page.
- Fixed all known TLS vulnerabilities affecting the backend server:
    - TLS min version increased to 1.2
    - Removed ALL but perfect-forward-secrecy ciphers
- Removed requirement of specifying `--trusted-ca-file` when using TLS on backend
- Prevented backend from loading server TLS configuration for http client
- Enforced uniform TLS configuration for all three backend components (apid, agentd, dashboardd)
- Set http client timeout to 15 seconds for sensuctl
- Round robin scheduling is now fully functional.
- Web UI offline state detection and and alert banner.

### Changed
- Asset downloading now uses buffered I/O.

### Fixed
- Check results sent via the agent socket now support handlers.
- `sensuctl user list` can now output yaml and wrapped-json
- Fixed bug with how long commands were displayed on check details page.
- Assets downloads no longer specify a client timeout.
- Fixed a bug where agent entity subscriptions would be communicated to the
  backend incorrectly. Due to the scheduler using the subscriptions from the
  HTTP header, this does not have any effect on scheduling.
- Web - Fixes issue where timeout value was not displayed.
- Fixed bug with how long commands were displayed on check details page.

### Removed
- Removed the concept of "edition" and the edition header.

## [5.2.1] - 2019-02-11

### Fixed
- Fixed a regression in the agent that would not allow proxy checks to be
run for subsequent executions.
### Added
- Web UI - support for labels and annotations

## [5.2.0] - 2019-02-06

### Added
- Added support for the following TLS related options to `sensuctl`:
`--trusted-ca-file` and `--insecure-skip-tls-verify`. This allows sensuctl
users to use a self-signed certificate without adding it to the operating
system's CA store, either by explicitly trusting the signer, or by disabling
TLS hostname verification.
- Added a generic watcher in the store.
- Added `RemoveProvider` method to authenticator.
- Check output truncation support has been added. Check output can be truncated
by adjusting the max_output_size and discard_output properties.
- Added ability to silence/unsilence from the event details page.
- Added support for wrapped resources in the API with `sensuctl create` &
`sensuctl edit`.
- Web UI - platform version displays on the entity details page.
- Web UI - include proxy request configuration on check details page.
- Web UI - display deregistration config on the entity details page.

### Changed
- Removed unused workflow `rel_build_and_test` in CircleCI config.
- Moved the `Provider` interface to `api/core/v2` package.
- Moved the `Authenticator` interface to `backend/authentication` package.
- Updated confirmation messages for sensuctl commands: `Created`, `Deleted` and
`Updated` instead of `OK`.
- Exported some functions and methods in the CLI client.
- The API authenticator now identifies providers by their name only.

### Fixed
- Check TTL failure events are now much more reliable, and will persist even
in the presence cluster member failures and cluster restarts.
- Fix snakeCase version of keys in typeMap for acronyms.
- Fixed a bug in keepalive processing that could result in a crash.
- Pin childprocess to v0.9.0 in CircleCI so fpm can be installed.
- Substitutions applied to command & hooks are now omitted from events.
- Fixes a bug where generic store methods assumed a namespace was provided for non-namespaced resources.
- Keepalive and check TTL database state is now properly garbage-collected on
entity deletion.
- Fixed a bug where `sensuctl version` required configuration files to exist.
- Updates the copy on the confirm disable dialog to accurately reflect the
operation.

## [5.1.1] - 2019-01-24

### Added
- Added the notion of authentication providers.

### Changed
- Improved logging for errors in proxy check requests.
- Updated Go version from 1.10 to 1.11.4.
- Refactoring of the internal authentication mechanism into a `basic`
authentication provider.
- Modified private generic store methods as public functions.
- Improved logging for errors in proxy check requests.
- Updated Go version from 1.10 to 1.11.4.
- Changed keepalive event to include check.output

### Fixed
- Fixed a bug where `sensuctl edit` was not removing the temp file it created.
- Fixed a bug where adhoc checks were not retrieving asset dependencies.
- Fixed a bug where check updates would cause the check to immediately fire.
- Fixed a bug where a bad line in check output would abort metric extraction.
An error is now logged instead, and extraction continues after a bad line is encountered.
- Keepalive events will now continue to fire after cluster restarts.
- Fixed a panic in the dashboardd shutdown routine.
- Fixed a bug where deleting a non-existent entity with sensuctl would not return an error.
- Web UI - toolbar menu buttons now switch with dark theme.
- Web UI - some buttons easier to see with dark theme.
- Agents will now take proxy entity names into consideration when guarding
against duplicate check requests.

### Changed
- Improved logging for errors in proxy check requests.
- Updated Go version from 1.10 to 1.11.4.

## [5.1.0] - 2018-12-18

### Added
- Support for the trusted-ca-file and insecure-skip-tls-verify flags in
  sensu-agent. These flags have the same meaning and use as their sensu-backend
  counterparts.

### Changed
- Default location for sensu-backend data has changed from /var/lib/sensu to
  /var/lib/sensu/sensu-backend. See release notes for more information.

### Fixed
- Keepalive and check TTL failure events now fire continuously until resolved.
- Listing an empty set of assets now correctly returns [] instead of null.
- Fixed API endpoint used by the CLI to create hooks via the 'sensuctl create'
  command. It's now possible to create objects of type 'Hook' with this command
  again.
- Firefox status icons not fully rendering

## [5.0.1] - 2018-12-12

### Changed
- Added --etcd-advertise-client-urls options to docker-compose.yaml sensu-backend start command

### Fixed
- Prevent a panic when using an external etcd cluster.
- Silences List in web ui sorted by ascending order; defaults to descending
- Reduces shuffling of items as events list updates
- Fixed error in UI where status value could not be coerced
- Copy local environment variables into execution context when running checks
- Ensure environment variables are joined with a semicolon on Windows
- Command arguments are no longer needlessly escaped on Windows
- Backend environments are now included in handler & mutator execution requests.

## [5.0.0] - 2018-11-30

### Added
- Add the `etcd-advertise-client-urls` config attribute to sensu-backend
- Support for multiple API versions added to sensuctl create
- Support for metadata added to wrapped resources (yaml, wrapped-json)
- Added the backend configuration attributes `api-listen-address` & `api-url`.
- Adds feedback when rerunning check[s] in the web app

### Removed
- Check subdue functionality has been disabled. Users that have checks with
subdues defined should delete and recreate the check. The subdue feature was
found to have issues, and we are re-working the feature for a future release.
- Filter when functionality has been disabled. Users that have filters with
'when' properties defined should delete and recreate the filter. Filter when
uses the same facility as check subdue for handling time windows.
- Removed event.Hooks and event.Silenced deprecated fields
- Extensions have been removed until we have time to revisit the feature.

### Changed
- Assets and checks environments are now merged, with a preference given to the
  values coming from the check's environment.
- Assets and handlers environments are now merged, with a preference given to the
  values coming from the handler's environment.
- Assets and mutators environments are now merged, with a preference given to the
  values coming from the mutator's environment.
- Metadata from wrappers and resources is now merged, with a preference given to
the values coming from the wrapper. Labels and annotations are deep-merged.
- Round-robin scheduling has been temporarily disabled.
- The dashboard now uses the `api-url` configuration attribute to connect to the
API.

### Fixed
- Fixed several resource leaks in the check scheduler.
- Fixed a bug in the dashboard where entities could not be silenced.
- Fix the `sensuctl cluster health` command.
- Fixed issue filtering by status on the events page
- Fixed interactive operations on entities in the CLI
- Removed rerun and check links for keepalives on event details page.
- Web UI - Made silencing language more clear on Silences List page
- Fixed a bug where resources from namespaces that share a common prefix, eg:
  "sensu" and "sensu-devel", could be listed together.
- Fixed a bug in the agent where the agent would deadlock after a significant
period of disconnection from the backend.
- Fixed a bug where logging events without checks would cause a nil panic.
- Removed the ability to rerun keepalives on the events list page
- A panic in keepalive/check ttl monitors causing a panic.
- Monitors are now properly namespaced in etcd.
- Updating a users groups will no longer corrupt their password
- Prevent empty error messages in sensuctl.
- Fixed a bug where keepalive failures could be influenced by check TTL
successes, and vice versa.
- Fixed a bug where check TTL events were not formed correctly.
- Fixed a web-ui bug causing the app to crash on window resize in FireFox

### Breaking Changes
- The backend configuration attributes `api-host` & `api-port` have been
replaced with `api-listen-address`.

## [2.0.0-beta.8-1] - 2018-11-15

### Added
- Assets are included on check details page.
- Adds links to view entities and checks from the events page.
- Added an agent/cmd package, migrated startup logic out of agent main
- Improved debug logging in pipeline filtering.
- Add object metadata to entities (including labels).
- Add filter query support for labels.
- Add support for setting labels on agents with the command line.
- The sensuctl tool now supports yaml.
- Add support for `--all-namespaces` flag in `sensuctl extension list`
subcommand.
- Added functionality to the dynamic synthesize function, allowing it to
flatten embedded and non-embedded fields to the top level.
- Added the sensuctl edit command.
- Added javascript filtering.

### Removed
- Govaluate is no longer part of sensu-go.

### Fixed
- Display appropriate fallback when an entity's lastSeen field is empty.
- Silences List in web ui sorted by ascending order
- Sorting button now works properly
- Fixed unresponsive silencing entry form begin date input.
- Removed lastSeen field from check summary
- Fixed a panic on the backend when handling keepalives from older agent versions.
- Fixed a bug that would prevent some keepalive failures from occurring.
- Improved event validation error messages.
- Improved agent logging for statsd events.
- Fixues issue with tooltip positioning.
- Fixed bug with toolbar menus collapsing into the overflow menu
- The agent now reconnects to the backend if its first connection attempt
  fails.
- Avoid infinite loop when code cannot be highlighted.

### Changes
- Deprecated the sensu-agent `--id` flag, `--name` should be used instead.

### Breaking Changes
- Environments and organizations have been replaced with namespaces.
- Removed unused asset metadata field.
- Agent subscriptions are now specified in the config file as an array instead
  instead of a comma-delimited list of strings.
- Extended attributes have been removed and replaced with labels. Labels are
string-string key-value pairs.
- Silenced `id`/`ID` field has changed to `name`/`Name`.
- Entity `id`/`ID` field has changed to `name`/`Name`.
- Entity `class`/`Class` field has changed to `entity_class`/`EntityClass`.
- Check `proxy_entity_id`/`ProxyEntityID` field has changed to `proxy_entity_name`/`ProxyEntityName`.
- Objects containing both a `name`/`Name` and `namespace`/`Namespace` field have been
replaced with `metadata`/`ObjectMeta` (which contains both of those fields).
- Role-based access control (RBAC) has been completely redesigned.
- Filter and token substitution variable names now match API naming. Most names
that were previously UpperCased are now lower_cased.
- Filter statements are now called expressions. Users should update their
filter definitions to use this new naming.

## [2.0.0-beta.7-1] - 2018-10-26

### Added
- Asset functionality for mutators and handlers.
- Web ui allows publishing and unpublishing on checks page.
- Web ui allows publishing and unpublishing on check details page.
- Web ui code highlighting added.

### fixed
- fixes exception thrown when web ui browser window is resized.

## [2.0.0-beta.6-2] - 2018-10-22

### Added
- Add windows/386 to binary gcs releases
- TLS authentication and encryption for etcd client and peer communication.
- Added a debug log message for interval timer initial offset.
- Added a privilege escalation test for RBAC.

### Removed
- Staging resources and configurations have been removed from sensu-go.
- Removed handlers/slack from sensu/sensu-go. It can now be found in
sensu/slack-handler.
- Removed the `Error` store and type.

### Changed
- Changed sensu-agent's internal asset manager to use BoltDB.
- Changed sensuctl title colour to use terminal's configured default for bold
text.
- The backend no longer forcibly binds to localhost.
- Keepalive intervals and timeouts are now configured in the check object of
keepalive events.
- The sensu-agent binary is now located at ./cmd/sensu-agent.
- Sensuctl no longer uses auto text wrapping.
- The backend no longer requires embedded etcd. External etcd instances can be
used by providing the --no-embed option. In this case, the client will dial
the URLs provided by --listen-client-urls.
- The sensu-agent binary is now located at ./cmd/sensu-agent.
- Sensuctl no longer uses auto text wrapping.
- The backend no longer requires embedded etcd. External etcd instances can be
used by providing the --no-embed option. In this case, the client will dial
the URLs provided by --listen-client-urls.
- Deprecated daemon `Status()` functions and `/info` (`/info` will be
re-implemented in https://github.com/sensu/sensu-go/issues/1739).
- The sensu-backend flags related to etcd are now all prefixed with `etcd` and
the older versions are now deprecated.
- Web ui entity recent events are sorted by last ok.
- etcd is now the last component to shutdown during a graceful shutdown.
- Web ui entity recent events are sorted by last ok
- Deprecated --custom-attributes in the sensu-agent command, changed to
--extended-attributes.
- Interfaced command execution and mocked it for testing.
- Updated the version of `libprotoc` used to 3.6.1.

### Fixed
- Fixed a bug in `sensuctl configure` where an output format called `none` could
  be selected instead of `tabular`.
- Fixes a bug in `sensuctl cluster health` so the correct error is handled.
- Fixed a bug where assets could not extract git tarballs.
- Fixed a bug where assets would not install if given cache directory was a
relative path.
- Fixed a bug where an agent's collection of system information could delay
sending of keepalive messages.
- Fixed a bug in nagios perfdata parsing.
- Etcd client URLs can now be a comma-separated list.
- Fixed a bug where output metric format could not be unset.
- Fixed a bug where the agent does not validate the ID at startup.
- Fixed a bug in `sensuctl cluster health` that resulted in an unmarshal
error in an unhealthy cluster.
- Fixed a bug in the web ui, removed references to keepaliveTimeout.
- Keepalive checks now have a history.
- Some keepalive events were misinterpreted as resolution events, which caused
these events to be handled instead of filtered.
- Some failing keepalive events were not properly emitted after a restart of
sensu-backend.
- The check output attribute is still present in JSON-encoded events even if
empty.
- Prevent an empty Path environment variable for agents on Windows.
- Fixed a bug in `sensuctl check update` interactive mode. Boolean defaults
were being displayed rather than the check's current values.
- Use the provided etcd client TLS information when the flag `--no-embed-etcd`
is used.
- Increase duration delta in TestPeriodicKeepalive integration test.
- Fixed some problems introduced by Go 1.11.

### Breaking Changes
- Removed the KeepaliveTimeout attribute from entities.

## [2.0.0-beta.4] - 2018-08-14

### Added
- Added the Sensu edition in sensuctl config view subcommand.
- List the supported resource types in sensuctl.
- Added agent ID and IP address to backend session connect/disconnect logs
- Licenses collection for RHEL Dockerfiles and separated RHEL Dockerfiles.

### Changed
- API responses are inspected after each request for the Sensu Edition header.
- Rename list-rules subcommand to info in sensuctl role commmand with alias
for backward compatibility.
- Updated gogo/protobuf and golang/protobuf versions.
- Health API now returns etcd alarms in addition to cluster health.

### Fixed
- Fixed agentd so it does not subscribe to empty subscriptions.
- Rules are now implicitly granting read permission to their configured
environment & organization.
- The splay_coverage attribute is no longer mandatory in sensuctl for proxy
check requests and use its default value instead.
- sensu-agent & sensu-backend no longer display help usage and duplicated error
message on startup failure.
- `Issued` & `History` are now set on keepalive events.
- Resolves a potential panic in `sensuctl cluster health`.
- Fixed a bug in InfluxDB metric parsing. The timestamp is now optional and
compliant with InfluxDB line protocol.
- Fixed an issue where adhoc checks would not be issued to all agents in a
clustered installation.

### Breaking Changes
- Corrects the check field `total_state-change` json tag to `total_state_change`.

## [2.0.0-beta.3-1] - 2018-08-02

### Added
- Added unit test coverage for check routers.
- Added API support for cluster management.
- Added sensuctl cluster member-list command.
- Added Sensu edition detection in sensuctl.
- Added sensuctl cluster member-add command.
- Added API client support for enterprise license management.
- Added a header to API calls that returns the current Sensu Edition.
- Added sensuctl cluster health command.

### Changed
- The Backend struct has been refactored to allow easier customization in
enterprise edition.
- Use etcd monitor instead of in-memory monitor.
- Refactoring of the cmd package for sensuctl to allow easier customization in
the enterprise edition.
- Upgrade dep to v0.5.0
- Added cluster health information to /health endpoint in sensu-backend.

### Fixed
- Fixed `sensuctl completion` help for bash and zsh.
- Fixed a bug in build.sh where versions for Windows and Mac OS were not
generated correctly.
- Display the name of extensions with table formatting in sensuctl.
- Fixed TLS issue that occurred when dashboard communicated with API.
- Check TTL now works with round robin checks.
- Format string for --format flag help now shows actual arguments.
- Push the sensu/sensu:nightly docker image to the Docker Hub.
- Replaced dummy certs with ones that won't expire until 100 years in the
future.
- Fixed a bug where clustered round robin check execution executed checks
too often.
- Catch errors in type assertions in cli.
- Fixed a bug where users could accidentally create invalid gRPC handlers.

### Removed
- Removed check subdue e2e test.
- Removed unused Peek method in the Ring data structure.

### Breaking Changes
- Removed deprecated import command.

## [2.0.0-beta.2] - 2018-06-28

### Added
- Performed an audit of events and checks. Added `event.HasCheck()` nil checks
prior to assuming the existence of said check.
- Added a Create method to the entities api.
- Added the ability to set round robin scheduling in sensuctl
- Added Output field to GRPC handlers
- Additional logging around handlers
- Accept additional time formats in sensuctl
- Entities can now be created via sensuctl.
- Added the format `wrapped-json` to sensuctl `configure`, `list` and `info`
commands, which is compatible with `sensuctl create`.
- Added debug event log with all event data.
- Added yml.example configurations for staging backend and agents.
- Added test resources in `testing/config/resources.json` to be used in staging.
- Added all missing configuration options to `agent.yml.example` and
`backend.yml.example`.
- Added environment variables to checks.
- Added logging redaction integration test.
- Added check token substitution integration test.
- Added the `sensuctl config view` subcommand.
- Added extension service configuration to staging resources.
- Added some documentation around extensions.
- Added Dockerfile.rhel to build RHEL containers.

### Changed
- Upgraded gometalinter to v2.
- Add logging around the Sensu event pipeline.
- Split out the docker commands in build script so that building images and
  pushing can be done separately.
- Migrated the InfluxDB handler from the sensu-go repository to
github.com/nikkiki/sensu-influxdb-handler
- Entry point for sensu-backend has been changed to
  `github.com/sensu/sensu-go/cmd/sensu-backend`
- Don't allow unknown fields in types that do not support custom attributes
when creating resources with `sensuctl create`.
- Provided additional context to metric event logs.
- Updated goversion in the appveyor configuration for minor releases.
- Use a default hostname if one cannot be retrieved.
- Return an error from `sensuctl configure` when the configured organization
or environment does not exist.
- Remove an unnecessary parameter from sensuctl environment create.
- The profile environment & organization values are used by default when
creating a resource with sensuctl.
- Migrated docker image to sensu Docker Hub organization from sensuapp.
- Use the sensu/sensu image instead of sensu/sensu-go in Docker Hub.

### Fixed
- Prevent panic when verifying if a metric event is silenced.
- Add logging around the Sensu event pipeline
- Marked silenced and hooks fields in event as deprecated
- Fixed a bug where hooks could not be created with `create -f`
- Metrics with zero-values are now displayed correctly
- Fix handler validation routine
- Fixed a small bug in the opentsdb transformer so that it trims trailing
whitespace characters.
- Sensu-agent logs an error if the statsd listener is unable to start due to an
invalid address or is stopped due to any other error.
- Fixed a bug where --organization and --environment flags were hidden for all
commands
- Fix a bug where environments could not be created with sensuctl create
- StatsD listener on Windows is functional
- Add version output for dev and nightly builds (#1320).
- Improve git version detection by directly querying for the most recent tag.
- Fixed `sensuctl create -f` for `Role`
- Fixed `sensuctl create -f` for `Event`
- Added validation for asset SHA512 checksum, requiring that it be at least 128
characters and therefore fixing a bug in sensuctl
- Silenced IDs are now generated when not set in `create -f` resources
- API requests that result in a 404 response are now logged
- Fixed a bug where only a single resource could be created with
`sensuctl create` at a time.
- Fixed a bug where environments couldn't be deleted if there was an asset in
the organization they reside in.
- Dashboard's backend reverse proxy now works with TLS certs are configured.
- Fixed a bug with the IN operator in query statements.
- Boolean fields with a value of `false` now appear in json format (removed
`omitempty` from protobufs).
- The sensuctl create command no longer prints a spurious warning when
non-default organizations or environments are configured.
- When installing assets, errors no longer cause file descriptors to leak, or
lockfiles to not be cleaned up.
- Fixed a bug where the CLI default for round robin checks was not appearing.
- Missing custom attributes in govaluate expressions no longer result in
an error being logged. Instead, a debug message is logged.
- Update AppVeyor API token to enable GitHub deployments.
- Allow creation of metric events via backend API.
- Fixed a bug where in some circumstances checks created with sensuctl create
would never fail.
- Fixed a goroutine leak in the ring.
- Fixed `sensuctl completion` help for bash and zsh.

### Removed
- Removed Linux/386 & Windows/386 e2e jobs on Travis CI & AppVeyor
- Removed check output metric extraction e2e test, in favor of more detailed
integration coverage.
- Removed the `leader` package
- Removed logging redaction e2e test, in favor of integration coverage.
- Removed check token substitution e2e test, in favor of integration coverage.
- Removed round robin scheduling e2e test.
- Removed proxy check e2e test.
- Removed check scheduling e2e test.
- Removed keepalive e2e test.
- Removed event handler e2e test.
- Removed `sensuctl` create e2e tests.
- Removed hooks e2e test.
- Removed assets e2e test.
- Removed agent reconnection e2e test.
- Removed extensions e2e test.

## [2.0.0-beta.1] - 2018-05-07
### Added
- Add Ubuntu 18.04 repository
- Support for managing mutators via sensuctl.
- Added ability to sort events in web UI.
- Add PUT support to APId for the various resource types.
- Added flags to disable the agent's API and Socket listeners
- Made Changelog examples in CONTRIBUTING.md more obvious
- Added cli support for setting environment variables in mutators and handlers.
- Added gRPC extension service definition.
- The slack handler now uses the iconURL & username flag parameters.
- Support for nightlies in build/packaging tooling.
- Added extension registry support to apid.
- Added extension registry to the store.
- Add sensuctl create command.
- Adds a statsd server to the sensu-agent which runs statsd at a configurable
flush interval and converts gostatsd metrics to Sensu Metric Format.
- Add event filtering to extensions.
- Proper 404 page for web UI.
- Add sensuctl extension command.
- Add extensions to pipelined.
- Added more tests surrounding the sensu-agent's statsd server and udp port.
- Add the `--statsd-event-handlers` flag to sensu-agent which configures the
event handlers for statsd metrics.
- Add default user with username "sensu" with global, read-only permissions.
- Add end-to-end test for extensions.
- Add configuration setting for backend and agent log level.
- Add extension package for building third-party Sensu extensions in Go.
- Add the `--statsd-disable` flag to sensu-agent which configures the
statsd listener. The listener is enabled by default.
- Added an influx-db handler for events containing metrics.
- Add 'remove-when' and 'set-when' subcommands to sensuctl filter command.
- Added the Transformer interface.
- Added a Graphite Plain Text transformer.
- Add support for `metric_format` and `metric_handlers` fields in the Check and
CheckConfig structs.
- Add CLI support for `metric_format` and `metric_handlers` fields in `sensuctl`.
- Add support for metric extraction from check output for `graphite_plaintext`
transformer.
- Added a OpenTSDB transformer.
- Add support for metric extraction from check output for `opentsdb_line`
- Added a Nagios performance data transformer.
- Add support for metric extraction from check output for `nagios_perfdata`
- Added an InfluxDB Line transformer.
- Add support for metric extraction from check output for `influxdb_line`
transformer.
- Add e2e test for metric extraction.

### Changed
- Changed the maximum number of open file descriptors on a system to from 1024
(default) to 65535.
- Increased the default etcd size limit from 2GB to 4GB.
- Move Hooks and Silenced out of Event and into Check.
- Handle round-robin scheduling in wizardbus.
- Added informational logging for failed entity keepalives.
- Replaced fileb0x with vfsgen for bundling static assets into binary. Nodejs 8+
and yarn are now dependencies for building the backend.
- Updated etcd to 3.3.2 from 3.3.1 to fix an issue with autocompaction settings.
- Updated and corrected logging style for variable fields.
- Build protobufs with go generate.
- Creating roles via sensuctl now supports passing flags for setting permissions
  rules.
- Removed -c (check) flag in sensuctl check execute command.
- Fix a deadlock in the monitor.
- Don't allow the bus to drop messages.
- Events list can properly be viewed on mobile.
- Updated Sirupsen/logrus to sirupsen/logrus and other applicable dependencies using the former.
- Set default log level to 'warn'.
- Optimize check marshaling.
- Silenced API only accepts 'id' parameter on DELETE requests.
- Disable gostatsd internal metric collection.
- Improved log entries produced by pipelined.
- Allow the InfluxDB handler to parse the Sensu metric for an InfluxDB field tag
and measurement.
- Removed organization and environment flags from create command.
- Changed `metric_format` to `output_metric_format`.
- Changed `metric_handlers` to `output_metric_handlers`.

### Fixed
- Terminate processes gracefully in e2e tests, allowing ports to be reused.
- Shut down sessions properly when agent connections are disrupted.
- Fixed shutdown log message in backend
- Stopped double-writing events in eventd
- Agents from different orgs/envs with the same ID connected to the same backend
  no longer overwrite each other's messagebus subscriptions.
- Fix the manual packaging process.
- Properly log the event being handled in pipelined
- The http_check.sh example script now hides its output
- Silenced entries using an asterisk can be deleted
- Improve json unmarshaling performance.
- Events created from the metrics passed to the statsd listener are no longer
swallowed. The events are sent through the pipeline.
- Fixed a bug where the Issued field was never populated.
- When creating a new statsd server, use the default flush interval if given 0.
- Fixed a bug where check and checkconfig handlers and subscriptions are null in rendered JSON.
- Allow checks and hooks to escape zombie processes that have timed out.
- Install all dependencies with `dep ensure` in build.sh.
- Fixed an issue in which some agents intermittently miss check requests.
- Agent statsd daemon listens on IPv4 for Windows.
- Include zero-valued integers in JSON output for all types.
- Check event entities now have a last_seen timestamp.
- Improved silenced entry display and UX.
- Fixed a small bug in the opentsdb transformer so that it trims trailing
whitespace characters.

## [2.0.0-nightly.1] - 2018-03-07
### Added
- A `--debug` flag on sensu-backend for enabling a pprof HTTP endpoint on localhost.
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
- Added GraphQL explorer to web UI.
- Added check occurrences and occurrences_watermark attributes from Sensu 1.x.
- Added issue template for GitHub.
- Added custom functions to evaluate a unix timestamp in govaluate.

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
- We no longer duplicate hook execution for types that fall into both an exit
code and severity (ex. 0, ok).
- Updated the sensuctl guidelines.
- Changed travis badge to use travis-ci.org in README.md.
- Govaluate's modifier tokens can now be optionally forbidden.
- Increase the stack size on Travis CI.
- Refactor store, queue and ring interfaces, and daemon I/O details.
- Separated global from local flags in sensuctl usage.

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
- Fixed HookList `hooks` validation and updated `type` validation message to
allow "0" as a valid type.
- Events' check statuses & execution times are now properly added to CheckHistory.
- Sensu v1 Check's with TTL, timeout and threshold values can now be imported
correctly.
- Use uint32 for status so it's not empty when marshalling.
- Automatically create a "default" environment when creating a new organization.

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
