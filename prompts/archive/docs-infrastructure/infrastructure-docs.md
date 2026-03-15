# Infrastructure documentation

Alright, it is time to add and change some much needed documentation.
While reading through everything, i noticed that there is not yet a good setup for the bots actual infrastructure.
When i say infrastructure, i mean all of the things that make the frontend user experience work.
Think of an event or interaction handler, or a good way to handle errors throughout the project.
All those kind of things need to be developed and functional before we start working on the main features of the bot.

All together, this infrastructure should be able to handle all of the features that this bot may bring, that includes post-mvp features.

## Event Handlers

Each of the following events should likely be their own script and maybe even their own package if the functionality gets too large, but all of them should be contained inside an `events` package.
Keep in mind that the following list of events may not be complete and i may be missing important requirements overall, if so, feel free to suggest more.

### InteractionCreate

This needs to handle any type of interaction that may be received, for example: 
- slash commands
- button interactions
- message context commands
- and so on...

This script should receive all types of interactions and then send the relevant information of the interaction to its definition and execution if it exists.

The way to orginize this is likely best to be done with a custom type system where a list of const types are added for each type of interaction (slash commands, buttons, etc.) and then each specific interaction (like one slash command) would be one key type inside the slash commands type list.
small example:

```go
// The value is the defined name of the command
type TSlashCommand string
// The value is the buttons custom_id
type TButton string

const (
    ComposeCreate TSlashCommand = "compose-create",
    ComposeEdit TSlashCommand = "compose-edit",
    // ...
)

const (
    ComposeCreatePost TButton = "button_compose-create_post",
    ComposeCreatePropose TButton = "button_compose-create_propose",
    ComposeCreateCancel TButton = "button_compose-create_cancel",
    // ...
)
```
It is likely also good to create a common convention in the interaction name id's where applicable, like done in the example with the buttons.

The example should not be taken as set in stone, it will likely not be in the same file and it may change some things based on if it fits better with the overall design of the infrastructure.

When using this structure, it is easier to maintain the locations of all interaction definitions with maps because the only thing the bot needs to identify an interaction is one of these types.

```go
// in ./internal/commands/commands.go
// Routes to all commands
type MCommandDefinitions map[TSlashCommand]SCommandDef

var CommandDefinitions MCommandDefinitions = MCommandDefinitions{
    ComposeCreate: {
        Definition: ComposeCreateDefinition,
        Execute: ComposeCreateExecute,
        // and any other relevant fields ...
    }
} 
```

### GuildCreate & GuildDelete

This event needs to be handled correctly too as the database needs to be updated based on these events.

### Error

Discord's api also has an error event that will be emited once an error occurs in discord itself.
This needs to be handled in multiple ways, like:
- Logging it in the terminal
- Informing a potential user that may have triggered the error that something gone wrong.
- possibly sending a formatted error embed to a logging channel (that more of a polish feature though)


# Instruction

It is your job to introduce this in all of this projects documentation.
Make sure to delegate tasks to Subagents where they apply, read the @.cursor/agents/INDEX.md to learn when to use Subagents and where they can assist you. You should essentially not have to do any actual work yourself and instead oversee the work that is done by the Subagents that you started.

Remember to also apply all applicable @.cursor/rules/ yourself at all times.