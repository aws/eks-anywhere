package clientutil_test

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestAddAnnotation(t *testing.T) {
	tests := []struct {
		name       string
		obj        client.Object
		key, value string
	}{
		{
			name:  "empty annotations",
			obj:   &corev1.ConfigMap{},
			key:   "my-annotation",
			value: "my-value",
		},
		{
			name: "non empty annotations",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"a": "b",
					},
				},
			},
			key:   "my-annotation",
			value: "my-value",
		},
		{
			name: "annotation present same value",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"my-annotation": "my-value",
					},
				},
			},
			key:   "my-annotation",
			value: "my-value",
		},
		{
			name: "annotation present diff value",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"my-annotation": "other-value",
					},
				},
			},
			key:   "my-annotation",
			value: "my-value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			clientutil.AddAnnotation(tt.obj, tt.key, tt.value)
			g.Expect(tt.obj.GetAnnotations()).To(HaveKeyWithValue(tt.key, tt.value))
		})
	}
}

func TestAddLabel(t *testing.T) {
	tests := []struct {
		name       string
		obj        client.Object
		key, value string
	}{
		{
			name:  "empty labels",
			obj:   &corev1.ConfigMap{},
			key:   "my-label",
			value: "my-value",
		},
		{
			name: "non empty labels",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"a": "b",
					},
				},
			},
			key:   "my-label",
			value: "my-value",
		},
		{
			name: "label present same value",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"my-label": "my-value",
					},
				},
			},
			key:   "my-label",
			value: "my-value",
		},
		{
			name: "label present diff value",
			obj: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"my-label": "other-value",
					},
				},
			},
			key:   "my-label",
			value: "my-value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			clientutil.AddLabel(tt.obj, tt.key, tt.value)
			g.Expect(tt.obj.GetLabels()).To(HaveKeyWithValue(tt.key, tt.value))
		})
	}
}
