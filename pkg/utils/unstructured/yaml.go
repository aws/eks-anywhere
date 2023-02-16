package unstructured

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/util/yaml"
)

func YamlToUnstructured(yamlObjects []byte) ([]unstructured.Unstructured, error) {
	// Using this CAPI util for now, not sure if we want to depend on it but it's well written
	return yaml.ToUnstructured(yamlObjects)
}

func UnstructuredToYaml(yamlObjects []unstructured.Unstructured) ([]byte, error) {
	// Using this CAPI util for now, not sure if we want to depend on it but it's well written
	return yaml.FromUnstructured(yamlObjects)
}

// StripNull removes all null fields from the provided yaml.
func StripNull(resources []byte) ([]byte, error) {
	uList, err := YamlToUnstructured(resources)
	if err != nil {
		return nil, fmt.Errorf("converting yaml to unstructured: %v", err)
	}
	for _, u := range uList {
		stripNull(u.Object)
	}
	return UnstructuredToYaml(uList)
}

func stripNull(m map[string]interface{}) {
	val := reflect.ValueOf(m)
	for _, key := range val.MapKeys() {
		v := val.MapIndex(key)
		if v.IsNil() {
			delete(m, key.String())
			continue
		}
		if t, ok := v.Interface().(map[string]interface{}); ok {
			stripNull(t)
		}
	}
}
