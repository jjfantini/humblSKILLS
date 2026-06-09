# Decisions

Reasoning memory. Append non-obvious approval, privacy, and routing choices.

Entry shape:

```markdown
### <YYYY-MM-DD> | <short title>
- Context: <situation>
- Options: <choices considered>
- Chose: <choice>
- Why: <reason>
- Result: <outcome or TBD>
```

---

### 2026-06-09 | Keep Writes Approval-Gated
- Context: TweetClaw can handle both read-only lookups and write-like X/Twitter
  actions through OpenClaw.
- Options: (A) call tools directly from any prompt, (B) require exact operator
  approval before write-like actions.
- Chose: B, exact operator approval.
- Why: Posting, replies, DMs, media upload, monitors, and webhooks change an
  account or external state.
- Result: The skill routes write-like jobs through explicit review before use.
