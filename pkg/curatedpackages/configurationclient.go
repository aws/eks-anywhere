package curatedpackages

import (
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"
)

func GenerateAllValidConfigurations(configs map[string]string) (string, error) {
	data := map[string]interface{}{}
	for key, val := range configs {
		if val != "" {
			keySegments := strings.Split(key, ".")
			parse(data, keySegments, 0, val)
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
