// Package semver parses semantic versions and compares them using SemVer 2.0.0
// precedence. Parsing is strict: versions require three unsigned decimal core
// components, without leading zeros, and do not accept whitespace or a v prefix.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is an immutable semantic version. Its zero value represents 0.0.0.
// Use Parse to construct other versions. Versions may be copied freely.
// Go's == operator compares the entire value, including build metadata; use
// Compare to test equality by precedence.
type Version struct {
	major, minor, patch uint64
	prerelease, build   string
}

// Parse parses a complete SemVer 2.0.0 version. Core components must fit uint64;
// numeric prerelease identifiers have no size limit. On failure, Parse returns
// the zero Version and an error describing the invalid input.
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
		if !numeric(part) || (len(part) > 1 && part[0] == '0') {
			return Version{}, fmt.Errorf("semver: invalid core component %q in %q", part, text)
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("semver: core component %q exceeds uint64 in %q", part, text)
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
		if id == "" {
			return false
		}
		for i := 0; i < len(id); i++ {
			c := id[i]
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
				return false
			}
		}
		if prerelease && len(id) > 1 && id[0] == '0' && numeric(id) {
			return false
		}
	}
	return true
}

// String returns the complete version, including prerelease and build metadata.
// For a successfully parsed version, it reproduces the input exactly.
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
// precedence, or a positive value if a follows b. Build metadata is ignored.
// A release follows all prereleases with the same major, minor, and patch.
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
		if xn && yn && len(x) != len(y) {
			// Valid numeric identifiers have no leading zeros, so length
			// determines order without converting arbitrarily large numbers.
			if len(x) < len(y) {
				return -1
			}
			return 1
		}
		if xn != yn {
			if xn {
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
