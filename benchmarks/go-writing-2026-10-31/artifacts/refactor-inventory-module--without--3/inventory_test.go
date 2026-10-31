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
	for _, name := range []string{"pear", "apple"} {
		if err := i.Add(name, 2, 50); err != nil {
			t.Fatal(err)
		}
	}
	for _, verbose := range []bool{false, true} {
		want := "apple: 2\npear: 2\n"
		if verbose {
			want = "apple: 2\n  price=50\npear: 2\n  price=50\n"
		}
		for n := 0; n < 20; n++ {
			if got := i.Report(verbose); got != want {
				t.Fatalf("Report(%v)=%q, want %q", verbose, got, want)
			}
		}
	}
	if got := New().Report(true); got != "" {
		t.Fatalf("empty report=%q", got)
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
	item, _ := i.Get("max")
	if item.Quantity != math.MaxInt64 {
		t.Fatal("failed restock changed quantity")
	}
	if err := i.Restock("max", 0); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"max"}); err != nil || total != math.MaxInt64 {
		t.Fatalf("boundary: %d, %v", total, err)
	}
	if err := i.Add("product", math.MaxInt64, 2); err != nil {
		t.Fatal(err)
	}
	for _, names := range [][]string{{"product"}, {"max", "max"}} {
		if total, err := i.TotalValue(names); total != 0 || !errors.Is(err, ErrOverflow) {
			t.Fatalf("%v: %d, %v", names, total, err)
		}
	}
	if err := i.Add("zero", 0, math.MaxInt64); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"zero"}); total != 0 || err != nil {
		t.Fatalf("zero: %d, %v", total, err)
	}
}

func TestValidationAndReplacement(t *testing.T) {
	var i Inventory
	for _, item := range []Item{{"", 1, 1}, {"a", -1, 1}, {"a", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v): %v", item, err)
		}
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"missing"}); total != 0 || !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %d, %v", total, err)
	}
	if _, ok := i.Get("missing"); ok {
		t.Fatal("found missing item")
	}
	if err := i.Add("a", 2, 3); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", -1); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if err := i.Add("a", 4, 5); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"a", "a"}); total != 40 || err != nil {
		t.Fatalf("duplicates: %d, %v", total, err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("empty: %d, %v", total, err)
	}
	data, err := json.Marshal(i.items["a"])
	if err != nil || string(data) != `{"name":"a","quantity":4,"price_cents":5}` {
		t.Fatalf("JSON: %s, %v", data, err)
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.csv")
	if _, err := Load(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	cases := []struct {
		text     string
		sentinel error
		numeric  error
	}{
		{"a,1", ErrInvalid, nil},
		{"a,no,2", ErrInvalid, strconv.ErrSyntax},
		{"a,1,no", ErrInvalid, strconv.ErrSyntax},
		{"a,9223372036854775808,1", ErrOverflow, strconv.ErrRange},
		{"a,1,9223372036854775808", ErrOverflow, strconv.ErrRange},
		{"a,-1,2", ErrInvalid, nil},
		{"a,1,-2", ErrInvalid, nil},
		{" ,1,2", ErrInvalid, nil},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			if err := os.WriteFile(path, []byte("\n"+tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			i, err := Load(path)
			if i != nil || !errors.Is(err, tc.sentinel) || !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("Load: %v, %v", i, err)
			}
			if tc.numeric != nil {
				var numErr *strconv.NumError
				if !errors.Is(err, tc.numeric) || !errors.As(err, &numErr) {
					t.Fatalf("numeric cause lost: %v", err)
				}
			}
		})
	}
	if err := os.WriteFile(path, []byte("\n apple , 2 , 50 \n\n apple,3,60\npear,0,0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := i.Report(true); got != "apple: 3\n  price=60\npear: 0\n  price=0\n" {
		t.Fatalf("loaded: %q", got)
	}
}
