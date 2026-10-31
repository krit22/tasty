package semver_test

import (
	"fmt"
	"strings"
	"testing"

	"example.com/semver"
)

func TestParseRoundTrip(t *testing.T) {
	for _, text := range []string{
		"0.0.0", "1.2.3", "18446744073709551615.18446744073709551615.18446744073709551615",
		"1.2.3-alpha.0.z-9", "1.2.3+001.build-01", "1.2.3-alpha+build.001",
		"1.2.3--+--", "1.2.3-0A.01a.AZaz09-",
		"1.2.3-" + strings.Repeat("9", 1000),
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err != nil {
				t.Fatalf("Parse(%q): %v", text, err)
			}
			if got := v.String(); got != text {
				t.Errorf("String() = %q, want %q", got, text)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	for _, text := range []string{
		"", "1", "1.2", "1.2.3.4", ".1.2", "1..2", "1.2.",
		"01.2.3", "1.02.3", "1.2.03", "v1.2.3", " 1.2.3", "1.2.3\n",
		"+1.2.3", "-1.2.3", "1.-2.3", "1.2.+3", "1.2.3-", "1.2.3+",
		"1.2.3-a..b", "1.2.3-.a", "1.2.3-a.", "1.2.3-01", "1.2.3-a.00",
		"1.2.3+a..b", "1.2.3+.a", "1.2.3+a.", "1.2.3+a+b", "1.2.3-a+b+c",
		"1.2.3-a_b", "1.2.3+a_b", "1.2.3-é", "1.2.3+é", "１.2.3", "1.2.3-\x00",
		"18446744073709551616.0.0", "0.18446744073709551616.0", "0.0.18446744073709551616",
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err == nil {
				t.Fatalf("Parse(%q) = %v, want error", text, v)
			}
			if v != (semver.Version{}) {
				t.Errorf("Parse(%q) returned nonzero version on failure: %v", text, v)
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

func TestCompare(t *testing.T) {
	ordered := []string{
		"0.0.0", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta",
		"1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1",
		"1.0.0", "1.0.1", "1.1.0", "2.0.0", "18446744073709551615.0.0",
	}
	for i, a := range ordered {
		for j, b := range ordered {
			got := semver.Compare(mustParse(t, a), mustParse(t, b))
			if i < j && got >= 0 || i == j && got != 0 || i > j && got <= 0 {
				t.Errorf("Compare(%q, %q) = %d; index order %d vs %d", a, b, got, i, j)
			}
		}
	}
	for _, pair := range [][2]string{
		{"1.0.0-9", "1.0.0-10"},
		{"1.0.0-" + strings.Repeat("9", 1000), "1.0.0-1" + strings.Repeat("0", 1000)},
		{"1.0.0-18446744073709551616", "1.0.0-18446744073709551617"},
		{"1.0.0-999", "1.0.0-A"}, {"1.0.0-Z", "1.0.0-a"},
		{"1.0.0-0", "1.0.0-0.0"},
	} {
		a, b := mustParse(t, pair[0]), mustParse(t, pair[1])
		if semver.Compare(a, b) >= 0 || semver.Compare(b, a) <= 0 {
			t.Errorf("expected %q to precede %q", pair[0], pair[1])
		}
	}
}

func TestBuildMetadataAndZeroValue(t *testing.T) {
	for _, base := range []string{"0.0.0", "1.2.3", "1.2.3-alpha.1"} {
		a := mustParse(t, base+"+a")
		for _, suffix := range []string{"", "+b", "+001"} {
			if got := semver.Compare(a, mustParse(t, base+suffix)); got != 0 {
				t.Errorf("build metadata affected precedence: Compare(%q, %q) = %d", a, base+suffix, got)
			}
		}
	}
	var zero semver.Version
	if zero.String() != "0.0.0" || semver.Compare(zero, mustParse(t, "0.0.0")) != 0 {
		t.Errorf("zero Version = %v, want 0.0.0", zero)
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
	for _, seed := range []string{"0.0.0", "1.2.3-alpha+build", "1.2.3-01", "1.2.3+", "18446744073709551616.0.0"} {
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
			t.Fatalf("version does not compare equal to itself: %q", text)
		}
	})
}
