package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
)

func TestValidateHostOSConfig(t *testing.T) {
	tests := []struct {
		name         string
		hostOSConfig *HostOSConfiguration
		osFamily     OSFamily
		wantErr      string
	}{
		{
			name:         "nil HostOSConfig",
			hostOSConfig: nil,
			osFamily:     Bottlerocket,
			wantErr:      "",
		},
		{
			name:         "empty HostOSConfig",
			hostOSConfig: &HostOSConfiguration{},
			wantErr:      "",
		},
		{
			name: "empty NTP servers",
			hostOSConfig: &HostOSConfiguration{
				NTPConfiguration: &NTPConfiguration{
					Servers: []string{},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "NTPConfiguration.Servers can not be empty",
		},
		{
			name: "invalid NTP servers",
			hostOSConfig: &HostOSConfiguration{
				NTPConfiguration: &NTPConfiguration{
					Servers: []string{
						"time-a.eks-a.aws",
						"not a valid ntp server",
						"also invalid",
						"udp://",
						"time-b.eks-a.aws",
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "ntp servers [not a valid ntp server, also invalid, udp://] is not valid",
		},
		{
			name: "valid NTP config",
			hostOSConfig: &HostOSConfiguration{
				NTPConfiguration: &NTPConfiguration{
					Servers: []string{
						"time-a.eks-a.aws",
						"time-b.eks-a.aws",
						"192.168.0.10",
						"2610:20:6f15:15::26",
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "empty Bottlerocket config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "empty Bottlerocket.Kuberentes config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kubernetes: &v1beta1.BottlerocketKubernetesSettings{},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "empty Bottlerocket.Kuberentes full valid config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
						AllowedUnsafeSysctls: []string{
							"net.core.somaxconn",
							"net.ipv4.ip_local_port_range",
						},
						ClusterDNSIPs: []string{
							"1.2.3.4",
							"5.6.7.8",
						},
						MaxPods: 100,
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "invalid Bottlerocket.Kuberentes.AllowedUnsafeSysctls",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
						AllowedUnsafeSysctls: []string{
							"net.core.somaxconn",
							"",
						},
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "BottlerocketConfiguration.Kubernetes.AllowedUnsafeSysctls can not have an empty string (\"\")",
		},
		{
			name: "invalid Bottlerocket.Kuberentes.ClusterDNSIPs",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
						ClusterDNSIPs: []string{
							"1.2.3.4",
							"not a valid IP",
						},
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "IP address [not a valid IP] in BottlerocketConfiguration.Kubernetes.ClusterDNSIPs is not a valid IP",
		},
		{
			name: "invalid Bottlerocket.Kuberentes.MaxPods",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kubernetes: &v1beta1.BottlerocketKubernetesSettings{
						MaxPods: -1,
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "BottlerocketConfiguration.Kubernetes.MaxPods can not be negative",
		},
		{
			name: "Bottlerocket config with non-Bottlerocket OSFamily",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{},
			},
			osFamily: Ubuntu,
			wantErr:  "BottlerocketConfiguration can only be used with osFamily: \"bottlerocket\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateHostOSConfig(tt.hostOSConfig, tt.osFamily)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
