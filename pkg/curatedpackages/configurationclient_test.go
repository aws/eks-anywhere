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
	invalidbp *packagesv1.BundlePackage
	configs   map[string]string
}

func newConfigurationTest(t *testing.T) *configurationTest {
	invalidbp := &packagesv1.BundlePackage{
		Source: packagesv1.BundlePackageSource{
			Versions: []packagesv1.SourceVersion{},
		},
	}

	config := map[string]string{
		"expose.tls.auto.commonName": "localhost",
		"expose.tls.enabled":         "false",
		"sourceRegistry":             "localhost:8080",
		"title":                      "",
		"subtitle":                   "",
	}

	return &configurationTest{
		WithT:     NewWithT(t),
		configs:   config,
		invalidbp: invalidbp,
	}
}

func TestGenerateAllValidConfigurationsSuccess(t *testing.T) {
	tt := newConfigurationTest(t)

	output, err := curatedpackages.GenerateAllValidConfigurations(tt.configs)
	tt.Expect(err).To(BeNil())

	expectedOutput := fmt.Sprintf(
		"%s:\n  %s:\n    %s:\n      %s: %s\n    %s: %s\n%s: %s\n",
		"expose", "tls", "auto", "commonName", "localhost", "enabled", "false",
		"sourceRegistry", "localhost:8080",
	)

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
