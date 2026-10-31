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
		"1.2.3-alpha.0.a-b.Z+001.build-2", "1.2.3--+--", "1.2.3-0alpha", "1.2.3+000",
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
	v, err := semver.Parse("12.34.56-rc.1+build.2")
	want := semver.Version{Major: 12, Minor: 34, Patch: 56, Prerelease: "rc.1", Build: "build.2"}
	if err != nil || v != want {
		t.Fatalf("Parse = %+v, %v; want %+v", v, err, want)
	}
	if got := (semver.Version{}).String(); got != "0.0.0" {
		t.Fatalf("zero String = %q", got)
	}
}

func TestInvalid(t *testing.T) {
	for _, text := range []string{
		"", "1", "1.2", "1.2.3.4", "v1.2.3", " 1.2.3", "1.2.3\n", "+1.2.3",
		"01.2.3", "1.02.3", "1.2.03", "-1.2.3", "1..3", "1.2.x",
		"18446744073709551616.0.0", "0.18446744073709551616.0", "0.0.18446744073709551616",
		"1.2.3-", "1.2.3+", "1.2.3-01", "1.2.3-a.00", "1.2.3-.a", "1.2.3-a.",
		"1.2.3-a..b", "1.2.3+..", "1.2.3+a.", "1.2.3+a+b", "1.2.3-a_b", "1.2.3+a_b",
		"1.2.3-é", "1.2.3+é", "１.2.3", "1.2.3-a\x00", "1.2.3+a b",
	} {
		t.Run(text, func(t *testing.T) {
			v, err := semver.Parse(text)
			if err == nil || v != (semver.Version{}) {
				t.Fatalf("Parse(%q) = %+v, %v; want zero and error", text, v, err)
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
		"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta",
		"1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0",
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
		{"1.0.0-9", "1.0.0-10"},
		{"1.0.0-" + strings.Repeat("9", 100), "1.0.0-1" + strings.Repeat("0", 100)},
		{"1.0.0-" + strings.Repeat("9", 100), "1.0.0-a"},
		{"1.0.0-A", "1.0.0-a"},
		{"1.0.0-123456789012345678901", "1.0.0-123456789012345678902"},
		{"0.0.0", "18446744073709551615.0.0"},
	} {
		a, b := mustParse(t, pair[0]), mustParse(t, pair[1])
		if semver.Compare(a, b) != -1 || semver.Compare(b, a) != 1 {
			t.Errorf("incorrect ordering for %q < %q", pair[0], pair[1])
		}
	}
	for _, pair := range [][2]string{{"1.2.3+a", "1.2.3+b"}, {"1.2.3", "1.2.3+001"}, {"1.2.3-a+x", "1.2.3-a+y"}} {
		if semver.Compare(mustParse(t, pair[0]), mustParse(t, pair[1])) != 0 {
			t.Errorf("build metadata changed precedence: %v", pair)
		}
	}
}

func Example() {
	a, _ := semver.Parse("1.2.3-rc.1+build.42")
	b, _ := semver.Parse("1.2.3")
	fmt.Println(a)
	fmt.Println(semver.Compare(a, b))
	// Output:
	// 1.2.3-rc.1+build.42
	// -1
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{"1.2.3", "0.0.0-a+001", "1.2.3-999999999999999999999", "", "1.2.3+"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		v, err := semver.Parse(text)
		if err != nil {
			return
		}
		if v.String() != text {
			t.Fatalf("round trip: %q != %q", v.String(), text)
		}
		if semver.Compare(v, v) != 0 {
			t.Fatal("version is not equal to itself")
		}
	})
}
