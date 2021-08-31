package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

const semverRegex = `^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

var semverRegexp = regexp.MustCompile(semverRegex)

type Version struct {
	Major, Minor, Patch       uint64
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

	v.Major, err = strconv.ParseUint(namedGroups["major"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid major version in semver %s: %v", version, err)
	}
	v.Minor, err = strconv.ParseUint(namedGroups["minor"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid minor version in semver %s: %v", version, err)
	}
	v.Patch, err = strconv.ParseUint(namedGroups["patch"], 10, 64)
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
