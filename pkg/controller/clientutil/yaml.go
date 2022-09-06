package clientutil

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/utils/unstructured"
)

func YamlToClientObjects(yamlObjects []byte) ([]client.Object, error) {
	unstructuredObjs, err := unstructured.YamlToUnstructured(yamlObjects)
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
