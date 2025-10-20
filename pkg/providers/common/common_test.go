package common_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	bootstrapv1 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

const (
	emptyBottlerocketConfig = `bottlerocket: {}`

	emptyKubernetesConfig = `bottlerocket:
  kubernetes: {}`

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

	kernelSysctlConfig = `bottlerocket:
  kernel:
    sysctlSettings:
      foo: bar`

	bootKernelConfig = `bottlerocket:
  boot:
    bootKernelParameters:
      foo:
      - abc
      - def`

	kubeConfig = `bottlerocket:
  kubernetes:
    cpuCFSQuota: false
    cpuManagerReconcilePeriod: 15s
    evictionHard:
      a: b
    maxPods: 100
    shutdownGracePeriod: 15s
    shutdownGracePeriodCriticalPods: 15s`
)

func TestGetCAPIBottlerocketSettingsConfig(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name           string
		config         *v1alpha1.HostOSConfiguration
		expected       string
		brKubeSettings *bootstrapv1.BottlerocketKubernetesSettings
	}{
		{
			name:           "nil config",
			config:         nil,
			brKubeSettings: nil,
			expected:       "",
		},
		{
			name:           "empty config",
			config:         &v1alpha1.HostOSConfiguration{},
			brKubeSettings: nil,
			expected:       "",
		},
		{
			name: "empty BR config",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{},
			},
			brKubeSettings: nil,
			expected:       emptyBottlerocketConfig,
		},
		{
			name: "empty kubernetes config",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{},
			},
			brKubeSettings: &bootstrapv1.BottlerocketKubernetesSettings{},
			expected:       emptyKubernetesConfig,
		},
		{
			name: "with allowed unsafe sysctls",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{},
			},
			brKubeSettings: &bootstrapv1.BottlerocketKubernetesSettings{
				AllowedUnsafeSysctls: []string{"foo", "bar"},
			},
			expected: allowedUnsafeSysctlsConfig,
		},
		{
			name: "with cluster dns IPs",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{},
			},
			brKubeSettings: &bootstrapv1.BottlerocketKubernetesSettings{
				ClusterDNSIPs: []string{"1.2.3.4", "5.6.7.8"},
			},
			expected: clusterDNSIPsConfig,
		},
		{
			name: "with max pods",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{},
			},
			brKubeSettings: &bootstrapv1.BottlerocketKubernetesSettings{
				MaxPods:     100,
				CPUCFSQuota: ptr.To(false),
				CPUManagerReconcilePeriod: &v1.Duration{
					Duration: 15 * time.Second,
				},
				EvictionHard: map[string]string{
					"a": "b",
				},
				ShutdownGracePeriod: &v1.Duration{
					Duration: 15 * time.Second,
				},
				ShutdownGracePeriodCriticalPods: &v1.Duration{
					Duration: 15 * time.Second,
				},
			},
			expected: kubeConfig,
		},
		{
			name: "with kernel sysctl config",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{
					Kernel: &bootstrapv1.BottlerocketKernelSettings{
						SysctlSettings: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			brKubeSettings: nil,
			expected:       kernelSysctlConfig,
		},
		{
			name: "with boot kernel parameters",
			config: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{
					Boot: &bootstrapv1.BottlerocketBootSettings{
						BootKernelParameters: map[string][]string{
							"foo": {
								"abc",
								"def",
							},
						},
					},
				},
			},
			brKubeSettings: nil,
			expected:       bootKernelConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := common.GetCAPIBottlerocketSettingsConfig(tt.config, tt.brKubeSettings)
			g.Expect(err).ToNot(HaveOccurred())
			if got != tt.expected {
				fmt.Println(got)
				fmt.Println(tt.expected)
			}

			g.Expect(got).To(Equal(tt.expected))
		})
	}
}

func TestGetExternalEtcdReleaseURL(t *testing.T) {
	g := NewWithT(t)
	testcases := []struct {
		name           string
		clusterVersion string
		etcdURL        string
		err            error
	}{
		{
			name:           "Pre-etcd url enabled version < 0.19.0",
			clusterVersion: "v0.15.2",
			etcdURL:        "",
			err:            nil,
		},
		{
			name:           "Etcd url enabled version = 0.19.0",
			clusterVersion: "v0.19.0",
			etcdURL:        test.VersionBundle().KubeDistro.EtcdURL,
		},
		{
			name:           "Etcd url enabled version = 0.19.0 with dev build",
			clusterVersion: "v0.19.0-dev+latest",
			etcdURL:        test.VersionBundle().KubeDistro.EtcdURL,
		},
		{
			name:           "Post etcd url enabled version > 0.19.0",
			clusterVersion: "v0.21.0",
			etcdURL:        test.VersionBundle().KubeDistro.EtcdURL,
		},
		{
			name:           "Invalid semver for eks-a cluster version",
			clusterVersion: "a.12.4.3.78",
			err:            fmt.Errorf("invalid semver for clusterVersion: invalid major version in semver a.12.4.3.78: strconv.ParseInt: parsing \"\": invalid syntax"),
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			eksaVersion := v1alpha1.EksaVersion(tt.clusterVersion)
			got, err := common.GetExternalEtcdReleaseURL(&eksaVersion, test.VersionBundle())
			if tt.err == nil {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err.Error()).To(Equal(tt.err.Error()))
			}
			g.Expect(got).To(Equal(tt.etcdURL))
		})
	}
}

func TestGetExternalEtcdReleaseURLWithNilEksaVersion(t *testing.T) {
	g := NewWithT(t)
	got, err := common.GetExternalEtcdReleaseURL(nil, test.VersionBundle())
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(got).To(BeEmpty())
}

func TestValidateBottlerocketKC(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name   string
		subErr error
		spec   *unstructured.Unstructured
		br     *bootstrapv1.BottlerocketKubernetesSettings
	}{
		{
			name: "cp config",
			spec: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind":       "KubeletConfiguration",
					"maxPods":    50,
					"apiVersion": "api",
				},
			},
			subErr: nil,
		},
		{
			name: "worker config",
			spec: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"maxPods":     50,
					"kind":        "KubeletConfiguration",
					"cpuCFSQuota": ptr.To(false),
					"cpuManagerReconcilePeriod": &v1.Duration{
						Duration: 15 * time.Second,
					},
				},
			},
			subErr: nil,
		},
		{
			name:   "nil kc config",
			spec:   nil,
			subErr: nil,
		},
		{
			name: "invalid cp config",
			spec: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"maxPodss": 50,
					"kind":     "KubeletConfiguration",
				},
			},

			subErr: errors.New("unknown field \"maxPodss\""),
		},
		{
			name: "invalid worker config",
			spec: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"maxPodss": 50,
					"kind":     "KubeletConfiguration",
				},
			},

			subErr: errors.New("unknown field \"maxPodss\""),
		},
		{
			name: "invalid worker config",
			spec: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"staticPodPath": "path",
					"kind":          "KubeletConfiguration",
				},
			},

			subErr: errors.New("unknown field \"staticPodPath\""),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(_ *testing.T) {
			_, err := common.ConvertToBottlerocketKubernetesSettings(tc.spec)
			if tc.subErr == nil {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tc.subErr.Error()))
			}
		})
	}
}
