# Inventory refactor

- Kept every exported name and function signature, including `Report(verbose bool)`, and preserved `Item` fields and JSON keys.
- Reports now sort by item name (Go string order) and use a string builder. Their line format, indentation, and trailing newlines are unchanged.
- File read errors now return with their underlying error available through `errors.Is` and `errors.As`.
- Invalid numeric fields now return errors with line and field context instead of panicking. They wrap both `ErrInvalid` and the original `strconv.NumError`, including syntax and out-of-range errors.
- Missing names in `TotalValue` return `ErrNotFound`. Multiplication, total addition, and restock quantity overflow return `ErrOverflow` before arithmetic can wrap. Failed totals return zero; failed restocks leave the item unchanged. Existing sentinels remain compatible with `errors.Is`.
- Simplified map lookup and made the zero-value `Inventory` usable for insertion.
- Added regression coverage for errors, arithmetic boundaries, unchanged state after failures, loading, replacement, duplicate totals, report ordering, and JSON shape.

## Deliberately retained behavior

`Add` replaces an existing item, permits zero quantity and price, and rejects empty names or negative numbers. Direct `Add` calls do not trim names. `Restock` checks existence before validating the amount. `TotalValue` counts repeated names repeatedly and returns zero for an empty selection. An item with an individually overflowing value can still be stored; overflow is checked when calculating value.

Loading still uses simple comma splitting, not quoted CSV: each nonblank line must have exactly three fields. Blank lines are skipped, fields are trimmed, signed base-10 integers are accepted, and the last occurrence of a duplicate name wins. Any load failure returns a nil inventory. No concurrency guarantees or nil-receiver support were added.
