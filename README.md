# Tasty

Sourced Go guidance for coding agents, with executable examples and reproducible evaluations.

**Experimental Go writing alpha.** Tasty currently ships one skill for writing,
debugging, and refactoring Go. Its examples and evaluation harness are tested; whether
it improves agent output is a separate question measured by the recorded comparisons.
Go review, porting, and TypeScript are later milestones, not current capabilities.

An agent can produce Go that compiles while losing error identities, changing a public
JSON shape, leaking request-scoped work, or adding unnecessary abstractions. Tasty
puts relevant design decisions and API contracts in the agent's context, then asks it
to verify the behavior. The agent still supplies its own model, tools, and permissions.

## Install

Install only the Go writing skill:

```bash
npx skills add krit22/tasty --skill go-best-practices
```

Or download the per-skill ZIP from [Releases](https://github.com/krit22/tasty/releases).
The archive contains a `go-best-practices/` folder with SKILL.md and its references.
Copy that folder into your agent's supported skill directory. For agents that support
the shared project location, that is `.agents/skills/go-best-practices/`; use your
host's own location where required. Restart or refresh skills if the host requires it.

Git-based installation also copies the maintainer evaluation files. The release ZIP
is the smaller payload and excludes those fixtures; the skill does not route agents
to the evaluation material.

Ask your agent to use `go-best-practices` when implementing or refactoring Go. For example:

> Use go-best-practices to refactor this package. Preserve its public API and explain
> any behavior changes and the checks you ran.

The skill is also eligible for automatic selection when the host supports it. Our
initial comparisons explicitly load the skill and do not establish discovery reliability.
Installing the skill does not add Go, linters, dependencies, or configuration to your project.

## What the skill changes

- It starts from the project's API, tests, supported Go version, and requested scope.
- It favors concrete designs and interfaces with a real consumer need.
- It preserves useful error contracts and distinguishes cleanup from garbage collection.
- It asks who owns cancellation, goroutine completion, and shared mutable state.
- It treats exported data fields, zero values, and clear bool parameters contextually.
- It distinguishes nil/empty data, ordering, numeric bounds, and other observable contracts.
- It uses tests and available tools, then reports what actually ran.

Detailed material is read only when relevant. The skill has no network service or
install-time code. See [PRIVACY.md](PRIVACY.md).

## Sources and editorial choices

The guidance synthesizes Effective Go, Go Code Review Comments, Google's Go style
documents, Uber's Go guide, and official Go language and package contracts. Specific
testing, cancellation, and error-wrapping articles support the corresponding topics.
These are original condensed notes, not copies of those manuals.

[The source registry](sources/sources.json) records URLs, reviewed sections, dates,
and contextual limits. [The rule registry](sources/rules.json) maps each named rule to
its sources, rationale, exceptions, and verification method. The skill's
[source decisions](skills/go-best-practices/references/source-decisions.md) explain how
company recommendations and project conventions are reconciled.

## Evidence

The initial suite has three tasks: a SemVer library, a JSONL CLI, and an inventory
refactor. Each includes independent contract tests, a known-good implementation,
and deliberate buildable regressions that the grader must detect. The inventory task
also checks preservation of exported plain data and an existing bool parameter.

There are 19 automated behavioral checks per complete three-task run. Qualitative
Markdown rubrics cover API design, meaningful candidate tests, and documentation;
they are reported separately and are not automatically counted as passed.

The [first comparison](benchmarks/go-writing-2026-10-31/README.md) completed 18 fresh
Codex CLI tasks: three cases, two conditions, three repetitions. Both conditions passed
57/57 automated behavioral checks. **This suite demonstrates no behavioral improvement
from the skill on the tested configuration.** The tasks appear too easy for this runner
to distinguish the conditions. This result does not assess every design recommendation,
other models, or automatic discovery. A separate blinded review also tied at 24/27
qualitative rubric checks per condition; it found missing API-contract documentation
in every inventory refactor.

A passing fixture suite validates the harness, not the effectiveness of the skill.
See [ROADMAP.md](ROADMAP.md) for the evidence gates and next experiments.

## Reproduce the local checks

Use Python 3.10+ and Go 1.26 or later. CI checks the documented examples and contracts
on Go 1.26.10 and 1.27.3. Set `TASTY_GO` if Go is not on PATH.

```bash
python3 scripts/check.py
python3 scripts/examples.py
python3 scripts/evaluate.py prove
python3 -m unittest discover -s tests
python3 scripts/package.py
```

The model comparison is a separate operation that uses your Codex CLI account:

```bash
python3 scripts/evaluate.py run --out /tmp/tasty-comparison --repetitions 3 --parallel 2
```

That schedules 18 fresh tasks. Each condition receives the same prompt and inputs;
only one receives the skill. Withheld tests are added later in a separate grading
workspace. Runner failures are recorded separately from candidate failures. See
[CONTRIBUTING.md](CONTRIBUTING.md) for the protocol and frozen-input policy.

## Build in public

The development sequence follows [BeautifulRust](https://github.com/morganlinton/BeautifulRust):
source synthesis, a writing skill, evaluation-driven revision, review plus enforcement,
porting, domain-specific guidance, and broader benchmarks. Tasty's Go implementation
and notes are original work. We are starting with the first complete experiment rather
than claiming every planned language and workflow is available.

Report a concrete agent mistake, harmful recommendation, missing exception, or install
problem through the issue templates. Contributions should connect advice to a source
and a reproducible task. See [CONTRIBUTING.md](CONTRIBUTING.md) and [CHANGELOG.md](CHANGELOG.md).

## License

Tasty's original skill text and tooling are MIT licensed; see [LICENSE](LICENSE).
Referenced publications retain their own terms. Attribution and source links are kept
with the guidance.
