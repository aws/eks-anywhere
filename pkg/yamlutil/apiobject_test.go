package yamlutil

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObjectLookupGetFromRef(t *testing.T) {
	g := NewWithT(t)
	o := ObjectLookup{}
	want := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      "my-secret",
		},
		Data: map[string][]byte{
			"username": []byte("test"),
			"password": []byte("test"),
		},
	}
	objRef := corev1.ObjectReference{
		Kind:       want.Kind,
		APIVersion: want.APIVersion,
		Name:       want.Name,
		Namespace:  want.Namespace,
	}

	o.add(want)

	otherSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      "my-other-secret",
		},
	}
	o.add(otherSecret)

	got := o.GetFromRef(objRef)
	g.Expect(got).To(Equal(want))
}
