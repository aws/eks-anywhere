package clientutil_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestClusterNameIndexerValid(t *testing.T) {
	g := NewWithT(t)
	c := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
	}
	name := clientutil.ClusterNameIndexer(c)
	g.Expect(name[0]).To(Equal(c.Name))
}

func TestClusterNameIndexerFail(t *testing.T) {
	g := NewWithT(t)
	c := &v1alpha1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
	}
	g.Expect(func() {
		clientutil.ClusterNameIndexer(c)
	}).To(Panic())
}
