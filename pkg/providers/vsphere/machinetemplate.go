package vsphere

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

func getMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*vspherev1.VSphereMachineTemplate, error) {
	m := &vspherev1.VSphereMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, errors.Wrap(err, "reading vSphereMachineTemplate")
	}

	return m, nil
}

func machineTemplateEqual(new, old *vspherev1.VSphereMachineTemplate) bool {
	return equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}
