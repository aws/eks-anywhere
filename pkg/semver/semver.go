package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const semverRegex = `^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

var semverRegexp = regexp.MustCompile(semverRegex)

type Version struct {
	Major, Minor, Patch       int64
	Prerelease, Buildmetadata string
}

func New(version string) (*Version, error) {
	matches := semverRegexp.FindStringSubmatch(version)
	namedGroups := make(map[string]string, len(matches))
	groupNames := semverRegexp.SubexpNames()
	for i, value := range matches {
		name := groupNames[i]
		if name != "" {
			namedGroups[name] = value
		}
	}

	v := &Version{}
	var err error

	v.Major, err = strconv.ParseInt(namedGroups["major"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid major version in semver %s: %v", version, err)
	}
	v.Minor, err = strconv.ParseInt(namedGroups["minor"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid minor version in semver %s: %v", version, err)
	}
	v.Patch, err = strconv.ParseInt(namedGroups["patch"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid patch version in semver %s: %v", version, err)
	}

	v.Prerelease = namedGroups["prerelease"]
	v.Buildmetadata = namedGroups["buildmetadata"]

	return v, nil
}

func (v *Version) SameMajor(v2 *Version) bool {
	return v.Major == v2.Major
}

func (v *Version) SameMinor(v2 *Version) bool {
	return v.SameMajor(v2) && v.Minor == v2.Minor
}

func (v *Version) SamePatch(v2 *Version) bool {
	return v.SameMinor(v2) && v.Patch == v2.Patch
}

func (v *Version) SamePrerelease(v2 *Version) bool {
	return v.SamePatch(v2) && v.Prerelease == v2.Prerelease
}

func (v *Version) Equal(v2 *Version) bool {
	return v.SamePrerelease(v2) && v.Buildmetadata == v2.Buildmetadata
}

func (v *Version) GreaterThan(v2 *Version) bool {
	return v.Compare(v2) == 1
}

func (v *Version) LessThan(v2 *Version) bool {
	return v.Compare(v2) == -1
}

func (v *Version) Compare(v2 *Version) int {
	if c := compare(v.Major, v2.Major); c != 0 {
		return c
	}
	if c := compare(v.Minor, v2.Minor); c != 0 {
		return c
	}
	if c := compare(v.Patch, v2.Patch); c != 0 {
		return c
	}
	return 0
}

// CompareBuildMetadata compares the build metadata of v and v2.
// The metadata is split in its identifiers and these compared one by one.
// Number identifiers are considered lower than strings.
// If one build metadata is a prefix of the other, the longer one is considered greater.
// -1 == v is less than v2.
// 0 == v is equal to v2.
// 1 == v is greater than v2.
// 2 == v is different than v2 (it is not possible to identify if lower or greater).
func (v *Version) CompareBuildMetadata(v2 *Version) int {
	return v.buildIdentifiers().compare(v2.buildIdentifiers())
}

func (v *Version) buildIdentifiers() identifiers {
	return newIdentifiers(strings.Split(v.Buildmetadata, "."))
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func compare(i, i2 int64) int {
	if i > i2 {
		return 1
	} else if i < i2 {
		return -1
	}
	return 0
}
