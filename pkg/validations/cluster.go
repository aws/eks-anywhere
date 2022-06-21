package validations

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/logger"
)

func ValidateCertForRegistryMirror(clusterSpec *cluster.Spec, tlsValidator TlsValidator) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}

	if cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify {
		logger.V(1).Info("Warning: skip registry certificate verification is enabled", "registryMirrorConfiguration.insecureSkipVerify", true)
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
		return fmt.Errorf("registry %s is using self-signed certs, please provide the certificate using caCertContent field. Or use insecureSkipVerify field to skip registry certificate verification", cluster.Spec.RegistryMirrorConfiguration.Endpoint)
	}

	if certContent != "" {
		if err = tlsValidator.ValidateCert(host, port, certContent); err != nil {
			return fmt.Errorf("invalid registry certificate: %v", err)
		}
	}

	return nil
}

func ValidateK8s123Support(clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.K8s123Support()) {
		if clusterSpec.Cluster.Spec.KubernetesVersion == v1alpha1.Kube123 {
			return fmt.Errorf("kubernetes version %s is not enabled. Please set the env variable %v", v1alpha1.Kube123, features.K8s123SupportEnvVar)
		}
	}
	return nil
}
