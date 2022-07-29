package curatedpackages_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

type configurationTest struct {
	*WithT
	validbp   *packagesv1.BundlePackage
	invalidbp *packagesv1.BundlePackage
}

func newConfigurationTest(t *testing.T) *configurationTest {
	validbp := &packagesv1.BundlePackage{
		Source: packagesv1.BundlePackageSource{
			Versions: []packagesv1.SourceVersion{
				{
					Configurations: []packagesv1.VersionConfiguration{
						{
							Name:     "schema",
							Default:  "H4sIAAAAAAAAA7WSP2/DIBDFd38KZHWMsZspytq9itqx6nC2zzYJ5lzAStMo372ASWXlj9QOHXm8e7/juGPCWPog6nTN0s7awazzvEMpKcOdyUAd9h1q5ANUO2jRcKdy6OGLFOwNr6jPTdVhD3xrSKWLkDYp80R/mUUj6TavNTQ2XxbLIntcxoSp2AorMZReNRENhyHcU7nFyk6axo9RaPSPeHNnpxgadYUv2Apj9SF14vsi8dZB04DaCjTOfLxpPuszmNOFagMs6DU2MErrryRVIDsydr0qVsXcYiotBivcWJztNTCYjhDWkGZxqDwNNaep9GcCf2niLvaJlAWhULMQe4EyY/lPtHPyBRA/BzLIrXSLpKCU/suuyCWRRFA30Q1Ig3czYbTkd7In9Qw93sj+zUfG+CQiUqhr4V8IcjNfntBKckq+Ab+VTzhCAwAA",
							Required: true,
						},
					},
				},
			},
		},
	}

	invalidbp := &packagesv1.BundlePackage{
		Source: packagesv1.BundlePackageSource{
			Versions: []packagesv1.SourceVersion{},
		},
	}

	return &configurationTest{
		WithT:     NewWithT(t),
		validbp:   validbp,
		invalidbp: invalidbp,
	}
}

func TestGetConfigurationsFromBundleSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.validbp)

	tt.Expect(err).To(BeNil())
	tt.Expect(len(configs)).To(Equal(5))
}

func TestGetConfigurationsFromBundleFail(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(nil)

	tt.Expect(err).NotTo(BeNil())
	tt.Expect(configs).To(BeNil())
}

func TestGetConfigurationsFromBundleWhenNoVersions(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.invalidbp)

	tt.Expect(err).NotTo(BeNil())
	tt.Expect(configs).To(BeNil())
}

func TestGetConfigurationsFromBundleFailWhenNoConfigs(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.invalidbp)

	tt.Expect(err).NotTo(BeNil())
	tt.Expect(configs).To(BeNil())
}

func TestUpdateConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.validbp)
	tt.Expect(err).To(BeNil())
	newConfigs := make(map[string]string)
	newConfigs["sourceRegistry"] = "127.0.0.1:8080"

	err = curatedpackages.UpdateConfigurations(configs, newConfigs)

	tt.Expect(err).To(BeNil())
	tt.Expect(configs["sourceRegistry"].Default).To(Equal(newConfigs["sourceRegistry"]))
}

func TestUpdateConfigurationsFail(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.validbp)
	tt.Expect(err).To(BeNil())
	newConfigs := make(map[string]string)
	newConfigs["registry"] = "127.0.0.1:8080"

	err = curatedpackages.UpdateConfigurations(configs, newConfigs)
	tt.Expect(err).NotTo(BeNil())
}

func TestGenerateAllValidConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.validbp)
	tt.Expect(err).To(BeNil())

	output, err := curatedpackages.GenerateAllValidConfigurations(configs)
	tt.Expect(err).To(BeNil())

	expectedOutput := fmt.Sprintf(
		"%s:\n  %s:\n    %s:\n      %s: %s\n    %s: %s\n%s: %s\n",
		"expose", "tls", "auto", "commonName", "localhost", "enabled", "false",
		"sourceRegistry", "localhost:8080",
	)

	tt.Expect(output).To(Equal(expectedOutput))
}

func TestGenerateDefaultConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs, err := curatedpackages.GetConfigurationsFromBundle(tt.validbp)
	tt.Expect(err).To(BeNil())

	output := curatedpackages.GenerateDefaultConfigurations(configs)
	expectedOutput := fmt.Sprintf("%s: %s\n",
		"sourceRegistry", "localhost:8080")

	tt.Expect(output).To(Equal(expectedOutput))
}

func TestParseConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)

	configs := []string{"registry=localhost:8080"}
	parsedConfigs, err := curatedpackages.ParseConfigurations(configs)

	tt.Expect(err).To(BeNil())
	tt.Expect(len(parsedConfigs)).To(Equal(1))
}

func TestParseConfigurationsFail(t *testing.T) {
	tt := newConfigurationTest(t)

	configs := []string{"registry"}
	parsedConfigs, err := curatedpackages.ParseConfigurations(configs)

	tt.Expect(err).NotTo(BeNil())
	tt.Expect(len(parsedConfigs)).To(Equal(0))
}
