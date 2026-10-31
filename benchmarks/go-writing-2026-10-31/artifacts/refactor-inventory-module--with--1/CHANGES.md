# Inventory refactor

- Kept all exported names, signatures (including `Report(verbose bool)`), Item fields, and JSON tags.
- Reports sort items by name and use a string builder while keeping the existing text format and trailing newlines.
- File failures now return wrapped errors with the path. Invalid numeric input returns errors with path, line, and field context and preserves the underlying strconv error. Syntax errors match ErrInvalid; out-of-range integers match ErrOverflow.
- Missing items in TotalValue return ErrNotFound instead of panicking. Restock checks quantity overflow before changing state; TotalValue checks both multiplication and addition and returns ErrOverflow. Failed totals return zero. Sentinel errors remain compatible with errors.Is.
- The Inventory zero value now supports Add. Get uses a single map lookup.
- Added regression tests for error identities, overflow boundaries, unchanged state after failed restocks, deterministic reports, JSON shape, and loading behavior.

Deliberately retained: Add replaces existing entries; zero quantities and prices are accepted; negative quantities and prices and empty names are invalid. Restock checks name existence before validating the amount. TotalValue counts repeated names repeatedly and returns zero for an empty selection. Add accepts names verbatim, while Load trims fields, skips whitespace-only lines, and requires exactly three comma-separated fields (without CSV quoting). Duplicate loaded names replace earlier entries. Numeric parsing remains base-10 signed int64. Empty files produce empty inventories. Oversized item values are accepted by Add and Load; value overflow is checked when TotalValue is requested. No concurrent-use guarantee is introduced.
