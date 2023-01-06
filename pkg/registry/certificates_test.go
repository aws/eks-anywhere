package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
)

func TestGetCertificatesSuccess(t *testing.T) {
	result, err := registry.GetCertificates("testdata/harbor.eksa.demo.crt")
	assert.NotNil(t, result)
	assert.NoError(t, err)
}

func TestGetCertificatesNothing(t *testing.T) {
	result, err := registry.GetCertificates("")
	assert.Nil(t, result)
	assert.NoError(t, err)
}

func TestGetCertificatesError(t *testing.T) {
	result, err := registry.GetCertificates("bogus.crt")
	assert.Nil(t, result)
	assert.EqualError(t, err, "error reading certificate file <bogus.crt>: open bogus.crt: no such file or directory")
}
