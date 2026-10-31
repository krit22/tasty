package inventory

import (
	"errors"
	"fmt"
	"os"
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
	i.items[name] = Item{Name: name, Quantity: quantity, PriceCents: price}
	return nil
}

func (i *Inventory) Get(name string) (Item, bool) {
	if _, ok := i.items[name]; ok {
		v := i.items[name]
		return v, true
	}
	return Item{}, false
}

func (i *Inventory) Restock(name string, amount int64) error {
	v, ok := i.items[name]
	if !ok {
		return ErrNotFound
	}
	if amount < 0 {
		return ErrInvalid
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
			panic("unknown item: " + name)
		}
		total += v.Quantity * v.PriceCents
	}
	return total, nil
}

func Load(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
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
			panic(err)
		}
		price, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			panic(err)
		}
		if err := i.Add(strings.TrimSpace(parts[0]), quantity, price); err != nil {
			return nil, err
		}
	}
	return i, nil
}

func (i *Inventory) Report(verbose bool) string {
	s := ""
	for name, v := range i.items {
		s = s + name + ": " + strconv.FormatInt(v.Quantity, 10) + "\n"
		if verbose {
			s = s + "  price=" + strconv.FormatInt(v.PriceCents, 10) + "\n"
		}
	}
	return s
}
