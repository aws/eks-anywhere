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

var helmTemplateEnvVars = map[string]string{
	"HELM_EXPERIMENTAL_OCI": "1",
}

type Helm struct {
	executable     Executable
	registryMirror string
}

type HelmOpt func(*Helm)

func WithRegistryMirror(mirror string) HelmOpt {
	return func(h *Helm) {
		h.registryMirror = mirror
	}
}

func NewHelm(executable Executable, opts ...HelmOpt) *Helm {
	h := &Helm{
		executable: executable,
	}

	for _, o := range opts {
		o(h)
	}

	return h
}

func (h *Helm) Template(ctx context.Context, ociURI, version, namespace string, values interface{}) ([]byte, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling values for helm template: %v", err)
	}

	result, err := h.executable.Command(
		ctx, "template", h.url(ociURI), "--version", version, insecureSkipVerifyFlag, "--namespace", namespace, "-f", "-",
	).WithStdIn(valuesYaml).WithEnvVars(helmTemplateEnvVars).Run()
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (h *Helm) PullChart(ctx context.Context, ociURI, version string) error {
	_, err := h.executable.Command(ctx, "pull", h.url(ociURI), "--version", version, insecureSkipVerifyFlag).
		WithEnvVars(helmTemplateEnvVars).Run()
	return err
}

func (h *Helm) PushChart(ctx context.Context, chart, registry string) error {
	logger.Info("Pushing", "chart", chart)
	_, err := h.executable.Command(ctx, "push", chart, registry, insecureSkipVerifyFlag).WithEnvVars(helmTemplateEnvVars).Run()
	return err
}

func (h *Helm) RegistryLogin(ctx context.Context, registry, username, password string) error {
	logger.Info("Logging in to helm registry", "registry", registry)
	_, err := h.executable.Command(ctx, "registry", "login", registry, "--username", username, "--password", password, "--insecure").WithEnvVars(helmTemplateEnvVars).Run()
	return err
}

func (h *Helm) SaveChart(ctx context.Context, ociURI, version, folder string) error {
	_, err := h.executable.Command(ctx, "pull", h.url(ociURI), "--version", version, insecureSkipVerifyFlag, "--destination", folder).
		WithEnvVars(helmTemplateEnvVars).Run()
	return err
}

func (h *Helm) url(originalURL string) string {
	return urls.ReplaceHost(originalURL, h.registryMirror)
}
