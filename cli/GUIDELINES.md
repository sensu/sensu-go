# UX Guidelines

The following serves as a living document, describing the ideals that should be
followed when adding, updating or expanding new sub-commands to `sensuctl`.

## Format

Sensuctl outputs details in one of two ways: a user friendly format and a
structured data format.

The data format output (at this time only JSON is supported) provides the end
user with a structured output that _should_ be easy to integrate into an
administrator's or engineer's tool-chain. Ideally error output is also
structured so that tooling can recover from say an update command failing due to
the given resource not being found.

The user-friendly out is the default provided to end-users and it's goal is to
be friendly, conventional and easy to parse by the _human_ eye. The rest of this
document will largely focus on this latter format.

## Management Commands

Management commands are those that expose a number of subcommands that act upon
a Sensu primitive. For example an 'asset' management command would expose
one subcommand for each action a user can perform on an asset or collection of
assets.

For the sake of consistency we ask that developers keep the naming of the
subcommands under a management command be consistently named. The following is
list of standard names for common subcommands.

- If you're intention is to add a command that adds a new resource to the system
  consider naming the command `create`. When the command adds an item to an
  existing resource prefix the command's name with `add` (eg. `add-role`,
  `add-subscription`).
- When adding a new command that removes a resource from the system consider
  naming the command `delete`. When the command removes an item from an existing
  resource prefix the command's name with `remove` (eg. `remove-role`,
  `remove-subscription`).
- When adding a new command that lists a collection of resources consider naming
  the command `list`. If the command lists a collection of associated resources
  prefix the command's name with `list` (eg. `list-roles`, `list-members`).
- When adding a new command that shows expanded details of a resource consider
  naming the command `info`.

## Colour
## Destructive Actions (TODO: Is there a better way to say dangerous actions?)

## Displaying Collections
## Displaying Expanded Details
## Format

## Output
## Command Output
## Auto-completion
