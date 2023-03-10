package common_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

const (
	emptyBottlerocketConfig = `bottlerocket: {}`

	emptyKubernetesConfig = `bottlerocket:
  kubernetes: {}`

	maxPodsConfig = `bottlerocket:
  kubernetes:
    maxPods: 100`

	allowedUnsafeSysctlsConfig = `bottlerocket:
  kubernetes:
    allowedUnsafeSysctls:
    - foo
    - bar`

	clusterDNSIPsConfig = `bottlerocket:
  kubernetes:
    clusterDNSIPs:
    - 1.2.3.4
    - 5.6.7.8`
)

func TestGetCAPIBottlerocketSettingsConfig(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name     string
		config   *v1alpha1.BottlerocketConfiguration
		expected string
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: "",
		},
		{
			name:     "empty config",
			config:   &v1alpha1.BottlerocketConfiguration{},
			expected: emptyBottlerocketConfig,
		},
		{
			name: "empty kubernetes config",
			config: &v1alpha1.BottlerocketConfiguration{
				Kubernetes: &v1beta1.BottlerocketKubernetesSettings{},
			},
			expected: emptyKubernetesConfig,
		},
		{
			name: "with allowed unsafe sysctls",
			config: &v1alpha1.BottlerocketConfiguration{
				Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
					AllowedUnsafeSysctls: []string{"foo", "bar"},
				},
			},
			expected: allowedUnsafeSysctlsConfig,
		},
		{
			name: "with cluster dns IPs",
			config: &v1alpha1.BottlerocketConfiguration{
				Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
					ClusterDNSIPs: []string{"1.2.3.4", "5.6.7.8"},
				},
			},
			expected: clusterDNSIPsConfig,
		},
		{
			name: "with max pods",
			config: &v1alpha1.BottlerocketConfiguration{
				Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
					MaxPods: 100,
				},
			},
			expected: maxPodsConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := common.GetCAPIBottlerocketSettingsConfig(tt.config)
			g.Expect(err).ToNot(HaveOccurred())
			if got != tt.expected {
				fmt.Println(got)
				fmt.Println(tt.expected)
			}

			g.Expect(got).To(Equal(tt.expected))
		})
	}
}
