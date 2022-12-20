package api

import (
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func CleanupPathsFromYaml(yamlContent []byte, paths []string) ([]byte, error) {
	m := make(map[string]interface{})

	if err := yaml.Unmarshal(yamlContent, &m); err != nil {
		return nil, err
	}

	deletePaths(m, paths)

	newYaml, err := yaml.Marshal(&m)
	if err != nil {
		return nil, err
	}

	return newYaml, nil
}

// CleanupPathsInObject unsets or nullifies the provided paths in a give kubernetes Object.
func CleanupPathsInObject[T any, PT interface {
	*T
	client.Object
}](obj PT, paths []string,
) error {
	clusterObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return errors.Wrap(err, "failed converting cluster to unstructured")
	}

	deletePaths(clusterObj, paths)

	*obj = *new(T)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(clusterObj, obj); err != nil {
		return errors.Wrap(err, "failed converting back to cluster from unstructured")
	}

	return nil
}

func deletePaths(m map[string]interface{}, paths []string) {
	for _, p := range paths {
		deletePath(m, strings.Split(p, "."))
	}
}

func deletePath(m map[string]interface{}, path []string) {
	currentElement := m
	for i := 0; i < len(path)-1; i++ {
		p := strings.TrimSuffix(path[i], "[]")
		isArray := len(path[i]) != len(p)

		e, ok := currentElement[p]
		if !ok {
			return
		}

		if isArray {
			arrayElement, arrayOk := e.([]interface{})
			if !arrayOk {
				return
			}

			for _, o := range arrayElement {
				currentElement, ok = o.(map[string]interface{})
				if !ok {
					continue
				}

				deletePath(currentElement, path[i+1:])
			}
			return
		}

		currentElement, ok = e.(map[string]interface{})
		if !ok {
			return
		}
	}

	delete(currentElement, path[len(path)-1])
}
