package reconciler

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/constants"
)

const BoostrapSecretName = "capas-manager-bootstrap-credentials"

func getSnowCredentials(ctx context.Context, cli client.Client) (credentials, caBundle []byte, err error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: constants.CapasSystemNamespace,
		Name:      BoostrapSecretName,
	}
	if err = cli.Get(ctx, secretKey, secret); err != nil {
		return nil, nil, err
	}

	return secret.Data["credentials"], secret.Data["ca-bundle"], nil
}
