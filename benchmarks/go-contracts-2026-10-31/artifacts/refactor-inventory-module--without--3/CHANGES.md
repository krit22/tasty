# Changes

- Preserved all exported names, signatures (including `Report(verbose bool)`),
  Item fields, and JSON keys.
- Reports now sort items by name and use a string builder. Both verbosity modes
  retain their original text format, indentation, and trailing newlines.
- File errors from Load now return with path context and preserve the underlying
  error for errors.Is/errors.As. Numeric parsing errors return with line and field
  context, preserving both ErrInvalid and the underlying strconv error. Invalid
  rows and invalid item values include line numbers. Failed loads return nil
  inventory rather than a partially loaded inventory.
- TotalValue returns ErrNotFound for missing names. Restock detects quantity
  overflow before mutation; TotalValue checks multiplication and accumulation
  overflow. Both return errors wrapping ErrOverflow. Failed totals return zero.
- Simplified map lookup and made the zero-value Inventory usable with Add.
- Added regression coverage for errors, overflow boundaries, unchanged state after
  failed restocks, loading, report ordering and formatting, and Item JSON.

## Deliberately retained behavior

Add replaces an existing item, and accepts zero quantities and prices. Names are
case-sensitive; Add does not trim names. Load trims fields, skips blank lines,
requires exactly three comma-separated fields (without CSV quoting), parses
signed base-10 int64 numbers, and lets the last duplicate name win. Numeric values
outside int64 return parsing errors; ErrOverflow identifies inventory arithmetic
that exceeds int64. Add may store individually valid quantities and prices whose
product will overflow; TotalValue detects that when requested.

TotalValue counts repeated names repeatedly and returns zero for an empty list.
Restock checks name existence before rejecting a negative amount. Get retains its
(Item, bool) missing-item convention. Sentinel errors remain compatible with
errors.Is. The package does not add concurrency guarantees or nil-receiver support.
