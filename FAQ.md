# Sensu Core 2.0 Frequently Asked Questions (FAQ)

## Should I use Sensu Core 2.0 in production?

It depends. Depending on your industry and the nature of the infrastructure, e.g. fintech, you may need to use Sensu 1.x, as it has been battle tested for 6 years, and it is commercially supported by Sensu Inc. However, Sensu 2.0 can be considered production ready for most users.

## Why the design and architecture changes? Where’s my RabbitMQ?

Sensu 1.x is a now a 6 year old design and architecture. Over the 6 years, we have identified its limitations and sources of friction for new users and experienced operators. By design, Sensu 2.0 has fewer moving parts than 1.x, but makes no compromises in terms of functionality and scalability. Sensu 2.0 provides 1.x feature parity and maintains the core principles of composability and extensibility.

## Why “Entities”?

Sensu 2.0 uses an updated data model to address the latest paradigm shift in infrastructure and software delivery like containers and “serverless”. Sensu is no longer confined to the concept of a client, we have “entities” that can can represent anything.

## Is the Sensu 2.0 API backwards compatible?

While Sensu 2.0 supports many of the existing 1.x features (1.x parity), the REST API had to change in order to support new functionality (e.g. RBAC authentication, configuration by API, Entities).

## Is the Sensu 2.0 event data format the same?

No, however, there are many similarities. Sensu 2.0 events are still JSON, but contain a few new and different key spaces, such as “entity” and “metrics”. These changes are for the better, you can now represent anything in Sensu events, and metrics are first class!

## How do I migrate from 1.x?

Sensu 2.0’s sensuctl utility will be able to migrate many 1.x configurations to the 2.x format. The final list of supported Sensu 1.x migration types is still in flux, although we’re aiming to support as many types as possible, including client-to-entity, check, mutator, filter, handler, and extension imports.

There are a couple of major differences in configuration that are particularly noteworthy: token substitution and Ruby eval. Token substitution in Sensu 2.0 is done with Go templating--existing token substitution strings will have to be rewritten. Sensu 2.0 also does not have access to a Ruby interpreter, so any filters using eval expressions will have to be rewritten using Sensu 2.0’s new expression language.

We acknowledge that many users have come up with creative and unique applications for Sensu 1.x. Heck, that’s one of Sensu’s primary strengths as a monitoring framework. Our goal is to support assisted migrations (using sensuctl) for as many of these as possible, and to document the remaining situations and “gotchas” where a seamless migration is simply not feasible. Regardless, we acknowledge that the migration path is not without risk, and will prioritize migration-affecting bugs highly going forward.

## Can I still use Configuration Management? (ALT: Do I still have to use Configuration Management?)

You can still use your CM tool of choice. There are parity projects happening for each tool under a `sensu2.0` branch on Ansible, Puppet, Chef, SaltStack repositories.
