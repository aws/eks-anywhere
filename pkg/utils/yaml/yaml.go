package yaml

import (
	"bytes"
	"fmt"

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
