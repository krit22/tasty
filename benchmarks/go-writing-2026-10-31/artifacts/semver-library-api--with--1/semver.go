// Package semver parses semantic versions and compares them using SemVer 2.0.0
// precedence. Parsing is strict: versions must contain major, minor, and patch
// numbers and must not have surrounding whitespace or a leading "v".
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is an immutable semantic version. Its zero value represents 0.0.0.
// Use Parse to construct other versions. Build metadata is preserved by String
// but ignored by Compare; Go's == operator does not implement precedence equality.
type Version struct {
	major, minor, patch uint64
	prerelease          string
	build               string
}

// Parse parses a complete SemVer 2.0.0 version. Core numbers must fit uint64;
// numeric prerelease identifiers have no size limit. On failure it returns the
// zero Version and an error.
func Parse(text string) (Version, error) {
	var v Version
	core, build, hasBuild := strings.Cut(text, "+")
	if hasBuild {
		if !validIdentifiers(build, false) {
			return Version{}, fmt.Errorf("semver: invalid build metadata in %q", text)
		}
		v.build = build
	}
	core, pre, hasPre := strings.Cut(core, "-")
	if hasPre {
		if !validIdentifiers(pre, true) {
			return Version{}, fmt.Errorf("semver: invalid prerelease in %q", text)
		}
		v.prerelease = pre
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("semver: expected major.minor.patch in %q", text)
	}
	numbers := [3]uint64{}
	for i, part := range parts {
		if !numeric(part) || (len(part) > 1 && part[0] == '0') {
			return Version{}, fmt.Errorf("semver: invalid core number %q in %q", part, text)
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("semver: core number overflows uint64 in %q", text)
		}
		numbers[i] = n
	}
	v.major, v.minor, v.patch = numbers[0], numbers[1], numbers[2]
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
	for {
		id, rest, more := strings.Cut(s, ".")
		if id == "" {
			return false
		}
		for i := 0; i < len(id); i++ {
			c := id[i]
			if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '-') {
				return false
			}
		}
		if prerelease && len(id) > 1 && id[0] == '0' && numeric(id) {
			return false
		}
		if !more {
			return true
		}
		s = rest
	}
}

// String returns the complete version, including prerelease and build metadata.
// For a successfully parsed version, it returns exactly the input to Parse.
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

// Compare returns a negative number if a precedes b, zero if they have equal
// precedence, or a positive number if a follows b. Build metadata is ignored.
// Numeric prerelease identifiers precede nonnumeric ones; nonnumeric identifiers
// are ordered by ASCII bytes. A release follows all prereleases of the same core.
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
	left, right := a.prerelease, b.prerelease
	for {
		x, xr, xm := strings.Cut(left, ".")
		y, yr, ym := strings.Cut(right, ".")
		if x != y {
			xn, yn := numeric(x), numeric(y)
			switch {
			case xn && !yn:
				return -1
			case !xn && yn:
				return 1
			case xn && yn:
				// Valid numeric identifiers have no leading zeros, so length decides
				// magnitude before lexical comparison, without any integer conversion.
				if len(x) < len(y) {
					return -1
				}
				if len(x) > len(y) {
					return 1
				}
			}
			return strings.Compare(x, y)
		}
		if !xm && !ym {
			return 0
		}
		if !xm {
			return -1
		}
		if !ym {
			return 1
		}
		left, right = xr, yr
	}
}
