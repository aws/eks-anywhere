package cloudstack_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/cmk"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

func TestRegistryGetWithNilExecConfig(t *testing.T) {
	g := NewGomegaWithT(t)
	executableBuilder := executables.NewLocalExecutablesBuilder()
	registry := cloudstack.NewValidatorFactory(cmk.NewCmkBuilder(executableBuilder), nil, false)
	_, err := registry.Get(nil)
	g.Expect(err).NotTo(BeNil())
}

func TestRegistryGetSuccess(t *testing.T) {
	g := NewGomegaWithT(t)
	executableBuilder := executables.NewLocalExecutablesBuilder()
	registry := cloudstack.NewValidatorFactory(cmk.NewCmkBuilder(executableBuilder), nil, false)
	validator, err := registry.Get(&decoder.CloudStackExecConfig{Profiles: []decoder.CloudStackProfileConfig{
		{
			Name:          "test",
			ApiKey:        "apikey",
			SecretKey:     "secretKey",
			ManagementUrl: "test-url",
		},
	}})
	g.Expect(err).To(BeNil())
	g.Expect(validator).ToNot(BeNil())
}
