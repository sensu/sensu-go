## User stories

1. A User may be a member of a Group.

2. A User may be a member of more than one Group.

3. A Group may be granted Read or Read/Write access to a specific object type within an Environment.

4. A Group may be granted Read or Read/Write access to a specific instance of an object within an Environment (overriding the group's permissions to that object type for this object).

5. An Agent has a User.

6. Agent has a default entity that is its hostname -- that entity is scoped in an Environment. The Agent only has authorization to modify that Entity or create Events associated with it within that Environment.

7. An Agent may be granted access in an automated or manual fashion to modify other Entities or create Events associated with those Entities in any Environment.

#### Examples

Given a single environment:

The Developers group has four checks with sensitive information in their event payloads. The Developers group has Read/Write access to those checks.

The Developers group also have five checks that anyone can read. The Developers group has Read/Write access to those checks.

Another group of users, Support, does not have Read or Write access to the sensitive checks, but does have Read access to the non-sensitive checks and associated events.

## Proposal

### Overview

In order to provide fine-grained access control, Sensu must have Organizations, Groups, Environments, Object Types, and Object Instances.

A Sensu User can be a member of multiple Organizations.

A Sensu User can be a member of multiple Groups within each Organization.

A Sensu Group may grant permissions to Object Types (i.e. checks) and Object Instances (i.e. check "check_lb") for one or more Environments within an Organization.

By default, every Sensu User, Entity, and Object (i.e. check) is a member of the Organization "default".

By default, every Entity, and Object (i.e. check) is a member of the Environment "default".

The Organization "default" has a Group "default", which grants read/write permissions to every Object Type in the Environment "default". A User is is NOT a member by default, however, it can become a member at creation time or later.

Object Types include checks, subscriptions, filters, mutators, and handlers.

A Sensu Environment may have one or more trusted SSL certificate chains, used to authenticate Agent Users (**WIP**).

#### Examples

```
-- Organization
              |
               --Group
                     |
                      --Environment
                                 |
                                  -- Object Type
                                  -- Object Instance

-- User
```

```
-- default (Organization)
         |
          --default (Group)
                  |
                   --default (Environment)
                           |
                            -- {checks: rw, filters: rw, mutators: rw, handlers: rw} (Object Type)
                            -- {checks: {check_lb: --}} Object Instance

-- portertech (User)
```

### Authentication

A Sensu User has a name (i.e. "portertech) and two means of authentication, a password, or a SSL client certificate.

A Sensu User has a session token, generated when a user successfully authenticates using a password, which expires after a configurable amount of time. A Sensu user session token may be used to make Sensu API requests.

Sensu Agent Users have SSL client certificate, used for authentication. If the Agent's Environment has a trusted SSL certificate chain, the chain is used to verify the SSL client certificate provided by the Agent. If verification fails or the Environment does not have a trusted SSL certificate chain, the Agent User's SSL client certificate is not set, and the User is marked as "unverified". Unverified Users can be verified by another Sensu User (operator), storing the provided SSL client certificate. The "unverified" workflow is intended to help users get started quickly. A Sensu cluster can opt-out of the "unverified" workflow, which will be a common step after configuring a trusted SSL certificate chain for an Environment.

### Service changes

#### Backend

The Sensu Backend MUST use a combination of "organization" and "environment" for name-spacing. This applies to the message bus (topic prefix, e.g. "default:default:subscription:webserver"), and the data store (etcd key prefix).

The Sensu Backend API is responsible for RBAC management and enforcing RBAC on an API request basis.

A Sensu Backend TLS websocket server is responsible for creating and authenticating Sensu Agent users. A Sensu Agent User uses a SSL client certificate for authentication. See [Authentication](#authentication) for more information.

#### Agent

The Sensu Agent must have a configurable Sensu RBAC User. The User uses a SSL client certificate, either one specified by a configured file path, or one generated on first run. Other than providing User information when connecting to the Sensu Backend websocket, the Agent remains ignorant of RBAC.