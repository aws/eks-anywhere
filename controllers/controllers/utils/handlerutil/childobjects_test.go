package handlerutil_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers/controllers/utils/handlerutil"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestChildObjectToClusters(t *testing.T) {
	testCases := []struct {
		testName     string
		obj          client.Object
		wantRequests []reconcile.Request
	}{
		{
			testName: "two clusters",
			obj: &anywherev1.OIDCConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-oidc",
					Namespace: "my-namespace",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: anywherev1.GroupVersion.String(),
							Kind:       anywherev1.ClusterKind,
							Name:       "my-cluster",
						},
						{
							APIVersion: anywherev1.GroupVersion.String(),
							Kind:       anywherev1.ClusterKind,
							Name:       "my-other-cluster",
						},
					},
				},
			},
			wantRequests: []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      "my-cluster",
						Namespace: "my-namespace",
					},
				},
				{
					NamespacedName: types.NamespacedName{
						Name:      "my-other-cluster",
						Namespace: "my-namespace",
					},
				},
			},
		},
		{
			testName: "no-clusters",
			obj: &anywherev1.OIDCConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-oidc",
					Namespace: "my-namespace",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: anywherev1.GroupVersion.String(),
							Kind:       "OtherObj",
							Name:       "my-obj",
						},
					},
				},
			},
			wantRequests: []reconcile.Request{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			handle := handlerutil.ChildObjectToClusters(logr.New(logf.NullLogSink{}))
			requests := handle(tt.obj)
			g.Expect(requests).To(Equal(tt.wantRequests))
		})
	}
}
