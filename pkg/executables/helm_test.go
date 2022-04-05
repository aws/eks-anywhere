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

func newHelmTest(t *testing.T) *helmTest {
	ctrl := gomock.NewController(t)
	e := mocks.NewMockExecutable(ctrl)
	return &helmTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		h:     executables.NewHelm(e),
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

func newHelmTemplateTest(t *testing.T) *helmTemplateTest {
	return &helmTemplateTest{
		helmTest: newHelmTest(t),
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
		tt.e, tt.ctx, "template", tt.ociURI, "--version", tt.version, "--insecure-skip-tls-verify", "--namespace", tt.namespace, "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, tt.values)).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateSuccessWithRegistryMirror(t *testing.T) {
	tt := newHelmTemplateTest(t)
	tt.h = executables.NewHelm(tt.e, executables.WithRegistryMirror("1.2.3.4:443"))
	expectCommand(
		tt.e, tt.ctx, "template", "oci://1.2.3.4:443/account/charts", "--version", tt.version, "--insecure-skip-tls-verify", "--namespace", tt.namespace, "-f", "-",
	).withStdIn(tt.valuesYaml).withEnvVars(tt.envVars).to().Return(*bytes.NewBuffer(tt.wantTemplateContent), nil)

	tt.Expect(tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, tt.values)).To(Equal(tt.wantTemplateContent), "helm.Template() should succeed return correct template content")
}

func TestHelmTemplateErrorYaml(t *testing.T) {
	tt := newHelmTemplateTest(t)
	values := func() {}

	_, gotErr := tt.h.Template(tt.ctx, tt.ociURI, tt.version, tt.namespace, values)
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
