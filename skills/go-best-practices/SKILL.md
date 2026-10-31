---
name: go-best-practices
description: Write, debug, and refactor idiomatic Go code, including packages, CLIs, HTTP clients, error handling, and concurrent work. Apply when implementing or changing Go code or go.mod. Use the relevant references for API contracts and design tradeoffs; preserve the project's behavior and supported Go version.
license: MIT
---

# Go best practices

Use published Go guidance to make concrete implementation decisions. Read the existing
code, callers, tests, and `go.mod` before changing them. A small tool, a reusable package,
and a long-running service have different needs. Preserve the user's requested scope.

## Establish the contract

- Identify observable inputs, outputs, ordering, errors, side effects, and compatibility
  requirements. Preserve those during refactoring. Pin undocumented behavior with a
  characterization test before changing the implementation. Flag suspicious behavior
  rather than silently changing it.
- Follow the module's `go` directive, build tags, dependency versions, and local patterns.
  Do not upgrade the toolchain or add a dependency just to use a preferred idiom.
- Language and standard-library contracts take precedence over company style guides.
  Local conventions settle reasonable style choices. Our [source decisions](references/source-decisions.md)
  distinguish recommendations from requirements. Rules have stable `GO-...` identifiers
  in the references so a reviewer can trace a recommendation to its source.

## Keep the design small

- **GO-SIMPLE:** Prefer concrete types and ordinary functions. Introduce an interface
  at the consumer when it needs substitution or a smaller contract, not one per struct.
  Accept useful standard contracts such as `io.Reader`; do not invent a parallel one.
  Generics should remove real duplication while keeping the algorithm understandable.
- **GO-API:** Design names in their package context. Avoid stuttering and catch-all
  `util` packages. Keep required functionality easy to find. Export only useful API.
  A usable zero value is valuable; constructors belong where initialization or validation
  requires them, not automatically on every type.
- **GO-DATA:** Exported fields on configuration, request/response, and plain-data structs
  can be appropriate. Encapsulate state when methods must preserve an invariant. Do not
  add getters, setters, builders, or wrapper types solely to imitate another language.
- **GO-OPTIONS:** Use an options struct or named operations when several positional
  arguments obscure meaning. A clear boolean argument does not require an enum. Preserve
  a public signature unless the task allows a change; explain unavoidable changes.
- Read [Effective Go](references/effective-go.md) for composition and zero values,
  [Google's guidance](references/google-go-style.md) for simplicity and API choices,
  and [review conventions](references/go-code-review-comments.md) for interfaces.

## Preserve error contracts

- **GO-ERROR:** Return errors for expected failures. Check an error before using the
  corresponding value. Do not discard a read, parse, write, scanner, or close error when
  it can change the result. `panic`, `log.Fatal`, and `os.Exit` do not belong on ordinary
  failure paths in reusable functions. CLI exit policy belongs at the program boundary.
- **GO-CAUSE:** Add useful context, such as the path or operation. Wrap with `%w` when
  callers should be able to inspect the cause with `errors.Is` or `errors.As`; wrapping
  exposes that cause as part of the API. Use `%v` when deliberately hiding it. Avoid
  matching error-message strings when semantic inspection is available.
- Choose a sentinel for an identity callers need, a structured error for actionable
  fields, and a plain error otherwise. A global error framework or custom type for
  every failure is unnecessary. Return a literal `nil` error on success rather than a
  typed nil pointer converted to the `error` interface.
- Log or return an error at the layer responsible for it. Avoid repeated logging while
  propagating the same failure. Cleanup errors must not erase the original failure.
- Read [standard-library contracts](references/stdlib-contracts.md) for wrapping,
  partial I/O, and JSON; [Uber's guidance](references/uber-go-style.md) covers ownership
  of errors and cleanup.

## Own resources and goroutines

- **GO-CLEANUP:** After successful acquisition, arrange cleanup on every exit path.
  In a loop, scope resource use in a helper if deferring until the entire function exits
  would retain many files or response bodies. Do not close a resource owned by a caller
  unless that is the documented contract. Check flush/close errors when they affect
  successful output. Go's garbage collector is not resource cleanup.
- **GO-CONTEXT:** Pass request-scoped `context.Context` through blocking downstream work,
  usually as the first parameter. Do not replace a caller's context with `Background()`
  to avoid cancellation. Release derived contexts with their cancel function. A context
  parameter alone proves nothing: the work must actually observe it.
- **GO-LIFETIME:** Every goroutine needs an owner, a completion or cancellation path,
  and a defined way for its owner to observe completion. Consider sends and receives
  when the other side exits early. The sender that owns channel completion normally
  closes it; receivers should not close shared channels to stop producers.
- **GO-SYNC:** Protect shared mutable state. Do not copy a used mutex or wait group.
  Choose channels or locks based on ownership and the contract; neither is a universal
  replacement for the other. Keep locks out of callbacks and blocking I/O where possible.
  Do not claim race freedom from one passing test run.
- For HTTP, propagate context into requests, reuse an appropriately configured client,
  close response bodies, and interpret status codes according to the API. Successful
  transport does not mean the server returned a successful status. Check the
  [concurrency and HTTP reference](references/concurrency-http.md) when this applies.

## Respect data and compatibility

- **GO-DATA-CONTRACT:** A nil slice, an empty slice, an absent key, and a zero value can
  have different observable meanings, especially in JSON. Preserve those distinctions.
  Do not sort a collection or change a map representation unless order is part of the
  contract. When deterministic output is required, define and test its tie-breaking.
- Slices and maps share backing storage. Decide whether a boundary lends, transfers,
  or copies mutable data. Do not accidentally mutate a caller through a retained alias.
- **GO-NUMBERS:** Choose integer widths from the contract. Conversions can truncate;
  ordinary integer arithmetic does not report overflow. Check bounds where external
  input or business invariants require it. Do not use float64 for exact integer values
  merely because JSON decoding defaults to it.
- When parsing external data, distinguish syntax from validation. Preserve unknown-field,
  whitespace, trailing-data, and size-limit behavior unless a change is requested.

## Keep public contracts discoverable

- **GO-DOCS:** When writing or changing an exported API, document its contract next
  to the declaration. Explain non-obvious results, errors callers can inspect, side
  effects, ordering, and ownership where relevant. During a refactor that changes
  failure behavior, update the affected API comments as well as the change notes.
  Describe what callers can rely on, not internal algorithms. Avoid boilerplate that
  merely repeats a name. See [Go doc comments](references/doc-comments.md).

## Test the behavior, then verify

- **GO-TEST:** Test the public behavior and failure paths: malformed input, missing
  resources, partial I/O, errors callers inspect, ordering, and cancellation when relevant.
  Prefer local fixtures, `t.TempDir`, `httptest`, and explicit synchronization. Use
  timeouts as deadlock safeguards rather than sleeps as evidence that work completed.
- Use table-driven tests when they clarify related cases. Keep failure messages useful.
  Follow the project's test framework; do not replace it to impose a preference.
  See [testing guidance](references/go-testing.md).
- Format changed Go files with `gofmt`. Run the relevant tests and `go vet` first;
  broaden to `go test ./...` for the module where practical. Use `go test -race ./...`
  for relevant concurrent code on a supported toolchain. Run the project's configured
  analyzers, including Staticcheck if available. Do not rewrite unrelated files or
  install tools into the user's project without a task-related reason.
- Report which checks actually ran, their outcomes, and any unavailable checks. A
  compiler, a linter, and tests provide different evidence; none establishes every
  design rule. When tools are unavailable, inspect the code and say what remains unverified.

## Before returning a change

Confirm that the requested behavior works, failures retain useful information,
resources have owners, concurrency has a termination story, and affected public
contracts are documented alongside their declarations. API changes should be
explicit. Keep explanations proportional to the task. Skip extra architecture and
documentation that do not help the user understand or use the change.
