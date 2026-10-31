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

type Item struct {
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

type Inventory struct{ items map[string]Item }

func New() *Inventory { return &Inventory{items: map[string]Item{}} }

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
	v, ok := i.items[name]
	return v, ok
}

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

func (i *Inventory) TotalValue(names []string) (int64, error) {
	var total int64
	for _, name := range names {
		v, ok := i.items[name]
		if !ok {
			return 0, fmt.Errorf("total value of %q: %w", name, ErrNotFound)
		}
		if v.PriceCents != 0 && v.Quantity > math.MaxInt64/v.PriceCents {
			return 0, fmt.Errorf("value of %q: %w", name, ErrOverflow)
		}
		value := v.Quantity * v.PriceCents
		if value > math.MaxInt64-total {
			return 0, fmt.Errorf("total value: %w", ErrOverflow)
		}
		total += value
	}
	return total, nil
}

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
			return nil, numericError(path, lineNo+1, "quantity", err)
		}
		price, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			return nil, numericError(path, lineNo+1, "price", err)
		}
		if err := i.Add(strings.TrimSpace(parts[0]), quantity, price); err != nil {
			return nil, fmt.Errorf("load %q line %d: %w", path, lineNo+1, err)
		}
	}
	return i, nil
}

// numericError preserves the parse cause as well as the inventory error category.
func numericError(path string, line int, field string, err error) error {
	kind := ErrInvalid
	if errors.Is(err, strconv.ErrRange) {
		kind = ErrOverflow
	}
	return fmt.Errorf("load %q line %d %s: %w: %w", path, line, field, kind, err)
}

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
