package curatedpackages

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/types"
)

func GetConfigurationsFromBundle(bundlePackage *packagesv1.BundlePackage) (map[string]*types.DisplayConfiguration, error) {
	configs := make(map[string]*types.DisplayConfiguration)
	if bundlePackage == nil {
		return nil, fmt.Errorf("non existent package")
	}
	if len(bundlePackage.Source.Versions) < 1 {
		return nil, fmt.Errorf("package %s doesn't have enough versions", bundlePackage.Name)
	}
	jsonSchema, err := getJsonSchemaFromBundle(bundlePackage)
	if err != nil {
		return nil, fmt.Errorf("parsing schema %v", err)
	}
	schemaObj := &types.Schema{}
	err = json.Unmarshal(jsonSchema, schemaObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling json schema to an object %v", err)
	}

	for key, prop := range schemaObj.Properties {
		configs[key] = &types.DisplayConfiguration{
			Type:        prop.Type,
			Default:     prop.Default,
			Description: prop.Description,
		}
	}

	for _, field := range schemaObj.Required {
		if v, found := configs[field]; found {
			v.Required = true
		}
	}
	return configs, nil
}

func UpdateConfigurations(originalConfigs map[string]*types.DisplayConfiguration, newConfigs map[string]string) error {
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

func GenerateAllValidConfigurations(configs map[string]*types.DisplayConfiguration) (string, error) {
	data := map[string]interface{}{}
	for key, val := range configs {
		if val.Default != "" || val.Required {
			keySegments := strings.Split(key, ".")
			parse(data, keySegments, 0, val.Default)
		}
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal configurations %v", data)
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

func GenerateDefaultConfigurations(configs map[string]*types.DisplayConfiguration) string {
	data := map[string]interface{}{}
	for key, val := range configs {
		if val.Required {
			keySegments := strings.Split(key, ".")
			parse(data, keySegments, 0, val.Default)
		}
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return ""
	}
	return string(out)
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

func DisplayConfigurationOptions(configs map[string]*types.DisplayConfiguration) {
	w := new(tabwriter.Writer)
	defer w.Flush()
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
	fmt.Fprintf(w, "%s\t%s\t%s\t \n", "Configuration", "Required", "Default")
	fmt.Fprintf(w, "%s\t%s\t%s\t \n", "-------------", "--------", "-------")
	for key, val := range configs {
		fmt.Fprintf(w, "%s\t%v\t%s\t \n", key, val.Required, val.Default)
	}
}

func getJsonSchemaFromBundle(bundlePackage *packagesv1.BundlePackage) ([]byte, error) {
	// The package configuration is gzipped and base64 encoded
	// When processing the configuration, the reverse occurs: base64 decode, then unzip
	configuration := bundlePackage.Source.Versions[0].Configurations[0].Default
	decodedConfiguration, err := base64.StdEncoding.DecodeString(configuration)
	if err != nil {
		return nil, fmt.Errorf("decoding configuration")
	}
	reader := bytes.NewReader(decodedConfiguration)
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error when uncompressing configurations %v", err)
	}

	output, err := ioutil.ReadAll(gzreader)
	if err != nil {
		return nil, fmt.Errorf("error reading configurations %v", err)
	}

	return output, nil
}
