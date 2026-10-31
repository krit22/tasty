# Effective Go: composition and usable APIs

Source: [Effective Go](https://go.dev/doc/effective_go), especially Names, Interfaces,
Embedding, and Initialization. Reviewed 2026-10-31. This older guide does not cover
modules or generics; check current language and package contracts for those.

**GO-API:** Read a name at its call site, including its package. Prefer a focused
package over a collection of unrelated helpers. Methods can belong on a concrete type
without a matching interface. Embedding promotes methods; it is not class inheritance.

**GO-DATA:** Useful zero values reduce initialization obligations. Add a constructor
when an object needs dependencies or must establish an invariant. Plain records can
remain plain records; exported fields are useful when callers should construct data.

**GO-SIMPLE:** Start with the data and operations the problem needs. Add indirection
when a consumer needs it. Preserve an existing initialization contract during a refactor.

These are API design considerations, not reasons to redesign working neighboring code.

```go
package example

import "fmt"

type Counter struct{ n int }

func (c *Counter) Add(n int) { c.n += n }
func (c *Counter) Value() int { return c.n }

func Example() {
	c := new(Counter) // The zero value already satisfies the contract.
	c.Add(2)
	fmt.Println(c.Value())
	// Output: 2
}
```
