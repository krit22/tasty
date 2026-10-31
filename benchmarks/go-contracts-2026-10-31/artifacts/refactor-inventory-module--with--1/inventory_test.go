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

func TestContracts(t *testing.T) {
	var i Inventory
	if err := i.Add("pear", 1, 20); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("apple", 2, 30); err != nil {
		t.Fatal(err)
	}
	if err := i.Add("pear", 3, 40); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		verbose bool
		want    string
	}{
		{false, "apple: 2\npear: 3\n"},
		{true, "apple: 2\n  price=30\npear: 3\n  price=40\n"},
	} {
		if got := i.Report(tc.verbose); got != tc.want {
			t.Fatalf("Report(%v) = %q, want %q", tc.verbose, got, tc.want)
		}
	}
	total, err := i.TotalValue([]string{"pear", "pear"})
	if err != nil || total != 240 {
		t.Fatalf("total=%d err=%v", total, err)
	}
	if total, err := i.TotalValue(nil); total != 0 || err != nil {
		t.Fatalf("empty total=%d err=%v", total, err)
	}
	if _, ok := i.Get("missing"); ok {
		t.Fatal("missing item found")
	}
	if New().Report(true) != "" {
		t.Fatal("nonempty report")
	}
	b, err := json.Marshal(Item{Name: "apple", Quantity: 2, PriceCents: 30})
	if err != nil || string(b) != `{"name":"apple","quantity":2,"price_cents":30}` {
		t.Fatalf("JSON=%s err=%v", b, err)
	}
}

func TestFailures(t *testing.T) {
	i := New()
	for _, tc := range []struct {
		name            string
		quantity, price int64
	}{{"", 0, 0}, {"a", -1, 0}, {"a", 0, -1}} {
		if err := i.Add(tc.name, tc.quantity, tc.price); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add(%+v): %v", tc, err)
		}
	}
	if err := i.Restock("missing", -1); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := i.TotalValue([]string{"missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := i.Add("a", math.MaxInt64, 1); err != nil {
		t.Fatal(err)
	}
	if err := i.Restock("a", -1); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if err := i.Restock("a", 1); !errors.Is(err, ErrOverflow) {
		t.Fatal(err)
	}
	if got, _ := i.Get("a"); got.Quantity != math.MaxInt64 {
		t.Fatalf("mutated: %+v", got)
	}
	if err := i.Restock("a", 0); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"a"}); total != math.MaxInt64 || err != nil {
		t.Fatalf("boundary=%d err=%v", total, err)
	}
	if total, err := i.TotalValue([]string{"a", "a"}); total != 0 || !errors.Is(err, ErrOverflow) {
		t.Fatalf("sum=%d err=%v", total, err)
	}
	if err := i.Add("a", math.MaxInt64, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := i.TotalValue([]string{"a"}); !errors.Is(err, ErrOverflow) {
		t.Fatal(err)
	}
	if err := i.Add("a", math.MaxInt64, 0); err != nil {
		t.Fatal(err)
	}
	if total, err := i.TotalValue([]string{"a"}); total != 0 || err != nil {
		t.Fatalf("zero price=%d err=%v", total, err)
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.txt")
	if _, err := Load(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v", err)
	}
	for _, tc := range []struct {
		input string
		cause error
	}{
		{"a,1", ErrInvalid}, {",1,2", ErrInvalid}, {"a,-1,2", ErrInvalid}, {"a,1,-2", ErrInvalid},
		{"a,no,2", strconv.ErrSyntax}, {"a,1,no", strconv.ErrSyntax},
		{"a,9223372036854775808,1", strconv.ErrRange}, {"a,1,9223372036854775808", strconv.ErrRange},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if err := os.WriteFile(path, []byte("\n"+tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := Load(path)
			if got != nil || !errors.Is(err, tc.cause) {
				t.Fatalf("Load=%v, %v; want %v", got, err, tc.cause)
			}
			if !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), path) {
				t.Fatalf("missing context: %v", err)
			}
		})
	}
	if err := os.WriteFile(path, []byte("\n apple , +2 , 30 \r\n\t\n apple,3,40\nzero,0,0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	i, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := i.Report(true); got != "apple: 3\n  price=40\nzero: 0\n  price=0\n" {
		t.Fatalf("loaded report=%q", got)
	}
}
