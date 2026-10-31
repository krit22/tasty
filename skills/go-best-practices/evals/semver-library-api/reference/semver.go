// Package semver parses versions and compares their SemVer precedence.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed version. Build metadata is retained but ignored by Compare.
type Version struct {
	core  [3]uint64
	pre   []string
	build string
}

// Parse accepts SemVer 2.0.0 text with uint64 core components.
func Parse(text string) (Version, error) {
	var v Version
	main, build, hasBuild := strings.Cut(text, "+")
	if hasBuild {
		if !identifiers(build, false) {
			return v, fmt.Errorf("invalid build metadata: %q", build)
		}
		v.build = build
	}
	core, pre, hasPre := strings.Cut(main, "-")
	if hasPre {
		if !identifiers(pre, true) {
			return v, fmt.Errorf("invalid prerelease: %q", pre)
		}
		v.pre = strings.Split(pre, ".")
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return v, fmt.Errorf("expected three core components: %q", text)
	}
	for i, part := range parts {
		if !numeric(part) || (len(part) > 1 && part[0] == '0') {
			return v, fmt.Errorf("invalid core component: %q", part)
		}
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return v, fmt.Errorf("core component %q: %w", part, err)
		}
		v.core[i] = n
	}
	return v, nil
}

func numeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func identifiers(s string, prerelease bool) bool {
	for _, part := range strings.Split(s, ".") {
		if part == "" {
			return false
		}
		for _, c := range part {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '-') {
				return false
			}
		}
		if prerelease && numeric(part) && len(part) > 1 && part[0] == '0' {
			return false
		}
	}
	return true
}

// String returns the original valid version text.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.core[0], v.core[1], v.core[2])
	if len(v.pre) > 0 {
		s += "-" + strings.Join(v.pre, ".")
	}
	if v.build != "" {
		s += "+" + v.build
	}
	return s
}

// Compare returns a negative value, zero, or a positive value by SemVer precedence.
func Compare(a, b Version) int {
	for i := range a.core {
		if a.core[i] < b.core[i] {
			return -1
		}
		if a.core[i] > b.core[i] {
			return 1
		}
	}
	if len(a.pre) == 0 && len(b.pre) > 0 {
		return 1
	}
	if len(b.pre) == 0 && len(a.pre) > 0 {
		return -1
	}
	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		x, y := a.pre[i], b.pre[i]
		if x == y {
			continue
		}
		nx, ny := numeric(x), numeric(y)
		if nx && !ny {
			return -1
		}
		if !nx && ny {
			return 1
		}
		if nx && ny && len(x) != len(y) {
			if len(x) < len(y) {
				return -1
			}
			return 1
		}
		if x < y {
			return -1
		}
		return 1
	}
	if len(a.pre) < len(b.pre) {
		return -1
	}
	if len(a.pre) > len(b.pre) {
		return 1
	}
	return 0
}
