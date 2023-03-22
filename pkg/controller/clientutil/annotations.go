package clientutil

import "sigs.k8s.io/controller-runtime/pkg/client"

// AddAnnotation adds an annotation to the given object.
// If the annotation already exists, it overwrites its value.
func AddAnnotation(o client.Object, key, value string) {
	a := o.GetAnnotations()
	if a == nil {
		a = make(map[string]string, 1)
	}
	a[key] = value
	o.SetAnnotations(a)
}

// AddLabel adds a label to the given object.
// If the label already exists, it overwrites its value.
func AddLabel(o client.Object, key, value string) {
	l := o.GetLabels()
	if l == nil {
		l = make(map[string]string, 1)
	}
	l[key] = value
	o.SetLabels(l)
}
