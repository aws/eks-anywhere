package executables

import (
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	helmPath               = "helm"
	insecureSkipVerifyFlag = "--insecure-skip-tls-verify"
)

var helmTemplateEnvVars = map[string]string{
	"HELM_EXPERIMENTAL_OCI": "1",
}

type Helm struct {
	Executable
}

func NewHelm(executable Executable) *Helm {
	return &Helm{
		Executable: executable,
	}
}

func (h *Helm) Template(ctx context.Context, ociURI, version, namespace string, values interface{}) ([]byte, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling values for helm template: %v", err)
	}

	result, err := h.Command(
		ctx, "template", ociURI, "--version", version, insecureSkipVerifyFlag, "--namespace", namespace, "-f", "-",
	).WithStdIn(valuesYaml).WithEnvVars(helmTemplateEnvVars).Run()
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (h *Helm) PullChart(ctx context.Context, ociURI, version string) error {
	_, err := h.Command(ctx, "pull", ociURI, "--version", version, insecureSkipVerifyFlag).
		WithEnvVars(helmTemplateEnvVars).Run()
	return err
}

func (h *Helm) PushChart(ctx context.Context, chart, registry string) error {
	logger.Info("Pushing to registry", "chart", chart, "registry", registry)
	_, err := h.Command(ctx, "push", chart, registry, insecureSkipVerifyFlag).WithEnvVars(helmTemplateEnvVars).Run()
	return err
}

func (h *Helm) RegistryLogin(ctx context.Context, registry, username, password string) error {
	_, err := h.Command(ctx, "registry", "login", registry, "--username", username, "--password", password, "--insecure").WithEnvVars(helmTemplateEnvVars).Run()
	return err
}
