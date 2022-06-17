package handlerutil_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers/controllers/utils/handlerutil"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestCAPIObjectToCluster(t *testing.T) {
	testCases := []struct {
		testName     string
		obj          client.Object
		wantRequests []reconcile.Request
	}{
		{
			testName:     "no eksa managed",
			obj:          &clusterv1.Cluster{},
			wantRequests: nil,
		},
		{
			testName: " missing namespace",
			obj: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterapi.EKSAClusterLabelName: "my-cluster",
					},
				},
			},
			wantRequests: nil,
		},
		{
			testName: "managed capi resource",
			obj: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterapi.EKSAClusterLabelName:      "my-cluster",
						clusterapi.EKSAClusterLabelNamespace: "my-namespace",
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
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			handle := handlerutil.CAPIObjectToCluster(logr.New(logf.NullLogSink{}))
			requests := handle(tt.obj)
			g.Expect(requests).To(Equal(tt.wantRequests))
		})
	}
}
