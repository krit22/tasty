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

var (
	ErrInvalid  = errors.New("invalid item")
	ErrNotFound = errors.New("item not found")
	ErrOverflow = errors.New("inventory overflow")
)

type Item struct {
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

// Inventory stores items by name. Its zero value is ready to use.
type Inventory struct{ items map[string]Item }

func New() *Inventory { return &Inventory{items: make(map[string]Item)} }

// Add inserts an item, replacing any existing item with the same name.
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

func (i *Inventory) Get(name string) (Item, bool) {
	item, ok := i.items[name]
	return item, ok
}

func (i *Inventory) Restock(name string, amount int64) error {
	item, ok := i.items[name]
	if !ok {
		return ErrNotFound
	}
	if amount < 0 {
		return ErrInvalid
	}
	if amount > math.MaxInt64-item.Quantity {
		return ErrOverflow
	}
	item.Quantity += amount
	i.items[name] = item
	return nil
}

// TotalValue sums the requested items; repeated names are counted repeatedly.
func (i *Inventory) TotalValue(names []string) (int64, error) {
	var total int64
	for _, name := range names {
		item, ok := i.items[name]
		if !ok {
			return 0, fmt.Errorf("item %q: %w", name, ErrNotFound)
		}
		if item.Quantity != 0 && item.PriceCents > math.MaxInt64/item.Quantity {
			return 0, fmt.Errorf("item %q: %w", name, ErrOverflow)
		}
		value := item.Quantity * item.PriceCents
		if value > math.MaxInt64-total {
			return 0, fmt.Errorf("item %q: %w", name, ErrOverflow)
		}
		total += value
	}
	return total, nil
}

// Load reads comma-separated name, quantity, and price fields. Blank lines are ignored.
func Load(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load inventory: %w", err)
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
		quantity, err := parseNumber(parts[1])
		if err != nil {
			return nil, fmt.Errorf("line %d quantity: %w", lineNo+1, err)
		}
		price, err := parseNumber(parts[2])
		if err != nil {
			return nil, fmt.Errorf("line %d price: %w", lineNo+1, err)
		}
		if err := i.Add(strings.TrimSpace(parts[0]), quantity, price); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo+1, err)
		}
	}
	return i, nil
}

func parseNumber(s string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	return value, nil
}

// Report lists items in name order, optionally including their prices.
func (i *Inventory) Report(verbose bool) string {
	names := make([]string, 0, len(i.items))
	for name := range i.items {
		names = append(names, name)
	}
	sort.Strings(names)

	var report strings.Builder
	for _, name := range names {
		item := i.items[name]
		fmt.Fprintf(&report, "%s: %d\n", name, item.Quantity)
		if verbose {
			fmt.Fprintf(&report, "  price=%d\n", item.PriceCents)
		}
	}
	return report.String()
}
