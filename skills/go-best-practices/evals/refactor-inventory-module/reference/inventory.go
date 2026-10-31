// Package inventory maintains quantities and integer-cent prices.
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

var ErrInvalid = errors.New("invalid item")
var ErrNotFound = errors.New("item not found")
var ErrOverflow = errors.New("inventory overflow")

// Item is plain inventory data; its JSON representation is part of the API.
type Item struct {
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

// Inventory stores items by name. It is not safe for concurrent mutation.
type Inventory struct{ items map[string]Item }

// New returns an empty inventory.
func New() *Inventory { return &Inventory{items: map[string]Item{}} }

// Add inserts or replaces an item, rejecting empty names and negative values.
func (i *Inventory) Add(name string, quantity, price int64) error {
	if name == "" || quantity < 0 || price < 0 {
		return ErrInvalid
	}
	i.items[name] = Item{Name: name, Quantity: quantity, PriceCents: price}
	return nil
}

// Get returns a copy of the named item, if present.
func (i *Inventory) Get(name string) (Item, bool) { v, ok := i.items[name]; return v, ok }

// Restock adds a nonnegative amount without changing state on overflow.
func (i *Inventory) Restock(name string, amount int64) error {
	v, ok := i.items[name]
	if !ok {
		return fmt.Errorf("restock %q: %w", name, ErrNotFound)
	}
	if amount < 0 {
		return ErrInvalid
	}
	if amount > math.MaxInt64-v.Quantity {
		return ErrOverflow
	}
	v.Quantity += amount
	i.items[name] = v
	return nil
}

// TotalValue totals the named entries, counting repeated names repeatedly.
func (i *Inventory) TotalValue(names []string) (int64, error) {
	var total int64
	for _, name := range names {
		v, ok := i.items[name]
		if !ok {
			return 0, fmt.Errorf("total %q: %w", name, ErrNotFound)
		}
		if v.PriceCents != 0 && v.Quantity > math.MaxInt64/v.PriceCents {
			return 0, ErrOverflow
		}
		value := v.Quantity * v.PriceCents
		if value > math.MaxInt64-total {
			return 0, ErrOverflow
		}
		total += value
	}
	return total, nil
}

// Load reads name,quantity,price lines, preserving replacement and whitespace policy.
func Load(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	i := New()
	for lineNo, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("%s line %d: %w", path, lineNo+1, ErrInvalid)
		}
		quantity, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s line %d quantity: %w", path, lineNo+1, err)
		}
		price, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s line %d price: %w", path, lineNo+1, err)
		}
		if err := i.Add(strings.TrimSpace(parts[0]), quantity, price); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, lineNo+1, err)
		}
	}
	return i, nil
}

// Report returns name-sorted quantities, adding price lines when verbose.
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
