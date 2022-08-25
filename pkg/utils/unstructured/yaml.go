package unstructured

import (
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
