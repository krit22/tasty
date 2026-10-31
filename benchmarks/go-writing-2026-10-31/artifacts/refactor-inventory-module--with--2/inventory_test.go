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

func TestRetainedBehavior(t *testing.T) {
	var i Inventory
	if err := i.Add("apple", 1, 5); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("apple", 2, 7); err != nil {
		t.Fatal(err)
	}
	if got, ok := i.Get("apple"); !ok || got != (Item{"apple", 2, 7}) {
		t.Fatalf("Get = %+v, %v", got, ok)
	}
	if got, ok := i.Get("missing"); ok || got != (Item{}) {
		t.Fatalf("missing Get = %+v, %v", got, ok)
	}
	if total, err := i.TotalValue([]string{"apple", "apple"}); total != 28 || err != nil {
		t.Fatalf("total = %d, %v", total, err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("empty total = %d, %v", total, err)
	}
	data, err := json.Marshal(Item{"apple", 2, 7})
	if err != nil || string(data) != `{"name":"apple","quantity":2,"price_cents":7}` {
		t.Fatalf("JSON = %s, %v", data, err)
	}
}

func TestReport(t *testing.T) {
	i := New()
	for _, name := range []string{"pear", "apple", "banana"} {
		if err := i.Add(name, 2, 3); err != nil {
			t.Fatal(err)
		}
	}
	for _, verbose := range []bool{false, true} {
		want := ""
		for _, name := range []string{"apple", "banana", "pear"} {
			want += name + ": 2\n"
			if verbose {
				want += "  price=3\n"
			}
		}
		if got := i.Report(verbose); got != want {
			t.Fatalf("Report(%v) = %q, want %q", verbose, got, want)
		}
	}
	if got := New().Report(true); got != "" {
		t.Fatalf("empty report = %q", got)
	}
}

func TestArithmeticErrors(t *testing.T) {
	i := New()
	for _, item := range []Item{{"quantity", math.MaxInt64, 1}, {"product", math.MaxInt64, 2}, {"one", 1, 1}, {"zero", 0, math.MaxInt64}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); err != nil {
			t.Fatal(err)
		}
	}
	if err := i.Restock("quantity", 1); !errors.Is(err, ErrOverflow) {
		t.Fatalf("Restock = %v", err)
	}
	if item, _ := i.Get("quantity"); item.Quantity != math.MaxInt64 {
		t.Fatalf("overflow changed item: %+v", item)
	}
	if err := i.Restock("quantity", 0); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		names []string
		want  error
	}{
		{[]string{"missing"}, ErrNotFound}, {[]string{"product"}, ErrOverflow}, {[]string{"quantity", "one"}, ErrOverflow},
	} {
		if _, err := i.TotalValue(tc.names); !errors.Is(err, tc.want) {
			t.Fatalf("TotalValue(%v) = %v, want %v", tc.names, err, tc.want)
		}
	}
	if total, err := i.TotalValue([]string{"quantity", "zero"}); total != math.MaxInt64 || err != nil {
		t.Fatalf("boundary total = %d, %v", total, err)
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing restock = %v", err)
	}
	if err := i.Restock("one", -1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("negative restock = %v", err)
	}
	for _, item := range []Item{{"", 1, 1}, {"negative", -1, 1}, {"negative", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v) = %v", item, err)
		}
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.csv")
	if _, err := Load(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file = %v", err)
	}
	for _, tc := range []struct {
		input string
		want  error
		parse error
	}{
		{"apple,1", ErrInvalid, nil}, {",1,2", ErrInvalid, nil}, {"apple,-1,2", ErrInvalid, nil}, {"apple,1,-2", ErrInvalid, nil},
		{"apple,no,2", ErrInvalid, strconv.ErrSyntax}, {"apple,1,no", ErrInvalid, strconv.ErrSyntax},
		{"apple,9223372036854775808,2", ErrOverflow, strconv.ErrRange}, {"apple,1,9223372036854775808", ErrOverflow, strconv.ErrRange},
	} {
		if err := os.WriteFile(path, []byte("\n"+tc.input), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := Load(path)
		if got != nil || !errors.Is(err, tc.want) || (tc.parse != nil && !errors.Is(err, tc.parse)) {
			t.Fatalf("Load(%q) = %v, %v", tc.input, got, err)
		}
		if !strings.Contains(err.Error(), "line 2") {
			t.Fatalf("missing line context: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte(" \n apple , 1 , 2 \r\napple,3,4\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := i.Get("apple"); got != (Item{"apple", 3, 4}) {
		t.Fatalf("loaded item = %+v", got)
	}
}
