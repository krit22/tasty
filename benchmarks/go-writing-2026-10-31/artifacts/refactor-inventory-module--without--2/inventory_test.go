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

func TestArithmeticBoundaries(t *testing.T) {
	i := New()
	i.Add("max", math.MaxInt64, 1)
	if err := i.Restock("max", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("restock: %v", err)
	}
	if item, _ := i.Get("max"); item.Quantity != math.MaxInt64 {
		t.Fatal("overflow changed quantity")
	}
	if err := i.Restock("max", 0); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"max"}); total != math.MaxInt64 || err != nil {
		t.Fatalf("boundary: %d, %v", total, err)
	}
	i.Add("one", 1, 1)
	i.Add("product", math.MaxInt64, 2)
	for _, names := range [][]string{{"max", "one"}, {"product"}, {"max", "max"}} {
		if total, err := i.TotalValue(names); total != 0 || !errors.Is(err, ErrOverflow) {
			t.Fatalf("%v: %d, %v", names, total, err)
		}
	}
	i.Add("zero", math.MaxInt64, 0)
	if total, err := i.TotalValue([]string{"zero"}); total != 0 || err != nil {
		t.Fatalf("zero: %d, %v", total, err)
	}
}

func TestMissingAndInvalid(t *testing.T) {
	i := New()
	if _, err := i.TotalValue([]string{"missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	for _, item := range []Item{{"", 0, 0}, {"a", -1, 0}, {"a", 0, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	i.Add("a", 2, 3)
	if err := i.Restock("a", -1); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("empty: %d %v", total, err)
	}
	if _, ok := i.Get("missing"); ok {
		t.Fatal("found missing item")
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file error: %v", err)
	}
	for _, tc := range []struct {
		text        string
		want, cause error
	}{
		{"apple,1", ErrInvalid, nil},
		{"apple,no,2", ErrInvalid, strconv.ErrSyntax},
		{"apple,1,no", ErrInvalid, strconv.ErrSyntax},
		{"apple,9223372036854775808,1", ErrOverflow, strconv.ErrRange},
		{"apple,1,9223372036854775808", ErrOverflow, strconv.ErrRange},
		{"apple,-1,2", ErrInvalid, nil},
		{"apple,1,-2", ErrInvalid, nil},
		{",1,2", ErrInvalid, nil},
	} {
		t.Run(tc.text, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "items")
			if err := os.WriteFile(path, []byte("\n"+tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			inv, err := Load(path)
			if inv != nil || !errors.Is(err, tc.want) || !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("inventory=%v err=%v", inv, err)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("lost parse cause: %v", err)
			}
		})
	}
}

func TestLoadAndReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items")
	if err := os.WriteFile(path, []byte(" z , 1 , 2 \n\n a,3,4\nz,5,6\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 20; n++ {
		if got := i.Report(false); got != "a: 3\nz: 5\n" {
			t.Fatalf("report: %q", got)
		}
		if got := i.Report(true); got != "a: 3\n  price=4\nz: 5\n  price=6\n" {
			t.Fatalf("verbose: %q", got)
		}
	}
	if total, err := i.TotalValue([]string{"a", "a"}); total != 24 || err != nil {
		t.Fatalf("duplicates: %d %v", total, err)
	}
	data, err := json.Marshal(Item{"a", 3, 4})
	if err != nil || string(data) != `{"name":"a","quantity":3,"price_cents":4}` {
		t.Fatalf("JSON: %s %v", data, err)
	}
}

func TestZeroValueInventory(t *testing.T) {
	var i Inventory
	if i.Report(true) != "" {
		t.Fatal("nonempty report")
	}
	if err := i.Add("a", 0, 0); err != nil {
		t.Fatal(err)
	}
	if item, ok := i.Get("a"); !ok || item.Name != "a" {
		t.Fatal("add failed")
	}
}
