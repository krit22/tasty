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
	i.Add("pear", 3, 40)
	i.Add("apple", 2, 50)
	for _, tc := range []struct {
		verbose bool
		want    string
	}{
		{false, "apple: 2\npear: 3\n"},
		{true, "apple: 2\n  price=50\npear: 3\n  price=40\n"},
	} {
		for n := 0; n < 20; n++ {
			if got := i.Report(tc.verbose); got != tc.want {
				t.Fatalf("Report(%v) = %q, want %q", tc.verbose, got, tc.want)
			}
		}
	}
	if got := New().Report(true); got != "" {
		t.Fatalf("empty report = %q", got)
	}
}

func TestArithmeticErrors(t *testing.T) {
	i := New()
	i.Add("full", math.MaxInt64, 1)
	if err := i.Restock("full", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("Restock: %v", err)
	}
	if item, _ := i.Get("full"); item.Quantity != math.MaxInt64 {
		t.Fatal("overflow changed quantity")
	}
	if err := i.Restock("full", 0); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("missing", 1); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := i.Restock("full", -1); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	i.Add("product", math.MaxInt64, 2)
	i.Add("one", 1, 1)
	i.Add("zero", 0, math.MaxInt64)
	for _, tc := range []struct {
		names []string
		want  error
	}{
		{[]string{"missing"}, ErrNotFound},
		{[]string{"product"}, ErrOverflow},
		{[]string{"full", "one"}, ErrOverflow},
		{[]string{"full", "full"}, ErrOverflow},
	} {
		if _, err := i.TotalValue(tc.names); !errors.Is(err, tc.want) {
			t.Fatalf("TotalValue(%v): %v", tc.names, err)
		}
	}
	if total, err := i.TotalValue([]string{"full", "zero"}); err != nil || total != math.MaxInt64 {
		t.Fatalf("boundary: %d, %v", total, err)
	}
	if total, err := i.TotalValue(nil); err != nil || total != 0 {
		t.Fatalf("empty: %d, %v", total, err)
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	for _, tc := range []struct {
		data        string
		want, cause error
	}{
		{"apple,1", ErrInvalid, nil},
		{"apple,no,1", ErrInvalid, strconv.ErrSyntax},
		{"apple,1,no", ErrInvalid, strconv.ErrSyntax},
		{"apple,9223372036854775808,1", ErrOverflow, strconv.ErrRange},
		{"apple,1,9223372036854775808", ErrOverflow, strconv.ErrRange},
		{"apple,-1,1", ErrInvalid, nil},
		{"apple,1,-1", ErrInvalid, nil},
		{",1,1", ErrInvalid, nil},
	} {
		t.Run(tc.data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "items")
			if err := os.WriteFile(path, []byte("\n"+tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			inv, err := Load(path)
			if inv != nil || !errors.Is(err, tc.want) {
				t.Fatalf("Load = %v, %v", inv, err)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("lost parse cause: %v", err)
			}
			if !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("missing line context: %v", err)
			}
		})
	}
}

func TestRetainedBehavior(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items")
	if err := os.WriteFile(path, []byte("\n apple , 2 , 50\n\t\napple,3,10\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"apple", "apple"}); total != 60 || err != nil {
		t.Fatalf("repeated names: %d, %v", total, err)
	}
	item, ok := i.Get("apple")
	if !ok || item != (Item{"apple", 3, 10}) {
		t.Fatalf("item: %+v, %v", item, ok)
	}
	data, err := json.Marshal(item)
	if err != nil || string(data) != `{"name":"apple","quantity":3,"price_cents":10}` {
		t.Fatalf("JSON: %s, %v", data, err)
	}
	if item, ok := i.Get("missing"); ok || item != (Item{}) {
		t.Fatalf("missing: %+v, %v", item, ok)
	}
	for _, item := range []Item{{"", 1, 1}, {"x", -1, 1}, {"x", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	var zero Inventory
	if err := zero.Add("zero", 0, 0); err != nil {
		t.Fatal(err)
	}
}
