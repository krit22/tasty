# Standard-library contracts: errors, data, and I/O

Sources, reviewed 2026-10-31:

- [errors](https://pkg.go.dev/errors) and [error wrapping](https://go.dev/blog/go1.13-errors).
- [io](https://pkg.go.dev/io), including partial reads/writes and ownership conventions.
- [encoding/json](https://pkg.go.dev/encoding/json), including exported fields, nil slices,
  and number decoding.
- [Go specification: numeric conversions](https://go.dev/ref/spec#Conversions).

**GO-CAUSE:** `errors.Is` and `errors.As` inspect wrapped errors. Wrapping deliberately
exposes a cause to callers. Do not promise an error identity and then destroy it while
adding context. A typed nil pointer inside an error interface is not a nil interface.

```go
package example

import (
	"errors"
	"fmt"
)

func Example() {
	notFound := errors.New("not found")
	err := fmt.Errorf("load catalog: %w", notFound)
	fmt.Println(errors.Is(err, notFound))
	// Output: true
}
```

**GO-DATA-CONTRACT:** JSON distinguishes null from an empty array. Preserve a documented
wire shape and public data fields. A cast or type assertion is not input validation.

```go
package example

import (
	"encoding/json"
	"fmt"
)

func Example() {
	var absent []string
	empty := []string{}
	a, _ := json.Marshal(absent) // These concrete values cannot fail to marshal.
	b, _ := json.Marshal(empty)
	fmt.Println(string(a), string(b))
	// Output: null []
}
```

**GO-ERROR / GO-CLEANUP:** Process the bytes returned with an I/O error according to
the operation's contract. Check scanner errors after iteration. Writers may fail;
flushing or closing output may be the point where a write failure becomes visible.

**GO-NUMBERS:** Integer conversions do not validate range and arithmetic does not return
overflow errors. Validate the range required by the API. JSON decoded into `any` uses
float64 for numbers by default; preserve exact integers with an appropriate strategy.
