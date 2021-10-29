package crypto_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/crypto"
)

func TestGenerateIamAuthSelfSignCertKeyPairSuccess(t *testing.T) {
	certGen := crypto.NewCertificateGenerator()
	_, _, err := certGen.GenerateIamAuthSelfSignCertKeyPair()
	if err != nil {
		t.Fatalf("certificategenerator.GenerateIamAuthSelfSignCertKeyPair()\n error = %v\n wantErr = nil", err)
	}
}
