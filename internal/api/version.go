package api

import (
	"fmt"
	"strconv"
	"strings"
)

// Version identifies a version of the Slurm REST API data_parser, with format
// v<major>.<minor>.<micro> (e.g. v0.0.44).
//
// It is used both to build request paths and to version-gate fields and
// features that only exist as of a specific version.
type Version struct {
	Major int
	Minor int
	Micro int
}

// ParseVersion interprets strings such as "v0.0.45" or "0.0.45".
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format %q (expected vX.Y.Z)", s)
	}

	var v Version
	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil {
		return Version{}, fmt.Errorf("invalid major in %q: %w", s, err)
	}
	if v.Minor, err = strconv.Atoi(parts[1]); err != nil {
		return Version{}, fmt.Errorf("invalid minor in %q: %w", s, err)
	}
	if v.Micro, err = strconv.Atoi(parts[2]); err != nil {
		return Version{}, fmt.Errorf("invalid micro in %q: %w", s, err)
	}
	return v, nil
}

// String returns the canonical "vX.Y.Z" representation.
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Micro)
}

// Compare returns -1, 0 or 1 depending on whether v is less than, equal to or
// greater than other.
func (v Version) Compare(other Version) int {
	switch {
	case v.Major != other.Major:
		return cmpInt(v.Major, other.Major)
	case v.Minor != other.Minor:
		return cmpInt(v.Minor, other.Minor)
	default:
		return cmpInt(v.Micro, other.Micro)
	}
}

// AtLeast reports whether v is greater than or equal to other. It is the base
// primitive for version-gating: a field is included only if the cluster
// version satisfies it.
func (v Version) AtLeast(other Version) bool {
	return v.Compare(other) >= 0
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// SupportedVersions are the versions srest knows how to consume, ordered from
// highest to lowest. They are used for auto-detection.
var SupportedVersions = []Version{
	{0, 0, 45},
	{0, 0, 44},
	{0, 0, 43},
	{0, 0, 42},
	{0, 0, 41},
	{0, 0, 40},
}
