package vsphere

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestFailureDomainsSpecSuccess(t *testing.T) {
	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: v1alpha1.Kube124,
				},
			},
			VSphereDatacenter: &anywherev1.VSphereDatacenterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: anywherev1.VSphereDatacenterConfigSpec{
					FailureDomains: []anywherev1.FailureDomain{
						{
							Name: "fd-1",
						},
					},
				},
			},
		},
	}

	logger := test.NewNullLogger()
	failureDomains, err := FailureDomainsSpec(logger, spec)
	assert.Nil(t, err)
	assert.True(t, len(failureDomains.Objects()) > 0)
	assert.Equal(t, failureDomains.Groups[0].VsphereDeploymentZone.Name, "test-test-fd-1")
}
