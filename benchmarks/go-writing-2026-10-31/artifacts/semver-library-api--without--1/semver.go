// Package semver parses semantic versions and compares them using SemVer 2.0.0
// precedence. Parsing is strict: a version must have three core components and
// may not have a prefix, surrounding whitespace, or leading zeroes in numeric
// core or prerelease identifiers. Build metadata is preserved but ignored when
// comparing versions.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a semantic version. Its zero value represents 0.0.0.
// Prerelease and Build contain dot-separated identifiers without the leading
// '-' or '+'. Values returned by Parse are valid; callers constructing or
// modifying a Version must follow the same SemVer syntax rules.
type Version struct {
	Major, Minor, Patch uint64
	Prerelease          string
	Build               string
}

// Parse parses a complete SemVer 2.0.0 version. Core components must fit uint64;
// numeric prerelease identifiers have no size limit. On failure it returns the
// zero Version and a non-nil error.
func Parse(text string) (Version, error) {
	invalid := func(reason string) (Version, error) {
		return Version{}, fmt.Errorf("semver: invalid version %q: %s", text, reason)
	}
	core := text
	var v Version
	if i := strings.IndexByte(core, '+'); i >= 0 {
		v.Build = core[i+1:]
		core = core[:i]
		if !validIdentifiers(v.Build, false) {
			return invalid("invalid build metadata")
		}
	}
	if i := strings.IndexByte(core, '-'); i >= 0 {
		v.Prerelease = core[i+1:]
		core = core[:i]
		if !validIdentifiers(v.Prerelease, true) {
			return invalid("invalid prerelease")
		}
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return invalid("expected major.minor.patch")
	}
	components := []*uint64{&v.Major, &v.Minor, &v.Patch}
	for i, part := range parts {
		if !numeric(part) || (len(part) > 1 && part[0] == '0') {
			return invalid("invalid core component")
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return invalid("core component exceeds uint64")
		}
		*components[i] = n
	}
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

// String returns the version including its prerelease and build metadata.
// For a Version returned by Parse, it reproduces the original input exactly.
func (v Version) String() string {
	s := strconv.FormatUint(v.Major, 10) + "." + strconv.FormatUint(v.Minor, 10) + "." + strconv.FormatUint(v.Patch, 10)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare returns -1, 0, or +1 according to whether a precedes, has equal
// precedence to, or follows b. Both versions must be valid. Build metadata is
// ignored, and a release follows all prereleases with the same core version.
func Compare(a, b Version) int {
	for i, x := range [...]uint64{a.Major, a.Minor, a.Patch} {
		y := [...]uint64{b.Major, b.Minor, b.Patch}[i]
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	if a.Prerelease == b.Prerelease {
		return 0
	}
	if a.Prerelease == "" {
		return 1
	}
	if b.Prerelease == "" {
		return -1
	}
	xs, ys := strings.Split(a.Prerelease, "."), strings.Split(b.Prerelease, ".")
	for i := 0; i < len(xs) && i < len(ys); i++ {
		x, y := xs[i], ys[i]
		if x == y {
			continue
		}
		xn, yn := numeric(x), numeric(y)
		if xn != yn {
			if xn {
				return -1
			}
			return 1
		}
		// Valid numeric identifiers have no leading zeroes, so length then
		// lexical order compares arbitrarily large integers exactly.
		if xn && len(x) != len(y) {
			if len(x) < len(y) {
				return -1
			}
			return 1
		}
		return strings.Compare(x, y)
	}
	if len(xs) < len(ys) {
		return -1
	}
	return 1
}
