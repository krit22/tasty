// Package semver parses semantic versions and compares them using SemVer 2.0.0
// precedence. Parsing is strict: versions must include major, minor, and patch
// components, with no whitespace or leading "v". Build metadata is preserved by
// String but ignored by Compare.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is an immutable semantic version. Its zero value is 0.0.0.
// Use Parse to construct other versions. Core components fit in uint64;
// numeric prerelease identifiers have no size limit.
type Version struct {
	major, minor, patch uint64
	prerelease, build   string
}

// Parse parses a complete SemVer 2.0.0 version. It rejects malformed versions
// and core components exceeding uint64. On error it returns the zero Version.
func Parse(text string) (Version, error) {
	var v Version
	core, build, hasBuild := strings.Cut(text, "+")
	if hasBuild && !validIdentifiers(build, false) {
		return Version{}, fmt.Errorf("semver: invalid build metadata in %q", text)
	}
	core, pre, hasPre := strings.Cut(core, "-")
	if hasPre && !validIdentifiers(pre, true) {
		return Version{}, fmt.Errorf("semver: invalid prerelease in %q", text)
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("semver: expected major.minor.patch in %q", text)
	}
	components := []*uint64{&v.major, &v.minor, &v.patch}
	for i, part := range parts {
		if !numeric(part) || len(part) > 1 && part[0] == '0' {
			return Version{}, fmt.Errorf("semver: invalid core component %q", part)
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("semver: core component %q exceeds uint64", part)
		}
		*components[i] = n
	}
	v.prerelease, v.build = pre, build
	return v, nil
}

func numeric(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func validIdentifiers(s string, prerelease bool) bool {
	for _, id := range strings.Split(s, ".") {
		if id == "" || prerelease && len(id) > 1 && id[0] == '0' && numeric(id) {
			return false
		}
		for i := 0; i < len(id); i++ {
			c := id[i]
			if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '-') {
				return false
			}
		}
	}
	return true
}

// String returns the version's SemVer representation, including build metadata.
func (v Version) String() string {
	s := strconv.FormatUint(v.major, 10) + "." + strconv.FormatUint(v.minor, 10) + "." + strconv.FormatUint(v.patch, 10)
	if v.prerelease != "" {
		s += "-" + v.prerelease
	}
	if v.build != "" {
		s += "+" + v.build
	}
	return s
}

// Compare returns a negative value if a precedes b, zero if they have equal
// precedence, and a positive value if a follows b. Build metadata is ignored.
func Compare(a, b Version) int {
	for _, pair := range [][2]uint64{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if a.prerelease == b.prerelease {
		return 0
	}
	if a.prerelease == "" {
		return 1
	}
	if b.prerelease == "" {
		return -1
	}
	left, right := strings.Split(a.prerelease, "."), strings.Split(b.prerelease, ".")
	for i := 0; i < len(left) && i < len(right); i++ {
		x, y := left[i], right[i]
		if x == y {
			continue
		}
		xn, yn := numeric(x), numeric(y)
		if xn && !yn {
			return -1
		}
		if !xn && yn {
			return 1
		}
		if xn && len(x) != len(y) {
			if len(x) < len(y) {
				return -1
			}
			return 1
		}
		return strings.Compare(x, y)
	}
	if len(left) < len(right) {
		return -1
	}
	return 1
}
