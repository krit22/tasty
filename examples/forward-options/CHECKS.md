This forward test used the alpha-1 skill before the documentation revision.

Independent forward test of `skills/go-best-practices/SKILL.md`.

Artifact directory: `examples/forward-options`.

I created a fresh standard-library-only package exposing the requested `Options`
fields and `Render([]string, bool) string`, added characterization tests, and then
followed the skill to add `ParseOptions([]byte) (Options, error)` in `options.go`.
The original implementation is saved as `render.go.before`; `render.go` remained
byte-for-byte identical. No project files were edited, and no project evals or
reference implementations were read.

The parser uses `json.Unmarshal`, rejects a negative limit, returns zero Options
on failure, and wraps JSON errors with `%w`. Tests cover successful values, zero
and missing fields, null, unknown fields, whitespace, negative limits, empty and
malformed input, wrong JSON types, fractional and overflowing integers, trailing
data, and `errors.As` access to syntax/type errors. Existing rendering behavior,
public field construction, and JSON tags are also exercised.

Every Go command below used Go 1.27.3 from
`the Go 1.27.3 executable`,
`GOCACHE=a temporary cache`, and `GOFLAGS=-buildvcs=false`.
The fixture's `go 1.22.0` directive was unchanged.

Actual checks, in order:

- Baseline `go test ./...`: passed before adding ParseOptions.
- New tests before implementation: `go test -run '^TestParseOptions' ./...`
  failed as expected because `render.ParseOptions` did not yet exist.
- `gofmt -w options.go options_test.go render.go render_test.go`: completed.
- `go test -count=1 -run '^TestParseOptions' ./...`: passed.
- `go vet ./...`: passed with no diagnostics.
- `cmp render.go render.go.before`: passed, confirming the existing API and
  implementation were untouched.
- `gofmt -l options.go options_test.go render.go render_test.go`: no output.
- `go list -m all`: printed only `example.com/render`, confirming no dependency
  modules were introduced.
- `go test -count=1 -v ./...`: all six top-level tests and all subtests passed.
- `command -v staticcheck`: no executable found. No analyzer configuration
  existed in the fresh fixture. Staticcheck was not installed or run.

The race detector was not run because this request introduces no concurrency.
An older Go runtime was not tested; compilation and execution used the supplied
1.27.3 toolchain only.

I encountered no concrete behavioral defect or instruction that overreached the
requested task. The instructions explicitly allowed plain exported configuration
fields and the existing boolean parameter, which helped avoid unnecessary API
changes. The reference reading was more work than this small implementation, but
the relevant digests were short and did not require fetching the repository,
installing tools, adding dependencies, or applying concurrency guidance.

This is a successful single-scenario behavioral check, not evidence that every
rule in the skill works across other Go tasks.

The working-file backup referenced above was used for the comparison and is not part of this published example.
