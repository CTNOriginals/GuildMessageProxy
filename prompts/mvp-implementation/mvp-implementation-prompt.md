# MVP Implementation - Project Leader Prompt

Implement the GuildMessageProxy MVP as defined in docs/roadmap/mvp-feature-plan.md.

## Goal

Create a working MVP with the core compose -> preview -> post flow and basic editing capability.

## Prerequisites

Read these files before planning:
1. docs/roadmap/mvp-feature-plan.md - complete MVP specification
2. docs/ARCHITECTURE.md - package layout and conventions
3. docs/ROUTE_MAP.md - routing and handler flows
4. internal/commands/registry.go - command registration system
5. internal/events/interaction_create.go - interaction routing

## Scope

### In Scope

1. **Compose Command Group**
   - `/compose create` - start a new draft with content input
   - `/compose post` - post the draft to a target channel
   - `/compose propose` - propose an edit to an existing proxied message

2. **Core Flows**
   - Compose -> Preview -> Post (ephemeral preview with Post/Cancel buttons)
   - Basic Edit (preview edited message with Apply/Cancel buttons)

3. **Handler Logic**
   - internal/handlers/compose.go - compose, preview, post logic
   - internal/handlers/preview.go - render ephemeral preview
   - internal/handlers/edit.go - edit proposal and apply

4. **Storage Integration**
   - Store proxy message metadata (guild_id, channel_id, message_id, owner_id, timestamp)
   - Update storage operations for edit tracking

### Out of Scope

- Voting / approval workflows
- Complex embed builders
- Persistent audit logs
- Multi-user editing permissions

## Implementation Plan

### Phase 1: Command Definitions

Add command definitions to internal/commands/:
- compose.go - command group with create, post, propose subcommands
- Wire into registry.go

### Phase 2: Handler Package

Create internal/handlers/:
- compose.go - create flow, preview generation
- post.go - webhook posting logic
- edit.go - edit proposal and apply logic
- preview.go - shared preview rendering

### Phase 3: Event Integration

Wire handlers into internal/events/interaction_create.go:
- Route compose commands to handlers
- Handle Post/Cancel/Apply buttons

### Phase 4: Storage Updates

Update internal/storage/:
- Extend GuildConfig if needed
- Add proxy message storage operations
- Track edit history minimally

### Phase 5: Verification

Build and verify:
- go build ./cmd/bot
- All linter checks pass
- Flows work end-to-end

## Documentation Updates

Update docs/ to reflect implementation:
- docs/ROUTE_MAP.md - document command flows
- docs/INDEX.md - mark MVP complete

## Deliverables

- Working /compose command group
- Ephemeral preview with buttons
- Webhook posting
- Basic edit flow
- All code compiles with no lint errors

## Constraints

- Use existing patterns from internal/commands/ and internal/events/
- Follow project conventions (var declarations, no m-dashes)
- Log appropriately using internal/logging/
- Handle Discord errors via RespondWithError
