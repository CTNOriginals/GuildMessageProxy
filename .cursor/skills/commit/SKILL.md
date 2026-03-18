---
name: commit
description: Stage and commit changes using conventional commits format. Use when the user asks to commit, save changes to git, or create a commit.
disable-model-invocation: true
---

# Commit

Stage and commit changes using conventional commits format. Apply the `conventional-commits` skill for message format details.

## Grouping

Commit changes in logical chunks:

- Group files in each commit so they relate to each other (e.g., feature code + its tests, or all docs for one change)
- Prefer small commits with few files per commit
- Exception: bulk changes that affect many files in the same way (e.g., a project-wide rename or format pass) may be one commit

## Self-contained commits

Each commit must be self-contained and able to function on its own. Do not commit anything that loses relevant context because that context is in another commit.

- If Y references or depends on something defined in X, include both X and Y in the same commit, or commit X before Y
- Do not commit Y without X when Y would be broken or meaningless without it
- Order commits so that definitions land before or with their references

## Workflow

1. Run `git status` to see what changed.
2. Stage related files together: `git add <paths>` (avoid `git add -A` when changes span multiple logical commits).
3. Commit with message: `git commit -m "type(scope): description"`.
4. For multi-line body: `git commit -m "type(scope): description" -m "body line 1" -m "body line 2"`.
5. Repeat steps 2-4 for each logical chunk until all changes are committed.

## Granular Line-Level Commits

When a file contains multiple unrelated changes, stage specific lines rather than the entire file:

### Using `git add -p` (Interactive Patch Mode)

Interactively choose which hunks to stage:

```bash
git add -p <file>
```

For each hunk, choose:
- `y` - stage this hunk
- `n` - do not stage this hunk
- `s` - split this hunk into smaller hunks
- `e` - manually edit the hunk

### Using `git add -e` (Edit Mode)

Manually specify which lines to stage by editing the diff:

```bash
git add -e <file>
```

Edit the diff to keep only the changes you want to stage. Lines removed from the edit will remain unstaged.

**Example workflow for granular commits:**

```bash
# File has both feature code and bug fix
git add -p src/handlers.go  # Stage only the feature hunks
git commit -m "feat(handlers): add user validation"

git add -p src/handlers.go  # Stage remaining bug fix hunks
git commit -m "fix(handlers): correct null pointer check"
```
