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
							Name:     "sourceRegistry",
							Default:  "localhost:8080",
							Required: true,
						},
						{
							Name:     "title",
							Default:  "",
							Required: false,
						},
						{
							Name:     "subtitle",
							Default:  "",
							Required: false,
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
	configs := curatedpackages.GetConfigurationsFromBundle(tt.validbp)

	tt.Expect(len(configs)).To(Equal(3))
}

func TestGetConfigurationsFromBundleFail(t *testing.T) {
	tt := newConfigurationTest(t)
	configs := curatedpackages.GetConfigurationsFromBundle(nil)

	tt.Expect(len(configs)).To(Equal(0))
}

func TestGetConfigurationsFromBundleFailWhenNoConfigs(t *testing.T) {
	tt := newConfigurationTest(t)
	configs := curatedpackages.GetConfigurationsFromBundle(tt.invalidbp)

	tt.Expect(len(configs)).To(Equal(0))
}

func TestUpdateConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs := curatedpackages.GetConfigurationsFromBundle(tt.validbp)
	newConfigs := make(map[string]string)
	newConfigs["sourceRegistry"] = "127.0.0.1:8080"

	err := curatedpackages.UpdateConfigurations(configs, newConfigs)

	tt.Expect(err).To(BeNil())
	tt.Expect(configs["sourceRegistry"].Default).To(Equal(newConfigs["sourceRegistry"]))
}

func TestUpdateConfigurationsFail(t *testing.T) {
	tt := newConfigurationTest(t)
	configs := curatedpackages.GetConfigurationsFromBundle(tt.validbp)
	newConfigs := make(map[string]string)
	newConfigs["registry"] = "127.0.0.1:8080"

	err := curatedpackages.UpdateConfigurations(configs, newConfigs)

	tt.Expect(err).NotTo(BeNil())
}

func TestGenerateAllValidConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs := curatedpackages.GetConfigurationsFromBundle(tt.validbp)

	output := curatedpackages.GenerateAllValidConfigurations(configs)
	expectedOutput := fmt.Sprintf("%s: \"%s\"\n",
		"sourceRegistry", "localhost:8080")

	tt.Expect(output).To(Equal(expectedOutput))
}

func TestGenerateDefaultConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)
	configs := curatedpackages.GetConfigurationsFromBundle(tt.validbp)

	output := curatedpackages.GenerateDefaultConfigurations(configs)
	expectedOutput := fmt.Sprintf("%s: \"%s\"\n",
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
