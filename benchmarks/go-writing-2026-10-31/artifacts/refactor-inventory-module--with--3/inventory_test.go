package inventory

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
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

func TestItemContract(t *testing.T) {
	var i Inventory
	for _, item := range []Item{{"", 1, 1}, {"x", -1, 1}, {"x", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v): %v", item, err)
		}
	}
	if err := i.Add("apple", 2, 50); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("apple", 3, 20); err != nil {
		t.Fatal(err)
	}
	item, ok := i.Get("apple")
	if !ok || item != (Item{"apple", 3, 20}) {
		t.Fatalf("Get: %+v, %v", item, ok)
	}
	data, err := json.Marshal(item)
	if err != nil || string(data) != `{"name":"apple","quantity":3,"price_cents":20}` {
		t.Fatalf("JSON: %s, %v", data, err)
	}
	if item, ok := i.Get("missing"); ok || item != (Item{}) {
		t.Fatalf("missing Get: %+v, %v", item, ok)
	}
	if total, err := i.TotalValue([]string{"apple", "apple"}); total != 120 || err != nil {
		t.Fatalf("duplicate total: %d, %v", total, err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("empty total: %d, %v", total, err)
	}
	if err := i.Restock("apple", -1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("negative restock: %v", err)
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing restock: %v", err)
	}
	if _, err := i.TotalValue([]string{"apple", "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing total: %v", err)
	}
}

func TestOverflow(t *testing.T) {
	for _, tc := range []struct {
		name            string
		quantity, price int64
		names           []string
		want            int64
		overflow        bool
	}{
		{"product", math.MaxInt64, 2, []string{"x"}, 0, true},
		{"sum", math.MaxInt64, 1, []string{"x", "x"}, 0, true},
		{"boundary", math.MaxInt64, 1, []string{"x"}, math.MaxInt64, false},
		{"zero quantity", 0, math.MaxInt64, []string{"x"}, 0, false},
		{"zero price", math.MaxInt64, 0, []string{"x"}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := New()
			if err := i.Add("x", tc.quantity, tc.price); err != nil {
				t.Fatal(err)
			}
			got, err := i.TotalValue(tc.names)
			if tc.overflow {
				if !errors.Is(err, ErrOverflow) {
					t.Fatalf("total: %d, %v", got, err)
				}
				return
			}
			if got != tc.want || err != nil {
				t.Fatalf("total: %d, %v; want %d", got, err, tc.want)
			}
		})
	}
	i := New()
	if err := i.Add("x", math.MaxInt64-1, 0); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("x", 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("x", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("restock overflow: %v", err)
	}
	item, _ := i.Get("x")
	if item.Quantity != math.MaxInt64 {
		t.Fatalf("overflow changed quantity: %d", item.Quantity)
	}
}

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name, data string
		sentinel   error
		numeric    bool
	}{
		{"valid", "\n apple , 2 , 50 \n\n apple,3,20\npear,0,0\n", nil, false},
		{"fields", "apple,1", ErrInvalid, false},
		{"empty name", ",1,2", ErrInvalid, false},
		{"negative", "apple,-1,2", ErrInvalid, false},
		{"quantity syntax", "apple,no,2", strconv.ErrSyntax, true},
		{"price syntax", "apple,1,no", strconv.ErrSyntax, true},
		{"quantity range", "apple,9223372036854775808,2", strconv.ErrRange, true},
		{"price range", "apple,1,9223372036854775808", strconv.ErrRange, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "items.csv")
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			i, err := Load(path)
			if tc.sentinel != nil {
				if i != nil || !errors.Is(err, tc.sentinel) {
					t.Fatalf("Load: %v, %v", i, err)
				}
				if tc.numeric {
					var numErr *strconv.NumError
					if !errors.As(err, &numErr) {
						t.Fatalf("numeric cause: %v", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := i.Report(true); got != "apple: 3\n  price=20\npear: 0\n  price=0\n" {
				t.Fatalf("loaded report: %q", got)
			}
		})
	}
	if i, err := Load(filepath.Join(t.TempDir(), "missing")); i != nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v, %v", i, err)
	}
	if i, err := Load(t.TempDir()); i != nil || err == nil {
		t.Fatalf("directory: %v, %v", i, err)
	}
}

func TestReport(t *testing.T) {
	i := New()
	if got := i.Report(true); got != "" {
		t.Fatalf("empty report: %q", got)
	}
	for _, name := range []string{"pear", "apple", "banana"} {
		if err := i.Add(name, 2, 50); err != nil {
			t.Fatal(err)
		}
	}
	for n := 0; n < 20; n++ {
		if got := i.Report(false); got != "apple: 2\nbanana: 2\npear: 2\n" {
			t.Fatalf("report: %q", got)
		}
		if got := i.Report(true); got != "apple: 2\n  price=50\nbanana: 2\n  price=50\npear: 2\n  price=50\n" {
			t.Fatalf("verbose report: %q", got)
		}
	}
}
