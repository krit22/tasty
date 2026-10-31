# Changes

- Preserved exported names, signatures (including `Report(verbose bool)`), Item fields, and JSON tags.
- Reports sort by item name using Go string ordering, retain their original text format, and use a string builder.
- Simplified map lookup and made the zero-value Inventory usable with Add.
- Load returns wrapped file errors instead of panicking. Invalid rows include line numbers; numeric errors also identify the field and preserve strconv causes for errors.Is/errors.As. Numeric syntax failures match ErrInvalid; out-of-range numbers match both ErrInvalid and ErrOverflow.
- TotalValue returns ErrNotFound for missing names. Restock detects quantity overflow; TotalValue checks multiplication and accumulation overflow. Errors remain compatible with errors.Is and the existing sentinels. Failed restocks leave items unchanged; failed totals return zero and an error. Failed loads return nil and an error.

## Deliberately retained behavior

Add replaces existing items, accepts zero quantities/prices, rejects empty names and negative numbers, and does not trim names. Load trims fields, skips blank lines, uses a simple three-field comma-separated format without CSV quoting, and lets later duplicate names replace earlier entries. TotalValue counts repeated names repeatedly and returns zero for an empty list. Restock checks name existence before rejecting a negative amount. Add permits items whose eventual total value may overflow; TotalValue reports that overflow when evaluated. Inventory remains unsynchronized.

Regression tests cover ordering and formatting, JSON shape, loading and error causes, validation, replacement, repeated names, missing items, arithmetic boundaries, and failed-operation state preservation.
