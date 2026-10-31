# Inventory changes

- File reads and numeric parsing in Load now return contextual errors instead of panicking. Wrapped file and strconv errors remain inspectable with errors.Is/errors.As. Invalid records retain ErrInvalid compatibility and include the path and line number.
- TotalValue returns zero with ErrNotFound for missing names. Restock checks quantity overflow; TotalValue checks both multiplication and accumulated value overflow. These failures wrap ErrOverflow. Failed restocks leave the item unchanged.
- Report(verbose bool) keeps its signature and text format and now sorts by item name in ascending Go string order. A strings.Builder avoids repeated copying.
- Get uses one map lookup, and a zero-value Inventory can now accept Add calls. API comments describe error and ordering contracts.
- Regression tests cover errors, arithmetic boundaries, reports, replacement, parsing conventions, and Item JSON.

## Deliberately retained behavior

Exported names, signatures, Item fields and JSON keys are unchanged. Add replaces an existing item rather than incrementing it, accepts zero quantities/prices, and does not trim names. It permits items whose individual value would overflow; TotalValue detects that overflow when requested. TotalValue counts repeated names repeatedly and returns zero for an empty selection. Restock checks for a missing name before validating the amount. Invalid Add/Restock inputs retain ErrInvalid.

Load skips whitespace-only lines, trims fields, requires exactly three comma-separated fields without CSV quoting, accepts the existing base-10 signed integer syntax, and replaces duplicate names with the last record. Numeric syntax/range errors retain strconv causes. Report retains trailing newlines, optional indented prices, and an empty string for an empty inventory. Inventory still requires external synchronization for concurrent reads and writes. No dependencies or module version changes were introduced.
