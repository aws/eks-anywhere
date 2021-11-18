package clusterapi_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/templater"
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
