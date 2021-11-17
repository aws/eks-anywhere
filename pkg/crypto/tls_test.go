package crypto_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/crypto"
)

var validCipherSuites = []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}

func TestSecureCipherSuiteNames(t *testing.T) {
	list := crypto.SecureCipherSuiteNames()
	if !reflect.DeepEqual(list, validCipherSuites) {
		assert.Equal(t, validCipherSuites, list, "expected cipher suites to be list of valid")
	}
}
