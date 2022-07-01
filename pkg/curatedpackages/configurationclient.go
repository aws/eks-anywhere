package curatedpackages

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

func GetConfigurationsFromBundle(bundlePackage *packagesv1.BundlePackage) map[string]packagesv1.VersionConfiguration {
	configs := make(map[string]packagesv1.VersionConfiguration)
	if bundlePackage == nil || len(bundlePackage.Source.Versions) < 1 {
		return configs
	}
	packageConfigurations := bundlePackage.Source.Versions[0].Configurations

	for _, config := range packageConfigurations {
		configs[config.Name] = config
	}
	return configs
}

func UpdateConfigurations(originalConfigs map[string]packagesv1.VersionConfiguration, newConfigs map[string]string) error {
	for key, val := range newConfigs {
		value, exists := originalConfigs[key]
		if !exists {
			return fmt.Errorf("invalid key: %s. please specify the correct configurations", key)
		}
		value.Default = val
		originalConfigs[key] = value
	}
	return nil
}

func GenerateAllValidConfigurations(configs map[string]packagesv1.VersionConfiguration) (string, error) {
	data := map[string]interface{}{}
	for key, val := range configs {
		if val.Default != "" || val.Required {
			keySegments := strings.Split(key, ".")
			parse(data, keySegments, 0, val.Default)
		}
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshall object %v", data)
	}
	return string(out), nil
}

func parse(data map[string]interface{}, keySegments []string, index int, val string) {
	if index >= len(keySegments) {
		return
	}
	key := keySegments[index]
	inner := map[string]interface{}{}
	if _, ok := data[key]; ok {
		inner = data[key].(map[string]interface{})
	}
	parse(inner, keySegments, index+1, val)
	if len(inner) == 0 {
		if bVal, err := strconv.ParseBool(val); err == nil {
			data[key] = bVal
		} else {
			data[key] = val
		}
	} else {
		data[key] = inner
	}
}

func GenerateDefaultConfigurations(configs map[string]packagesv1.VersionConfiguration) string {
	b := new(bytes.Buffer)
	for key, val := range configs {
		if val.Required {
			fmt.Fprintf(b, "%s: \"%s\"\n", key, val.Default)
		}
	}
	return b.String()
}

func ParseConfigurations(configs []string) (map[string]string, error) {
	parsedConfigurations := make(map[string]string)

	for _, c := range configs {
		keyval := strings.Split(c, "=")
		if len(keyval) < 2 {
			return nil, fmt.Errorf("please specify %s as key=value", c)
		}
		key, val := keyval[0], keyval[1]
		parsedConfigurations[key] = val
	}
	return parsedConfigurations, nil
}

func DisplayConfigurationOptions(configs map[string]packagesv1.VersionConfiguration) {
	w := new(tabwriter.Writer)
	defer w.Flush()
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
	fmt.Fprintf(w, "%s\t%s\t%s\t \n", "Configuration", "Required", "Default")
	fmt.Fprintf(w, "%s\t%s\t%s\t \n", "-------------", "--------", "-------")
	for key, val := range configs {
		fmt.Fprintf(w, "%s\t%v\t%s\t \n", key, val.Required, val.Default)
	}
}
