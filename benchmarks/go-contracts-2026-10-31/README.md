# Follow-up: affected public API contracts — 2026-10-31

**The focused documentation assertion improved on this development case.** The blinded
reviewer passed it for 3/3 outputs with alpha 2 and 0/3 outputs without the skill.
Both conditions passed all 21 behavioral checks. This is not held-out evidence of a
general improvement in Go coding quality.

## Why this comparison exists

The [first comparison](../go-writing-2026-10-31/README.md) tied on all behavioral and
qualitative totals. All six inventory refactors lacked important public API contract
documentation. We added one instruction and one digest based on
[Go Doc Comments](https://go.dev/doc/comment), then repeated that case with fresh controls.
The [protocol](../../experiments/api-contracts.md) describes the treatment and stopping
condition. We did not change the task, starting files, tests, or rubric.

## Results

| Observation | Without skill | With alpha 2 |
| --- | ---: | ---: |
| Completed model tasks | 3 | 3 |
| Behavioral checks | 21/21 | 21/21 |
| Candidate builds and test suites | 3/3 | 3/3 |
| Appropriate API rubric | 3/3 | 3/3 |
| Meaningful tests rubric | 3/3 | 3/3 |
| Clear contracts rubric | 0/3 | 3/3 |
| **Total qualitative checks** | **6/9** | **9/9** |

The controls again left non-obvious exported contracts undocumented. The revised-skill
outputs added declaration comments covering relevant errors, replacement semantics,
ordering, repeated-name totals, and failed-operation behavior. The reviewer noted some
broad wording in Load comments but accepted the combined explanation and change notes.
See the exact evidence rather than treating the pass labels as definitive judgments.

## Evidence and reproduction

- [results.json](results.json) records the independent behavioral assessments, runtime
  metadata, skill and artifact hashes, and reported usage.
- `artifacts/` contains the generated Go code, candidate tests, change notes, and final
  responses used for the assessment.
- [qualitative-blind.json](qualitative-blind.json) preserves all 18 rubric decisions
  with file/line evidence. [blind-key.json](blind-key.json) reveals the mapping afterward.
- The same frozen [evaluation lock](../../evals.lock.json) was used in both comparisons.
- The assessment tool hashes match the scripts at the
  [documentation-revision commit](https://github.com/krit22/tasty/tree/be95a093fb88c31a852bbc745ff715494ee57110/scripts).

With Go on PATH, reproduce a behavioral assessment locally:

```bash
python3 scripts/evaluate.py grade \
  --case refactor-inventory-module \
  --workspace benchmarks/go-contracts-2026-10-31/artifacts/refactor-inventory-module--with--1
```

No model call is needed for this command. To generate fresh candidate outputs, use
`evaluate.py run --case refactor-inventory-module` with an empty output directory.

## Method and limitations

Six fresh tasks used Codex CLI 0.163.0 with user config ignored and the default model;
the runner did not report its model identifier. Conditions were shuffled and ran with
at most two concurrent tasks. Go 1.27.3 executed the assessments. VCS stamping was
disabled consistently from the start of this follow-up.

A new fresh-context reviewer saw six anonymous candidates, the unchanged task, and the
three original rubrics. It did not receive the skill, previous results, treatment
description, or condition key. The key was applied after grading. Candidate text did
not explicitly identify Tasty or the skill. The judgments are one model's static review;
it did not rerun behavior, audit reported commands, or see the original supplied files
for independently checking test preservation.

This is a development case selected because of its previous failure. It is not an
unseen task, a cross-agent benchmark, a discovery test, or a statistically established
effect. The independent reviewer also differs from the baseline reviewer in context,
so compare the fresh conditions within this follow-up rather than treating all rounds
as one controlled experiment. No behavioral regression was observed on these tests;
that does not establish correctness on all inputs.

## Reported usage

| Metric, across three runs | Without skill | With alpha 2 |
| --- | ---: | ---: |
| Input tokens, including cached | 310,657 | 439,564 |
| Cached input tokens | 263,168 | 372,352 |
| Output tokens | 9,736 | 10,204 |
| Sum of elapsed seconds | 231.25 | 245.19 |

Input usage was about 41.5% higher with the skill. These are runner-reported tokens,
not prices, and timing includes tools and scheduling. The result supports this specific
documentation change; it does not establish that the additional context is worthwhile
for every task. Future evidence should cover unseen work and substantive review,
cancellation, resource, and concurrency failures.
