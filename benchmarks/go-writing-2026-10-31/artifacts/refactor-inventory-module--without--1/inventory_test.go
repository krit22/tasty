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

func TestReport(t *testing.T) {
	i := New()
	for _, item := range []Item{{"pear", 3, 20}, {"apple", 2, 50}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		verbose bool
		want    string
	}{{false, "apple: 2\npear: 3\n"}, {true, "apple: 2\n  price=50\npear: 3\n  price=20\n"}} {
		for range 20 {
			if got := i.Report(tc.verbose); got != tc.want {
				t.Fatalf("Report(%v) = %q; want %q", tc.verbose, got, tc.want)
			}
		}
	}
	if got := New().Report(true); got != "" {
		t.Fatalf("empty report = %q", got)
	}
}

func TestArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name            string
		quantity, price int64
		names           []string
		want            int64
		err             error
	}{
		{"maximum", math.MaxInt64, 1, []string{"x"}, math.MaxInt64, nil},
		{"product overflow", math.MaxInt64, 2, []string{"x"}, 0, ErrOverflow},
		{"sum overflow", math.MaxInt64, 1, []string{"x", "x"}, 0, ErrOverflow},
		{"zero quantity", 0, math.MaxInt64, []string{"x"}, 0, nil},
		{"zero price", math.MaxInt64, 0, []string{"x"}, 0, nil},
		{"duplicates", 2, 3, []string{"x", "x"}, 12, nil},
		{"empty", 2, 3, nil, 0, nil},
		{"missing", 2, 3, []string{"x", "absent"}, 0, ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := New()
			if err := i.Add("x", tc.quantity, tc.price); err != nil {
				t.Fatal(err)
			}
			got, err := i.TotalValue(tc.names)
			if got != tc.want || !errors.Is(err, tc.err) {
				t.Fatalf("TotalValue = %d, %v; want %d, %v", got, err, tc.want, tc.err)
			}
		})
	}
}

func TestMutations(t *testing.T) {
	var i Inventory
	if err := i.Add("x", math.MaxInt64-1, 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("x", 1); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		amount int64
		want   error
	}{
		{"x", 1, ErrOverflow}, {"x", -1, ErrInvalid},
		{"absent", 1, ErrNotFound}, {"absent", -1, ErrNotFound},
	} {
		if err := i.Restock(tc.name, tc.amount); !errors.Is(err, tc.want) {
			t.Fatalf("Restock(%q, %d) = %v", tc.name, tc.amount, err)
		}
	}
	if got, _ := i.Get("x"); got.Quantity != math.MaxInt64 {
		t.Fatalf("failed restock changed item: %+v", got)
	}
	for _, item := range []Item{{"", 1, 1}, {"x", -1, 1}, {"x", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v) = %v", item, err)
		}
	}
	if got, _ := i.Get("x"); got.Quantity != math.MaxInt64 {
		t.Fatalf("failed Add changed item: %+v", got)
	}
	if err := i.Add("x", 0, 0); err != nil {
		t.Fatal(err)
	}
	if got, ok := i.Get("x"); !ok || got != (Item{Name: "x"}) {
		t.Fatalf("replacement = %+v, %v", got, ok)
	}
	if got, ok := i.Get("absent"); ok || got != (Item{}) {
		t.Fatalf("missing = %+v, %v", got, ok)
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.txt")
	if inv, err := Load(path); inv != nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file = %v, %v", inv, err)
	}
	for _, tc := range []struct {
		name, data, field string
		numeric           bool
	}{
		{"columns", "x,1\n", "line 1", false},
		{"quantity syntax", "\nx,no,1\n", "line 2 quantity", true},
		{"price syntax", "x,1,no\n", "line 1 price", true},
		{"quantity range", "x,9223372036854775808,1", "quantity", true},
		{"price range", "x,1,9223372036854775808", "price", true},
		{"negative quantity", "x,-1,1", "line 1", false},
		{"negative price", "x,1,-1", "line 1", false},
		{"empty name", " ,1,1", "line 1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			inv, err := Load(path)
			if inv != nil || !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("Load = %v, %v", inv, err)
			}
			if tc.numeric {
				var numErr *strconv.NumError
				if !errors.As(err, &numErr) {
					t.Fatalf("numeric error lost: %v", err)
				}
			}
		})
	}
	if err := os.WriteFile(path, []byte("\n apple , 2 , 50 \r\n\t\npear,0,0\napple,3,60\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := i.Report(true); got != "apple: 3\n  price=60\npear: 0\n  price=0\n" {
		t.Fatalf("loaded report = %q", got)
	}
}

func TestItemJSON(t *testing.T) {
	data, err := json.Marshal(Item{"apple", 2, 50})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"name":"apple","quantity":2,"price_cents":50}` {
		t.Fatalf("JSON = %s", data)
	}
}
