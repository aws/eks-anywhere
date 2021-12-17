package crypto_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/crypto"
)

var validCipherSuitesString = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"

func TestSecureCipherSuiteNames(t *testing.T) {
	string := crypto.SecureCipherSuitesString()
	if !reflect.DeepEqual(string, validCipherSuitesString) {
		assert.Equal(t, validCipherSuitesString, string, "cipher suites don't match")
	}
}
