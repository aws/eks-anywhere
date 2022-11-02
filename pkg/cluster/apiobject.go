package cluster

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// APIObject represents a kubernetes API object.
type APIObject interface {
	runtime.Object
	GetName() string
}

type ObjectLookup map[string]APIObject

// GetFromRef searches in a ObjectLookup for an APIObject referenced by a anywherev1.Ref.
func (o ObjectLookup) GetFromRef(apiVersion string, ref anywherev1.Ref) APIObject {
	return o[keyForRef(apiVersion, ref)]
}

func (o ObjectLookup) add(obj APIObject) {
	o[keyForObject(obj)] = obj
}

func keyForRef(apiVersion string, ref anywherev1.Ref) string {
	return key(apiVersion, ref.Kind, ref.Name)
}

func key(apiVersion, kind, name string) string {
	// this assumes we don't allow to have objects in multiple namespaces
	return fmt.Sprintf("%s%s%s", apiVersion, kind, name)
}

func keyForObject(o APIObject) string {
	return key(o.GetObjectKind().GroupVersionKind().GroupVersion().String(), o.GetObjectKind().GroupVersionKind().Kind, o.GetName())
}
