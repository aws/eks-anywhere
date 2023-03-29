package cloudstack

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// GetMachineTemplate gets a TinkerbellMachineTemplate object using the provided client
// If the object doesn't exist, it returns a NotFound error.
func GetMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*cloudstackv1.CloudStackMachineTemplate, error) {
	m := &cloudstackv1.CloudStackMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, errors.Wrap(err, "reading cloudstackMachineTemplate")
	}

	return m, nil
}

// machineTemplateEqual returns a boolean indicating whether the provided CloudStackMachineTemplates are equal.
func machineTemplateEqual(new, old *cloudstackv1.CloudStackMachineTemplate) bool {
	return equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}
