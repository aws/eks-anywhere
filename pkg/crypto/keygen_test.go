package crypto_test

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/crypto"
)

func TestGenerateSSHKeyPairSuccess(t *testing.T) {
	privateFile := "private_file"
	publicFile := "public_file"
	user := "user"
	dir := "testdata"

	_, writer := test.NewWriter(t)
	c, err := crypto.NewKeyGenerator(writer)
	if err != nil {
		t.Fatalf("failed creating new key generator error = %v", err)
	}

	bytes, err := c.GenerateSSHKeyPair(dir, dir, privateFile, publicFile, user)
	if err != nil || bytes == nil {
		t.Fatalf("GenerateSSHKeyPair() error = %v wantErr = nil", err)
	}
}
