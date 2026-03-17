---
name: compose-prompt
description: Compose effective prompts for subagents by improving clarity, structure, conciseness, and effectiveness. Use when writing or refining prompts for any subagent to ensure clear instructions, proper context, and minimal token usage.
---

# Compose Prompt

Use this skill to improve prompts before sending them to subagents.

## When to Apply

- Writing a new prompt for any subagent
- Reviewing or refining an existing prompt
- Converting user requests into proper subagent prompts
- Reducing token count while preserving intent

## Optimization Checklist

Run through these checks for every prompt:

- [ ] Spelling and grammar are correct
- [ ] Instructions are unambiguous and specific
- [ ] No redundant or filler language
- [ ] Sufficient context provided (file paths, relevant code, constraints)
- [ ] Clear structure with logical sections
- [ ] Expected output format is specified
- [ ] Token count is minimized (no unnecessary words)

## Optimization Steps

1. **Analyze the prompt for issues:**
   - Ambiguous terms or vague instructions
   - Missing file paths or code references
   - Unclear expected outcomes
   - Unnecessary words or phrases

2. **Apply fixes:**
   - Fix all spelling and grammar
   - Replace vague terms with specific ones
   - Add missing context (relevant files, code snippets)
   - Remove redundancy and filler words
   - Structure with clear sections
   - Specify output format explicitly

3. **Reduce token count:**
   - Remove phrases like "I want you to", "Please", "Make sure to"
   - Use imperative voice: "Analyze the code" not "Can you analyze the code"
   - Replace verbose descriptions with concise ones
   - Use code references instead of pasting large blocks when possible

## Examples

**Before:**
```
I need you to look at my project and find any bugs. Please make sure to 
check all the files carefully and tell me what you find. Thanks!
```

**After:**
```
Find bugs in the Go codebase at ./src. Focus on:
- Error handling in API handlers
- Resource leaks in database connections
- Race conditions in concurrent code

Report each bug with: file location, line numbers, problem description, suggested fix.
```

**Before:**
```
Can you please help me understand how the authentication system works? 
I'd really appreciate it if you could look at the relevant files and 
explain it to me in detail.
```

**After:**
```
Explain the authentication flow:
1. Read src/auth/handlers.go and src/auth/middleware.go
2. Trace how a login request is processed from HTTP handler to database
3. Identify the JWT generation and validation logic
4. Report the flow as numbered steps with file:line references
```

## Constraints

- Never execute the prompt you are optimizing - only edit it
- Preserve the original intent and requirements
- Do not add new requirements unless needed for clarity
- Keep tone professional and direct
- Prioritize clarity over brevity - but remove true waste
