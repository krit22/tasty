# Go doc comments

Source: [Go Doc Comments](https://go.dev/doc/comment), especially Funcs and Types.
Reviewed 2026-10-31.

**GO-DOCS:** Put API documentation beside the exported declaration. Explain the
operation's observable contract: returns, relevant errors, special cases, and side
effects. Type comments can explain zero-value meaning and concurrency guarantees.
Leave implementation details inside the function when necessary.

Tasty applies this during implementation and refactoring of affected APIs. Change
notes help reviewers understand a transition; declaration comments help callers use
the resulting API. Keep both accurate without expanding into unrelated documentation
cleanup or adding comments that only repeat identifiers.
