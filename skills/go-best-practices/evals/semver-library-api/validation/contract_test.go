package semver_test

import (
	semver "example.com/semver"
	"testing"
)

func TestTastyRoundTrip(t *testing.T) {
	for _, text := range []string{"0.0.0", "1.2.3", "1.2.3-alpha.1+build.5", "18446744073709551615.0.0", "1.0.0+001"} {
		v, err := semver.Parse(text)
		if err != nil || v.String() != text {
			t.Fatalf("Parse(%q) round trip = %q, %v", text, v.String(), err)
		}
	}
}

func TestTastyInvalidInput(t *testing.T) {
	for _, text := range []string{"", "1", "1.2", "v1.2.3", "01.2.3", "1.02.3", "1.2.03", "-1.2.3", "1.2.3-", "1.2.3+", "1.2.3-a..b", "1.2.3-01", "1.2.3+a..b", "1.2.3-a_b", "1.2.3-β", " 1.2.3", "1.2.3\n", "1.2.3+build+again"} {
		if _, err := semver.Parse(text); err == nil {
			t.Errorf("Parse(%q) accepted invalid input", text)
		}
	}
}

func TestTastyCoreOverflow(t *testing.T) {
	for _, text := range []string{"18446744073709551616.0.0", "0.18446744073709551616.0", "0.0.18446744073709551616"} {
		if _, err := semver.Parse(text); err == nil {
			t.Errorf("Parse(%q) accepted overflowing core component", text)
		}
	}
}

func TestTastyPrecedence(t *testing.T) {
	sequence := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "2.0.0"}
	for i, a := range sequence {
		va, err := semver.Parse(a)
		if err != nil {
			t.Fatal(err)
		}
		for j, b := range sequence {
			vb, err := semver.Parse(b)
			if err != nil {
				t.Fatal(err)
			}
			got := semver.Compare(va, vb)
			if (i < j && got >= 0) || (i == j && got != 0) || (i > j && got <= 0) {
				t.Errorf("Compare(%q, %q) = %d", a, b, got)
			}
		}
	}
}

func TestTastyBuildMetadata(t *testing.T) {
	a, err := semver.Parse("1.2.3-rc.1+build.1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := semver.Parse("1.2.3-rc.1+build.99")
	if err != nil {
		t.Fatal(err)
	}
	if semver.Compare(a, b) != 0 {
		t.Fatal("build metadata changed precedence")
	}
}

func TestTastyLongPrereleaseNumbers(t *testing.T) {
	a, err := semver.Parse("1.0.0-99999999999999999999999999999999999999")
	if err != nil {
		t.Fatal(err)
	}
	b, err := semver.Parse("1.0.0-100000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if semver.Compare(a, b) >= 0 {
		t.Fatal("numeric prerelease comparison overflowed or compared lexically")
	}
}
