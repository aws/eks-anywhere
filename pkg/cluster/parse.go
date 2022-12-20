package cluster

import (
	"fmt"
	"os"
)

// ParseConfig reads yaml file with at least one Cluster object and generates the corresponding Config
// using the default package config manager.
func ParseConfigFromFile(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cluster config file: %v", err)
	}

	return ParseConfig(content)
}

// ParseConfig reads yaml manifest with at least one Cluster object and generates the corresponding Config
// using the default package config manager.
func ParseConfig(yamlManifest []byte) (*Config, error) {
	return manager().Parse(yamlManifest)
}
