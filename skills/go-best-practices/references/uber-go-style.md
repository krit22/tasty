# Uber Go style: own failures and cleanup

Source: [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md),
especially Errors, Defer to Clean Up, Don't fire-and-forget goroutines, and Copy Slices
and Maps at Boundaries. Reviewed 2026-10-31. Organization-specific choices are not
universal requirements; existing contracts and current standard-library APIs matter.

**GO-ERROR:** Handle a failure where the program can decide what to do with it. Repeated
logging as the same error travels upward obscures ownership. A library generally
returns failures to callers, leaving the application to decide exit policy.

**GO-CLEANUP:** Pair acquisition with cleanup. A defer near acquisition makes early
returns easier to reason about. In loops, consider how long resources stay open.

**GO-LIFETIME:** Starting work also creates a shutdown obligation. The owner must
know how the goroutine stops and how to wait for it. Cancellation without completion
observation can still leave work running.

**GO-DATA-CONTRACT:** Maps and slices can share mutable storage. Decide whether callers
may mutate a retained value; copy where isolation is part of the contract. Do not copy
indiscriminately when borrowing or shared ownership is intentional.
