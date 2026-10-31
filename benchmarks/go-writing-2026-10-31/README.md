# First Go writing comparison — 2026-10-31

**No behavioral improvement was demonstrated on this suite.** Both conditions passed
all 57 automated checks. This is a ceiling result for these tasks and this agent
configuration; it does not establish that the skill improves coding quality generally.

## Behavioral results

Each case ran three times per condition. The same 19 checks were applied to each full
three-case repetition, giving 57 checks per condition and 114 total observations.

| Case | Without skill | With skill |
| --- | ---: | ---: |
| SemVer library | 18/18 | 18/18 |
| JSONL CLI | 18/18 | 18/18 |
| Inventory refactor | 21/21 | 21/21 |
| **Total** | **57/57** | **57/57** |
| Candidate builds | 9/9 | 9/9 |
| Candidate + contract test suites | 9/9 | 9/9 |
| Runner errors | 0 | 0 |

[results.json](results.json) contains each assertion, evidence, run duration, reported
usage, evaluation hash, and hashes of the exact skill files supplied. Generated Go
sources, tests, change notes, and final responses are published under `artifacts/`.
Qualitative rubrics are separate from these behavioral scores.

## Method

- Three ordinary task prompts: a library API, a CLI, and a compatible refactor.
- Each task/condition ran three times in a fresh workspace: 18 completed model runs.
- Codex CLI 0.163.0, default model, with user config ignored. The runner did not report
  a model identifier; this limits exact model-level reproducibility. We do not infer one.
- Go 1.27.3. Task modules target Go 1.26; their final assessments use local Go tooling,
  a temporary cache, `GOWORK=off`, `GOTOOLCHAIN=local`, and `GOFLAGS=-buildvcs=false`.
- Conditions received the same task and starting files. The skill condition additionally
  received the skill and reference text and an explicit instruction to read them.
- Run order was shuffled with a recorded seed, with at most two concurrent tasks.
- The coding agent did not receive grader tests or reference implementations in its
  workspace. After generation, the independent contract tests were added to a copy.
- Reference implementations passed all 19 checks. Six buildable mutations verified
  detection of lost error causes, overflow, incorrect precedence, scanner limits,
  and silenced partial failures.
- Case inputs, validation tests, rubrics, references, and mutations were frozen in
  [evals.lock.json](../../evals.lock.json) before the model comparison.

Fresh workspaces are not hermetic containers. Host agent behavior beyond the explicitly
supplied content is not comprehensively controlled. Skill discovery was not measured.

## Assessment provenance

The run generator initially encountered host Git metadata that prevented VCS stamping
for some CLI assessments. That is a tooling issue, not a candidate defect. Every final
published candidate was assessed again with VCS stamping disabled; tasks, tests, skill
content, and candidate code were unchanged. Generation records were preserved locally.

The recorded grader hash matches `scripts/evaluate.py` in the
[first public commit](https://github.com/krit22/tasty/blob/79b18fd79531bc422c0b9d1ccac5116dddaac4fb/scripts/evaluate.py).
The current grader has the same grading behavior with more concise command-line summary
output. The shared helper hash and assessment flags are also recorded in results.json.
The export grades the exact sources it publishes, rather than a different working copy.

To reproduce one result, with Go on PATH:

```bash
python3 scripts/evaluate.py grade \
  --case semver-library-api \
  --workspace benchmarks/go-writing-2026-10-31/artifacts/semver-library-api--without--1
```

To reproduce the comparison with your own account, use the command in the repository
README and an empty output directory. Defaults may resolve to a different model later.
Running the local grader does not make model calls.

## Reported usage

| Metric, across nine runs | Without skill | With skill |
| --- | ---: | ---: |
| Input tokens, including cached | 1,135,012 | 1,368,146 |
| Cached input tokens | 984,576 | 1,184,768 |
| Output tokens | 32,762 | 31,252 |
| Sum of run elapsed seconds | 638.76 | 654.23 |

Input increased by about 20.5% without a behavioral-score increase. Token counts are
runner-reported usage, not a monetary cost calculation. Timing includes agent and
tool execution, varies with scheduling, and is not an isolated model-speed benchmark.

## Qualitative review

An independent fresh-context model reviewed all 18 anonymized candidates against the
three frozen qualitative rubrics. It saw source, candidate tests, change notes, and final
responses, but no condition labels, key, skill text, or reference implementations. The
key was applied only after the reviewer returned its decisions.

| Qualitative assertion | Without skill | With skill |
| --- | ---: | ---: |
| Appropriate API | 9/9 | 9/9 |
| Meaningful candidate tests | 9/9 | 9/9 |
| Clear contracts and verification | 6/9 | 6/9 |
| **Total** | **24/27** | **24/27** |

[qualitative-blind.json](qualitative-blind.json) preserves all 54 judgments and file/line
evidence. [blind-key.json](blind-key.json) reveals the mapping after grading. All six
failures were inventory refactors lacking documentation of non-obvious public API
behavior. Both conditions showed this gap.

This is one model's static judgment, not a deterministic result or a population estimate.
The reviewer did not independently rerun behavior or audit the reported tool commands.
Original supplied files were not available to it for checking test preservation.
These scores remain separate from the behavioral results.

The [forward test](../../examples/forward-options/CHECKS.md) is another separate check:
one fresh-context agent followed the skill while adding a parser without changing an
existing API. It passed, but had no no-skill control and proves no comparative benefit.

## What this changes next

Preserve the saturated suite and its result. Establish realistic failures in review,
request cancellation, resource ownership, and concurrent shutdown before extending the
guidance. Freeze new cases separately and repeat the comparison. Do not select a new
rubric merely because it favors the skill or change these tasks after seeing scores.

One agent configuration, three tasks, three repetitions, incomplete environmental
control, and a small hand-written test set limit generalization. Passing tests do not
prove the absence of bugs. Tasty remains an experimental alpha.
