package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAndValidateOIDCConfig(t *testing.T) {
	tests := []struct {
		testName       string
		fileName       string
		refName        string
		wantOIDCConfig *OIDCConfig
		wantErr        bool
	}{
		{
			testName:       "file doesn't exist",
			fileName:       "testdata/fake_file.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "not parseable file",
			fileName:       "testdata/not_parseable_oidcconfig.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
		{
			testName: "refName doesn't match",
			fileName: "testdata/cluster_1_19_oidc.yaml",
			refName:  "wrongName",
			wantErr:  true,
		},
		{
			testName: "valid OIDC",
			fileName: "testdata/cluster_1_19_oidc.yaml",
			refName:  "eksa-unit-test",
			wantOIDCConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OIDCConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: OIDCConfigSpec{
					ClientId:     "id11",
					GroupsClaim:  "claim1",
					GroupsPrefix: "prefix-for-groups",
					IssuerUrl:    "https://mydomain.com/issuer",
					RequiredClaims: []OIDCConfigRequiredClaim{
						{
							Claim: "sub",
							Value: "test",
						},
					},
					UsernameClaim:  "username-claim",
					UsernamePrefix: "username-prefix",
				},
			},
			wantErr: false,
		},
		{
			testName:       "empty client id",
			fileName:       "testdata/cluster_invalid_oidc_null_clientid.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "null issuer url",
			fileName:       "testdata/cluster_invalid_oidc_null_issuer_url.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "invalid issuer url",
			fileName:       "testdata/cluster_invalid_oidc_bad_issuer_url.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "issuer url non https",
			fileName:       "testdata/cluster_invalid_oidc_issuer_url_non_https.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
		{
			testName:       "extra required claims",
			fileName:       "testdata/cluster_oidc_extra_required_claims.yaml",
			wantOIDCConfig: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetAndValidateOIDCConfig(tt.fileName, tt.refName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAndValidateOIDCConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantOIDCConfig) {
				t.Fatalf("GetAndValidateOIDCConfig() = %#v, want %#v", got, tt.wantOIDCConfig)
			}
		})
	}
}
