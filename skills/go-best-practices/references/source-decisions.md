# How Tasty selects guidance

These are Tasty's editorial decisions, not additional Go language rules.

1. Official language and package contracts resolve behavioral questions. Existing
   project compatibility and user requirements determine what may change.
2. Established ecosystem conventions guide new code. Google and Uber documents supply
   reasoning, but their organizational preferences do not override a project's needs.
3. Recommendations carry exceptions. Exported data fields, useful zero values, clear
   bool parameters, and concrete return types are legitimate Go patterns.
4. Introduce interfaces, generics, constructors, and custom errors only for an actual
   contract. We deliberately avoid copying Rust's trait and error-type requirements.
5. Preserve behavior while refactoring. A defect fix can change behavior when requested;
   identify it and test it. Don't hide API breaks under a style cleanup.
6. A source supports a rule; it does not establish that the skill improves model output.
   That requires separately recorded evaluations. Formatting and citation scores are
   reported separately from behavioral correctness.

The maintainer source and rule registry is in `sources/` in the repository. Release
skill archives include these self-contained digests and source links so they work
without fetching the repository or another skill.
