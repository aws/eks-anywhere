package templater

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const objectSeparator string = "\n---\n"

func AppendYamlResources(resources ...[]byte) []byte {
	separator := []byte(objectSeparator)

	size := 0
	for _, resource := range resources {
		size += len(resource) + len(separator)
	}

	b := make([]byte, 0, size)
	for _, resource := range resources {
		b = append(b, resource...)
		b = append(b, separator...)
	}

	return b
}

func ObjectsToYaml(objs ...runtime.Object) ([]byte, error) {
	r := [][]byte{}
	for _, o := range objs {
		b, err := yaml.Marshal(o)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal object: %v", err)
		}
		r = append(r, b)
	}
	return AppendYamlResources(r...), nil
}
