package clientutil

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ObjectsToClientObjects[T client.Object](objs []T) []client.Object {
	runtimeObjs := make([]client.Object, 0, len(objs))
	for _, o := range objs {
		runtimeObjs = append(runtimeObjs, o)
	}

	return runtimeObjs
}
