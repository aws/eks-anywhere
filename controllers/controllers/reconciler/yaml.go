package reconciler

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func YamlToClientObjects(yamlObjects []byte) ([]client.Object, error) {
	unstructuredObjs, err := YamlToUnstructured(yamlObjects)
	if err != nil {
		return nil, err
	}

	objs := make([]client.Object, 0, len(unstructuredObjs))
	// Use a numbered loop to avoid problems when retrieving the pointer
	for i := range unstructuredObjs {
		objs = append(objs, &unstructuredObjs[i])
	}

	return objs, nil
}

func YamlToUnstructured(yamlObjects []byte) ([]unstructured.Unstructured, error) {
	// Using this CAPI util for now, not sure if we want to depend on it but it's well written
	return yaml.ToUnstructured(yamlObjects)
}
