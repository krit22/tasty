---
tags: [go-best-practices, library]
max_turns: 30
timeout_seconds: 900
---

Build a dependency-free Go package named semver for parsing and comparing semantic
version strings using SemVer 2.0.0 precedence. Use these entrypoints:

- Parse(text string) (Version, error)
- (Version).String() string
- Compare(a, b Version) int, returning a negative value, zero, or a positive value

Core major/minor/patch components must fit uint64; reject overflow. Prerelease numeric
identifiers can be arbitrarily long. Build metadata does not affect precedence.
Invalid input returns an error. Include usable API documentation and tests. Keep the
package practical and dependency-free. A go.mod is already present.
