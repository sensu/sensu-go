## User stories

1. An external resource can have an Entity.

2. A Check can be associated with an Entity, automatically adding the Entity to the Check's check request Events.

3. An Agent will not add it's Agent Entity to an Event if another Entity is already present.

#### Examples

A network router has an Entity "gateway-01".

A User configures a check to run a SNMP on an Agent, monitoring "gateway-01" eth-0.

The resulting Event contains the Entity "gateway-01".

## Proposal

Sensu needs the ability to associate an Agent check execution with another Entity (not just the Agent's Entity).

A Sensu Check may have an "entity" configuration attribute, specifying an Entity by ID (or name?), allowing the Check to have a direct relationship with an Entity. When a check specifies an "entity", the Sensu Backend will include the Entity in check request Events for the Check. If the Sensu Backend is unable to find the entity when creating a check request Event, an error is logged, and the check request Event is not published.

### Authorization

Sensu Agents who recieve check request Events containing an Entity are assumed to be granted permission to produce an Event on its behalf. This bypasses RBAC permissions granted to a Agent's User by one or more Groups.
