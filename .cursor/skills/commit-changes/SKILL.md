---
name: commit
description: Stage and commit changes using conventional commits format. Use when the user asks to commit, save changes to git, or create a commit.
---

# Commit

Use this skill when the user asks to commit, save changes, or create a git commit.

For commit message format (type, scope, description, examples), see `.cursor/rules/conventional-commits.mdc`.

## Workflow

1. Run `git status` to see what changed.
2. Stage files: `git add <paths>` or `git add -A`.
3. Commit with message: `git commit -m "type(scope): description"`.
4. For multi-line body: `git commit -m "type(scope): description" -m "body line 1" -m "body line 2"`.
