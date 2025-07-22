package framework

import (
	_ "embed"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

//go:embed testdata/tinkerbell/custom_config.yaml
var customTinkerbellConfigYAML []byte

// GetCustomTinkerbellConfig returns a custom TinkerbellTemplateConfig.
func GetCustomTinkerbellConfig(tinkerbellLBIP string) (*anywherev1.TinkerbellTemplateConfig, error) {
	// Replace placeholders with actual values using string replacement
	configContent := string(customTinkerbellConfigYAML)
	localIP, err := networkutils.GetLocalIP()
	if err != nil {
		return nil, err
	}
	configContent = strings.ReplaceAll(configContent, "__TINKERBELL_LOCAL_IP__", localIP.String())
	configContent = strings.ReplaceAll(configContent, "__TINKERBELL_LB_IP__", tinkerbellLBIP)

	// Parse the YAML into TinkerbellTemplateConfig
	var config anywherev1.TinkerbellTemplateConfig
	if err := yaml.Unmarshal([]byte(configContent), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template config: %v", err)
	}

	return &config, nil
}
