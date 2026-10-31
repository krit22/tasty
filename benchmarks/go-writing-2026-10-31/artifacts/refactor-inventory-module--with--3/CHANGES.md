# Inventory refactor

- `Load` returns file and numeric parsing errors instead of panicking. Wrapped
  errors retain their causes for `errors.Is` and `errors.As`; numeric and item
  validation errors include the line number, and numeric errors identify the field.
- `TotalValue` returns an error wrapping `ErrNotFound` for a missing name.
- `Restock` checks quantity addition, and `TotalValue` checks both multiplication
  and accumulation before arithmetic. Overflow returns an error wrapping
  `ErrOverflow`. A failed restock leaves the item unchanged; failed totals return
  zero and an error.
- `Report(verbose bool)` sorts by item name in ascending Go string order and uses
  a string builder. Its signature, line format, verbose price lines, and final
  newline remain unchanged. Empty inventories still produce an empty string.
- `Get` uses one map lookup. The zero value of `Inventory` now supports `Add` by
  initializing its map when needed.

## Deliberately retained behavior

All exported names, function signatures, `Item` fields, and JSON tags are unchanged.
`Add` replaces an existing item, accepts zero quantities and prices, and rejects
empty names and negative numbers with `ErrInvalid`. It does not trim names.
`Restock` checks for a missing item before validating its amount and accepts zero.
`Get` returns an item copy and a presence flag.

`TotalValue` counts repeated names repeatedly and returns zero for an empty list.
`Add` may store an item whose value exceeds int64; the overflow is reported when
that item's value is requested. Existing sentinel errors remain discoverable with
`errors.Is`.

The file format remains three comma-separated fields per nonblank line, without
CSV quoting. Blank lines are skipped, surrounding field whitespace is trimmed,
and later duplicate names replace earlier ones. Decimal int64 parsing remains
unchanged, including acceptance of a leading plus sign. Parse failures retain
`strconv` errors (syntax or range), rather than being relabeled as arithmetic
`ErrOverflow`. Invalid field counts and invalid item data retain `ErrInvalid`.
Load failures return no partial inventory.

The package remains unsynchronized; callers must coordinate concurrent mutation.

## Validation

Regression tests cover replacement, JSON shape, missing names, invalid input,
file errors, numeric syntax/range errors, overflow boundaries, unchanged state on
failed restocking, repeated names, zero values, and both sorted report modes.
`go test ./...` and `go vet ./...` pass using the supplied Go toolchain and caches.
