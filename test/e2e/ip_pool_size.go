package e2e

import (
	_ "embed"
	"fmt"
	"regexp"

	"sigs.k8s.io/yaml"
)

//go:embed IP_POOL_SIZE.yaml
var ipPoolSizeFile []byte

const defaultIPPoolSize = 1

// IPPoolRequirement maps a compiled regex pattern to the number of IPs needed.
type IPPoolRequirement struct {
	re         *regexp.Regexp
	ipPoolSize int
}

// ipPoolEntry is used for YAML unmarshalling.
type ipPoolEntry struct {
	Pattern    string `json:"pattern"`
	IPPoolSize int    `json:"ipPoolSize"`
}

// LoadIPPoolRequirements parses and compiles the embedded IP pool size YAML config.
func LoadIPPoolRequirements() ([]IPPoolRequirement, error) {
	var entries []ipPoolEntry
	err := yaml.Unmarshal(ipPoolSizeFile, &entries)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal IP pool size yaml: %v", err)
	}

	requirements := make([]IPPoolRequirement, 0, len(entries))
	for _, entry := range entries {
		re, err := regexp.Compile(entry.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern %q: %v", entry.Pattern, err)
		}
		requirements = append(requirements, IPPoolRequirement{
			re:         re,
			ipPoolSize: entry.IPPoolSize,
		})
	}
	return requirements, nil
}

// GetIPPoolSize returns the IP pool size for a given test name.
// It checks the test name against each pattern in order and returns
// the IP pool size for the first match. If no pattern matches, it returns
// defaultIPPoolSize (1).
func GetIPPoolSize(testName string, requirements []IPPoolRequirement) int {
	for _, req := range requirements {
		if req.re.MatchString(testName) {
			return req.ipPoolSize
		}
	}
	return defaultIPPoolSize
}
