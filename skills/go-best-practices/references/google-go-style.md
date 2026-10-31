# Google Go style: choose clarity in context

Sources: [core guide](https://google.github.io/styleguide/go/guide),
[decisions](https://google.github.io/styleguide/go/decisions), and
[best practices](https://google.github.io/styleguide/go/best-practices).
Reviewed 2026-10-31. These documents distinguish their own normative guidance from
contextual suggestions; their organizational conventions are not Go language rules.

**GO-SIMPLE:** Prefer a direct design a reader can follow. An abstraction needs to
improve the current program, not merely make it shorter. Clarity takes precedence
over concision. Explain non-obvious reasons, not what an obvious line already says.

**GO-API:** Use package context to avoid repetition. Prefer coherent packages and APIs
over generic organizational containers. Follow local naming and test conventions
unless they cause a concrete problem.

**GO-OPTIONS:** Group arguments when the call becomes ambiguous. Keep a simple, clear
existing signature when it serves its callers. Broad style cleanup is not a feature
requirement.

**GO-TEST:** Test the behavior promised to users. Use fixtures and focused helpers to
make failures understandable. Changes should remain proportionate to the request.
