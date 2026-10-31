# Inventory refactor

- File reads and numeric parsing now return contextual errors instead of panicking. Load errors identify the path, line, and numeric field where applicable. Numeric syntax errors match ErrInvalid; out-of-range integers match ErrOverflow. Wrapped file and strconv causes remain available through errors.Is/errors.As.
- TotalValue returns ErrNotFound for missing names. Restock checks quantity overflow; TotalValue checks both multiplication and accumulated value overflow. Failed restocks leave the item unchanged; failed totals return zero. Existing sentinels remain compatible with errors.Is.
- Report(verbose bool) sorts names lexicographically and uses a strings.Builder. Its signature, line formatting, verbose price lines, and trailing newline are retained.
- Get uses one map lookup. Add initializes a zero-value Inventory on first use.

Deliberately retained: Add replaces an existing item rather than accumulating stock; zero quantities and prices are valid; Add keeps names exactly as supplied. Load trims fields, skips whitespace-only lines, accepts CRLF, replaces duplicate names, and uses the existing simple three-field comma format (no CSV quoting). Load returns no inventory on failure. TotalValue counts repeated requested names repeatedly and returns zero for an empty selection. Restock checks for a missing name before rejecting a negative amount. Item fields, integer types, and JSON tags are unchanged. Inventory still requires callers to synchronize concurrent mutations.

Regression tests cover these behaviors, deterministic reports, JSON shape, file and parsing failures, sentinel/cause inspection, arithmetic boundaries, and unchanged state after failed restocking.
