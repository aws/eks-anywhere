// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

const semverRegex = `^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)`

var semverRegexp = regexp.MustCompile(semverRegex)

type Version struct {
	Major, Minor, Patch int64
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

	return v, nil
}

func (v *Version) GreaterThan(v2 *Version) bool {
	return v.Compare(v2) == 1
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

func compare(i, i2 int64) int {
	if i > i2 {
		return 1
	} else if i < i2 {
		return -1
	}
	return 0
}
