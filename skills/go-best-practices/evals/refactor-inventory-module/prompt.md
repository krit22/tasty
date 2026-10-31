---
tags: [go-best-practices, refactor]
max_turns: 30
timeout_seconds: 900
---

Refactor this inventory package into practical, idiomatic Go. Keep its exported names,
function signatures, Item fields and JSON shape, and behavior on valid inputs. Keep
Report(verbose bool); existing callers use it. Make report ordering deterministic by
item name. Fix expected-input failures so file and numeric errors return to callers,
missing names return ErrNotFound, and quantity/value overflow returns ErrOverflow.
Preserve errors.Is compatibility with the existing sentinel errors. Add regression
tests and write CHANGES.md explaining the changes and any behavior deliberately retained.

Input: inventory.go, inventory_test.go, go.mod.
