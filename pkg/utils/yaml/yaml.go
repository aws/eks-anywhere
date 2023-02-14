package yaml

import (
	"bytes"
	"fmt"
	"reflect"

	unstructuredutil "github.com/aws/eks-anywhere/pkg/utils/unstructured"
	"sigs.k8s.io/yaml"
)

// Join joins YAML resources into a single YAML document. It does not validate individual
// resources.
func Join(resources [][]byte) []byte {
	return bytes.Join(resources, []byte("\n---\n"))
}

// Serialize serializes objects into YAML documents.
func Serialize[T any](objs ...T) ([][]byte, error) {
	r := [][]byte{}
	for _, o := range objs {
		b, err := yaml.Marshal(o)
		if err != nil {
			return nil, fmt.Errorf("marshalling object: %v", err)
		}
		r = append(r, b)
	}
	return r, nil
}

// Strips null fields from the yaml.
func StripNull(resources []byte) ([]byte, error) {
	uList, err := unstructuredutil.YamlToUnstructured(resources)
	if err != nil {
		return nil, fmt.Errorf("stripping object of null fields")
	}

	for _, u := range uList {
		stripNulls(u.Object)
	}

	return unstructuredutil.UnstructuredToYaml(uList)
}

func stripNulls(m map[string]interface{}) {
	val := reflect.ValueOf(m)
	for _, key := range val.MapKeys() {
		v := val.MapIndex(key)
		if v.IsNil() {
			delete(m, key.String())
			continue
		}
		switch t := v.Interface().(type) {
		// If key is a JSON object (Go Map), use recursion to go deeper
		case map[string]interface{}:
			stripNulls(t)
		}
	}
}
