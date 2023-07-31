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
			name: "empty Bottlerocket.Kubernetes config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kubernetes: &v1beta1.BottlerocketKubernetesSettings{},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "empty Bottlerocket.Kubernetes full valid config",
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
			name: "invalid Bottlerocket.Kubernetes.AllowedUnsafeSysctls",
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
			name: "invalid Bottlerocket.Kubernetes.ClusterDNSIPs",
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
			name: "invalid Bottlerocket.Kubernetes.MaxPods",
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
		{
			name: "valid kernel config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kernel: &v1beta1.BottlerocketKernelSettings{
						SysctlSettings: map[string]string{
							"vm.max_map_count":         "262144",
							"fs.file-max":              "65535",
							"net.ipv4.tcp_mtu_probing": "1",
						},
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "invalid kernel key value",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Kernel: &v1beta1.BottlerocketKernelSettings{
						SysctlSettings: map[string]string{
							"": "262144",
						},
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "sysctlSettings key cannot be empty",
		},
		{
			name: "valid bootSettings config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Boot: &v1beta1.BottlerocketBootSettings{
						BootKernelParameters: map[string][]string{
							"console": {
								"tty0",
								"ttyS0,115200n8",
							},
						},
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "invalid bootSettings config",
			hostOSConfig: &HostOSConfiguration{
				BottlerocketConfiguration: &BottlerocketConfiguration{
					Boot: &v1beta1.BottlerocketBootSettings{
						BootKernelParameters: map[string][]string{
							"": {
								"tty0",
								"ttyS0,115200n8",
							},
						},
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "bootKernelParameters key cannot be empty",
		},
		{
			name: "valid cert bundle",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: []certBundle{
					{
						Name: "bundle1",
						Data: `-----BEGIN CERTIFICATE-----
MIIFrTCCA5WgAwIBAgIUFXtDg6MEuAA0Ns1Ah2pfVKDC/nIwDQYJKoZIhvcNAQEN
BQAwZjELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdTZWF0dGxl
MQ0wCwYDVQQKDARFa3NhMRIwEAYDVQQLDAlQZXJzb25hbGwxFTATBgNVBAMMDDEw
LjgwLjE0OC41NjAeFw0yMzAyMTEwMDQ0MjRaFw0zMzAyMDgwMDQ0MjRaMGYxCzAJ
BgNVBAYTAlVTMQswCQYDVQQIDAJXQTEQMA4GA1UEBwwHU2VhdHRsZTENMAsGA1UE
CgwERWtzYTESMBAGA1UECwwJUGVyc29uYWxsMRUwEwYDVQQDDAwxMC44MC4xNDgu
NTYwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDpVLHBdFr7fh+PRQnZ
PlyXbIDfxdKCNZaNfxOl5I2eKW5m+zennt7N+nM1/5oDcVoIzVVCFkSmztl5GNr7
zyKfqDb/q6wZSQTOreNALBEh6redLnzc6OkQYnlFFuLcTuWLqTHdMoEJozbW+9K+
Z3lxKU92FvsTDaZKCT8NWwnKoTeXhEZtOF0KnJmzdQUztL7mNjSn53qUv/6WHwBG
/8F7elkJYP98jhhYkpKgPSnpoDuay3zmQxsFXvh9+j9GztODoroZgkNBuooObsnE
CQEFJLGZ03XAkyaumzfjSD5Ma4QQZPy0VwV2NHL3ngec+wxUH7u+FqupWbQZtJP6
+I3jSGGmhI2G/NIJZD0jiytlR1YmoUYM5qHl/VvrcqKoMEIGgF1ktYup9NAdGzLA
AItBbjDY8Vl+TGMC0vDbrHtVReYOWfx7TBo5nPnBjC1yTtnoYIk8EkJaPY9AuI5V
/WJ866PrPo5dPw/EVu3kuVzG5VoYb+iqY/qENnDnNeP0rCYF/tgna8c6sBkJG2y9
vo8ZkLC22J1CTWNqpiLBg0B9Jn2WIGFKNlavj8RiLSjNMDDj0EzS8mxZqlE+gn+C
MsASCpOAtIKPlr+x+Bfd0xaG1HyzaJV/cFHW8p6+8UvE9Nl0SeTWd8RtG36zXxif
WBDWiUJ9clQSp0F4Yunxts3PFQIDAQABo1MwUTAdBgNVHQ4EFgQUT78n048MqxDw
DdtI+841JN+OQpEwHwYDVR0jBBgwFoAUT78n048MqxDwDdtI+841JN+OQpEwDwYD
VR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQ0FAAOCAgEAI3TT8ykTsX4Ra8asa8gl
DqVsz8t5HjOpA2nfAmO9GETg+StEW0j2FGglu+S4OhMx4TN4rYAPY0C78+5GHQwu
UAKxAVXigLXG2TmgszJ1OIBXIRdMlyc3gVm9j5Odcrm0WSGmSe47K8IgEVwlGEoS
CEUfKbMlVzEaPlP7YUOtiElzYi8D4ht9JfmRPi7PjUjVrWw0My+XPrFA4siU8dfh
BbvN9ybHmgltFqeZEqEv2wA9l6EpY71sArWsw+k5OYb2tXiXBxXY9LJzNDdJsNW9
8EP5rFIPvtoZMlE9qAHqy1kkqxcjvhcD3SD8zxJKCDLCJnVPG01k2siwlA46jcTT
xFZhEttbFcASpURjxkvBndXCktN2myWwokNGlf1hosxk2lG5DcySwHAhjXLJV6r2
l7X/CZpR38n/FXAKwiMQAPFZsLRU/EWBPTDlD1zjQSH8weGEj8+e9tQuTm+QsYrb
aJW3puZ84fEgYu/QMjGTuJzd+ZswMcLLyyn4Sm9nvchE8SdEUiF6L0Lc8+qgwmJU
idxqPeX4DMweDcskpZDPbfI6jnNorvGiWaLAYEJ4ntc3SP/lvbwXXOJLvhnRP+Ov
zphcd/PRLS7VpAhWOVbulPjB8DkX0PmvgaCeiTDuajMxq6ve64v+dCwUcqqbamC2
OelAabtJKd8B2BUsR7JRIN8=
-----END CERTIFICATE-----`,
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "valid nil certBunldes",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: nil,
			},
			osFamily: Bottlerocket,
			wantErr:  "",
		},
		{
			name: "invalid cert no data",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: []certBundle{
					{
						Name: "bundle1",
						Data: "",
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "failed to parse certificate PEM",
		},
		{
			name: "invalid cert bundle no name",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: []certBundle{
					{
						Name: "",
						Data: "ABCDEF",
					},
					{
						Name: "bundle2",
						Data: "123456",
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "certBundles name cannot be empty",
		},
		{
			name: "invalid cert bundle wrong data type",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: []certBundle{
					{
						Name: "bundle1",
						Data: "QUJDREVm",
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "failed to parse certificate PEM",
		},
		{
			name: "invalid cert bundle wrong data type",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: []certBundle{
					{
						Name: "bundle1",
						Data: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQ0NnS0NBZ0VBbFJ1Um5UaFVqVTgvcHJ3WXhidHkKV1BUOXBVUkkzbGJzS01pQjZGbi9WSE9LRTEzcDREOHhnT0NBRHBkUmFnZFQ2bjRldHI5YXR6REtVU3ZwTXRSMwpDUDVub05jOTdXaU5DZ2dCalZXaHM3c3pFZTh1Z3lxRjIzWHdwSFE2dVYxTEtINTBtOTJNYk9XZkN0alU5cC94CnFoTnBRUTFBWmhxTnk1R2V2YXA1azhYelJtalNsZE5BRlpNWTdZdjNHaStueUN3R3dwVnRCVXdodUx6Z05GSy8KeUR0dzJXY1dtVVU3TnVDOFE2TVd2UGVieFZ0Q2ZWcC9pUVU2cTYweXl0NmFHT0JraEFYMExwS0FFaEtpZGl4WQpuUDlQTlZCdnhndTNYWjRQMzZnWlY2K3VtbUtkQlZuYzNOcXdCTHU1K0NjZFJkdXNtSFBIZDVwSGY0LzM4WjMvCjZxVTJhL2ZQdld6Y2VWVEVnWjQ3UWpGTVRDVG1Dd050Mjljdmk3elplUXpqdHdRZ240aXBOOU5pYlJIL0F4L3EKVGJJekhmckoxeGEyUnRlV1NkRmp3dHhpOUMyMEhVa2pYU2VJNFlselFNSDBmUFg2S0NFN2FWZVBUT25CNjlJLwphOS9xOTZEaVhaYWp3bHBxM3dGY3RyczFvWHFCcDVEVnJDSWo4aFUyd05nQjdMdFExbUN0c1l6Ly9oZWFpMEs5ClBoRTRYNmhpRTBZbWVBWmpSMHVIbDhNLzVhVzl4Q29KNzIrMTJrS3BXQWEwU0ZSV0x5NkZlak5ZQ1lwa3VwVkoKeWVjTGsvNEwxVzBsNmpRUVpuV0VyWFpZZTBQTkZjbXdHWHkxUmVwODNrZkJSTktSeTV0dm9jYWxMbHdYTGRVawpBSVUrMkdLanlUM2lNdXpaeHhGeFBGTUNBd0VBQVE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==",
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "failed to parse certificate",
		},
		{
			name: "more than one cert in one bundle",
			hostOSConfig: &HostOSConfiguration{
				CertBundles: []certBundle{
					{
						Name: "bundle1",
						Data: `-----BEGIN CERTIFICATE-----
MIIFrTCCA5WgAwIBAgIUFXtDg6MEuAA0Ns1Ah2pfVKDC/nIwDQYJKoZIhvcNAQEN
BQAwZjELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdTZWF0dGxl
MQ0wCwYDVQQKDARFa3NhMRIwEAYDVQQLDAlQZXJzb25hbGwxFTATBgNVBAMMDDEw
LjgwLjE0OC41NjAeFw0yMzAyMTEwMDQ0MjRaFw0zMzAyMDgwMDQ0MjRaMGYxCzAJ
BgNVBAYTAlVTMQswCQYDVQQIDAJXQTEQMA4GA1UEBwwHU2VhdHRsZTENMAsGA1UE
CgwERWtzYTESMBAGA1UECwwJUGVyc29uYWxsMRUwEwYDVQQDDAwxMC44MC4xNDgu
NTYwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDpVLHBdFr7fh+PRQnZ
PlyXbIDfxdKCNZaNfxOl5I2eKW5m+zennt7N+nM1/5oDcVoIzVVCFkSmztl5GNr7
zyKfqDb/q6wZSQTOreNALBEh6redLnzc6OkQYnlFFuLcTuWLqTHdMoEJozbW+9K+
Z3lxKU92FvsTDaZKCT8NWwnKoTeXhEZtOF0KnJmzdQUztL7mNjSn53qUv/6WHwBG
/8F7elkJYP98jhhYkpKgPSnpoDuay3zmQxsFXvh9+j9GztODoroZgkNBuooObsnE
CQEFJLGZ03XAkyaumzfjSD5Ma4QQZPy0VwV2NHL3ngec+wxUH7u+FqupWbQZtJP6
+I3jSGGmhI2G/NIJZD0jiytlR1YmoUYM5qHl/VvrcqKoMEIGgF1ktYup9NAdGzLA
AItBbjDY8Vl+TGMC0vDbrHtVReYOWfx7TBo5nPnBjC1yTtnoYIk8EkJaPY9AuI5V
/WJ866PrPo5dPw/EVu3kuVzG5VoYb+iqY/qENnDnNeP0rCYF/tgna8c6sBkJG2y9
vo8ZkLC22J1CTWNqpiLBg0B9Jn2WIGFKNlavj8RiLSjNMDDj0EzS8mxZqlE+gn+C
MsASCpOAtIKPlr+x+Bfd0xaG1HyzaJV/cFHW8p6+8UvE9Nl0SeTWd8RtG36zXxif
WBDWiUJ9clQSp0F4Yunxts3PFQIDAQABo1MwUTAdBgNVHQ4EFgQUT78n048MqxDw
DdtI+841JN+OQpEwHwYDVR0jBBgwFoAUT78n048MqxDwDdtI+841JN+OQpEwDwYD
VR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQ0FAAOCAgEAI3TT8ykTsX4Ra8asa8gl
DqVsz8t5HjOpA2nfAmO9GETg+StEW0j2FGglu+S4OhMx4TN4rYAPY0C78+5GHQwu
UAKxAVXigLXG2TmgszJ1OIBXIRdMlyc3gVm9j5Odcrm0WSGmSe47K8IgEVwlGEoS
CEUfKbMlVzEaPlP7YUOtiElzYi8D4ht9JfmRPi7PjUjVrWw0My+XPrFA4siU8dfh
BbvN9ybHmgltFqeZEqEv2wA9l6EpY71sArWsw+k5OYb2tXiXBxXY9LJzNDdJsNW9
8EP5rFIPvtoZMlE9qAHqy1kkqxcjvhcD3SD8zxJKCDLCJnVPG01k2siwlA46jcTT
xFZhEttbFcASpURjxkvBndXCktN2myWwokNGlf1hosxk2lG5DcySwHAhjXLJV6r2
l7X/CZpR38n/FXAKwiMQAPFZsLRU/EWBPTDlD1zjQSH8weGEj8+e9tQuTm+QsYrb
aJW3puZ84fEgYu/QMjGTuJzd+ZswMcLLyyn4Sm9nvchE8SdEUiF6L0Lc8+qgwmJU
idxqPeX4DMweDcskpZDPbfI6jnNorvGiWaLAYEJ4ntc3SP/lvbwXXOJLvhnRP+Ov
zphcd/PRLS7VpAhWOVbulPjB8DkX0PmvgaCeiTDuajMxq6ve64v+dCwUcqqbamC2
OelAabtJKd8B2BUsR7JRIN8=
-----END CERTIFICATE-----

-----BEGIN CERTIFICATE-----
MIIFrTCCA5WgAwIBAgIUFXtDg6MEuAA0Ns1Ah2pfVKDC/nIwDQYJKoZIhvcNAQEN
BQAwZjELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAldBMRAwDgYDVQQHDAdTZWF0dGxl
MQ0wCwYDVQQKDARFa3NhMRIwEAYDVQQLDAlQZXJzb25hbGwxFTATBgNVBAMMDDEw
LjgwLjE0OC41NjAeFw0yMzAyMTEwMDQ0MjRaFw0zMzAyMDgwMDQ0MjRaMGYxCzAJ
BgNVBAYTAlVTMQswCQYDVQQIDAJXQTEQMA4GA1UEBwwHU2VhdHRsZTENMAsGA1UE
CgwERWtzYTESMBAGA1UECwwJUGVyc29uYWxsMRUwEwYDVQQDDAwxMC44MC4xNDgu
NTYwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDpVLHBdFr7fh+PRQnZ
PlyXbIDfxdKCNZaNfxOl5I2eKW5m+zennt7N+nM1/5oDcVoIzVVCFkSmztl5GNr7
zyKfqDb/q6wZSQTOreNALBEh6redLnzc6OkQYnlFFuLcTuWLqTHdMoEJozbW+9K+
Z3lxKU92FvsTDaZKCT8NWwnKoTeXhEZtOF0KnJmzdQUztL7mNjSn53qUv/6WHwBG
/8F7elkJYP98jhhYkpKgPSnpoDuay3zmQxsFXvh9+j9GztODoroZgkNBuooObsnE
CQEFJLGZ03XAkyaumzfjSD5Ma4QQZPy0VwV2NHL3ngec+wxUH7u+FqupWbQZtJP6
+I3jSGGmhI2G/NIJZD0jiytlR1YmoUYM5qHl/VvrcqKoMEIGgF1ktYup9NAdGzLA
AItBbjDY8Vl+TGMC0vDbrHtVReYOWfx7TBo5nPnBjC1yTtnoYIk8EkJaPY9AuI5V
/WJ866PrPo5dPw/EVu3kuVzG5VoYb+iqY/qENnDnNeP0rCYF/tgna8c6sBkJG2y9
vo8ZkLC22J1CTWNqpiLBg0B9Jn2WIGFKNlavj8RiLSjNMDDj0EzS8mxZqlE+gn+C
MsASCpOAtIKPlr+x+Bfd0xaG1HyzaJV/cFHW8p6+8UvE9Nl0SeTWd8RtG36zXxif
WBDWiUJ9clQSp0F4Yunxts3PFQIDAQABo1MwUTAdBgNVHQ4EFgQUT78n048MqxDw
DdtI+841JN+OQpEwHwYDVR0jBBgwFoAUT78n048MqxDwDdtI+841JN+OQpEwDwYD
VR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQ0FAAOCAgEAI3TT8ykTsX4Ra8asa8gl
DqVsz8t5HjOpA2nfAmO9GETg+StEW0j2FGglu+S4OhMx4TN4rYAPY0C78+5GHQwu
UAKxAVXigLXG2TmgszJ1OIBXIRdMlyc3gVm9j5Odcrm0WSGmSe47K8IgEVwlGEoS
CEUfKbMlVzEaPlP7YUOtiElzYi8D4ht9JfmRPi7PjUjVrWw0My+XPrFA4siU8dfh
BbvN9ybHmgltFqeZEqEv2wA9l6EpY71sArWsw+k5OYb2tXiXBxXY9LJzNDdJsNW9
8EP5rFIPvtoZMlE9qAHqy1kkqxcjvhcD3SD8zxJKCDLCJnVPG01k2siwlA46jcTT
xFZhEttbFcASpURjxkvBndXCktN2myWwokNGlf1hosxk2lG5DcySwHAhjXLJV6r2
l7X/CZpR38n/FXAKwiMQAPFZsLRU/EWBPTDlD1zjQSH8weGEj8+e9tQuTm+QsYrb
aJW3puZ84fEgYu/QMjGTuJzd+ZswMcLLyyn4Sm9nvchE8SdEUiF6L0Lc8+qgwmJU
idxqPeX4DMweDcskpZDPbfI6jnNorvGiWaLAYEJ4ntc3SP/lvbwXXOJLvhnRP+Ov
zphcd/PRLS7VpAhWOVbulPjB8DkX0PmvgaCeiTDuajMxq6ve64v+dCwUcqqbamC2
OelAabtJKd8B2BUsR7JRIN8=
-----END CERTIFICATE-----`,
					},
				},
			},
			osFamily: Bottlerocket,
			wantErr:  "",
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
