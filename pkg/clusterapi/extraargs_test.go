package clusterapi_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestOIDCToExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		oidc     *v1alpha1.OIDCConfig
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "no oidc",
			oidc:     nil,
			want:     clusterapi.ExtraArgs{},
		},
		{
			testName: "minimal oidc with zero values",
			oidc: &v1alpha1.OIDCConfig{
				Spec: v1alpha1.OIDCConfigSpec{
					ClientId:       "my-client-id",
					IssuerUrl:      "https://mydomain.com/issuer",
					RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{{}},
					GroupsClaim:    "",
				},
			},
			want: clusterapi.ExtraArgs{
				"oidc-client-id":  "my-client-id",
				"oidc-issuer-url": "https://mydomain.com/issuer",
			},
		},
		{
			testName: "minimal oidc with nil values",
			oidc: &v1alpha1.OIDCConfig{
				Spec: v1alpha1.OIDCConfigSpec{
					ClientId:       "my-client-id",
					IssuerUrl:      "https://mydomain.com/issuer",
					RequiredClaims: nil,
				},
			},
			want: clusterapi.ExtraArgs{
				"oidc-client-id":  "my-client-id",
				"oidc-issuer-url": "https://mydomain.com/issuer",
			},
		},
		{
			testName: "full oidc",
			oidc: &v1alpha1.OIDCConfig{
				Spec: v1alpha1.OIDCConfigSpec{
					ClientId:     "my-client-id",
					IssuerUrl:    "https://mydomain.com/issuer",
					GroupsClaim:  "claim1",
					GroupsPrefix: "prefix-for-groups",
					RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{{
						Claim: "sub",
						Value: "test",
					}},
					UsernameClaim:  "username-claim",
					UsernamePrefix: "username-prefix",
				},
			},
			want: clusterapi.ExtraArgs{
				"oidc-client-id":       "my-client-id",
				"oidc-groups-claim":    "claim1",
				"oidc-groups-prefix":   "prefix-for-groups",
				"oidc-issuer-url":      "https://mydomain.com/issuer",
				"oidc-required-claim":  "sub=test",
				"oidc-username-claim":  "username-claim",
				"oidc-username-prefix": "username-prefix",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.OIDCToExtraArgs(tt.oidc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OIDCToExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtraArgsAddIfNotEmpty(t *testing.T) {
	tests := []struct {
		testName  string
		e         clusterapi.ExtraArgs
		k         string
		v         string
		wantAdded bool
		wantV     string
	}{
		{
			testName:  "add string",
			e:         clusterapi.ExtraArgs{},
			k:         "key",
			v:         "value",
			wantAdded: true,
			wantV:     "value",
		},
		{
			testName:  "add empty string",
			e:         clusterapi.ExtraArgs{},
			k:         "key",
			v:         "",
			wantAdded: false,
			wantV:     "",
		},
		{
			testName: "add present string",
			e: clusterapi.ExtraArgs{
				"key": "value_old",
			},
			k:         "key",
			v:         "value_new",
			wantAdded: true,
			wantV:     "value_new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tt.e.AddIfNotEmpty(tt.k, tt.v)

			gotV, gotAdded := tt.e[tt.k]
			if tt.wantAdded != gotAdded {
				t.Errorf("ExtraArgs.AddIfNotZero() wasAdded = %v, wantAdded %v", gotAdded, tt.wantAdded)
			}

			if gotV != tt.wantV {
				t.Errorf("ExtraArgs.AddIfNotZero() gotValue = %v, wantValue %v", gotV, tt.wantV)
			}
		})
	}
}

func TestExtraArgsToPartialYaml(t *testing.T) {
	tests := []struct {
		testName string
		e        clusterapi.ExtraArgs
		want     templater.PartialYaml
	}{
		{
			testName: "valid args",
			e: clusterapi.ExtraArgs{
				"oidc-client-id":       "my-client-id",
				"oidc-groups-claim":    "claim1,claim2",
				"oidc-groups-prefix":   "prefix-for-groups",
				"oidc-issuer-url":      "https://mydomain.com/issuer",
				"oidc-required-claim":  "hd=example.com,sub=test",
				"oidc-signing-algs":    "ES256,HS256",
				"oidc-username-claim":  "username-claim",
				"oidc-username-prefix": "username-prefix",
			},
			want: templater.PartialYaml{
				"oidc-client-id":       "my-client-id",
				"oidc-groups-claim":    "claim1,claim2",
				"oidc-groups-prefix":   "prefix-for-groups",
				"oidc-issuer-url":      "https://mydomain.com/issuer",
				"oidc-required-claim":  "hd=example.com,sub=test",
				"oidc-signing-algs":    "ES256,HS256",
				"oidc-username-claim":  "username-claim",
				"oidc-username-prefix": "username-prefix",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := tt.e.ToPartialYaml(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtraArgs.ToPartialYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAwsIamAuthExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		awsiam   *v1alpha1.AWSIamConfig
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "no aws iam",
			awsiam:   nil,
			want:     clusterapi.ExtraArgs{},
		},
		{
			testName: "with aws iam config",
			awsiam:   &v1alpha1.AWSIamConfig{},
			want: clusterapi.ExtraArgs{
				"authentication-token-webhook-config-file": "/etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.AwsIamAuthExtraArgs(tt.awsiam); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AwsIamAuthExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodIAMConfigExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		podIAM   *v1alpha1.PodIAMConfig
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "no pod IAM config",
			podIAM:   nil,
			want:     nil,
		},
		{
			testName: "with pod IAM config",
			podIAM:   &v1alpha1.PodIAMConfig{ServiceAccountIssuer: "https://test"},
			want: clusterapi.ExtraArgs{
				"service-account-issuer": "https://test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.PodIAMAuthExtraArgs(tt.podIAM); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodIAMAuthExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolvConfExtraArgs(t *testing.T) {
	tests := []struct {
		testName   string
		resolvConf *v1alpha1.ResolvConf
		want       clusterapi.ExtraArgs
	}{
		{
			testName:   "default",
			resolvConf: &v1alpha1.ResolvConf{Path: ""},
			want:       map[string]string{},
		},
		{
			testName:   "with custom resolvConf file",
			resolvConf: &v1alpha1.ResolvConf{Path: "mypath"},
			want: clusterapi.ExtraArgs{
				"resolv-conf": "mypath",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.ResolvConfExtraArgs(tt.resolvConf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolvConfExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecureTlsCipherSuitesExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "default",
			want: clusterapi.ExtraArgs{
				"tls-cipher-suites": crypto.SecureCipherSuitesString(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.SecureTlsCipherSuitesExtraArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecureTlsCipherSuitesExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecureEtcdTlsCipherSuitesExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "default",
			want: clusterapi.ExtraArgs{
				"cipher-suites": crypto.SecureCipherSuitesString(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecureEtcdTlsCipherSuitesExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCgroupDriverCgroupfsExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "default",
			want: clusterapi.ExtraArgs{
				"cgroup-driver": "cgroupfs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.CgroupDriverCgroupfsExtraArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CgroupDriverCgroupfsExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCgroupDriverSystemdExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "default",
			want: clusterapi.ExtraArgs{
				"cgroup-driver": "systemd",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.CgroupDriverSystemdExtraArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CgroupDriverSystemdExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeLabelsExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		wnc      v1alpha1.WorkerNodeGroupConfiguration
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "no labels",
			wnc: v1alpha1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(3),
			},
			want: clusterapi.ExtraArgs{},
		},
		{
			testName: "with labels",
			wnc: v1alpha1.WorkerNodeGroupConfiguration{
				Count:  ptr.Int(3),
				Labels: map[string]string{"label1": "foo", "label2": "bar"},
			},
			want: clusterapi.ExtraArgs{
				"node-labels": "label1=foo,label2=bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.WorkerNodeLabelsExtraArgs(tt.wnc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WorkerNodeLabelsExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCpNodeLabelsExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		cpc      v1alpha1.ControlPlaneConfiguration
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "no labels",
			cpc: v1alpha1.ControlPlaneConfiguration{
				Count: 3,
			},
			want: clusterapi.ExtraArgs{},
		},
		{
			testName: "with labels",
			cpc: v1alpha1.ControlPlaneConfiguration{
				Count:  3,
				Labels: map[string]string{"label1": "foo", "label2": "bar"},
			},
			want: clusterapi.ExtraArgs{
				"node-labels": "label1=foo,label2=bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.ControlPlaneNodeLabelsExtraArgs(tt.cpc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControlPlaneNodeLabelsExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppend(t *testing.T) {
	tests := []struct {
		testName string
		e        clusterapi.ExtraArgs
		a        clusterapi.ExtraArgs
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "initially empty",
			e:        clusterapi.ExtraArgs{},
			a: clusterapi.ExtraArgs{
				"key1": "value1",
			},
			want: clusterapi.ExtraArgs{
				"key1": "value1",
			},
		},
		{
			testName: "initially not empty",
			e: clusterapi.ExtraArgs{
				"key1": "value1",
			},
			a: clusterapi.ExtraArgs{
				"key2": "value2",
			},
			want: clusterapi.ExtraArgs{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			testName: "append nil extraArgs",
			e: clusterapi.ExtraArgs{
				"key1": "value1",
			},
			a: nil,
			want: clusterapi.ExtraArgs{
				"key1": "value1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := tt.e.Append(tt.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtraArgs.Append() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeCIDRMaskExtraArgs(t *testing.T) {
	nodeCidrMaskSize := new(int)
	*nodeCidrMaskSize = 28
	tests := []struct {
		testName       string
		clusterNetwork *v1alpha1.ClusterNetwork
		want           clusterapi.ExtraArgs
	}{
		{
			testName:       "no cluster network config",
			clusterNetwork: nil,
			want:           nil,
		},
		{
			testName: "no nodes config",
			clusterNetwork: &v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{CidrBlocks: []string{"test", "test"}},
			},
			want: nil,
		},
		{
			testName: "with nodes config",
			clusterNetwork: &v1alpha1.ClusterNetwork{
				Nodes: &v1alpha1.Nodes{CIDRMaskSize: nodeCidrMaskSize},
			},
			want: clusterapi.ExtraArgs{
				"node-cidr-mask-size": "28",
			},
		},
		{
			testName: "with nodes config empty",
			clusterNetwork: &v1alpha1.ClusterNetwork{
				Nodes: &v1alpha1.Nodes{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.NodeCIDRMaskExtraArgs(tt.clusterNetwork); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeCIDRMaskExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEtcdEncryptionExtraArgs(t *testing.T) {
	tests := []struct {
		name           string
		etcdEncryption *[]v1alpha1.EtcdEncryption
		want           clusterapi.ExtraArgs
	}{
		{
			name:           "nil config",
			etcdEncryption: nil,
			want:           clusterapi.ExtraArgs{},
		},
		{
			name:           "empty config",
			etcdEncryption: &[]v1alpha1.EtcdEncryption{},
			want:           clusterapi.ExtraArgs{},
		},
		{
			name: "one config",
			etcdEncryption: &[]v1alpha1.EtcdEncryption{
				{
					Providers: []v1alpha1.EtcdEncryptionProvider{
						{
							KMS: &v1alpha1.KMS{
								Name:                "config1",
								SocketListenAddress: "unix:///var/run/kmsplugin/socket1-new.sock",
							},
						},
						{
							KMS: &v1alpha1.KMS{
								Name:                "config2",
								SocketListenAddress: "unix:///var/run/kmsplugin/socket1-old.sock",
							},
						},
					},
					Resources: []string{
						"secrets",
						"crd1.anywhere.eks.amazonsaws.com",
					},
				},
			},
			want: clusterapi.ExtraArgs{
				"encryption-provider-config": "/etc/kubernetes/enc/encryption-config.yaml",
			},
		},
		{
			name: "multiple configs",
			etcdEncryption: &[]v1alpha1.EtcdEncryption{
				{
					Providers: []v1alpha1.EtcdEncryptionProvider{
						{
							KMS: &v1alpha1.KMS{
								Name:                "config1",
								SocketListenAddress: "unix:///var/run/kmsplugin/socket1-new.sock",
							},
						},
						{
							KMS: &v1alpha1.KMS{
								Name:                "config2",
								SocketListenAddress: "unix:///var/run/kmsplugin/socket1-old.sock",
							},
						},
					},
					Resources: []string{
						"secrets",
						"crd1.anywhere.eks.amazonsaws.com",
					},
				},
				{
					Providers: []v1alpha1.EtcdEncryptionProvider{
						{
							KMS: &v1alpha1.KMS{
								Name:                "config3",
								SocketListenAddress: "unix:///var/run/kmsplugin/socket2-new.sock",
							},
						},
						{
							KMS: &v1alpha1.KMS{
								Name:                "config4",
								SocketListenAddress: "unix:///var/run/kmsplugin/socket2-old.sock",
							},
						},
					},
					Resources: []string{
						"configmaps",
						"crd2.anywhere.eks.amazonsaws.com",
					},
				},
			},
			want: clusterapi.ExtraArgs{
				"encryption-provider-config": "/etc/kubernetes/enc/encryption-config.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			if got := clusterapi.EtcdEncryptionExtraArgs(tt.etcdEncryption); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EtcdEncryptionExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeatureGatesExtraArgs(t *testing.T) {
	tests := []struct {
		testName string
		features []string
		want     clusterapi.ExtraArgs
	}{
		{
			testName: "no feature gates",
			features: []string{},
			want:     nil,
		},
		{
			testName: "single feature gate",
			features: []string{"feature1=true"},
			want: clusterapi.ExtraArgs{
				"feature-gates": "feature1=true",
			},
		},
		{
			testName: "multiple feature gates",
			features: []string{"feature1=true", "feature2=false", "feature3=true"},
			want: clusterapi.ExtraArgs{
				"feature-gates": "feature1=true,feature2=false,feature3=true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if got := clusterapi.FeatureGatesExtraArgs(tt.features...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FeatureGatesExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
