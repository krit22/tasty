# Inventory refactor

- Kept all exported names, signatures (including `Report(verbose bool)`), Item fields, and JSON tags.
- Simplified map lookup and used a string builder for reports. Reports now sort by item name in Go string order, preserving the existing line format and trailing newlines.
- Load returns wrapped file errors instead of panicking. Invalid records and negative values preserve `ErrInvalid`; malformed numbers also preserve the underlying parsing error. Numeric literals outside int64 return `ErrOverflow` and preserve `strconv.ErrRange`. Record errors include line numbers and numeric errors identify the field.
- Missing names in TotalValue return `ErrNotFound`. Restock checks quantity overflow before mutation; TotalValue checks both multiplication and addition overflow. Errors support `errors.Is`, including wrapped filesystem and numeric causes. TotalValue returns zero on failure, without a partial total.
- A zero-value Inventory now supports Add through lazy map initialization.

## Deliberately retained behavior

Add replaces an existing item, including its price and quantity. Zero quantities and prices are valid. Names are only required to be nonempty; Add does not trim them. Load trims fields, skips whitespace-only lines, requires exactly three comma-separated fields, and replaces duplicate names with the last record. It retains the simple unquoted record format rather than introducing CSV quoting rules. Failed loads return no inventory.

TotalValue counts repeated names repeatedly and returns zero for an empty list. Restock checks existence before validating a negative amount. Report's boolean still controls price lines. Inventory remains an in-memory type without concurrent access guarantees.

Regression tests cover arithmetic boundaries and unchanged state after overflow, missing names, invalid inputs, file and parsing error chains, loading semantics, deterministic reports in both modes, JSON shape, repeated-name totals, and zero-value initialization.
