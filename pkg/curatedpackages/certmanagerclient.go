package curatedpackages

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/constants"
)

type CertManagerClient struct {
	kubeConfig string
	kubectl    KubectlRunner
}

func NewCertManagerClient(kubectl KubectlRunner, kubeConfig string) *CertManagerClient {
	return &CertManagerClient{
		kubeConfig: kubeConfig,
		kubectl:    kubectl,
	}
}

// CertManagerExists checks if cert-manager exists in any namespace in the cluster.
func (cmc *CertManagerClient) CertManagerExists(ctx context.Context) (bool, error) {
	// Note although we passed in a namespace parameter in the kubectl command, the GetResource command will be
	// performed in all namespaces since CRDs are not bounded by namespaces.
	found, err := cmc.kubectl.GetResource(ctx, "crd", "certificates.cert-manager.io", cmc.kubeConfig, constants.CertManagerNamespace)

	return found, err
}
