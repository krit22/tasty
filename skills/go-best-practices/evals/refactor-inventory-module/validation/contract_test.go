package inventory_test

import (
	"encoding/json"
	"errors"
	i "example.com/inventory"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTastyExistingAPI(t *testing.T) {
	v := i.Item{Name: "apple", Quantity: 2, PriceCents: 50}
	data, err := json.Marshal(v)
	if err != nil || string(data) != "{\"name\":\"apple\",\"quantity\":2,\"price_cents\":50}" {
		t.Fatalf("wire shape %s err=%v", data, err)
	}
	inv := i.New()
	if err := inv.Add("apple", 2, 50); err != nil {
		t.Fatal(err)
	}
	if got, ok := inv.Get("apple"); !ok || got != v {
		t.Fatalf("Get()=%+v,%v", got, ok)
	}
	if err := inv.Add("apple", 3, 60); err != nil {
		t.Fatal(err)
	}
	if got, _ := inv.Get("apple"); got.Quantity != 3 {
		t.Fatal("duplicate Add must replace")
	}
	if got, _ := inv.TotalValue([]string{"apple", "apple"}); got != 360 {
		t.Fatalf("repeated requested name total=%d", got)
	}
}

func TestTastyMissingName(t *testing.T) {
	inv := i.New()
	if _, err := inv.TotalValue([]string{"unknown"}); !errors.Is(err, i.ErrNotFound) {
		t.Fatalf("missing total error=%v", err)
	}
	if err := inv.Restock("unknown", 1); !errors.Is(err, i.ErrNotFound) {
		t.Fatalf("missing restock error=%v", err)
	}
}

func TestTastyFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	_, err := i.Load(path)
	if !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), path) {
		t.Fatalf("file error must retain cause/path: %v", err)
	}
}

func TestTastyNumericError(t *testing.T) {
	for _, text := range []string{"apple,nope,50\n", "apple,2,nope\n", "apple,9223372036854775808,1\n"} {
		path := filepath.Join(t.TempDir(), "items.csv")
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := i.Load(path); err == nil {
			t.Fatalf("accepted invalid number: %q", text)
		}
	}
}

func TestTastyOverflow(t *testing.T) {
	inv := i.New()
	if err := inv.Add("huge", math.MaxInt64, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := inv.TotalValue([]string{"huge"}); !errors.Is(err, i.ErrOverflow) {
		t.Fatalf("multiplication overflow error=%v", err)
	}
	if err := inv.Restock("huge", 1); !errors.Is(err, i.ErrOverflow) {
		t.Fatalf("quantity overflow error=%v", err)
	}
	if item, _ := inv.Get("huge"); item.Quantity != math.MaxInt64 {
		t.Fatal("failed restock mutated quantity")
	}
	if err := inv.Add("huge", math.MaxInt64, 1); err != nil {
		t.Fatal(err)
	}
	if err := inv.Add("small", 1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := inv.TotalValue([]string{"huge", "small"}); !errors.Is(err, i.ErrOverflow) {
		t.Fatalf("sum overflow error=%v", err)
	}
}

func TestTastyDeterministicReport(t *testing.T) {
	inv := i.New()
	for _, name := range []string{"zebra", "apple", "pear"} {
		if err := inv.Add(name, 2, 50); err != nil {
			t.Fatal(err)
		}
	}
	for n := 0; n < 10; n++ {
		if got := inv.Report(false); got != "apple: 2\npear: 2\nzebra: 2\n" {
			t.Fatalf("report=%q", got)
		}
	}
	if got := inv.Report(true); got != "apple: 2\n  price=50\npear: 2\n  price=50\nzebra: 2\n  price=50\n" {
		t.Fatalf("verbose report=%q", got)
	}
}

func TestTastyPreservedValidation(t *testing.T) {
	inv := i.New()
	for _, input := range []struct {
		name       string
		qty, price int64
	}{{"", 1, 1}, {"a", -1, 1}, {"a", 1, -1}} {
		if err := inv.Add(input.name, input.qty, input.price); !errors.Is(err, i.ErrInvalid) {
			t.Fatalf("invalid Add error=%v", err)
		}
	}
	if err := inv.Add("a", 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := inv.Restock("a", -1); !errors.Is(err, i.ErrInvalid) {
		t.Fatalf("negative restock error=%v", err)
	}
}
