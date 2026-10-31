package inventory

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ErrInvalid indicates an empty name or a negative quantity or price.
var ErrInvalid = errors.New("invalid item")

// ErrNotFound indicates a name absent from the inventory.
var ErrNotFound = errors.New("item not found")

// ErrOverflow indicates a quantity or total value that exceeds int64.
var ErrOverflow = errors.New("inventory overflow")

// Item holds a named quantity and its unit price in cents.
type Item struct {
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

// Inventory stores items by name. Its zero value is ready to use.
// Concurrent access requires external synchronization.
type Inventory struct{ items map[string]Item }

// New returns an empty inventory.
func New() *Inventory { return &Inventory{items: map[string]Item{}} }

// Add replaces the named item. Empty names and negative numbers return ErrInvalid.
func (i *Inventory) Add(name string, quantity, price int64) error {
	if name == "" || quantity < 0 || price < 0 {
		return ErrInvalid
	}
	if i.items == nil {
		i.items = make(map[string]Item)
	}
	i.items[name] = Item{Name: name, Quantity: quantity, PriceCents: price}
	return nil
}

// Get returns the named item and whether it exists.
func (i *Inventory) Get(name string) (Item, bool) {
	v, ok := i.items[name]
	return v, ok
}

// Restock adds a nonnegative amount. It returns ErrNotFound for missing names,
// ErrInvalid for negative amounts, and ErrOverflow without mutation on overflow.
func (i *Inventory) Restock(name string, amount int64) error {
	v, ok := i.items[name]
	if !ok {
		return ErrNotFound
	}
	if amount < 0 {
		return ErrInvalid
	}
	if amount > math.MaxInt64-v.Quantity {
		return fmt.Errorf("restock %q: %w", name, ErrOverflow)
	}
	v.Quantity += amount
	i.items[name] = v
	return nil
}

// TotalValue sums the requested items, counting repeated names repeatedly.
// Missing names return ErrNotFound; product or sum overflow returns ErrOverflow.
// On error, the returned total is zero.
func (i *Inventory) TotalValue(names []string) (int64, error) {
	var total int64
	for _, name := range names {
		v, ok := i.items[name]
		if !ok {
			return 0, fmt.Errorf("total value %q: %w", name, ErrNotFound)
		}
		if v.PriceCents != 0 && v.Quantity > math.MaxInt64/v.PriceCents {
			return 0, fmt.Errorf("value of %q: %w", name, ErrOverflow)
		}
		value := v.Quantity * v.PriceCents
		if value > math.MaxInt64-total {
			return 0, fmt.Errorf("total value at %q: %w", name, ErrOverflow)
		}
		total += value
	}
	return total, nil
}

// Load reads unquoted name,quantity,price rows, trimming each field and ignoring
// blank lines. Later rows replace earlier names. Errors include path and line
// context and wrap filesystem, numeric parse, or ErrInvalid causes.
func Load(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load %q: %w", path, err)
	}
	i := New()
	for lineNo, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("load %q line %d: %w", path, lineNo+1, ErrInvalid)
		}
		quantity, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("load %q line %d quantity: %w", path, lineNo+1, err)
		}
		price, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("load %q line %d price: %w", path, lineNo+1, err)
		}
		if err := i.Add(strings.TrimSpace(parts[0]), quantity, price); err != nil {
			return nil, fmt.Errorf("load %q line %d: %w", path, lineNo+1, err)
		}
	}
	return i, nil
}

// Report lists items in ascending name order, with a trailing newline per row.
// If verbose is true, each item also includes its unit price on a second line.
func (i *Inventory) Report(verbose bool) string {
	names := make([]string, 0, len(i.items))
	for name := range i.items {
		names = append(names, name)
	}
	sort.Strings(names)
	var report strings.Builder
	for _, name := range names {
		v := i.items[name]
		fmt.Fprintf(&report, "%s: %d\n", name, v.Quantity)
		if verbose {
			fmt.Fprintf(&report, "  price=%d\n", v.PriceCents)
		}
	}
	return report.String()
}
