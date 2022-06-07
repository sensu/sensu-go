# Sensu-core Agents and Sensu-go

The investigation done as a part of [EDGEMEM-1509](https://ctl.atlassian.net/browse/EDGEMEM-1509) has lead to three possiblities of how to deal with current sensu-core agents with sensu-go, they are:

1. Borrow ideas from geographic regionality.

The idea here is to separate Watcher into two stacks; A sensu stack that is all the required parts to run sensu, and a Watcher stack that is composed of the Watcher APIs. This would allow us to keep one sensu stack running sensu-core, and allow agents' connections and message processing to continue in the same manner as today. We would need to modify the Watcher APIs to indenitify and call the correct sensu stack to retrieve data about agent connection status, current results, events, and suppressions. This also allows us a simplified manner to scale Watcher. As the number of agent connections grows, we can just spin up new sensu stack clusters to handle increased agent/message counts. This also lends itself to being able to better handle customer environments that are locked down as we could drop a sensu stack into their environment and create a VPN connection to that stack so that the Watcher APIs can still interact with it.

2. Implement Asset Management.

Implementing Asset Management will allow us to create a fairly simple service to read messages from rabbitMQ and then forward these messages to sensu-go agent(s) that we own, and control. Those agents can deliver the messages to the sensu-go backend. This then buys us the ability to take advantage of this feature from a product perspective to allow registering "assets" to better report moniotring data in a more human consumable manner. For example, often times operations uses the agent running on a management appliance to monitor cloud native services. This feature would allow for that agent to return results under an "asset" instead so when events are created, they are associated with that asset instead of the management appliance itself.

It should be noted that while "aliasing" current sensu-core agents through sensu-go agents that we own may not be a giant lift, the work to facilitate this as a usable feature to be consumed by customers/operations is a significantly bigger ask, so this would really just be laying the ground work for the Asset Management feature set.

3. Modify Sensu-Go to read and process messages from rabbitMQ.

This is included for the sake of completeness, but this is probably the least attractive option. We do not intend for there to be sensu-core agents in the long term once sensu-go is ready for the show so making modifications to sensu-go that may make it more difficult to update sensu-go makes little sense. It also does not buy us any addition functionality like the first two options.
