## Permissions and Safety Notes - GuildMessageProxy

This document captures early thoughts about risks, guardrails, and governance for GuildMessageProxy.
It is intentionally lightweight and will be refined as features mature.

---

### 1. Key Risks

- **Impersonation and trust**
  - Proxy messages are posted via webhooks, which support custom avatar and username per message. Users could configure messages to appear as if another user said them.
  - Clear attribution and governance are needed to mitigate impersonation risk.
- **Abuse and spam**
  - High-volume or automated use of proxy messages could flood channels.
  - Malicious users could repeatedly edit messages to evade moderation.
- **Sensitive content**
  - NSFW, harassment, or other policy-violating content could be proxied by the bot, amplifying impact.
- **Moderation confusion**
  - Moderators may not immediately know who actually initiated a proxy message or who last edited it.

---

### 2. MVP Guardrail Ideas

These are candidate constraints for the MVP; they do not all need to be implemented at once, but they guide design.

- **Who can create proxied messages**
  - Restrict the `/compose` command group to:
    - Users with specific roles (e.g. staff or trusted contributors), or
    - Everyone, but only in approved channels.
- **Who can edit**
  - MVP: only the original requester can edit their proxy message.
  - Future: allow staff roles or a small whitelist of roles to edit important messages.
- **Attribution**
  - Always show who requested and last edited the message:
    - Example: footer text like "Requested by @User • Last edited by @Editor".
  - Do not attempt to fully mimic another user's identity in the UI without clear attribution.
- **Basic rate limiting**
  - Limit how often a single user can:
    - Create new proxy messages in a given time window.
    - Edit the same message in quick succession.
  - Surface friendly error messages when limits are hit.

---

### 3. Operational Safety Practices (Future)

These items are good candidates for later iterations and should be kept in mind while designing the MVP:

- **Logging**
  - Send a summary of each proxy message and edit to a dedicated log channel (configurable per guild).
  - Include who requested/edited, what channel was targeted, and timestamps.
- **Approval workflows**
  - Require a certain number of staff approvals before:
    - Posting a proxy message in sensitive channels.
    - Editing or deleting “pinned” proxy messages (e.g. rules, announcements).
- **Configuration defaults**
  - Choose conservative defaults for new guilds:
    - Limit usage to staff-only roles initially.
    - Require explicit opt-in to allow broader usage.

This document should evolve alongside `mvp-feature-plan.md` and any future feature-specific docs (e.g. for voting and approvals).

