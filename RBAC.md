## Proposal

### Overview

In order to provide the fine-grained access control that our current customers have expressed a need for, Sensu must use organizations, environments, and check/agent subscriptions to define role-based access control. A Sensu user can be a member of multiple organizations, each with one or more environments, with each having zero or more subscriptions to further restrict a user's capabilities. By default, every Sensu user is a member of the "default" organization, with a "default" environment, with no subscriptions (ANY/ALL). A Sensu user can be for a Sensu Agent(s), a tool (i.e. Slackbot), or a human operator.

```
-- organization
              |
               --environment
                           |
                            -- subscription

-- user
```

### Users

A Sensu user has a name (i.e. "portertech) and two means of authentication, a password, or a SSL client certificate. A Sensu user also has a session token, generated when a user successfully authenticates using a password, which expires after a configurable amount of time. A Sensu user session token may be used to make Sensu API requests. When using a SSL client certificate for authentication, the Sensu Backend compares the certificate with the one stored for the user.

User membership:

- One or more "organization"
- One or more "environment" in "organization"
- One or more "subscription" in "environment"

### Entities

A Sensu entity has an "organization" and "environment", which default to "default". This allows Sensu RBAC to limit a user's allowed actions for it (e.g. create, view, update, delete).

### Checks

A Sensu check has an "organization" and "environment", which default to "default". This allows Sensu RBAC to limit a user's allowed actions for it (e.g. create, view, update, delete).

**The same applies to filters, mutators, and handlers.**

### Backend

The Sensu Backend MUST use a combination of "organization" and "environment" for name-spacing. This applies to the message bus topics (e.g. "default:default:subscription:webserver"), and the data store (key prefix - **GREP?**).

The Sensu Backend API is responsible for RBAC user management and enforcing RBAC on a request basis.

A Sensu Backend TLS websocket server is responsible authenticating Sensu Agent users. A Sensu Agent user uses a SSL client certificate for authentication. If a Backend encounters a new Sensu Agent user (nonexistent), a Sensu RBAC user is created and marked as "unverified", and the Agent must wait for another Sensu user to verify it before it can subscribe or send events. When another Sensu user verifies a "unverified" Agent user, the Agent users's SSL client certificate is stored and further websocket connection attempts are authorized.

### Agent

The Sensu Agent must have a configurable Sensu RBAC user. The user uses a SSL client certificate, either one specified by a configured file path, or one generated on first run. Other than providing user information when connecting to the Sensu Backend websocket, the Agent remains ignorant of RBAC.