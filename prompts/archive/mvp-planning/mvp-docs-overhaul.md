Lead a complete documentation overhaul for this project. The docs in ./docs/ may not accurately reflect the actual implementation state.

Plan and execute this in phases. The phases below are a starting framework - add, remove, or restructure phases as you see fit to accomplish the goal efficiently:

**Phase 1: Discovery**
Delegate to a researcher agent to:
- Read all documentation in ./docs/ and ./docs/roadmap/
- Read all Go source files to establish ground truth
- Identify discrepancies between documented "planned" vs actual "exists" state

**Phase 2: Analysis**
Review the research findings and determine:
- Which docs need updates to reflect current implementation
- Which implemented features lack documentation
- How the MVP scope should be replanned based on actual progress

**Phase 3: Documentation Update**
Delegate to a documenter agent to update any and all documentation files in ./docs/ as needed to reflect the current implementation state. This includes but is not limited to:
- Entry points and indexes
- Project structure and architecture docs
- Roadmap and planning documents
- Any other files where the documented state diverges from the actual code

**Phase 4: Verification**
If significant changes were made, delegate to a reviewer agent to verify:
- Updated docs accurately reflect the actual code
- No new inaccuracies were introduced
- Cross-references between docs remain consistent

**Phase 5: Synthesis**
Produce a report at `./prompts/mvp-planning/mvp-docs-overhaul_report.md` containing:
1. Summary of all documentation changes
2. Remaining gaps between docs and code
3. Recommended next implementation priorities
4. Replanned MVP scope based on current progress

Oversee each phase, resolve blockers, and ensure the final docs accurately represent the codebase.