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

func TestReportAndReplacement(t *testing.T) {
	var i Inventory
	for _, item := range []Item{{"pear", 3, 80}, {"apple", 2, 50}, {"pear", 4, 90}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		verbose bool
		want    string
	}{
		{false, "apple: 2\npear: 4\n"},
		{true, "apple: 2\n  price=50\npear: 4\n  price=90\n"},
	} {
		if got := i.Report(tc.verbose); got != tc.want {
			t.Fatalf("Report(%v)=%q, want %q", tc.verbose, got, tc.want)
		}
	}
	total, err := i.TotalValue([]string{"apple", "apple"})
	if total != 200 || err != nil {
		t.Fatalf("repeated total=%d, %v", total, err)
	}
	data, err := json.Marshal(Item{"apple", 2, 50})
	if err != nil || string(data) != `{"name":"apple","quantity":2,"price_cents":50}` {
		t.Fatalf("JSON=%s, %v", data, err)
	}
}

func TestArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name            string
		quantity, price int64
		names           []string
		want            int64
		wantErr         error
	}{
		{"limit", math.MaxInt64, 1, []string{"x"}, math.MaxInt64, nil},
		{"product overflow", math.MaxInt64, 2, []string{"x"}, 0, ErrOverflow},
		{"sum overflow", math.MaxInt64, 1, []string{"x", "x"}, 0, ErrOverflow},
		{"zero price", math.MaxInt64, 0, []string{"x"}, 0, nil},
		{"zero quantity", 0, math.MaxInt64, []string{"x"}, 0, nil},
		{"missing", 1, 1, []string{"x", "missing"}, 0, ErrNotFound},
		{"empty", 1, 1, nil, 0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := New()
			if err := i.Add("x", tc.quantity, tc.price); err != nil {
				t.Fatal(err)
			}
			got, err := i.TotalValue(tc.names)
			if got != tc.want || !errors.Is(err, tc.wantErr) {
				t.Fatalf("total=%d, %v; want %d, %v", got, err, tc.want, tc.wantErr)
			}
		})
	}
	i := New()
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
	}{{"x", 1, ErrOverflow}, {"x", -1, ErrInvalid}, {"missing", -1, ErrNotFound}} {
		if err := i.Restock(tc.name, tc.amount); !errors.Is(err, tc.want) {
			t.Fatalf("Restock=%v, want %v", err, tc.want)
		}
	}
	if item, _ := i.Get("x"); item.Quantity != math.MaxInt64 {
		t.Fatalf("failed restock mutated item: %+v", item)
	}
	for _, item := range []Item{{"", 1, 1}, {"x", -1, 1}, {"x", 1, -1}} {
		if err := i.Add(item.Name, item.Quantity, item.PriceCents); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Add=%v", err)
		}
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory.csv")
	if i, err := Load(path); i != nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file: %v, %v", i, err)
	}
	for _, tc := range []struct {
		name, content string
		wantErr       error
	}{
		{"valid", "\n apple , 2 , 50 \n\t\n apple,3,60\npear,0,0\n", nil},
		{"columns", "apple,2", ErrInvalid},
		{"name", ",2,3", ErrInvalid},
		{"negative quantity", "apple,-1,3", ErrInvalid},
		{"negative price", "apple,1,-3", ErrInvalid},
		{"quantity syntax", "apple,no,3", strconv.ErrSyntax},
		{"price syntax", "apple,1,no", strconv.ErrSyntax},
		{"quantity range", "apple,9223372036854775808,3", strconv.ErrRange},
		{"price range", "apple,1,9223372036854775808", strconv.ErrRange},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			i, err := Load(path)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Load error=%v, want %v", err, tc.wantErr)
			}
			if err != nil {
				if i != nil {
					t.Fatal("partial inventory returned")
				}
				return
			}
			if got := i.Report(false); got != "apple: 3\npear: 0\n" {
				t.Fatalf("loaded report=%q", got)
			}
		})
	}
}
