package executables

import (
	"context"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

const (
	helmPath               = "helm"
	insecureSkipVerifyFlag = "--insecure-skip-tls-verify"
)

type Helm struct {
	executable     Executable
	registryMirror string
	env            map[string]string
	insecure       bool
}

type HelmOpt func(*Helm)

func WithRegistryMirror(mirror string) HelmOpt {
	return func(h *Helm) {
		h.registryMirror = mirror
	}
}

func WithInsecure() HelmOpt {
	return func(h *Helm) {
		h.insecure = true
	}
}

// join the default and the provided maps together.
func WithEnv(env map[string]string) HelmOpt {
	return func(h *Helm) {
		for k, v := range env {
			h.env[k] = v
		}
	}
}

func NewHelm(executable Executable, opts ...HelmOpt) *Helm {
	h := &Helm{
		executable: executable,
		env: map[string]string{
			"HELM_EXPERIMENTAL_OCI": "1",
		},
		insecure: false,
	}

	for _, o := range opts {
		o(h)
	}

	return h
}

func (h *Helm) Template(ctx context.Context, ociURI, version, namespace string, values interface{}, kubeVersion string) ([]byte, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling values for helm template: %v", err)
	}

	params := []string{"template", h.url(ociURI), "--version", version, "--namespace", namespace, "--kube-version", kubeVersion}
	params = h.addInsecureFlagIfProvided(params)
	params = append(params, "-f", "-")

	result, err := h.executable.Command(ctx, params...).WithStdIn(valuesYaml).WithEnvVars(h.env).Run()
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (h *Helm) PullChart(ctx context.Context, ociURI, version string) error {
	params := []string{"pull", h.url(ociURI), "--version", version}
	params = h.addInsecureFlagIfProvided(params)
	_, err := h.executable.Command(ctx, params...).
		WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) PushChart(ctx context.Context, chart, registry string) error {
	logger.Info("Pushing", "chart", chart)
	params := []string{"push", chart, registry}
	params = h.addInsecureFlagIfProvided(params)
	_, err := h.executable.Command(ctx, params...).WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) RegistryLogin(ctx context.Context, registry, username, password string) error {
	logger.Info("Logging in to helm registry", "registry", registry)
	params := []string{"registry", "login", registry, "--username", username, "--password", password}
	if h.insecure {
		params = append(params, "--insecure")
	}
	_, err := h.executable.Command(ctx, params...).WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) SaveChart(ctx context.Context, ociURI, version, folder string) error {
	params := []string{"pull", h.url(ociURI), "--version", version, "--destination", folder}
	params = h.addInsecureFlagIfProvided(params)
	_, err := h.executable.Command(ctx, params...).
		WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) InstallChartFromName(ctx context.Context, ociURI, kubeConfig, name, version string) error {
	params := []string{"install", name, ociURI, "--version", version, "--kubeconfig", kubeConfig}
	params = h.addInsecureFlagIfProvided(params)
	_, err := h.executable.Command(ctx, params...).
		WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, namespace string, values []string) error {
	valueArgs := GetHelmValueArgs(values)
	params := []string{"install", chart, ociURI, "--version", version}
	params = append(params, valueArgs...)
	params = append(params, "--kubeconfig", kubeconfigFilePath)
	if len(namespace) > 0 {
		params = append(params, "--create-namespace", "--namespace", namespace)
	}
	params = h.addInsecureFlagIfProvided(params)

	logger.Info("Installing helm chart on cluster", "chart", chart, "version", version)
	_, err := h.executable.Command(ctx, params...).WithEnvVars(h.env).Run()
	return err
}

// InstallChartWithValuesFile installs a helm chart with the provided values file and waits for the chart deployment to be ready
// The default timeout for the chart to reach ready state is 5m.
func (h *Helm) InstallChartWithValuesFile(ctx context.Context, chart, ociURI, version, kubeconfigFilePath, valuesFilePath string) error {
	params := []string{"install", chart, ociURI, "--version", version, "--values", valuesFilePath, "--kubeconfig", kubeconfigFilePath, "--wait"}
	params = h.addInsecureFlagIfProvided(params)
	_, err := h.executable.Command(ctx, params...).WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) ListCharts(ctx context.Context, kubeconfigFilePath string) ([]string, error) {
	params := []string{"list", "-q", "--kubeconfig", kubeconfigFilePath}
	out, err := h.executable.Command(ctx, params...).WithEnvVars(h.env).Run()
	if err != nil {
		return nil, err
	}
	charts := strings.FieldsFunc(out.String(), func(c rune) bool {
		return c == '\n'
	})
	return charts, nil
}

func (h *Helm) addInsecureFlagIfProvided(params []string) []string {
	if h.insecure {
		return append(params, insecureSkipVerifyFlag)
	}
	return params
}

func (h *Helm) url(originalURL string) string {
	return urls.ReplaceHost(originalURL, h.registryMirror)
}

func GetHelmValueArgs(values []string) []string {
	valueArgs := []string{}
	for _, value := range values {
		valueArgs = append(valueArgs, "--set", value)
	}

	return valueArgs
}
