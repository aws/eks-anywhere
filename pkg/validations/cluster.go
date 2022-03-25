package validations

import (
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
)

func ValidateCertForRegistryMirror(clusterSpec *cluster.Spec, tlsValidator TlsValidator, provider providers.Provider) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}

	insecureSkip := cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify
	if insecureSkip && provider.Name() != constants.SnowProviderName {
		return errors.New("insecureSkipVerify is only supported for snow provider")
	}

	host, port := cluster.Spec.RegistryMirrorConfiguration.Endpoint, cluster.Spec.RegistryMirrorConfiguration.Port
	selfSigned, err := tlsValidator.HasSelfSignedCert(host, port)
	if err != nil {
		return fmt.Errorf("validating registry mirror endpoint: %v", err)
	}
	if selfSigned {
		logger.V(1).Info(fmt.Sprintf("Warning: registry mirror endpoint %s is using self-signed certs", cluster.Spec.RegistryMirrorConfiguration.Endpoint))
	}

	if insecureSkip {
		logger.V(1).Info("Warning: skip registry certificate verification is enabled", "registryMirrorConfiguration.insecureSkipVerify", true)
		return nil
	}

	certContent := cluster.Spec.RegistryMirrorConfiguration.CACertContent
	if certContent == "" && selfSigned {
		return fmt.Errorf("registry %s is using self-signed certs, please provide the certificate using caCertContent field. Or use insecureSkipVerify field to skip registry certificate verification", cluster.Spec.RegistryMirrorConfiguration.Endpoint)
	}

	if certContent != "" {
		err := tlsValidator.ValidateCert(host, port, certContent)
		if err != nil {
			return fmt.Errorf("invalid registry certificate: %v", err)
		}
	}

	return nil
}
