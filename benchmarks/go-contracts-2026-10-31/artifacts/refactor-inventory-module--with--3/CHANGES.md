# Inventory refactor

- Preserved exported names, signatures (including `Report(verbose bool)`), Item fields, and JSON tags.
- Reports now sort by item name and use a string builder, retaining both output formats and trailing newlines.
- Load returns contextual file and numeric parsing errors instead of panicking. Wrapped errors support `errors.Is` and `errors.As`; invalid rows continue to match ErrInvalid. Failed loads return no partial inventory.
- TotalValue returns ErrNotFound for missing names. Restock checks quantity overflow; TotalValue checks multiplication and accumulation overflow. These failures match ErrOverflow, return no partial total, and do not mutate items.
- The Inventory zero value now supports Add. Get uses one map lookup. API comments describe error and ordering contracts.

Deliberately retained: Add replaces an existing item, zero quantities/prices are valid, and only empty names and negative numbers are rejected. Restock checks for a missing item before validating the amount. TotalValue counts duplicate names repeatedly and returns zero for an empty list. Load ignores blank lines, trims fields, requires exactly three comma-separated fields (no CSV quoting), parses signed decimal int64 values, and lets later duplicate rows replace earlier ones. Numeric syntax/range errors retain strconv causes; they are distinct from arithmetic ErrOverflow. Inventory still requires external synchronization for concurrent use.

Regression tests cover report ordering and formatting, replacement, JSON shape, input validation, file and parsing errors, sentinel matching, arithmetic boundaries, and unchanged state after failed restocks.
