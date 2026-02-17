package yamlutil

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// APIObject represents a kubernetes API object.
type APIObject interface {
	runtime.Object
	GetName() string
}

// ObjectLookup allows to search APIObjects by a unique key composed of apiVersion, kind, and name.
type ObjectLookup map[string]APIObject

// GetFromRef searches in a ObjectLookup for an APIObject referenced by a corev1.ObjectReference.
func (o ObjectLookup) GetFromRef(ref corev1.ObjectReference) APIObject {
	return o[keyForRef(ref)]
}

// GetFromContractVersionedRef searches for an APIObject referenced by a v1beta2 ContractVersionedObjectReference.
func (o ObjectLookup) GetFromContractVersionedRef(ref clusterv1beta2.ContractVersionedObjectReference) APIObject {
	for _, obj := range o {
		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk.Group == ref.APIGroup && gvk.Kind == ref.Kind && obj.GetName() == ref.Name {
			return obj
		}
	}
	return nil
}

func (o ObjectLookup) add(obj APIObject) {
	o[keyForObject(obj)] = obj
}

func NewObjectLookupBuilder() *ObjectLookupBuilder {
	return &ObjectLookupBuilder{
		lookup: ObjectLookup{},
	}
}

// ObjectLookupBuilder allows to construct an ObjectLookup and add APIObjects to it.
type ObjectLookupBuilder struct {
	lookup ObjectLookup
}

// Add acumulates an API object that will be included in the built ObjectLookup.
func (o *ObjectLookupBuilder) Add(objs ...APIObject) *ObjectLookupBuilder {
	for _, obj := range objs {
		o.lookup.add(obj)
	}
	return o
}

// Build constructs and returns an ObjectLookup
// After this method is called, the builder is reset and loses track
// of all previously added objects.
func (o *ObjectLookupBuilder) Build() ObjectLookup {
	l := o.lookup
	o.lookup = ObjectLookup{}
	return l
}

// Key builds the yaml object key.
func Key(apiVersion, kind, name string) string {
	// this assumes we don't allow to have objects in multiple namespaces
	return fmt.Sprintf("%s%s%s", apiVersion, kind, name)
}

func keyForRef(ref corev1.ObjectReference) string {
	return Key(ref.APIVersion, ref.Kind, ref.Name)
}

func keyForObject(o APIObject) string {
	return Key(o.GetObjectKind().GroupVersionKind().GroupVersion().String(), o.GetObjectKind().GroupVersionKind().Kind, o.GetName())
}
