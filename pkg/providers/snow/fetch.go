package snow

import (
	"context"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func oldWorkerMachineTemplate(ctx context.Context, kubeClient kubernetes.Client, md *clusterv1.MachineDeployment) (*snowv1.AWSSnowMachineTemplate, error) {
	if md == nil {
		return nil, nil
	}

	mt := &snowv1.AWSSnowMachineTemplate{}
	err := kubeClient.Get(ctx, md.Spec.Template.Spec.InfrastructureRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}

func getMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*snowv1.AWSSnowMachineTemplate, error) {
	m := &snowv1.AWSSnowMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, errors.Wrap(err, "fetching snowMachineTemplate")
	}

	return m, nil
}
