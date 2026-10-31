# Follow-up: documenting affected API contracts

## Observation

The original 18-run suite tied on behavior (57/57 each) and qualitative checks
(24/27 each). All six inventory refactors failed the existing clear-contracts rubric:
important exported operations lacked comments describing their non-obvious contract.
The original results and frozen rubric are retained unchanged.

## Treatment

Alpha 2 adds one focused instruction and a source digest based on Go Doc Comments:
when changing an exported API, describe observable returns, errors, ordering, ownership,
and side effects where relevant, beside the declaration. Change notes do not replace
API comments. Avoid unrelated boilerplate. No inventory-specific answer or test result
is put in the skill.

## Comparison

- Repeat only `refactor-inventory-module`, with fresh control and treatment outputs.
- Three repetitions in each condition: six model tasks, same runner and Go release.
- Keep task, input files, contract tests, and the three qualitative rubrics unchanged.
- Primary observation: the already existing clear-contracts rubric, independently
  graded with anonymous labels. Also report all other rubric and behavioral results.
- Retain build/test success and the existing API and data-shape constraints.
- Publish candidate source, evidence, usage, and the blind key after grading.
- Stop after this comparison regardless of the result; do not tune repeatedly against
  the same case until a favorable score appears.

## Interpretation boundary

This task was selected because its failure was observed during development. It is not
held-out validation. An improvement would show that a targeted instruction can address
this particular documentation gap, not that Tasty improves arbitrary Go work. Stronger
claims still need unseen tasks, additional agents, and the roadmap's broader evidence.
