package snow

import (
	"context"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func getMachineTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*snowv1.AWSSnowMachineTemplate, error) {
	m := &snowv1.AWSSnowMachineTemplate{}
	if err := client.Get(ctx, name, namespace, m); err != nil {
		return nil, errors.Wrap(err, "fetching snowMachineTemplate")
	}

	return m, nil
}
