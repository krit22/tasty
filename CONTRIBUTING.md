# Contributing

Every recommendation needs a published source, a reason, and an applicability boundary.
Before proposing a rule, show the actual agent failure it would help prevent. Preserve
valid exceptions; a company convention is not a language requirement.

Skills live under `skills/`. Keep entrypoints focused and route deeper material into
references. Keep the package usable without another skill or a network connection.
Use standard Agent Skills metadata and agent-neutral instructions.

Run local checks before proposing a change:

```bash
python3 scripts/check.py
python3 scripts/examples.py
python3 scripts/evaluate.py prove
python3 -m unittest discover -s tests
python3 scripts/package.py
```

Go must be on PATH, or set `TASTY_GO` to the Go executable. Python 3.10+ is required.
The tooling uses temporary workspaces and a cache under `/tmp/tasty-go-cache`; override
the latter with `TASTY_CACHE`. It sets `GOTOOLCHAIN=local`, `GOWORK=off`, and
`GOFLAGS=-buildvcs=false` for reproducible fixtures independent of host Git metadata.

## Evaluations

The `case.yaml` files contain JSON, a YAML subset readable without additional Python
dependencies. Only `files/` and the ordinary task prompt are supplied to the coding
agent. `validation/` holds withheld behavioral tests. `reference/` holds a known-good
implementation for testing the grader, not the answer given to an evaluated agent.
`mutations.json` defines deliberate, buildable defects the grader must detect.

The checks verify behavior, not source-code spelling. The separate Markdown rubrics
need an evidence-backed review and are never silently included in the automated score.
Do not weaken a test because an agent fails it. Fix a genuinely incorrect test openly.

`evals.lock.json` freezes all case inputs, tests, rubrics, references, and mutations.
For an intentional evaluation revision, explain the change, run `python3 scripts/freeze.py`,
and commit the new lock with the changed files. This creates a new evaluation revision;
do not aggregate its results with old ones. Existing benchmark evidence remains immutable.

For a real comparison (uses model calls):

```bash
python3 scripts/evaluate.py run --out /tmp/tasty-comparison --repetitions 3 --parallel 2
```

The output directory must be empty. Preserve runner failures separately from completed
candidate failures. Record the agent version and model when known; do not invent a
model identity when a runner does not report it. Inspect any evidence before publishing.
Explicit skill loading measures content effectiveness, not automatic skill discovery.

## Releases

Update VERSION and CHANGELOG, run the checks, and build the per-skill ZIP. The ZIP has
the skill folder as its root and excludes deliberately faulty evaluation material.
Publish only claims supported by the recorded evidence. TypeScript, review, porting,
and any MCP delivery remain future milestones until individually implemented and tested.
