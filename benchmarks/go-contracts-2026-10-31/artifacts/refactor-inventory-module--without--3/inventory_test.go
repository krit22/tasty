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

func TestValidationAndReplacement(t *testing.T) {
	var i Inventory
	for _, item := range []Item{{Name: "", Quantity: 1}, {Name: "a", Quantity: -1}, {Name: "a", PriceCents: -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v): %v", item, err)
		}
	}
	if err := i.Add("a", 2, 3); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("a", 4, 5); err != nil {
		t.Fatal(err)
	}
	if got, ok := i.Get("a"); !ok || got != (Item{"a", 4, 5}) {
		t.Fatalf("Get: %+v, %v", got, ok)
	}
	if got, ok := i.Get("missing"); ok || got != (Item{}) {
		t.Fatalf("missing Get: %+v, %v", got, ok)
	}
	if err := i.Restock("a", -1); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"a", "a"}); total != 40 || err != nil {
		t.Fatalf("total=%d err=%v", total, err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("total=%d err=%v", total, err)
	}
	if _, err := i.TotalValue([]string{"a", "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestOverflow(t *testing.T) {
	i := New()
	if err := i.Add("a", math.MaxInt64-1, 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", 1); !errors.Is(err, ErrOverflow) {
		t.Fatal(err)
	}
	if got, _ := i.Get("a"); got.Quantity != math.MaxInt64 {
		t.Fatalf("overflow mutated item: %+v", got)
	}
	if total, err := i.TotalValue([]string{"a"}); total != math.MaxInt64 || err != nil {
		t.Fatalf("total=%d err=%v", total, err)
	}
	for _, tc := range []struct {
		name            string
		quantity, price int64
		names           []string
	}{
		{"product", math.MaxInt64, 2, []string{"a"}},
		{"sum", math.MaxInt64, 1, []string{"a", "a"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := i.Add("a", tc.quantity, tc.price); err != nil {
				t.Fatal(err)
			}
			if _, err := i.TotalValue(tc.names); !errors.Is(err, ErrOverflow) {
				t.Fatal(err)
			}
		})
	}
	for _, item := range []Item{{"a", math.MaxInt64, 0}, {"a", 0, math.MaxInt64}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); err != nil {
			t.Fatal(err)
		}
		if total, err := i.TotalValue([]string{"a"}); total != 0 || err != nil {
			t.Fatalf("total=%d err=%v", total, err)
		}
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(filepath.Join(dir, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, data string
		numeric    error
	}{
		{"fields", "a,1", nil}, {"name", ",1,2", nil},
		{"negative quantity", "a,-1,2", nil}, {"negative price", "a,1,-2", nil},
		{"quantity syntax", "a,no,2", strconv.ErrSyntax}, {"price syntax", "a,1,no", strconv.ErrSyntax},
		{"quantity range", "a,9223372036854775808,1", strconv.ErrRange},
		{"price range", "a,1,9223372036854775808", strconv.ErrRange},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "input")
			if err := os.WriteFile(path, []byte("\n"+tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			inv, err := Load(path)
			if inv != nil || !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("inventory=%v err=%v", inv, err)
			}
			if tc.numeric != nil && !errors.Is(err, tc.numeric) {
				t.Fatalf("lost numeric error: %v", err)
			}
		})
	}
	path := filepath.Join(dir, "valid")
	if err := os.WriteFile(path, []byte(" \n apple , 2 , 50\r\nbanana,0,0\napple,3,60\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := i.Report(true); got != "apple: 3\n  price=60\nbanana: 0\n  price=0\n" {
		t.Fatalf("report=%q", got)
	}
}

func TestReportAndJSON(t *testing.T) {
	i := New()
	if i.Report(false) != "" || i.Report(true) != "" {
		t.Fatal("nonempty report")
	}
	for _, name := range []string{"zebra", "apple", "middle"} {
		if err := i.Add(name, 2, 50); err != nil {
			t.Fatal(err)
		}
	}
	for n := 0; n < 20; n++ {
		if got := i.Report(false); got != "apple: 2\nmiddle: 2\nzebra: 2\n" {
			t.Fatalf("report=%q", got)
		}
		if got := i.Report(true); got != "apple: 2\n  price=50\nmiddle: 2\n  price=50\nzebra: 2\n  price=50\n" {
			t.Fatalf("report=%q", got)
		}
	}
	data, err := json.Marshal(Item{"apple", 2, 50})
	if err != nil || string(data) != `{"name":"apple","quantity":2,"price_cents":50}` {
		t.Fatalf("json=%s err=%v", data, err)
	}
}
