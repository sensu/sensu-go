# UX Guidelines

The following serves as a living document, describing the ideals that should be
followed when adding, updating or expanding new sub-commands to `sensuctl`.

## Table of Contents

<details>
  <summary>Expand</summary>
- [Format](#format)
</details>

## Format

Sensuctl outputs details in one of two ways: a user friendly format and a
structured data format.

#### Structured

The structure data output (at this time only JSON is supported) provides the end
user with a structured output that _should_ be easy to integrate into an
administrator's or engineer's tool-chain.

While most of the following document will not cover this format, be mindful of
how you would like the output of the command so that you could easily implement
it in a script you hypothetically could be writing.

Imagine any edge cases that could occur and how you could communicate it to the
end-user so that it could be handled appropriately. As such, ideally error
output is also structured so that tooling can recover, from, for instance an
update command failing due to the given resource not being found.

#### User Friendly

The user-friendly out is the default provided to end-users and it's goal is to
be friendly, conventional and easy to parse by the _human_ eye. The rest of this
document will largely focus on this latter format.

## Management Commands

Management commands are those that expose a number of subcommands that act upon
a Sensu primitive. For example an 'asset' management command would expose
one subcommand for each action a user can perform on an asset or collection of
assets.

In the interest of keeping a simple identifiable standard for our the ease of
our end users, we ask that developers keep the naming of the subcommands
consistent. The following is a list of standard names for common subcommands.

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

## Displaying Collections
## Displaying Expanded Details
## Creating Resources
## Updating Resources
## Deleting Resources

## Unrecoverable Actions

## Loading
## Colour
## Format
## Errors

## Output
## Command Output
## Auto-completion
