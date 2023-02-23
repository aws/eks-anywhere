package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestValidateHostOSConfig(t *testing.T) {
	tests := []struct {
		name         string
		hostOSConfig *HostOSConfiguration
		wantErr      string
	}{
		{
			name:         "nil HostOSConfig",
			hostOSConfig: nil,
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
			wantErr: "NTPConfiguration.Servers can not be empty",
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
			wantErr: "ntp servers [not a valid ntp server, also invalid, udp://] is not valid",
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
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateHostOSConfig(tt.hostOSConfig)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}
