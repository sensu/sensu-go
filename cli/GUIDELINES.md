# UX Guidelines

The following serves as a living document, describing the ideals that should be
followed when adding, updating, or expanding new sub-commands to `sensuctl`.

## Table of Contents

- [Format](#format)
  - [Structured](#structured)
  - [User Friendly](#user-friendly)
- [Colour](#colour)
- [Feedback](#feedback)
  - [Activity Indicators](#activity-indicators)
  - [Results](#results)
  - [Errors](#errors)
- [Unrecoverable Operations](#unrecoverable-operations)
- [Elements](#elements)
  - [Tables](#tables)
  - [Lists](#lists)
  - [Interactive Prompts](#interactive-prompts)
  - [Colour Elems](#colour-elems)
  - [Auto-completion](#auto-completion)
- [Management Commands](#management-commands)
- [Subcommand Conventions](#subcommand-conventions)
  - [List](#list)
  - [Info](#info)
  - [Create](#create)
  - [Update](#update)
  - [Delete](#delete)

## Format

Sensuctl outputs details in one of two ways: a user friendly format and a
structured data format.

#### Structured

The intention of our structured data format (at this time JSON) is to provide
the end user with consistent and predictable results in a well understood
format. Ideally always formatted (pretty printing & colour) in a conventional
manner.

Try to imagine any edge cases that could occur and how you could communicate
them to the end-user so that the user is not confused. For instance, if an
update subcommand is failing due to the given resource not being found; output
should be clear and predictable enough that they can quickly ascertain what type
of error occurred, on what resource and what steps they can take to continue.

Most of the following document will **not** cover this format.

#### User Friendly

The user friendly output is the default provided to end-users and it's goal is
to be friendly, conventional, and easy to parse by the _human_ eye. The rest of
this document will largely focus on this latter format.

## Colour

When used appropriately and in a consistent manner, colour is a powerful tool
for accurately communicating and drawing an end-user's eye to important details.
The following are some simple rules to consider when formatting your content:

- Use bold for emphasis. Examples of places where emphasis may help are:
  important words or phrases that the user shouldn't miss, field names, table
  headings.
- Use green and blue to draw users eye to primary details.
- Use blue as an indicator that that the value is the resource's primary
  identifier.
- Use red sparingly as it is the most eye catching colour; it is reserved for
  operations that are unrecoverable or things that are critical.

No element should ever be designed so that colour is required to understand
what it is trying to communicate. It is not a guarantee that the user's
terminal will support colour, nor that they are physically able to see or
differentiate between your colours. For instance, it would not be okay to omit
that a message is an error because you've used to the colour red.

The following sections will expand upon their usage in specific scenarios.

## Feedback

`sensuctl` should always **feel** kinetic; or put in other words, whenever an
end user executes a command, no matter their system or circumstances they should
feel as if the command has responded to their requested action. To use an
example, what if you were playing Diablo and your nine-foot tall muscular
Barbarian, giant cudgel in hand, swung at a poor unsuspecting demon, and the
game didn't immediately respond by flinging the enemy across the map? You would
quickly decry the game's poor netcode. Of course this isn't the case.
Accompanied by a satisfying _thunk_, the demon goes flying, the swing kicks up
dust, even ruptures nearby barrels, and you feel godlike.

With this in mind...

### Activity Indicators

Not all end-user internet connections are made equal, nor can we assume all
connections between our users and Sensu installations will always be low
latency. More over some commands consist of _many_ operations that can take time
to commit. As such, whenever some activity is taking place, whether determinate
or not, it should be accompanied by some form of activity indicator.

Types of activities and expected feedback:

- When an activity is determinate (eg. uploading an asset with known size)
  ideally a loading indicator that displays the current progress is displayed.
- When an activity's duration is indeterminate a spinner should be displayed.
  It is important that the loading indicator should animate so that the end-user
  does not believe that the command has frozen.

With activities where progress is indeterminate, where possible, it is ideal
that some information about the current process is exposed to the user. In this
way the user is not confused as to whether anything is actually happening. As an
example, the import command consists of many operations that occur in sequence;
as the operations are taking place they are printed to the screen as they
happen. This is not only so the user can see the product of the operation, but
so that they know that the command is continuing to execute.

One caveat is that you should be mindful of the messages you display, confusion
or suspicion can easily arise from language that is too vague or sometimes even
if the language is too detailed.

### Results

Every command should return either a relevant affirmative message or a
descriptive error. Without this feedback for most end-users it will be confusing
as to whether anything actually occurred, whether they should attempt to run the
command again, or if the correct action took place. No additional operations
should be required by the user to be confident that their intended interaction
was successful.

Unless there is good reason to do otherwise, the following should be true of any
command:

- Affirmative messages should generally one or two words, in the form of a title
  and should use the colour green. For example, a command that updates an entity
  should result in a message that says "Updated."
- For the sake of tooling that uses `sensuctl`, ensure that the exit code is `0`
  on success and `1` when failures occur.

More specific conventions regarding typical commands (create, update, etc.) can
be found below.

### Errors

Errors should be brief, explicit and clear.

Rules to consider:

- Where possible, when they occur they should be accompanied by a suggestion on
  how to rectify the issue.
  - When applicable include link to the documentation "for further reading."
- Use red colour to communicate their importance.
- Error messages should be printed to STDERR.
- If error results in the command stopping the exit code should be `1`.

## Unrecoverable Operations

While our end-users are likely awesome, fun, beautiful people, when designing
for humans, we need to keep _our_ fallibilities in mind. As such, where
possible any command where the result could lead to heartache should be
recoverable. While noble, this is not always plausible; in these cases where
there is no recovery or where there could be far reaching implications, the
command should always be guarded by a confirmation prompt.

Example unrecoverable operations:

- Deleting a namespace that is not empty.
- Operation that could result in an ops individual or team being paged at 2AM.
- Updating resource where non-trivial side-effects may be present.

The following are some base rules for handling confirmation:

- When applicable, the prompt should ask them to write out the name of the
  affected resource. In this way, it forces them to confirm that they are acting
  on the correct resource.
- Confirmation prompts can and should use the colour red in their communication,
  this is to drive home their importance.
- For scripts, a flag should be present on the command to allow the prompt to be
  skipped, this way the command can easily be used in tooling.
  - In the interest of making sure the action in is intentful however, the flag
    is ideally **not** consistent across all commands. We do not want to be in
    the position where we train the user to use the same flag all the time.

## Elements

### Tables

TODO

### Lists

TODO

### Interactive Prompts

Interactive prompts are useful for users that try Sensu for the first time or
don't know or care what gets created, so they don't need to know the
exact structure and details of a given resource.

The goal behind the interactive prompt is to create a valid resource with just
enough questions, in order to cover the required and most common fields without
overwhelming the user. Once created, this resource can be customized using the
appropriate management subcommands.

The interactive prompt needs to be explicitly called with a specific flag and
should not be used as the fallback if no flags were provided; instead the help
usage should be presented to the user.

### Colour Elems

For ease of formatting text in a consistent and correct manner the
`elements/globals` pacakage defines a number constants.

### Auto-completion

For ease of use `sensuctl` includes shell extensions, an end-user may add to
their shell of choice, if they want command auto-completion.

For the most part the implementation of the auto-completion is transparent,
however, when implementing a new subcommand be mindful of which flags your
command exposes that may be required. If there are some that are required you
should use the `MarkFlagRequired` method to declare it's importance.

```golang
cmd := cobra.Command{Use: "create NAME", Short: "Create a new doggo"}
cmd.Flags().Int("age", 1, "in months the amount of time the dog/puppy has been alive")
cmd.Flags().Int("legs", 4, "number of legs the puppy has")
cmd.MarkFlagRequired("age") // annotates 'age' flag for tooling
```

## Management Commands

Management commands are those that expose a number of subcommands that act upon
a Sensu primitive. For example, an 'asset' management command would expose
one subcommand for each action a user can perform on an asset or collection of
assets.

In the interest of keeping a simple and identifiable standard for our the ease
of our end users, we ask that developers keep the naming of the subcommands
consistent. The following is a list of standard names for common subcommands.

- If you're intention is to add a command that adds a new resource to the system
  consider naming the command `create`. When the command adds an item to a list
  of an existing resource, prefix the command's name with `add` (eg. `add-role`,
  `add-subscription`).
- When adding a new command that removes a resource from the system consider
  naming the command `delete`. When the command removes an item from an existing
  resource, prefix the command's name with `remove` (eg. `remove-role`,
  `remove-subscription`).
- When adding a new command that lists a collection of resources consider naming
  the command `list`. If the command lists a collection of associated resources,
  prefix the command's name with `list` (eg. `list-roles`, `list-members`).
- When adding a new command that shows expanded details of a resource consider
  naming the command `info`.
- When the command updates an item from an existing resource, prefix
  the command's name with `set` (eg. `set-command`, `set-subscriptions`).

## Subcommand Conventions

### Create

- Support creation through an interactive mode. [details](#interactive-prompts).
- Support creation through arguments and flags.
- Validate input before firing POST request.
- As much as possible return helpful error messages.
- "Created." message should be returned on success.

### Delete

TODO

### Info

- List all fields of resource in a legible consistent manner.
- Use a list element to display all values of resource. [details](#lists).
- Highlight primary identifier by using the colour blue. [details](#colour-elems).
- Highlight any values important values with red. (eg. failing check status.) [details](#colour-elems).
- Comma separate values when displaying slices with limited number of entries.
- Use a list element when displaying slices with many entries or entries with
  long names.

### Lists

- The "at a glance" view.
- Use a table element. [details](#table).
- Highlight primary identifier by using the colour blue. [details](#colour-elems).
- Highlight any values important values with red. (eg. failing check status.) [details](#colour-elems).
- colspan is limited; limit columns to only the most important ones.

### Update

TODO

## Inspiration / Previous Art

- `kubectl`

    Probably the most comparable tool; very nice to work with.

- `yarn`

    Very friendly, great loading states and results.

- `heroku`

    Similar in that it has lots of data to display; does excellent job.
