# Tasty development sequence

The project follows BeautifulRust's method: sourced guidance, a writing skill,
evaluation-driven revision, review and enforcement, porting, domain exceptions,
broader evidence, and distribution. The Go skill text and evaluation implementation
are original Tasty material; BeautifulRust is the methodological reference.

| Stage | Acceptance evidence | Status |
| --- | --- | --- |
| Go writing alpha | Sourced skill; three frozen cases; executable checks; recorded baseline | Complete: 18 runs; both conditions 57/57 behavioral checks |
| Revision from observed failures | Narrow changes tied to actual outcomes, followed by a new recorded comparison | Complete: focused docs assertion 3/3 with, 0/3 without; behavior 21/21 each |
| Go review | Correct prioritized findings; clean controls; locations and cited sources | Planned |
| Enforcement preset | Every advertised diagnostic demonstrated by a failing fixture and clean counterexample | Planned |
| Go porting | C parser and Python graph ports with independent behavior comparisons | Planned |
| HTTP/concurrency guidance | Cancellation and lifecycle feature/review cases, with valid patterns left alone | Planned |
| Broader benchmark | Two agent configurations, repeated runs, independent tasks, failures and costs published | Planned |
| TypeScript | Repeat source research and evaluation independently; reuse proven tooling | Later |

The initial baseline has three tasks, two conditions, and three repetitions: 18 runs.
Record automated behavior separately from qualitative rubrics. Explicit loading and
natural discovery are separate experiments. A stronger model or easy tasks can leave
little headroom; equal scores are evidence of no demonstrated benefit on those checks.

That ceiling occurred in the first comparison. Preserve these results. Before claiming broad behavioral benefit, investigate realistic review, cancellation,
resource ownership, and concurrency
tasks where ordinary agent output actually fails. Freeze a new suite separately rather
than changing this suite to manufacture a favorable comparison. Alpha 2 separately addresses the observed documentation gap in a targeted six-run
comparison. The current alpha remains an installable experiment, not a claim of general
improvement on unseen tasks.

For a later stable claim, predeclare a target of at least 20% fewer failed substantive
assertions overall, improvement in both tested agent configurations, no reduction in
build/test success, no unresolved recurring skill-induced critical regression, and no
false high-severity findings on clean review controls. These are acceptance targets,
not predicted outcomes. Narrow the claim or revise the content if evidence misses them.
