# Changelog

## 0.1.0-alpha.2 — 2026-10-31

- Add focused Go doc-comment guidance for affected public API contracts, sourced from
  the Go project's documentation. The first blinded review found missing contract
  comments in all six inventory refactors, with and without the original skill.
- Repeat the frozen inventory task and rubric as a separate six-run experiment.
  The original baseline and its negative result remain unchanged.
- In that follow-up, the blinded documentation assertion passed in 3/3 skill runs
  and 0/3 controls. Behavior passed 21/21 in each condition. This is an in-sample
  development result, not a general improvement claim.

## 0.1.0-alpha.1 — 2026-10-31

- First Go writing/refactoring skill with source-linked, contextual guidance.
- Source registry and rule-to-source traceability.
- Three frozen evaluation cases: SemVer library, JSONL CLI, and inventory refactor.
- Withheld contract tests, known-good implementations, and deliberate regression checks.
- Local validation, runnable documentation examples, deterministic ZIP packaging, and
  a runner for real with/without-skill comparisons.
- Experimental release. Behavioral evaluation and qualitative review are reported
  separately; general effectiveness has not been established.

- Initial evidence: 18 completed model tasks; both conditions passed 57/57 behavior
  checks and 24/27 independently reviewed qualitative checks. No improvement was shown.
