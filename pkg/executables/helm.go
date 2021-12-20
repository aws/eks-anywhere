package executables

import (
	"context"
	"fmt"

	"sigs.k8s.io/yaml"
)

type Helm struct {
	executable Executable
}

func NewHelm(executable Executable) *Helm {
	return &Helm{
		executable: executable,
	}
}

func (h *Helm) Template(ctx context.Context, ociURI, version, namespace string, values interface{}) ([]byte, error) {
	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling values for helm template: %v", err)
	}

	result, err := h.executable.ExecuteWithStdin(ctx, valuesYaml, "template", ociURI, "--version", version, "--namespace", namespace, "-f", "-")
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
