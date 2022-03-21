package validations

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
)

func ValidateCertForRegistryMirror(clusterSpec *cluster.Spec, tlsValidator TlsValidator) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}

	host, port := cluster.Spec.RegistryMirrorConfiguration.Endpoint, cluster.Spec.RegistryMirrorConfiguration.Port
	selfSigned, err := tlsValidator.HasSelfSignedCert(host, port)
	if err != nil {
		return fmt.Errorf("validating registry mirror endpoint: %v", err)
	}
	if selfSigned {
		logger.V(1).Info(fmt.Sprintf("Warning: registry mirror endpoint %s is using self-signed certs", cluster.Spec.RegistryMirrorConfiguration.Endpoint))
	}

	certContent := cluster.Spec.RegistryMirrorConfiguration.CACertContent
	if certContent == "" && selfSigned {
		return fmt.Errorf("registry %s is using self-signed certs, please provide the certificate using caCertContent field", cluster.Spec.RegistryMirrorConfiguration.Endpoint)
	}

	if certContent != "" {
		err := tlsValidator.ValidateCert(host, port, certContent)
		if err != nil {
			return fmt.Errorf("invalid registry certificate: %v", err)
		}
	}

	return nil
}
