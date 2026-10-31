# Request cancellation and HTTP ownership

Sources: [context](https://pkg.go.dev/context), [net/http](https://pkg.go.dev/net/http),
[sync](https://pkg.go.dev/sync), and [Pipelines and cancellation](https://go.dev/blog/pipelines).
Reviewed 2026-10-31. Check the project's Go version before using newer APIs.

**GO-CONTEXT:** Connect request-scoped work to its caller's context. Call cancellation
functions when the derived context's work ends. A goroutine blocked on a channel send
needs an exit route if consumers can leave early. Cancellation signals a request to
stop; the owner still needs a way to observe completion.

**GO-LIFETIME / GO-SYNC:** State who starts, stops, and joins each goroutine. Close a
channel only when its producer side knows no further sends can occur. Guard shared
mutable state and do not copy synchronization primitives after use.

**GO-CLEANUP:** HTTP clients should reuse transports and close response bodies after a
successful request. Configure bounds from the actual operation. A non-2xx status is
not a transport error; interpret it explicitly. Avoid reading an unbounded remote body
when the application has a meaningful size limit.

Test with `httptest` or a controlled transport. Synchronize events rather than sleeping
and hoping a goroutine reached the relevant line. A timeout can bound a failed test.
