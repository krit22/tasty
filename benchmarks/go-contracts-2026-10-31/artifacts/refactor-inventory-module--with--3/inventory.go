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

// ErrInvalid indicates an empty name or a negative quantity, price, or restock amount.
var ErrInvalid = errors.New("invalid item")

// ErrNotFound indicates that a requested item does not exist.
var ErrNotFound = errors.New("item not found")

// ErrOverflow indicates that a quantity or total value exceeds int64.
var ErrOverflow = errors.New("inventory overflow")

// Item describes a named stock quantity and its unit price in cents.
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

// Add inserts or replaces an item, returning ErrInvalid for invalid fields.
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

// Get returns the item and whether it exists.
func (i *Inventory) Get(name string) (Item, bool) {
	v, ok := i.items[name]
	return v, ok
}

// Restock increases an item's quantity. A missing name takes precedence over an
// invalid amount. On ErrInvalid or ErrOverflow the item remains unchanged.
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

// TotalValue sums the requested items, counting repeated names each time.
// It returns zero and ErrNotFound or ErrOverflow if the calculation fails.
func (i *Inventory) TotalValue(names []string) (int64, error) {
	var total int64
	for _, name := range names {
		v, ok := i.items[name]
		if !ok {
			return 0, fmt.Errorf("item %q: %w", name, ErrNotFound)
		}
		if v.Quantity != 0 && v.PriceCents > math.MaxInt64/v.Quantity {
			return 0, fmt.Errorf("value of %q: %w", name, ErrOverflow)
		}
		value := v.Quantity * v.PriceCents
		if value > math.MaxInt64-total {
			return 0, fmt.Errorf("total including %q: %w", name, ErrOverflow)
		}
		total += value
	}
	return total, nil
}

// Load reads comma-separated name, quantity, and price rows, ignoring blank lines.
// Fields are trimmed and later duplicate names replace earlier rows. Errors wrap
// file or numeric causes; invalid rows wrap ErrInvalid. No partial inventory is returned.
func Load(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read inventory %q: %w", path, err)
	}
	i := New()
	for lineNo, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("line %d: %w", lineNo+1, ErrInvalid)
		}
		quantity, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d quantity: %w", lineNo+1, err)
		}
		price, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d price: %w", lineNo+1, err)
		}
		if err := i.Add(strings.TrimSpace(parts[0]), quantity, price); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo+1, err)
		}
	}
	return i, nil
}

// Report lists items in ascending name order, with unit prices when verbose is true.
func (i *Inventory) Report(verbose bool) string {
	names := make([]string, 0, len(i.items))
	for name := range i.items {
		names = append(names, name)
	}
	sort.Strings(names)
	var s strings.Builder
	for _, name := range names {
		v := i.items[name]
		fmt.Fprintf(&s, "%s: %d\n", name, v.Quantity)
		if verbose {
			fmt.Fprintf(&s, "  price=%d\n", v.PriceCents)
		}
	}
	return s.String()
}
