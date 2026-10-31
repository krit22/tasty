package semver

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseRoundTrip(t *testing.T) {
	for _, input := range []string{
		"0.0.0", "1.2.3", "18446744073709551615.18446744073709551615.18446744073709551615",
		"1.2.3-alpha.0.a-1+001.build-7", "1.2.3--+--", "1.2.3+000.01",
		"1.2.3-" + strings.Repeat("9", 1000),
	} {
		t.Run(input, func(t *testing.T) {
			v, err := Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			if got := v.String(); got != input {
				t.Fatalf("String() = %q, want %q", got, input)
			}
		})
	}
	v, err := Parse("12.34.56-rc.1+build.7")
	if err != nil || v != (Version{Major: 12, Minor: 34, Patch: 56, Prerelease: "rc.1", Build: "build.7"}) {
		t.Fatalf("unexpected fields: %+v, %v", v, err)
	}
	if (Version{}).String() != "0.0.0" {
		t.Fatal("invalid zero value")
	}
}

func TestInvalid(t *testing.T) {
	for _, input := range []string{
		"", "1", "1.2", "1.2.3.4", "v1.2.3", " 1.2.3", "1.2.3\n",
		"01.2.3", "1.02.3", "1.2.03", "+1.2.3", "-1.2.3", "1..3",
		"18446744073709551616.0.0", "0.18446744073709551616.0", "0.0.18446744073709551616",
		"1.2.3-", "1.2.3+", "1.2.3-01", "1.2.3-a.00", "1.2.3-a..b",
		"1.2.3-.a", "1.2.3-a.", "1.2.3+foo..bar", "1.2.3+a+b",
		"1.2.3-a_b", "1.2.3+é", "１.2.3", "1.2.3-a\x00", "1.2.3+foo/bar",
	} {
		t.Run(input, func(t *testing.T) {
			v, err := Parse(input)
			if err == nil || v != (Version{}) {
				t.Fatalf("Parse(%q) = %+v, %v; want zero value and error", input, v, err)
			}
		})
	}
}

func TestPrecedence(t *testing.T) {
	ordered := []string{
		"0.9.9", "1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta",
		"1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1",
		"1.0.0", "1.0.1", "1.1.0", "2.0.0", "18446744073709551615.0.0",
	}
	for i, a := range ordered {
		for j, b := range ordered {
			va, _ := Parse(a)
			vb, _ := Parse(b)
			want := 0
			if i < j {
				want = -1
			}
			if i > j {
				want = 1
			}
			if got := Compare(va, vb); got != want {
				t.Errorf("Compare(%s, %s) = %d, want %d", a, b, got, want)
			}
		}
	}
	for _, pair := range [][2]string{
		{"1.0.0-99999999999999999999", "1.0.0-100000000000000000000"},
		{"1.0.0-100000000000000000001", "1.0.0-100000000000000000002"},
		{"1.0.0-999999999999999999999999", "1.0.0-A"},
		{"1.0.0-A", "1.0.0-a"}, {"1.0.0-a", "1.0.0-a.0"},
	} {
		a, _ := Parse(pair[0])
		b, _ := Parse(pair[1])
		if Compare(a, b) != -1 || Compare(b, a) != 1 {
			t.Errorf("incorrect ordering: %v", pair)
		}
	}
}

func TestBuildIgnored(t *testing.T) {
	for _, base := range []string{"1.2.3", "1.2.3-rc.1"} {
		a, _ := Parse(base + "+abc")
		for _, suffix := range []string{"", "+def", "+001"} {
			b, _ := Parse(base + suffix)
			if Compare(a, b) != 0 || Compare(b, a) != 0 {
				t.Fatal("build metadata changed precedence")
			}
		}
	}
}

func Example() {
	a, err := Parse("1.2.3-rc.1+build.7")
	if err != nil {
		panic(err)
	}
	b, err := Parse("1.2.3")
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
	fmt.Println(Compare(a, b))
	// Output:
	// 1.2.3-rc.1+build.7
	// -1
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{"1.2.3", "1.0.0-alpha.1+build", "", "1.2.3-01"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		v, err := Parse(input)
		if err != nil {
			return
		}
		if v.String() != input {
			t.Fatalf("round trip changed %q to %q", input, v.String())
		}
		if Compare(v, v) != 0 {
			t.Fatal("version does not equal itself")
		}
	})
}
