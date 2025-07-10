package framework

import (
	_ "embed"

	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

//go:embed testdata/tinkerbell/custom_config.yaml
var customTinkerbellConfigYAML []byte

// GetCustomTinkerbellConfig is a public wrapper for getCustomTinkerbellConfig.
func GetCustomTinkerbellConfig() (*anywherev1.TinkerbellTemplateConfig, error) {
	// Parse the YAML into a TinkerbellTemplateConfig object
	config := &anywherev1.TinkerbellTemplateConfig{}
	if err := yaml.Unmarshal(customTinkerbellConfigYAML, config); err != nil {
		return nil, err
	}

	return config, nil
}
