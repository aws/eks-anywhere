package yamlutil

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func keyForRef(ref corev1.ObjectReference) string {
	return key(ref.APIVersion, ref.Kind, ref.Name)
}

func key(apiVersion, kind, name string) string {
	// this assumes we don't allow to have objects in multiple namespaces
	return fmt.Sprintf("%s%s%s", apiVersion, kind, name)
}

func keyForObject(o APIObject) string {
	return key(o.GetObjectKind().GroupVersionKind().GroupVersion().String(), o.GetObjectKind().GroupVersionKind().Kind, o.GetName())
}
