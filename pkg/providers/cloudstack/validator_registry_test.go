package cloudstack_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

func getFromRegistry(registry cloudstack.ValidatorRegistry, execConfig *decoder.CloudStackExecConfig) (*cloudstack.Validator, error) {
	return registry.Get(execConfig)
}

func TestRegistryGetWithNilExecConfig(t *testing.T) {
	g := NewGomegaWithT(t)
	registry := cloudstack.NewValidatorFactory(executables.NewLocalExecutablesBuilder(), nil, false)
	_, err := getFromRegistry(registry, nil)
	g.Expect(err).NotTo(BeNil())
}

func TestRegistryGetSuccess(t *testing.T) {
	g := NewGomegaWithT(t)
	registry := cloudstack.NewValidatorFactory(executables.NewLocalExecutablesBuilder(), nil, false)
	_, err := getFromRegistry(registry, &decoder.CloudStackExecConfig{Profiles: []decoder.CloudStackProfileConfig{
		{
			Name:          "test",
			ApiKey:        "apikey",
			SecretKey:     "secretKey",
			ManagementUrl: "test-url",
		},
	}})
	g.Expect(err).To(BeNil())
}
