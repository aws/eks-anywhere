package yamlutil_test

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

func TestObjectLookupGetFromRef(t *testing.T) {
	g := NewWithT(t)
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

	otherSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      "my-other-secret",
		},
	}

	o := yamlutil.NewObjectLookupBuilder().Add(want, otherSecret).Build()

	got := o.GetFromRef(objRef)
	g.Expect(got).To(Equal(want))
}
