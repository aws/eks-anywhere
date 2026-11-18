package framework

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestLicenseTokenForCluster(t *testing.T) {
	t.Setenv(LicenseTokenEnvVar, "current-management")
	t.Setenv(LicenseToken2EnvVar, "current-workload")
	t.Setenv(StagingLicenseTokenEnvVar, "released-management")
	t.Setenv(StagingLicenseToken2EnvVar, "released-workload")
	for _, tt := range []struct {
		name       string
		management string
		released   bool
		want       string
	}{
		{"current management", "", false, "current-management"},
		{"current explicit self-management", "cluster", false, "current-management"},
		{"current workload", "management", false, "current-workload"},
		{"released management", "", true, "released-management"},
		{"released workload", "management", true, "released-workload"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			c := &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
				Spec:       v1alpha1.ClusterSpec{ManagementCluster: v1alpha1.ManagementCluster{Name: tt.management}},
			}
			g.Expect(licenseTokenForCluster(c, tt.released)).To(Equal(tt.want))
		})
	}
}
