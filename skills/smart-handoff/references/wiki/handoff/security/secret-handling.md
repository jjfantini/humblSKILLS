---
title: "Redact Secrets, Transfer Access Instructions Instead"
context: handoff
category: security
concept: secret-handling
description: "A handoff doc is a plaintext file on disk that gets pasted into another agent - a credential in it is a credential leaked"
tags: security, secrets, redaction, keychain, 1password, pii
sources: []
last_ingested: 2026-08-03
command: scripts/scan-secrets.sh
---

## The Threat Model

A handoff doc lands in `/tmp`, gets pasted into a second agent's context, and
often into a third harness's telemetry. Any credential written into it is
now in at least three places the user did not choose. Treat the document as
**public within the machine**.

## Always Redact

Run the scan before declaring the doc ready — never rely on eyeballing it:

```bash
bash scripts/scan-secrets.sh <path-to-handoff>.md
```

Exit 1 means a live-shaped credential is present. Fix it and re-run. `FAIL`
covers API keys (Anthropic, OpenAI, UploadThing, Stripe, Google), GitHub
tokens and PATs, AWS access keys, Slack tokens, JWTs, private-key blocks,
bearer headers, and credentials embedded in URLs. `WARN` covers inline
`key = "value"` assignments, email addresses, and phone numbers — judgment
calls, so read them rather than ignoring them.

## Transfer Access, Not Values

Replace every value with the command that retrieves it:

```markdown
## Access & Secrets

- **GitHub (personal account)** — `GH_TOKEN=$(gh auth token --user <login>)`
- **UploadThing (files app)** —
  `security find-generic-password -s uploadthing -a app-files -w`
- **Staging DB URL** — `op read "op://Engineering/staging-db/url"`
```

## When a Token Genuinely Must Cross

Sometimes the receiving agent needs a value the user holds only in this
session's context — a one-off token pasted into chat, a short-lived session
key. Do **not** write it into the doc. Instead:

1. **Flag it, unprompted.** "Step 4 needs the Linear API token you pasted
   earlier. I won't put it in the handoff."
2. **Offer keychain transit.** The user runs this themselves in a real
   terminal, so the value never enters an agent transcript:
   ```bash
   security add-generic-password -s handoff-transit -a linear-api -U \
     -T /usr/bin/security -w
   ```
   Note: that interactive prompt gets no tty through an agent's shell
   passthrough and silently stores a bare newline. It must be run by the user.
3. **Reference the account, not the value, in the doc.**
   ```markdown
   - **Linear API token** — `security find-generic-password -s handoff-transit -a linear-api -w`
     (transit-only; delete after use)
   ```
4. **Instruct deletion in the doc**, since transit storage is not storage:
   ```bash
   security delete-generic-password -s handoff-transit -a linear-api
   ```

If the user says the secret is already in 1Password or the keychain
permanently, skip the transit dance and just cite the `op read` /
`find-generic-password` path.

## Never Do This

### Incorrect

```markdown
- The API key is `sk-ant-api03-xxxxEXAMPLExxxx`, export it as ANTHROPIC_API_KEY
- DB: postgres://admin:hunter2@db.example.com:5432/app
```

### Correct

```markdown
- **Anthropic key** — `op read "op://Personal/anthropic/api-key"`, export as
  `ANTHROPIC_API_KEY`
- **DB** — `op read "op://Engineering/app-db/connection-string"`
  (user `admin`, host `db.example.com:5432`, database `app`)
```

Host, port, username, and database name are configuration and may stay. The
password may not.

## Sources

- (none) - authored from the smart-handoff design session.
