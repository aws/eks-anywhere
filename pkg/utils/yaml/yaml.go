package yaml

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
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

// SplitDocuments function splits content into individual document parts represented as byte slices.
func SplitDocuments(r io.Reader) ([][]byte, error) {
	resources := make([][]byte, 0)

	yr := apiyaml.NewYAMLReader(bufio.NewReader(r))
	for {
		d, err := yr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		resources = append(resources, d)
	}

	return resources, nil
}
