package docker

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// GetMachineTemplate gets a DockerMachineTemplate object using the provided client
// If the object doesn't exist, it returns a NotFound error.
func GetMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*dockerv1.DockerMachineTemplate, error) {
	m := &dockerv1.DockerMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, errors.Wrap(err, "reading dockerMachineTemplate")
	}

	return m, nil
}

// MachineTemplateEqual returns a boolean indicating whether or not the provided DockerMachineTemplates are equal.
func MachineTemplateEqual(new, old *dockerv1.DockerMachineTemplate) bool {
	return equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}
