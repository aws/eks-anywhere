package tinkerbell

import (
	"context"

	"github.com/pkg/errors"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// GetMachineTemplate gets a TinkerbellMachineTemplate object using the provided client
// If the object doesn't exist, it returns a NotFound error.
func GetMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*tinkerbellv1.TinkerbellMachineTemplate, error) {
	m := &tinkerbellv1.TinkerbellMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, errors.Wrap(err, "reading tinkerbellMachineTemplate")
	}

	return m, nil
}

// machineTemplateEqual returns a boolean indicating whether the provided TinkerbellMachineTemplates are equal.
func machineTemplateEqual(new, old *tinkerbellv1.TinkerbellMachineTemplate) bool {
	return equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}
