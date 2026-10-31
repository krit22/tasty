package semver_test

import (
	"fmt"
	"strings"
	"testing"

	"example.com/semver"
)

func TestParse(t *testing.T) {
	for _, text := range []string{
		"0.0.0", "1.2.3", "18446744073709551615.18446744073709551615.18446744073709551615",
		"1.2.3-alpha.0.z-1+001.build-2", "1.0.0--", "1.0.0+000", "1.0.0-0alpha",
		"1.0.0-" + strings.Repeat("9", 1000),
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err != nil {
				t.Fatal(err)
			}
			if got := v.String(); got != text {
				t.Errorf("String() = %q, want %q", got, text)
			}
		})
	}
	if got := (semver.Version{}).String(); got != "0.0.0" {
		t.Errorf("zero String() = %q", got)
	}
}

func TestParseInvalid(t *testing.T) {
	for _, text := range []string{
		"", "1", "1.2", "1.2.3.4", "v1.2.3", " 1.2.3", "1.2.3\n", "1.2.3 ",
		"01.2.3", "1.02.3", "1.2.03", "-1.2.3", "+1.2.3", "1..3", "a.2.3",
		"18446744073709551616.0.0", "0.18446744073709551616.0", "0.0.18446744073709551616",
		"1.2.3-", "1.2.3+", "1.2.3-01", "1.2.3-a.00", "1.2.3-a..b", "1.2.3-.a",
		"1.2.3-a.", "1.2.3+foo..bar", "1.2.3+a+b", "1.2.3-a_b", "1.2.3+é", "1.2.3-１２",
		"1.2.3-a\x00", "1.2.3-+build", "1.2.3+build.",
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err == nil {
				t.Fatalf("Parse(%q) = %v, want error", text, v)
			}
			if v != (semver.Version{}) {
				t.Errorf("Parse(%q) returned nonzero Version on error", text)
			}
		})
	}
}

func mustParse(t *testing.T, text string) semver.Version {
	t.Helper()
	v, err := semver.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestPrecedence(t *testing.T) {
	ordered := []string{
		"0.0.0", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta",
		"1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0",
		"18446744073709551615.0.0",
	}
	for i, a := range ordered {
		for j, b := range ordered {
			got := semver.Compare(mustParse(t, a), mustParse(t, b))
			if i < j && got >= 0 || i == j && got != 0 || i > j && got <= 0 {
				t.Errorf("Compare(%q, %q) = %d; index order %d, %d", a, b, got, i, j)
			}
		}
	}
	for _, pair := range [][2]string{
		{"1.0.0-9", "1.0.0-10"},
		{"1.0.0-18446744073709551616", "1.0.0-18446744073709551617"},
		{"1.0.0-" + strings.Repeat("9", 1000), "1.0.0-A"},
		{"1.0.0-A", "1.0.0-a"},
	} {
		a, b := mustParse(t, pair[0]), mustParse(t, pair[1])
		if semver.Compare(a, b) >= 0 || semver.Compare(b, a) <= 0 {
			t.Errorf("expected %q < %q", pair[0], pair[1])
		}
	}
}

func TestMetadataIgnored(t *testing.T) {
	for _, base := range []string{"1.2.3", "1.2.3-alpha.1"} {
		a := mustParse(t, base)
		b := mustParse(t, base+"+001.foo")
		c := mustParse(t, base+"+other")
		if semver.Compare(a, b) != 0 || semver.Compare(b, c) != 0 {
			t.Errorf("metadata changed precedence for %q", base)
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
	fmt.Println(a.String())
	fmt.Println(semver.Compare(a, b) < 0)
	// Output:
	// 1.2.3-rc.1+build.42
	// true
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{"1.2.3", "0.0.0-a+build", "1.0.0-01", "", "1.0.0--"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		v, err := semver.Parse(text)
		if err != nil {
			return
		}
		if v.String() != text {
			t.Fatalf("round trip = %q, want %q", v.String(), text)
		}
		if semver.Compare(v, v) != 0 {
			t.Fatal("version does not compare equal to itself")
		}
	})
}
