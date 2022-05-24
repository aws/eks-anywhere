package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateOIDCConfig(t *testing.T) {
	tests := []struct {
		testName   string
		oidcConfig *OIDCConfig
		wantErr    bool
	}{
		{
			testName: "valid OIDC",
			oidcConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       OIDCConfigKind,
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
			testName: "empty client id",
			oidcConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       OIDCConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: OIDCConfigSpec{
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
			wantErr: true,
		},
		{
			testName: "null issuer url",
			oidcConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       OIDCConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: OIDCConfigSpec{
					ClientId:     "id11",
					GroupsClaim:  "claim1",
					GroupsPrefix: "prefix-for-groups",
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
			wantErr: true,
		},
		{
			testName: "invalid issuer url",
			oidcConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       OIDCConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: OIDCConfigSpec{
					ClientId:     "id11",
					GroupsClaim:  "claim1",
					GroupsPrefix: "prefix-for-groups",
					IssuerUrl:    "mydomain./issuer",
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
			wantErr: true,
		},
		{
			testName: "issuer url non https",
			oidcConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       OIDCConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: OIDCConfigSpec{
					ClientId:     "id11",
					GroupsClaim:  "claim1",
					GroupsPrefix: "prefix-for-groups",
					IssuerUrl:    "http://mydomain.com/issuer",
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
			wantErr: true,
		},
		{
			testName: "extra required claims",
			oidcConfig: &OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       OIDCConfigKind,
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
						{
							Claim: "sub2",
							Value: "test2",
						},
					},
					UsernameClaim:  "username-claim",
					UsernamePrefix: "username-prefix",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := tt.oidcConfig.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("OIDCConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
