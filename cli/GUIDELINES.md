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

The intention of our structure data format (at this time only JSON is supported)
is to provide the end user with consistent predictable results that can easily
be integrated into an administrator's or engineer's tool-chain.

While most of the following document will **not** cover this format, when
developing a new command, try to mindful of how you would like the command to
function if you were trying to implement it into your own tool-chain.

Try to imagine any edge cases that could occur and how you could communicate it
to the end-user so that it could be handled appropriately. As such, ideally
errors are also output in a predictable form so that any tooling could easily
recover from and rectify any issue. For instance, an update command failing due
to the given resource not being found; if output is predictable the
administrator could easily have their tools recover by creating said resource
instead.

#### User Friendly

The user-friendly out is the default provided to end-users and it's goal is to
be friendly, conventional and easy to parse by the _human_ eye. The rest of this
document will largely focus on this latter format.

## Colour

- Use colour to draw the end-users eye to important details
- Use bold for emphasis; table headers, keys, etc.
- Use blue and green to draw users eye to primary details.
- Red should be used sparingly as it is the most eye catching colour; should
  only be used for actions that are unrecoverable or things that critical.
- No element should ever be designed so that colour is required to understand
  what it is trying to communicate. It is not a guarantee that the user's
  terminal will support colour. For instance it would not be okay to omit that a
  message is an error because you've used to the color red.
- Following sections will detail usage in different different scenarios.

## Loading

`sensuctl` should always **feel** kinetic; or put in other words, whenever an
end user executes a command, no matter their system or circumstances they should
feel as if the command has responded to their requested action. To use an
example, if you were playing Diablo and your nine-foot tall muscular Barbarian,
giant cudgel in hand; swung at a poor unsuspecting demon and the game didn't
immediately respond by flinging the enemy across the map? You would quickly
decry the game's poor netcode. Of course this isn't the case. Accompanied by a
satisfying thunk the demon goes flying, the swing even ruptures nearby barrels,
and you feel godlike.

With this in mind...

Not all end-users internet connections are made equally, nor can we assume all
connections between our users and Sensu installations will always be low
latency. More over some commands consist of _many_ operations that can take time
to commit. As such, whenever some activity is taking place, whether determinate
or not, it should be accompanied by some form of activity indicator.

Types of activities and expected feedback:

- When an activity is determinate (eg. uploading asset with known size) ideally
  a loading indicator that displays the current progress is displayed.
- When activity is duration indeterminate an spinner should be displayed.
  It is important that the loading indicator should animate so that the end-user
  does not believe that the command has frozen.

Especially with activities where progress is indeterminate, where possible, it
is ideal that some information about the current process is exposed to the user.
In this way the user is not confused as to whether anything is actually
happening. As an example, the import command consists of many operations that
occur in sequence; as the operations are taking place they are printed to the
screen as they happen. This is not only so the user can see that the result but
so that they know that the command is continuing to execute.

One caveat is that you should be mindful of the messages you display, confusion or suspicion can easily arise from language that is too vague or too explicit.

## Errors

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

## Output
## Command Output
## Auto-completion
