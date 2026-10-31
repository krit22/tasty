# Inventory changes

- Preserved exported names, signatures (including `Report(verbose bool)`), Item fields, and JSON tags.
- Reports now sort by item name, retaining existing line formatting and verbose prices. A strings.Builder avoids repeated string copying.
- File reads and numeric parsing now return errors with path and line/field context where applicable. Wrapped causes remain inspectable with errors.Is/errors.As. Numeric syntax and range errors retain strconv causes.
- Missing names in TotalValue now return ErrNotFound. Restock checks quantity overflow; TotalValue checks both multiplication and addition overflow and returns ErrOverflow. Failed restocks leave the item unchanged; failed totals return zero. Existing sentinel identities remain intact.
- Simplified map lookup and made the zero-value Inventory usable for Add. Documented API behavior and added regression coverage.

## Deliberately retained behavior

Add replaces existing items rather than accumulating quantities. Empty names and negative quantities/prices are invalid; zero is allowed. Add does not require an item's total value to fit int64: TotalValue validates that calculation when requested. Restock checks for a missing name before validating its amount. TotalValue counts duplicate requested names, and an empty list totals zero.

Load still reads simple, unquoted comma-separated rows with exactly three fields, trims surrounding field whitespace, skips blank lines, accepts signed decimal integers as strconv.ParseInt does, and lets later duplicate names replace earlier rows. It returns no partial inventory on failure. Numeric parse errors expose strconv's syntax/range causes rather than being reclassified as ErrInvalid or arithmetic ErrOverflow. Empty inventories report an empty string. Inventory access still requires external synchronization when concurrent.
