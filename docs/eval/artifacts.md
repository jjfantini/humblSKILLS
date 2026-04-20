# Eval artifacts

Each iteration writes artifacts under:

`$XDG_STATE_HOME/humblskills/evals/<skill>/iteration-N/`

Typical layout:

```text
iteration-N/
├── benchmark.json      # cross-section stats + deltas
├── trajectory.json     # per-session time series (smart arm compounds here)
├── report.html         # single-file Plotly dashboard
├── report.md           # plaintext mirror (PR-friendly)
├── report.json         # machine-readable
├── smart_skill/
│   └── session-NN/
│       ├── outputs/              # files the agent wrote
│       ├── transcript.txt      # full agent transcript
│       ├── timing.json         # tokens, duration, cost
│       ├── metrics.json        # tool-call counts + brain reads
│       ├── brain-snapshot-before/
│       └── brain-snapshot-after/ # feeds session N+1 for smart arm
├── flat_skill/...
└── no_skill/...
```

Iterations are **persistent** and **append-only**. Use `humblskills eval prune` to cap how many iterations you keep per skill.
