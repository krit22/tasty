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
		"1.2.3-alpha.1", "1.2.3-0.A-z.0a.-+001.build-9", "1.2.3+000.abc-def",
		"1.2.3-" + strings.Repeat("9", 1000),
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err != nil {
				t.Fatal(err)
			}
			if got := v.String(); got != text {
				t.Fatalf("String() = %q, want %q", got, text)
			}
		})
	}
	v, err := semver.Parse("12.34.56-rc.1+sha.001")
	if err != nil || v != (semver.Version{Major: 12, Minor: 34, Patch: 56, Prerelease: "rc.1", Build: "sha.001"}) {
		t.Fatalf("unexpected fields: %+v, %v", v, err)
	}
	if got := (semver.Version{}).String(); got != "0.0.0" {
		t.Fatalf("zero String() = %q", got)
	}
}

func TestParseInvalid(t *testing.T) {
	for _, text := range []string{
		"", "1", "1.2", "1.2.3.4", "v1.2.3", " 1.2.3", "1.2.3\n",
		"01.2.3", "1.02.3", "1.2.03", "+1.2.3", "-1.2.3", "1.-2.3", "1.2.-3",
		".2.3", "1..3", "1.2.", "a.2.3", "1.2.٣",
		"18446744073709551616.0.0", "0.18446744073709551616.0", "0.0.18446744073709551616",
		"1.2.3-", "1.2.3+", "1.2.3-+build", "1.2.3-alpha..1", "1.2.3-.a", "1.2.3-a.",
		"1.2.3-01", "1.2.3-alpha.00", "1.2.3-a_b", "1.2.3-α", "1.2.3-a b",
		"1.2.3+build..1", "1.2.3+.a", "1.2.3+a.", "1.2.3+a+b", "1.2.3+a_b", "1.2.3+a\x00",
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err == nil || v != (semver.Version{}) {
				t.Fatalf("Parse(%q) = %+v, %v; want zero and error", text, v, err)
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
	ordered := []string{
		"0.9.9", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta",
		"1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1",
		"1.0.0", "1.0.1", "1.1.0", "2.0.0", "18446744073709551615.0.0",
	}
	for i, a := range ordered {
		for j, b := range ordered {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := semver.Compare(mustParse(t, a), mustParse(t, b)); got != want {
				t.Errorf("Compare(%q, %q) = %d, want %d", a, b, got, want)
			}
		}
	}
	for _, pair := range [][2]string{
		{"1.0.0-99999999999999999999", "1.0.0-100000000000000000000"},
		{"1.0.0-100000000000000000000", "1.0.0-100000000000000000001"},
		{"1.0.0-" + strings.Repeat("9", 1000), "1.0.0-a"},
		{"1.0.0-Z", "1.0.0-a"}, {"1.0.0-a", "1.0.0-a.0"},
	} {
		a, b := mustParse(t, pair[0]), mustParse(t, pair[1])
		if semver.Compare(a, b) != -1 || semver.Compare(b, a) != 1 {
			t.Errorf("wrong ordering: %q < %q", pair[0], pair[1])
		}
	}
	for _, base := range []string{"1.2.3", "1.2.3-alpha.1"} {
		if semver.Compare(mustParse(t, base+"+abc"), mustParse(t, base+"+001")) != 0 ||
			semver.Compare(mustParse(t, base), mustParse(t, base+"+abc")) != 0 {
			t.Errorf("build metadata affected precedence for %s", base)
		}
	}
}

func Example() {
	a, err := semver.Parse("1.2.3-rc.1+build.42")
	if err != nil {
		panic(err)
	}
	b, _ := semver.Parse("1.2.3")
	fmt.Println(a)
	fmt.Println(semver.Compare(a, b))
	// Output:
	// 1.2.3-rc.1+build.42
	// -1
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"0.0.0", "1.2.3-rc.1+build.42", "1.2.3-01", "", "18446744073709551616.0.0"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := semver.Parse(s)
		if err != nil {
			if v != (semver.Version{}) {
				t.Fatal("nonzero result on error")
			}
			return
		}
		if v.String() != s {
			t.Fatalf("round trip changed %q to %q", s, v.String())
		}
		if semver.Compare(v, v) != 0 {
			t.Fatal("comparison is not reflexive")
		}
	})
}
