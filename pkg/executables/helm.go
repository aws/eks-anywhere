package executables

import (
	"context"
	"fmt"

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
}

type HelmOpt func(*Helm)

func WithRegistryMirror(mirror string) HelmOpt {
	return func(h *Helm) {
		h.registryMirror = mirror
	}
}

// join the default and the provided maps together
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

	result, err := h.executable.Command(
		ctx, "template", h.url(ociURI), "--version", version, insecureSkipVerifyFlag, "--namespace", namespace, "--kube-version", kubeVersion, "-f", "-",
	).WithStdIn(valuesYaml).WithEnvVars(h.env).Run()
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (h *Helm) PullChart(ctx context.Context, ociURI, version string) error {
	_, err := h.executable.Command(ctx, "pull", h.url(ociURI), "--version", version, insecureSkipVerifyFlag).
		WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) PushChart(ctx context.Context, chart, registry string) error {
	logger.Info("Pushing", "chart", chart)
	_, err := h.executable.Command(ctx, "push", chart, registry, insecureSkipVerifyFlag).WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) RegistryLogin(ctx context.Context, registry, username, password string) error {
	logger.Info("Logging in to helm registry", "registry", registry)
	_, err := h.executable.Command(ctx, "registry", "login", registry, "--username", username, "--password", password, "--insecure").WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) SaveChart(ctx context.Context, ociURI, version, folder string) error {
	fmt.Println(h.env)
	_, err := h.executable.Command(ctx, "pull", h.url(ociURI), "--version", version, insecureSkipVerifyFlag, "--destination", folder).
		WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) InstallChartFromName(ctx context.Context, ociURI, kubeConfig, name, version string) error {
	_, err := h.executable.Command(ctx, "install", name, ociURI, "--version", version, insecureSkipVerifyFlag, "--kubeconfig", kubeConfig).
		WithEnvVars(h.env).Run()
	return err
}

func (h *Helm) InstallChart(ctx context.Context, chart, ociURI, version, kubeconfigFilePath string, values []string) error {
	valueArgs := GetHelmValueArgs(values)
	params := []string{"install", chart, ociURI, "--version", version, insecureSkipVerifyFlag}
	params = append(params, valueArgs...)
	params = append(params, "--kubeconfig", kubeconfigFilePath)

	logger.Info("Installing helm chart on cluster", "chart", chart, "version", version)
	_, err := h.executable.Command(ctx, params...).WithEnvVars(h.env).Run()
	return err
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
