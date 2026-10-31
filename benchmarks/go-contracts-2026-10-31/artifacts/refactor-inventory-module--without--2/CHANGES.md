# Changes

- Kept all exported names, signatures, Item fields, and JSON tags, including Report(verbose bool).
- Reports sort items by name and use a strings.Builder while preserving the existing text format and trailing newlines.
- Load returns file and numeric errors instead of panicking, with file or line/field context. Wrapped file and parse errors remain available through errors.Is/errors.As. Invalid numeric syntax wraps ErrInvalid; numeric range errors wrap ErrOverflow.
- TotalValue returns ErrNotFound for missing names and ErrOverflow for multiplication or total overflow. Restock checks quantity overflow before changing the item. Sentinel errors remain compatible with errors.Is.
- Simplified Get and made the zero-value Inventory usable through Add.
- Added regression coverage for error propagation, overflow boundaries, unchanged state after failed restocking, deterministic reports, loading, and the JSON shape.

## Deliberately retained behavior

Add replaces an existing item rather than accumulating quantities. Empty names and negative quantities/prices are invalid; zero quantities/prices are valid. Add preserves names literally, while Load trims surrounding whitespace from fields and ignores blank lines. Load uses a simple three-field comma-separated format without CSV quoting and lets later duplicate names replace earlier entries. Failed loads return no partial inventory.

TotalValue counts repeated names repeatedly and returns zero for an empty list. Values are checked when calculating totals, so an item whose individual value exceeds int64 can still be stored. Restock checks for a missing name before validating the amount. Get retains its (Item, bool) missing-item convention. Reports retain the optional indented price lines. Inventory remains unsynchronized and requires external coordination for concurrent mutation.
