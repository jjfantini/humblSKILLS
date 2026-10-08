GIL_YOUN_LEE | 2026-09-21 10:23:03 UTC | #1

### Where does the bug appear (feature/product)?
Cursor CLI

### Describe the Bug
On the latest CLI (2026.09.18-9a7762b, Ultra, already up to date), --help and the --list-models tip still tell you to pass quoted bracket overrides:

--model 'claude-opus-4-8[context=1m,effort=high,fast=false]'

Using that syntax with --print is rejected before the model starts. Same result with claude-sonnet-5-high[context=1m].

The allowlist only has flat ids. --list-models labels claude-sonnet-5-high as "Claude Sonnet 5 1M", but a successful --print session on that id reports: Claude Sonnet 5 300K High No Thinking.

Related older threads:
https://forum.cursor.com/t/cursor-cli-agent-doesnt-support-square-bracket-model-name-variants/163905
https://forum.cursor.com/t/selecting-opus5-1m-context-through-cli-args-isnt-possible/166873

The second thread was marked fixed on 9 Sep 2026, with the note that you need context=1m in the --model string. On this build that string is not in the allowlist, so --print never gets that far.

### Steps to Reproduce
1. agent --version  ->  2026.09.18-9a7762b
2. agent --help still shows the bracket example above. agent --list-models ends with the same tip.
3. Run:

agent --print --output-format stream-json --trust --sandbox enabled --model 'claude-sonnet-5-high[context=1m]' --workspace /tmp/any -- 'Return {"ok":true} only.'

4. Repeat with the help example: --model 'claude-opus-4-8[context=1m,effort=high,fast=false]'
5. Then run the same command with the flat id --model claude-sonnet-5-high and inspect the first system/init event.

### Expected Behavior
Either --print accepts the documented '...[context=1m,...]' form and actually opens a 1M window, or --help / --list-models stop advertising that form and do not label a 300K session as 1M.



### Operating System
MacOS

### Version Information
CLI Version         2026.09.18-9a7762b
Latest              2026.09.18-9a7762b (up to date)
Subscription Tier   Ultra
OS                  darwin (arm64)
Shell               zsh

### For AI issues: which model did you use?
claude-sonnet-5-high (flat id). Documented --model 'claude-sonnet-5-high[context=1m]' and help example 'claude-opus-4-8[context=1m,effort=high,fast=false]' are rejected.

### For AI issues: add Request ID with privacy disabled
1e88b374-0604-4859-bb06-9ea907e5fa47  (successful --print on flat claude-sonnet-5-high; the rejected bracket calls never start, so they have no request id)

### Additional Information
Actual results from this machine just now:

Bracket forms: exit 1, empty stdout, stderr:
Cannot use this model: claude-sonnet-5-high[context=1m]. Available models: ... claude-sonnet-5-high ...
(no bracket ids, no bare claude-opus-4-8)

Same reject for the exact help example claude-opus-4-8[context=1m,effort=high,fast=false].

Flat id --print init event:
{"type":"system","subtype":"init","model":"Claude Sonnet 5 300K High No Thinking",...}

--list-models line for that id:
claude-sonnet-5-high - Claude Sonnet 5 1M

This is a headless --print gap. Interactive /model still has a Context 300K/1M control, but scripts cannot select 1M the way the help text describes.

### Does this stop you from using Cursor
No - Cursor works, but with this issue

-------------------------

mohitjain | 2026-09-21 10:41:15 UTC | #8

Hey @GIL_YOUN_LEE, thanks for the detailed report. You're right on both counts, and it's nothing on your end.

`--model` only accepts the full variant string (base id plus every parameter, in order), not a partial `[context=1m]` or the shortened `--help` example. These work with `--print` and open a real 1M window (I confirmed the first `system`/`init` event reports 1M):

```
--model 'claude-sonnet-5[thinking=false,context=1m,effort=high]'
--model 'claude-sonnet-5[thinking=true,context=1m,effort=high]'
--model 'claude-opus-4-8[thinking=true,context=1m,effort=high,fast=false]'
```

The flat `claude-sonnet-5-high` id resolves to the 300K variant, so that's what your session ran. The "1M" text next to it in `--list-models` is a labeling mistake, not a 1M session. We've let the team know about both the help/list text and that gap, and it's something we're tracking.

Could you try one of the strings above and reply with the `model` value from the first `system`/`init` event? It should read 1M. One heads-up: a 1M window uses more usage per request than 300K.

-------------------------

