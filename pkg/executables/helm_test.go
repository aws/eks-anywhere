package executables_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/mocks"
)

type helmTest struct {
	*WithT
	ctx     context.Context
	h       *executables.Helm
	e       *mocks.MockExecutable
	envVars map[string]string
}

func newHelmTest(t *testing.T, opts ...executables.HelmOpt) *helmTest {
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

func newHelmTemplateTest(t *testing.T, opts ...executables.HelmOpt) *helmTemplateTest {
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
		tt.e, tt.ctx, "template", tt.ociURI, "--version", tt.version, "--insecure-skip-tls-verify", "--namespace", tt.namespace, "--kube-version", "1.22", "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, tt.values, "1.22")).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateSuccessWithRegistryMirror(t *testing.T) {
	tt := newHelmTemplateTest(t, executables.WithRegistryMirror("1.2.3.4:443"))
	ociRegistryMirror := "oci://1.2.3.4:443/account/charts"
	expectCommand(
		tt.e, tt.ctx, "template", ociRegistryMirror, "--version", tt.version, "--insecure-skip-tls-verify", "--namespace", tt.namespace, "--kube-version", "1.22", "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, ociRegistryMirror, tt.version, tt.namespace, tt.values, "1.22")).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateSuccessWithEnv(t *testing.T) {
	tt := newHelmTemplateTest(t, executables.WithEnv(map[string]string{
		"HTTPS_PROXY": "test1",
	}))
	expectedEnv := map[string]string{
		"HTTPS_PROXY":           "test1",
		"HELM_EXPERIMENTAL_OCI": "1",
	}
	expectCommand(
		tt.e, tt.ctx, "template", tt.ociURI, "--version", tt.version, "--insecure-skip-tls-verify", "--namespace", tt.namespace, "--kube-version", "1.22", "-f", "-",
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
		tt.e, tt.ctx, "pull", url, "--version", version, "--insecure-skip-tls-verify", "--destination", destinationFolder,
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.SaveChart(tt.ctx, url, version, destinationFolder)).To(Succeed())
}

func TestHelmInstallChartSuccess(t *testing.T) {
	tt := newHelmTest(t)
	chart := "chart"
	url := "url"
	version := "1.1"
	kubeconfig := "/root/.kube/config"
	values := []string{"key1=value1"}
	expectCommand(
		tt.e, tt.ctx, "install", chart, url, "--version", version, "--insecure-skip-tls-verify", "--set", "key1=value1", "--kubeconfig", kubeconfig,
	).withEnvVars(tt.envVars).to().Return(bytes.Buffer{}, nil)

	tt.Expect(tt.h.InstallChart(tt.ctx, chart, url, version, kubeconfig, values)).To(Succeed())
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
