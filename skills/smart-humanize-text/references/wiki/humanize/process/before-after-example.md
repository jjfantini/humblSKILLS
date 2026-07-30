---
title: "Full Before/After Example: AI-Coding Essay"
context: humanize
category: process
concept: before-after-example
description: "Worked example: AI-sounding paragraph rewritten into human prose, with change log."
tags: process, example, before-after, walkthrough
sources:
  - "references/raw/Wikipedia_Signs_of_AI_writing.pdf"
last_ingested: 2026-04-17
---

## Full Before/After Example

End-to-end demonstration: an AI-sounding essay on AI-assisted coding, rewritten by stripping tells and adding voice.

### Before (AI-sounding)

> AI-assisted coding serves as an enduring testament to the transformative potential of large language models, marking a pivotal moment in the evolution of software development. In today's rapidly evolving technological landscape, these groundbreaking tools are reshaping how engineers ideate, iterate, and deliver, underscoring their vital role in modern workflows.
>
> At its core, the value proposition is clear: streamlining processes, enhancing collaboration, and fostering alignment. It's not just about autocomplete; it's about unlocking creativity at scale, ensuring that organizations can remain agile while delivering seamless, intuitive, and powerful experiences. The tool serves as a catalyst. The assistant functions as a partner. The system stands as a foundation.
>
> Industry observers have noted that adoption has accelerated from hobbyist experiments to enterprise-wide rollouts. Additionally, the ability to generate documentation, tests, and refactors showcases how AI can contribute to better outcomes, highlighting the intricate interplay between automation and human judgment.
>
> Despite challenges typical of emerging technologies - including hallucinations, bias, and accountability - the ecosystem continues to thrive. In conclusion, the future looks bright as we continue this journey toward excellence.

### After (human)

> AI coding assistants can make you faster at the boring parts. Not everything. Definitely not architecture.
>
> They're great at boilerplate: config files, test scaffolding, repetitive refactors. They're also great at sounding right while being wrong. I've accepted suggestions that compiled, passed lint, and still missed the point because I stopped paying attention.
>
> People I talk to land in two camps. Some use it like autocomplete for chores and review every line. Others disable it after it keeps suggesting patterns they don't want. Both feel reasonable.
>
> The productivity metrics are slippery. GitHub says Copilot users "accept 30% of suggestions," but acceptance isn't correctness, and correctness isn't value. If you don't have tests, you're basically guessing.

### What was fixed

- Killed "enduring testament," "pivotal moment," "evolving landscape," "groundbreaking," "vital role," "showcases," "intricate interplay," "fostering," "enhancing"
- Replaced "serves as / functions as / stands as" with "is" or cut entirely
- Removed negative parallelism ("It's not just X; it's Y")
- Removed rule-of-three patterns and synonym cycling
- Removed false range ("from X to Y, from A to B")
- Removed "Despite challenges... continues to thrive" formula
- Removed generic conclusion ("the future looks bright")
- Removed filler ("At its core," "Additionally")
- Added first-person, opinions, specific claims, varied sentence length
- Let sentences start with "Not," "They're," "Both," "If"

## Sources

- `references/raw/Wikipedia_Signs_of_AI_writing.pdf` - "Full Before & After Example" section
