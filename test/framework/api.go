package framework

import (
	"strings"

	"sigs.k8s.io/yaml"
)

var incompatiblePathsForVersion = map[string][]string{
	"v0.6.1": {"spec.clusterNetwork.dns"},
}

func cleanUpClusterForVersion(clusterYaml []byte, version string) ([]byte, error) {
	return cleanupPathsFromYaml(clusterYaml, incompatiblePathsForVersion[version])
}

func cleanupPathsFromYaml(yamlContent []byte, paths []string) ([]byte, error) {
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

func deletePaths(m map[string]interface{}, paths []string) {
	for _, p := range paths {
		deletePath(m, strings.Split(p, "."))
	}
}

func deletePath(m map[string]interface{}, path []string) {
	currentElement := m
	for i := 0; i < len(path)-1; i++ {
		e, ok := currentElement[path[i]]
		if !ok {
			return
		}

		currentElement, ok = e.(map[string]interface{})
		if !ok {
			return
		}
	}

	delete(currentElement, path[len(path)-1])
}
