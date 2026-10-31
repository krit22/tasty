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

func TestRestockOverflow(t *testing.T) {
	i := New()
	if err := i.Add("apple", math.MaxInt64, 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("apple", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("err=%v", err)
	}
	item, _ := i.Get("apple")
	if item.Quantity != math.MaxInt64 {
		t.Fatalf("quantity changed: %d", item.Quantity)
	}
	if err := i.Restock("apple", 0); err != nil {
		t.Fatal(err)
	}
}

func TestTotalValueErrorsAndBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name            string
		quantity, price int64
		names           []string
		want            int64
		err             error
	}{
		{"missing", 1, 1, []string{"missing"}, 0, ErrNotFound},
		{"product overflow", math.MaxInt64, 2, []string{"apple"}, 0, ErrOverflow},
		{"sum overflow", math.MaxInt64, 1, []string{"apple", "apple"}, 0, ErrOverflow},
		{"maximum", math.MaxInt64, 1, []string{"apple"}, math.MaxInt64, nil},
		{"zero quantity", 0, math.MaxInt64, []string{"apple"}, 0, nil},
		{"zero price", math.MaxInt64, 0, []string{"apple"}, 0, nil},
		{"duplicates", 2, 3, []string{"apple", "apple"}, 12, nil},
		{"empty", 1, 1, nil, 0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := New()
			if err := i.Add("apple", tc.quantity, tc.price); err != nil {
				t.Fatal(err)
			}
			got, err := i.TotalValue(tc.names)
			if got != tc.want || !errors.Is(err, tc.err) {
				t.Fatalf("got %d, %v; want %d, %v", got, err, tc.want, tc.err)
			}
		})
	}
}

func TestLoadErrors(t *testing.T) {
	if got, err := Load(filepath.Join(t.TempDir(), "missing")); got != nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, %v", got, err)
	}
	for _, tc := range []struct {
		name, input string
		cause       error
	}{
		{"fields", "apple,1", ErrInvalid},
		{"quantity syntax", "apple,no,1", strconv.ErrSyntax},
		{"price syntax", "apple,1,no", strconv.ErrSyntax},
		{"quantity range", "apple,9223372036854775808,1", strconv.ErrRange},
		{"price range", "apple,1,9223372036854775808", strconv.ErrRange},
		{"negative quantity", "apple,-1,1", ErrInvalid},
		{"negative price", "apple,1,-1", ErrInvalid},
		{"empty name", ",1,1", ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "inventory")
			if err := os.WriteFile(path, []byte("\n"+tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := Load(path)
			if got != nil || !errors.Is(err, tc.cause) || !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("got %v, %v", got, err)
			}
		})
	}
}

func TestLoadAndReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory")
	if err := os.WriteFile(path, []byte(" banana , 2 , 30 \n\n apple,1,5\nbanana,4,10\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 20; n++ {
		if got := i.Report(false); got != "apple: 1\nbanana: 4\n" {
			t.Fatalf("report=%q", got)
		}
		if got := i.Report(true); got != "apple: 1\n  price=5\nbanana: 4\n  price=10\n" {
			t.Fatalf("report=%q", got)
		}
	}
	item, _ := i.Get("banana")
	data, err := json.Marshal(item)
	if err != nil || string(data) != `{"name":"banana","quantity":4,"price_cents":10}` {
		t.Fatalf("JSON=%s err=%v", data, err)
	}
	if New().Report(true) != "" {
		t.Fatal("nonempty report")
	}
}

func TestValidationAndZeroValue(t *testing.T) {
	var i Inventory
	for _, item := range []Item{{"", 1, 1}, {"apple", -1, 1}, {"apple", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("err=%v", err)
		}
	}
	if err := i.Add("apple", 2, 3); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("apple", -1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
	if got, ok := i.Get("missing"); ok || got != (Item{}) {
		t.Fatalf("got %v, %v", got, ok)
	}
	if err := i.Add("apple", 0, 0); err != nil {
		t.Fatal(err)
	}
	if got, _ := i.Get("apple"); got.Quantity != 0 || got.PriceCents != 0 {
		t.Fatalf("got %v", got)
	}
}
