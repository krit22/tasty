# Go review conventions: interfaces and request scope

Source: [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), especially
Contexts, Interfaces, Pass Values, and Synchronous Functions. Reviewed 2026-10-31.

**GO-SIMPLE:** Let the consumer define the interface it needs. Returning a concrete type
often leaves callers more flexibility. Do not create a producer-side interface only
to enable hypothetical mocking. Existing abstractions may be justified by real users.

**GO-CONTEXT:** Pass request context through the call chain, normally as the first
parameter. Avoid storing it in long-lived service objects. A background task can have
a separate lifetime, but its owner must explicitly define that lifetime.

**GO-LIFETIME:** Synchronous functions make completion and error ownership easier to
understand. Callers can arrange concurrency when needed. When an API starts background
work itself, document and implement shutdown and completion.

Read the actual package documentation before treating a convention as a hard API rule.
