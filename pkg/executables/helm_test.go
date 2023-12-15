package executables_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

type helmTest struct {
	*WithT
	ctx     context.Context
	h       helm.Client
	e       *mocks.MockExecutable
	envVars map[string]string
}

func newHelmTest(t *testing.T, opts ...helm.Opt) *helmTest {
	ctrl := gomock.NewController(t)
	e := mocks.NewMockExecutable(ctrl)
	return &helmTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		h:     executables.NewHelm(e, opts...),
		e:     e,
		envVars: map[string]string{
			"HELM_EXPERIMENTAL_OCI": "1",
		},
	}
}

type helmTemplateTest struct {
	*helmTest
	values                     interface{}
	valuesYaml                 []byte
	ociURI, version, namespace string
	wantTemplateContent        []byte
}

func newHelmTemplateTest(t *testing.T, opts ...helm.Opt) *helmTemplateTest {
	return &helmTemplateTest{
		helmTest: newHelmTest(t, opts...),
		values: map[string]string{
			"key1": "values1",
			"key2": "values2",
		},
		valuesYaml: []byte(`key1: values1
key2: values2
`,
		),
		ociURI:              "oci://public.ecr.aws/account/charts",
		version:             "1.1.1",
		namespace:           "kube-system",
		wantTemplateContent: []byte("template-content"),
	}
}

func TestHelmTemplateSuccess(t *testing.T) {
	tt := newHelmTemplateTest(t)
	expectCommand(
		tt.e, tt.ctx, "template", tt.ociURI, "--version", tt.version, "--namespace", tt.namespace, "--kube-version", "1.22", "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, tt.values, "1.22")).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateSuccessWithInsecure(t *testing.T) {
	tt := newHelmTemplateTest(t, helm.WithInsecure())
	expectCommand(
		tt.e, tt.ctx, "template", tt.ociURI, "--version", tt.version, "--namespace", tt.namespace, "--kube-version", "1.22", "--insecure-skip-tls-verify", "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, tt.values, "1.22")).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateSuccessWithRegistryMirror(t *testing.T) {
	tt := newHelmTemplateTest(t, helm.WithRegistryMirror(&registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:443",
	}))
	ociRegistryMirror := "oci://1.2.3.4:443/account/charts"
	expectCommand(
		tt.e, tt.ctx, "template", ociRegistryMirror, "--version", tt.version, "--namespace", tt.namespace, "--kube-version", "1.22", "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, ociRegistryMirror, tt.version, tt.namespace, tt.values, "1.22")).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateSuccessWithEnv(t *testing.T) {
	tt := newHelmTemplateTest(t, helm.WithProxyConfig(map[string]string{
		"HTTPS_PROXY": "test1",
	}))
	expectedEnv := map[string]string{
		"HTTPS_PROXY":           "test1",
		"HELM_EXPERIMENTAL_OCI": "1",
	}
	expectCommand(
		tt.e, tt.ctx, "template", tt.ociURI, "--version", tt.version, "--namespace", tt.namespace, "--kube-version", "1.22", "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(expectedEnv).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, tt.values, "1.22")).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateErrorYaml(t *testing.T) {
	tt := newHelmTemplateTest(t)
	values := func() {}

	_, gotErr := tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, values, "1.22")
	tt.Expect(gotErr).To(HaveOccurred(), "helm.Template() should fail marshalling values to yaml")
	tt.Expect(gotErr).To(MatchError(ContainSubstring("failed marshalling values for helm template: error marshaling into JSON")))
}

func TestHelmSaveChartSuccess(t *testing.T) {
	tt := newHelmTest(t)
	url := "url"
	version := "1.1"
	destinationFolder := "folder"
	expectCommand(
		tt.e, tt.ctx, "pull", url, "--version", version, "--destination", destinationFolder,
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.SaveChart(tt.ctx, url, version, destinationFolder)).To(Succeed())
}

func TestHelmSaveChartSuccessWithInsecure(t *testing.T) {
	tt := newHelmTest(t, helm.WithInsecure())
	url := "url"
	version := "1.1"
	destinationFolder := "folder"
	expectCommand(
		tt.e, tt.ctx, "pull", url, "--version", version, "--destination", destinationFolder, "--insecure-skip-tls-verify",
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.SaveChart(tt.ctx, url, version, destinationFolder)).To(Succeed())
}

func TestHelmSkipCRDs(t *testing.T) {
	tt := newHelmTest(t)
	url := "url"
	version := "1.1"
	kubeconfig := "kubeconfig"
	chart := "chart"
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--skip-crds", "--kubeconfig", kubeconfig, "--create-namespace", "--namespace", constants.EksaPackagesName,
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChart(tt.ctx, chart, url, version, kubeconfig, constants.EksaPackagesName, "", true, nil)).To(Succeed())
}

func TestHelmInstallChartSuccess(t *testing.T) {
	tt := newHelmTest(t)
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	values := []string{"key1=value1"}
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--set", "key1=value1", "--kubeconfig", kubeconfig, "--create-namespace", "--namespace", "eksa-packages",
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChart(tt.ctx, chart, url, version, kubeconfig, "eksa-packages", "", false, values)).To(Succeed())
}

func TestHelmInstallChartSuccessWithValuesFile(t *testing.T) {
	tt := newHelmTest(t)
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	values := []string{"key1=value1"}
	valuesFileName := "values.yaml"
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--set", "key1=value1", "--kubeconfig", kubeconfig, "--create-namespace", "--namespace", "eksa-packages", "-f", valuesFileName,
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChart(tt.ctx, chart, url, version, kubeconfig, "eksa-packages", valuesFileName, false, values)).To(Succeed())
}

func TestHelmInstallChartSuccessWithInsecure(t *testing.T) {
	tt := newHelmTest(t, helm.WithInsecure())
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	values := []string{"key1=value1"}
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--set", "key1=value1", "--kubeconfig", kubeconfig, "--create-namespace", "--namespace", "eksa-packages", "--insecure-skip-tls-verify",
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChart(tt.ctx, chart, url, version, kubeconfig, "eksa-packages", "", false, values)).To(Succeed())
}

func TestHelmInstallChartSuccessWithInsecureAndValuesFile(t *testing.T) {
	tt := newHelmTest(t, helm.WithInsecure())
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	values := []string{"key1=value1"}
	valuesFileName := "values.yaml"
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--set", "key1=value1", "--kubeconfig", kubeconfig, "--create-namespace", "--namespace", "eksa-packages", "-f", valuesFileName, "--insecure-skip-tls-verify",
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChart(tt.ctx, chart, url, version, kubeconfig, "eksa-packages", valuesFileName, false, values)).To(Succeed())
}

func TestHelmGetValueArgs(t *testing.T) {
	tests := []struct {
		testName       string
		values         []string
		wantValuesArgs []string
	}{
		{
			testName:       "single Helm value override",
			values:         []string{"key1=value1"},
			wantValuesArgs: []string{"--set", "key1=value1"},
		},
		{
			testName:       "multiple Helm value overrides",
			values:         []string{"key1=value1", "key2=value2", "key3=value3"},
			wantValuesArgs: []string{"--set", "key1=value1", "--set", "key2=value2", "--set", "key3=value3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if gotValuesArgs := executables.GetHelmValueArgs(tt.values); !sliceEqual(gotValuesArgs, tt.wantValuesArgs) {
				t.Errorf("GetHelmValueArgs() = %v, want %v", gotValuesArgs, tt.wantValuesArgs)
			}
		})
	}
}

func TestHelmInstallChartWithValuesFileSuccess(t *testing.T) {
	tt := newHelmTest(t)
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	valuesFileName := "values.yaml"
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--values", valuesFileName, "--kubeconfig", kubeconfig, "--wait",
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChartWithValuesFile(tt.ctx, chart, url, version, kubeconfig, valuesFileName)).To(Succeed())
}

func TestHelmInstallChartWithValuesFileSuccessWithInsecure(t *testing.T) {
	tt := newHelmTest(t, helm.WithInsecure())
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	valuesFileName := "values.yaml"
	expectCommand(
		tt.e, tt.ctx, "upgrade", "--install", chart, url, "--version", version, "--values", valuesFileName, "--kubeconfig", kubeconfig, "--wait", "--insecure-skip-tls-verify",
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChartWithValuesFile(tt.ctx, chart, url, version, kubeconfig, valuesFileName)).To(Succeed())
}

func TestHelmListCharts(t *testing.T) {
	tt := newHelmTest(t, helm.WithInsecure())
	kubeconfig := "/root/.kube/config"
	t.Run("Normal functionality", func(t *testing.T) {
		output := []byte("eks-anywhere-packages\n")
		expected := []string{"eks-anywhere-packages"}
		expectCommand(tt.e, tt.ctx, "list", "-q", "--kubeconfig", kubeconfig).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(output), nil)
		tt.Expect(tt.h.ListCharts(tt.ctx, kubeconfig)).To(Equal(expected))
	})

	t.Run("Empty output", func(t *testing.T) {
		expected := []string{}
		expectCommand(tt.e, tt.ctx, "list", "-q", "--kubeconfig", kubeconfig).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)
		tt.Expect(tt.h.ListCharts(tt.ctx, kubeconfig)).To(Equal(expected))
	})

	t.Run("Errored out", func(t *testing.T) {
		output := errors.New("Error")
		var expected []string
		expectCommand(tt.e, tt.ctx, "list", "-q", "--kubeconfig", kubeconfig).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, output)
		result, err := tt.h.ListCharts(tt.ctx, kubeconfig)
		tt.Expect(err).To(HaveOccurred())
		tt.Expect(result).To(Equal(expected))
	})
}

func TestHelmDelete(s *testing.T) {
	kubeconfig := "/root/.kube/config"

	s.Run("Success", func(t *testing.T) {
		tt := newHelmTest(s)
		installName := "test-install"
		expectCommand(tt.e, tt.ctx, "delete", installName, "--kubeconfig", kubeconfig).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)
		err := tt.h.Delete(tt.ctx, kubeconfig, installName, "")
		tt.Expect(err).NotTo(HaveOccurred())
	})

	s.Run("passes the namespace, if present", func(t *testing.T) {
		tt := newHelmTest(s)
		testNamespace := "testing"
		installName := "test-install"
		expectCommand(tt.e, tt.ctx, "delete", installName, "--kubeconfig", kubeconfig, "--namespace", testNamespace).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)
		err := tt.h.Delete(tt.ctx, kubeconfig, installName, testNamespace)
		tt.Expect(err).NotTo(HaveOccurred())
	})

	s.Run("passes the insecure skip flag", func(t *testing.T) {
		tt := newHelmTest(t, helm.WithInsecure())
		installName := "test-install"
		expectCommand(tt.e, tt.ctx, "delete", installName, "--kubeconfig", kubeconfig, "--insecure-skip-tls-verify").withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)
		err := tt.h.Delete(tt.ctx, kubeconfig, installName, "")
		tt.Expect(err).NotTo(HaveOccurred())
	})

	s.Run("returns errors from the helm executable", func(t *testing.T) {
		tt := newHelmTest(s)
		installName := "test-install"
		expectCommand(tt.e, tt.ctx, "delete", installName, "--kubeconfig", kubeconfig).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, fmt.Errorf("test error"))
		err := tt.h.Delete(tt.ctx, kubeconfig, installName, "")
		tt.Expect(err).To(HaveOccurred())
	})
}

func TestHelmRegistryLoginSuccess(t *testing.T) {
	tt := newHelmTest(t)
	registry := "1.2.3.4:5050"
	username := "username"
	password := "password"

	expectCommand(tt.e, tt.ctx, "registry", "login", registry, "--username", username, "--password-stdin").withEnvVars(tt.envVars).withStdIn([]byte(password)).to().Return(bytes.Buffer{}, nil)
	tt.Expect(tt.h.RegistryLogin(tt.ctx, registry, username, password)).To(Succeed())
}

func TestHelmRegistryLoginSuccessWithInsecure(t *testing.T) {
	tt := newHelmTest(t, helm.WithInsecure())
	registry := "1.2.3.4:5050"
	username := "username"
	password := "password"

	expectCommand(tt.e, tt.ctx, "registry", "login", registry, "--username", username, "--password-stdin", "--insecure").withEnvVars(tt.envVars).withStdIn([]byte(password)).to().Return(bytes.Buffer{}, nil)
	tt.Expect(tt.h.RegistryLogin(tt.ctx, registry, username, password)).To(Succeed())
}
