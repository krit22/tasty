// Package semver parses and compares semantic versions according to SemVer 2.0.0.
// Versions must contain exactly three decimal core components, with optional
// prerelease identifiers and build metadata. Prefixes such as "v" and surrounding
// whitespace are not accepted.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a semantic version. The zero value represents 0.0.0.
// Prerelease and Build contain dot-separated identifiers without the leading
// '-' or '+'. Values returned by Parse are valid; callers constructing Version
// literals are responsible for using valid identifiers.
type Version struct {
	Major, Minor, Patch uint64
	Prerelease          string
	Build               string
}

// Parse parses a complete SemVer 2.0.0 version. Core components must fit uint64;
// numeric prerelease identifiers have no size limit. On error it returns the
// zero Version and an error.
func Parse(text string) (Version, error) {
	var v Version
	core := text
	if i := strings.IndexByte(core, '+'); i >= 0 {
		v.Build = core[i+1:]
		core = core[:i]
		if !validIdentifiers(v.Build, false) {
			return Version{}, fmt.Errorf("semver: invalid build metadata in %q", text)
		}
	}
	if i := strings.IndexByte(core, '-'); i >= 0 {
		v.Prerelease = core[i+1:]
		core = core[:i]
		if !validIdentifiers(v.Prerelease, true) {
			return Version{}, fmt.Errorf("semver: invalid prerelease in %q", text)
		}
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("semver: expected major.minor.patch in %q", text)
	}
	values := []*uint64{&v.Major, &v.Minor, &v.Patch}
	for i, part := range parts {
		if !numeric(part) || (len(part) > 1 && part[0] == '0') {
			return Version{}, fmt.Errorf("semver: invalid core component %q", part)
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("semver: core component %q: %w", part, err)
		}
		*values[i] = n
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

// String returns the version's SemVer representation, including build metadata.
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

// Compare returns -1, 0, or +1 according to the SemVer precedence of a and b.
// Build metadata is ignored. A release has higher precedence than a prerelease
// with the same core version. Both arguments must be valid versions.
func Compare(a, b Version) int {
	for _, pair := range [][2]uint64{{a.Major, b.Major}, {a.Minor, b.Minor}, {a.Patch, b.Patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
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
	as, bs := strings.Split(a.Prerelease, "."), strings.Split(b.Prerelease, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		x, y := as[i], bs[i]
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
		// Valid numeric identifiers have no leading zeroes, so length then
		// lexical order compares integers without imposing an integer limit.
		if xn && yn && len(x) != len(y) {
			if len(x) < len(y) {
				return -1
			}
			return 1
		}
		return strings.Compare(x, y)
	}
	if len(as) < len(bs) {
		return -1
	}
	return 1
}
