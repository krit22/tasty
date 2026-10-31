# Tests that establish behavior

Sources: [Go Test Comments](https://go.dev/wiki/TestComments),
[testing](https://pkg.go.dev/testing), and [race detector](https://go.dev/doc/articles/race_detector).
Reviewed 2026-10-31.

**GO-TEST:** Exercise success and failure contracts. Inspect error semantics rather
than incidental wording when the API exposes error identity or types. Show the input,
actual result, and expected result in a failure. Use table-driven cases when that makes
relationships clearer; independent tests can be clearer for distinct behaviors.

Use temporary directories and local services. Explicit synchronization makes concurrent
behavior reproducible. Avoid assertions that merely mirror implementation branches.

The race detector reports races along executed paths. A passing run does not prove
that all schedules or code paths are safe. Test cancellation and completion explicitly
even if the race detector stays quiet.

Follow the project's existing test framework. The source's preference for standard
testing facilities is not a reason to replace a working third-party framework.
