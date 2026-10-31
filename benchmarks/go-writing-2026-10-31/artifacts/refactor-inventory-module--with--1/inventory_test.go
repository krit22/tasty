package inventory

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestHappyPath(t *testing.T) {
	i := New()
	if err := i.Add("apple", 2, 50); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("apple", 1); err != nil {
		t.Fatal(err)
	}
	total, err := i.TotalValue([]string{"apple"})
	if err != nil || total != 150 {
		t.Fatalf("total=%d err=%v", total, err)
	}
}

func TestInventoryContracts(t *testing.T) {
	var i Inventory
	if err := i.Add("pear", 1, 3); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("apple", 2, 5); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("pear", 4, 7); err != nil {
		t.Fatal(err)
	}
	if got := i.Report(false); got != "apple: 2\npear: 4\n" {
		t.Fatalf("report = %q", got)
	}
	if got := i.Report(true); got != "apple: 2\n  price=5\npear: 4\n  price=7\n" {
		t.Fatalf("verbose report = %q", got)
	}
	if total, err := i.TotalValue([]string{"pear", "pear"}); total != 56 || err != nil {
		t.Fatalf("total=%d err=%v", total, err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("empty total=%d err=%v", total, err)
	}
	if got, ok := i.Get("missing"); ok || got != (Item{}) {
		t.Fatalf("missing=%+v, %v", got, ok)
	}
	if _, err := i.TotalValue([]string{"pear", "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing total: %v", err)
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing restock: %v", err)
	}
	for _, item := range []Item{{Name: ""}, {Name: "bad", Quantity: -1}, {Name: "bad", PriceCents: -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v): %v", item, err)
		}
	}
	if err := i.Restock("pear", -1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("negative restock: %v", err)
	}
	item, _ := i.Get("pear")
	data, err := json.Marshal(item)
	if err != nil || string(data) != `{"name":"pear","quantity":4,"price_cents":7}` {
		t.Fatalf("JSON=%s err=%v", data, err)
	}
}

func TestOverflow(t *testing.T) {
	i := New()
	if err := i.Add("max", math.MaxInt64, 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("max", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("restock: %v", err)
	}
	if item, _ := i.Get("max"); item.Quantity != math.MaxInt64 {
		t.Fatalf("failed restock changed item: %+v", item)
	}
	if err := i.Restock("max", 0); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"max"}); total != math.MaxInt64 || err != nil {
		t.Fatalf("boundary total=%d err=%v", total, err)
	}
	if _, err := i.TotalValue([]string{"max", "max"}); !errors.Is(err, ErrOverflow) {
		t.Fatalf("sum overflow: %v", err)
	}
	if err := i.Add("product", math.MaxInt64, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := i.TotalValue([]string{"product"}); !errors.Is(err, ErrOverflow) {
		t.Fatalf("product overflow: %v", err)
	}
	if err := i.Add("free", math.MaxInt64, 0); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"free"}); total != 0 || err != nil {
		t.Fatalf("free total=%d err=%v", total, err)
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.txt")
	if _, err := Load(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	cases := []struct {
		name, input string
		want, cause error
	}{
		{"fields", "apple,1", ErrInvalid, nil},
		{"quantity syntax", "apple,no,2", ErrInvalid, strconv.ErrSyntax},
		{"price syntax", "apple,1,no", ErrInvalid, strconv.ErrSyntax},
		{"quantity range", "apple,9223372036854775808,1", ErrOverflow, strconv.ErrRange},
		{"price range", "apple,1,9223372036854775808", ErrOverflow, strconv.ErrRange},
		{"negative quantity", "apple,-1,2", ErrInvalid, nil},
		{"negative price", "apple,1,-2", ErrInvalid, nil},
		{"empty name", " ,1,2", ErrInvalid, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte("\n"+tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			inv, err := Load(path)
			if inv != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Load=%v, %v; want %v", inv, err, tc.want)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("missing parse cause: %v", err)
			}
			if !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("missing line context: %v", err)
			}
		})
	}
	if err := os.WriteFile(path, []byte("\n apple , +2 , 03\r\n\t\n apple,4,5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := i.Report(true); got != "apple: 4\n  price=5\n" {
		t.Fatalf("loaded report=%q", got)
	}
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	i, err = Load(path)
	if err != nil || i == nil || i.Report(false) != "" {
		t.Fatalf("empty load=%v, %v", i, err)
	}
}
