package semver_test

import (
	"fmt"
	"strings"
	"testing"

	"example.com/semver"
)

func TestParseRoundTrip(t *testing.T) {
	for _, s := range []string{"0.0.0", "1.2.3", "18446744073709551615.18446744073709551615.18446744073709551615", "1.2.3-alpha.1", "1.2.3-0.0a.-.a-b+001.foo-bar", "1.2.3+001", "1.2.3-" + strings.Repeat("9", 1000)} {
		t.Run(s, func(t *testing.T) {
			v, err := semver.Parse(s)
			if err != nil {
				t.Fatal(err)
			}
			if got := v.String(); got != s {
				t.Fatalf("String() = %q, want %q", got, s)
			}
		})
	}
	var zero semver.Version
	if zero.String() != "0.0.0" {
		t.Fatalf("zero String() = %q", zero.String())
	}
}

func TestParseInvalid(t *testing.T) {
	inputs := []string{"", "1", "1.2", "1.2.3.4", "v1.2.3", " 1.2.3", "1.2.3\n", "01.2.3", "1.02.3", "1.2.03", "-1.2.3", "+1.2.3", "1..3", "1.2.x", "1.2.3-", "1.2.3+", "1.2.3-01", "1.2.3-a.00", "1.2.3-a..b", "1.2.3-.a", "1.2.3-a.", "1.2.3+a..b", "1.2.3+a.", "1.2.3+a+b", "1.2.3-a_b", "1.2.3+é", "1.2.3-α", "1.2.3-١", "1.2.3\x00"}
	for i := 0; i < 3; i++ {
		parts := []string{"1", "2", "3"}
		parts[i] = "18446744073709551616"
		inputs = append(inputs, strings.Join(parts, "."))
	}
	for _, s := range inputs {
		t.Run(s, func(t *testing.T) {
			v, err := semver.Parse(s)
			if err == nil {
				t.Fatalf("Parse(%q) = %v, want error", s, v)
			}
			if v != (semver.Version{}) {
				t.Fatalf("Parse(%q) returned nonzero value on error", s)
			}
		})
	}
}

func mustParse(t *testing.T, s string) semver.Version {
	t.Helper()
	v, err := semver.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestCompare(t *testing.T) {
	ordered := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0", "18446744073709551615.0.0"}
	for i, a := range ordered {
		for j, b := range ordered {
			got := semver.Compare(mustParse(t, a), mustParse(t, b))
			if (i < j && got >= 0) || (i == j && got != 0) || (i > j && got <= 0) {
				t.Errorf("Compare(%q, %q) = %d", a, b, got)
			}
		}
	}
	for _, pair := range [][2]string{{"1.0.0-9", "1.0.0-10"}, {"1.0.0-" + strings.Repeat("9", 1000), "1.0.0-1" + strings.Repeat("0", 1000)}, {"1.0.0-123456789012345678901", "1.0.0-123456789012345678902"}, {"1.0.0-999999999999999999999999", "1.0.0--"}, {"1.0.0-A", "1.0.0-a"}} {
		a, b := mustParse(t, pair[0]), mustParse(t, pair[1])
		if semver.Compare(a, b) >= 0 || semver.Compare(b, a) <= 0 {
			t.Errorf("expected %q < %q", pair[0], pair[1])
		}
	}
	for _, pair := range [][2]string{{"1.2.3+a", "1.2.3+b"}, {"1.2.3", "1.2.3+001"}, {"1.2.3-alpha+x", "1.2.3-alpha+y"}} {
		if got := semver.Compare(mustParse(t, pair[0]), mustParse(t, pair[1])); got != 0 {
			t.Errorf("Compare(%q, %q) = %d, want 0", pair[0], pair[1], got)
		}
	}
}

func Example() {
	a, err := semver.Parse("1.2.3-rc.1+build.42")
	if err != nil {
		panic(err)
	}
	b, err := semver.Parse("1.2.3")
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	fmt.Println(semver.Compare(a, b) < 0)
	// Output:
	// 1.2.3-rc.1+build.42
	// true
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"1.2.3", "0.0.0-0+a", "1.0.0-alpha.1", "1.2.3-01", ""} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := semver.Parse(s)
		if err != nil {
			return
		}
		if v.String() != s {
			t.Fatalf("round trip = %q, want %q", v.String(), s)
		}
		if semver.Compare(v, v) != 0 {
			t.Fatal("version differs from itself")
		}
	})
}
