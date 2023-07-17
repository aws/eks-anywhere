package framework

import (
	"context"
	"os"
	"testing"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// providerTemplateNameGenerator is an interface for getting template name for each provider.
type providerTemplateNameGenerator interface {
	envVarForTemplate(os OS, eksDName string) string
	defaultNameForTemplate(os OS, eksDName string) string
	defaultEnvVarForTemplate(os OS, kubeVersion anywherev1.KubernetesVersion) string
	searchTemplate(ctx context.Context, template string) (string, error)
}

// templateCache is a map of template name.
type templateRegistry struct {
	cache     map[string]string
	generator providerTemplateNameGenerator
}

// templateForRelease tries to find a suitable template for a particular eks-a release, k8s version and OS family.
// It follows these steps:
//
// 1. Look for explicit configuration through an env var: "T_{provider}_TEMPLATE_{osFamily}_{eks-d version}"
// eg. T_CLOUDSTACK_TEMPLATE_REDHAT_KUBERNETES_1_23_EKS_22, T_VSPHERE_TEMPLATE_REDHAT_KUBERNETES_1_23_EKS_22
// This should be used for explicit configuration, mostly in local development for overrides.
//
// 2. If not present, look for a template if the default templates: "{eks-d version}-{osFamily}"
// eg. kubernetes-1-23-eks-22-redhat (CloudStack),  /SDDC-Datacenter/vm/Templates/kubernetes-1-23-eks-22-redhat (vSphere)
// This is what should be used most of the time in CI, the explicit configuration is not present but the right template has already been
// imported to cloudstack.
//
// 3. If the template doesn't exist, default to the value of the default template env vars: eg. "T_CLOUDSTACK_TEMPLATE_REDHAT_1_23".
// This is a catch all condition. Mostly for edge cases where the bundle has been updated with a new eks-d version, but the
// the new template hasn't been imported yet. It also preserves backwards compatibility.
func (tc *templateRegistry) templateForRelease(t *testing.T, release *releasev1.EksARelease, kubeVersion anywherev1.KubernetesVersion, operatingSystem OS) string {
	t.Helper()
	versionsBundle := readVersionsBundles(t, release, kubeVersion)
	eksDName := versionsBundle.EksD.Name

	templateEnvVarName := tc.generator.envVarForTemplate(operatingSystem, eksDName)
	cacheKey := templateEnvVarName
	if template, ok := tc.cache[cacheKey]; ok {
		t.Logf("Template for release found in cache, using %s template.", template)
		return template
	}

	template, ok := os.LookupEnv(templateEnvVarName)
	if ok && template != "" {
		t.Logf("Env var %s is set, using %s template", templateEnvVarName, template)
		tc.cache[cacheKey] = template
		return template
	}
	t.Logf("Env var %s is not set, trying default generated template name", templateEnvVarName)

	// Env var is not set, try default template name
	template = tc.generator.defaultNameForTemplate(operatingSystem, eksDName)
	if template != "" {
		foundTemplate, err := tc.generator.searchTemplate(context.Background(), template)
		if err != nil {
			t.Logf("Failed checking if default template exists: %v", err)
		}

		if foundTemplate != "" {
			t.Logf("Default template for release exists, using %s template.", template)
			tc.cache[cacheKey] = template
			return template
		}
		t.Logf("Default template %s for release doesn't exit.", template)
	}
	// Default template doesn't exist, try legacy generic env var
	// It is not guaranteed that this template will work for the given release, if they don't match the
	// same ekd-d release, the test will fail. This is just a catch all last try for cases where the new template
	// hasn't been imported with its own name but the default one matches the same eks-d release.
	templateEnvVarName = tc.generator.defaultEnvVarForTemplate(operatingSystem, kubeVersion)
	template, ok = os.LookupEnv(templateEnvVarName)
	if !ok || template == "" {
		t.Fatalf("Env var %s for default template is not set, can't determine which template to use", templateEnvVarName)
	}

	t.Logf("Env var %s is set, using %s template. There are no guarantees this template will be valid. Cluster validation might fail.", templateEnvVarName, template)

	tc.cache[cacheKey] = template
	return template
}
