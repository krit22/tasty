# Inventory refactor

- Kept all exported names, method and function signatures (including `Report(verbose bool)`), Item fields, and JSON keys.
- Reports now sort by item name and use a string builder. Quantity lines, optional price lines, and trailing newlines retain their existing format.
- File read failures and numeric parse failures now return contextual errors instead of panicking. Load errors identify the line and numeric field where applicable. Wrapped errors retain `errors.Is` compatibility with filesystem/strconv errors and existing sentinels; malformed numeric fields also match `ErrInvalid`.
- Missing names in TotalValue now return `ErrNotFound`. Restock checks quantity overflow before mutation; TotalValue checks both multiplication and accumulation and returns `ErrOverflow`. TotalValue returns zero on error, with no partial total.
- Simplified map lookup and made the zero-value Inventory usable for Add.
- Added regression coverage for errors, arithmetic boundaries, unchanged state after failed restocking, report ordering/format, loading, validation, and JSON shape.

## Deliberately retained behavior

Add replaces an existing item rather than accumulating its quantity. Names passed directly to Add are not trimmed; only the empty string is invalid. Zero quantity and price are valid; negative numbers are invalid. Restock checks for a missing item before validating the amount. Get returns a zero Item and false for missing names. TotalValue counts repeated names repeatedly and returns zero for an empty selection.

Load retains the simple three-comma-separated-field format (no CSV quoting), trims surrounding field whitespace, skips blank lines, and lets later duplicate names replace earlier ones. It returns no partial inventory on failure. Numeric text outside int64 range returns the strconv range error (also matching ErrInvalid); ErrOverflow describes inventory arithmetic overflow. Add and Load permit individually valid quantity/price pairs whose product exceeds int64; TotalValue detects that overflow when requested.
