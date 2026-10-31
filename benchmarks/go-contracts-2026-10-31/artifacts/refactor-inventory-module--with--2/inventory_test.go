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

func TestReportsAndReplacement(t *testing.T) {
	var i Inventory
	for _, item := range []Item{{"pear", 3, 40}, {"apple", 9, 1}, {"apple", 2, 50}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		verbose bool
		want    string
	}{
		{false, "apple: 2\npear: 3\n"},
		{true, "apple: 2\n  price=50\npear: 3\n  price=40\n"},
	} {
		if got := i.Report(tc.verbose); got != tc.want {
			t.Fatalf("Report(%v) = %q, want %q", tc.verbose, got, tc.want)
		}
	}
	if total, err := i.TotalValue([]string{"apple", "apple"}); err != nil || total != 200 {
		t.Fatalf("repeated total = %d, %v", total, err)
	}
	if total, err := i.TotalValue(nil); err != nil || total != 0 {
		t.Fatalf("empty total = %d, %v", total, err)
	}
	b, err := json.Marshal(Item{"apple", 2, 50})
	if err != nil || string(b) != `{"name":"apple","quantity":2,"price_cents":50}` {
		t.Fatalf("JSON = %s, %v", b, err)
	}
	if got := New().Report(true); got != "" {
		t.Fatalf("empty report = %q", got)
	}
}

func TestValidationAndMissing(t *testing.T) {
	i := New()
	for _, item := range []Item{{"", 1, 1}, {"a", -1, 1}, {"a", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v): %v", item, err)
		}
	}
	if _, ok := i.Get("missing"); ok {
		t.Fatal("missing Get succeeded")
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing Restock: %v", err)
	}
	if total, err := i.TotalValue([]string{"missing"}); total != 0 || !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing total = %d, %v", total, err)
	}
	if err := i.Add("a", 1, 2); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", -1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("negative Restock: %v", err)
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
		{"product", math.MaxInt64, 2, []string{"a"}, 0, true},
		{"sum", math.MaxInt64, 1, []string{"a", "a"}, 0, true},
		{"boundary", math.MaxInt64, 1, []string{"a"}, math.MaxInt64, false},
		{"zero price", math.MaxInt64, 0, []string{"a"}, 0, false},
		{"zero quantity", 0, math.MaxInt64, []string{"a"}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := New()
			if err := i.Add("a", tc.quantity, tc.price); err != nil {
				t.Fatal(err)
			}
			got, err := i.TotalValue(tc.names)
			if got != tc.want || errors.Is(err, ErrOverflow) != tc.overflow || (!tc.overflow && err != nil) {
				t.Fatalf("total = %d, %v", got, err)
			}
		})
	}
	i := New()
	if err := i.Add("a", math.MaxInt64-1, 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("overflow restock: %v", err)
	}
	if got, _ := i.Get("a"); got.Quantity != math.MaxInt64 {
		t.Fatalf("overflow changed item: %+v", got)
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.csv")
	if _, err := Load(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	for _, tc := range []struct {
		name, data string
		want       error
	}{
		{"fields", "a,1", ErrInvalid},
		{"empty name", ",1,2", ErrInvalid},
		{"negative", "a,-1,2", ErrInvalid},
		{"quantity syntax", "a,no,2", strconv.ErrSyntax},
		{"price syntax", "a,1,no", strconv.ErrSyntax},
		{"quantity range", "a,9223372036854775808,1", strconv.ErrRange},
		{"price range", "a,1,9223372036854775808", strconv.ErrRange},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			i, err := Load(path)
			if i != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Load = %v, %v; want %v", i, err, tc.want)
			}
		})
	}
	if err := os.WriteFile(path, []byte("\n a , 2 , 3 \r\n \t\n a,4,5\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := i.Get("a"); !ok || got != (Item{"a", 4, 5}) {
		t.Fatalf("loaded item = %+v, %v", got, ok)
	}
}
