package oidc_test

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/oidc"
	"github.com/aws/eks-anywhere/internal/test"
)

func TestGenerateMinimalProvider(t *testing.T) {
	tests := []struct {
		testName          string
		issuerURL         string
		wantDiscoveryFile string
	}{
		{
			testName:          "s3 provider",
			issuerURL:         "https://s3-us-west-2.amazonaws.com/oidc-test-84hfke94j49",
			wantDiscoveryFile: "testdata/s3_discovery_expected.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := oidc.GenerateMinimalProvider(tt.issuerURL)
			if err != nil {
				t.Fatalf("GenerateServerFiles() error = %v, want nil", err)
			}

			test.AssertContentToFile(t, string(got.Discovery), tt.wantDiscoveryFile)

			block, _ := pem.Decode(got.PrivateKey)
			_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				t.Fatalf("GenerateServerFiles() = PrivateKey is not a valid rsa key: %v", err)
			}

			gotKeys := make(map[string][]map[string]string)
			err = json.Unmarshal(got.Keys, &gotKeys)
			if err != nil {
				t.Errorf("GenerateServerFiles() = Keys is not a valid json: %v", err)
			}

			keys := gotKeys["keys"]

			if len(keys) != 1 {
				t.Errorf("GenerateServerFiles() = len(Keys.Keys) = %d, want 1", len(keys))
			}

			key := keys[0]

			if key["kid"] != got.KeyID {
				t.Errorf("GenerateServerFiles() = Keys.keys[0].kid = %s, want %s", key["kid"], got.KeyID)
			}

			if key["kty"] != "RSA" {
				t.Errorf("GenerateServerFiles() = Keys.keys[0].kty = %s, want %s", key["kty"], "RSA")
			}

			if key["alg"] != "RS256" {
				t.Errorf("GenerateServerFiles() = Keys.keys[0].alg = %s, want %s", key["alg"], "RS256")
			}

			if key["use"] != "sig" {
				t.Errorf("GenerateServerFiles() = Keys.keys[0].use = %s, want %s", key["use"], "sig")
			}

			if key["e"] != "AQAB" {
				t.Errorf("GenerateServerFiles() = Keys.keys[0].e = %s, want %s", key["e"], "AQAB")
			}
		})
	}
}
