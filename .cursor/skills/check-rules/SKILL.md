---
name: check-rules
description: Audit your own recent changes in this conversation against project rules. Read all .cursor/rules/ files and verify recent edits comply with established conventions. Use when the user asks to check rules or when reviewing your own recent work for compliance.
---

# Check Rules

Audit your recent changes against all project rules in .cursor/rules/.

## Audit Workflow

1. **Read all rules** - Read each file in .cursor/rules/ and extract key constraints
2. **Identify recent changes** - Review edits in this conversation; group by file or rule category
3. **Check compliance** - For each rule, verify your changes follow it (scope/glob patterns indicate which files a rule governs)
4. **Report findings** - If violations found, document file path, line numbers, rule, and issue. Otherwise confirm compliance. Do not fix violations.
