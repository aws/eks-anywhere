package validations

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

// ValidateOSForRegistryMirror checks if the OS is valid for the provided registry mirror configuration.
func ValidateOSForRegistryMirror(clusterSpec *cluster.Spec, provider providers.Provider) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}

	machineConfigs := provider.MachineConfigs(clusterSpec)
	if !cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify || machineConfigs == nil {
		return nil
	}

	for _, mc := range machineConfigs {
		if mc.OSFamily() == v1alpha1.Bottlerocket {
			return errors.New("InsecureSkipVerify is not supported for bottlerocket")
		}
	}
	return nil
}

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
	authorityUnknown, err := tlsValidator.IsSignedByUnknownAuthority(host, port)
	if err != nil {
		return fmt.Errorf("validating registry mirror endpoint: %v", err)
	}
	if authorityUnknown {
		logger.V(1).Info(fmt.Sprintf("Warning: registry mirror endpoint %s is using self-signed certs", cluster.Spec.RegistryMirrorConfiguration.Endpoint))
	}

	certContent := cluster.Spec.RegistryMirrorConfiguration.CACertContent
	if certContent == "" && authorityUnknown {
		return fmt.Errorf("registry %s is using self-signed certs, please provide the certificate using caCertContent field. Or use insecureSkipVerify field to skip registry certificate verification", cluster.Spec.RegistryMirrorConfiguration.Endpoint)
	}

	if certContent != "" {
		if err = tlsValidator.ValidateCert(host, port, certContent); err != nil {
			return fmt.Errorf("invalid registry certificate: %v", err)
		}
	}

	return nil
}

// ValidateAuthenticationForRegistryMirror checks if REGISTRY_USERNAME and REGISTRY_PASSWORD is set if authenticated registry mirrors are used.
func ValidateAuthenticationForRegistryMirror(clusterSpec *cluster.Spec) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration != nil && cluster.Spec.RegistryMirrorConfiguration.Authenticate {
		_, _, err := config.ReadCredentials()
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateManagementClusterName checks if the management cluster specified in the workload cluster spec is valid.
func ValidateManagementClusterName(ctx context.Context, k KubectlClient, mgmtCluster *types.Cluster, mgmtClusterName string) error {
	cluster, err := k.GetEksaCluster(ctx, mgmtCluster, mgmtClusterName)
	if err != nil {
		return err
	}
	if cluster.IsManaged() {
		return fmt.Errorf("%s is not a valid management cluster", mgmtClusterName)
	}
	return nil
}

// ValidateManagementClusterBundlesVersion checks if management cluster's bundle version
// is greater than or equal to the bundle version used to upgrade a workload cluster.
func ValidateManagementClusterBundlesVersion(ctx context.Context, k KubectlClient, mgmtCluster *types.Cluster, workload *cluster.Spec) error {
	cluster, err := k.GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name)
	if err != nil {
		return err
	}

	if cluster.Spec.BundlesRef == nil {
		return fmt.Errorf("management cluster bundlesRef cannot be nil")
	}

	mgmtBundles, err := k.GetBundles(ctx, mgmtCluster.KubeconfigFile, cluster.Spec.BundlesRef.Name, cluster.Spec.BundlesRef.Namespace)
	if err != nil {
		return err
	}

	if mgmtBundles.Spec.Number < workload.Bundles.Spec.Number {
		return fmt.Errorf("cannot upgrade workload cluster with bundle spec.number %d while management cluster %s is on older bundle spec.number %d", workload.Bundles.Spec.Number, mgmtCluster.Name, mgmtBundles.Spec.Number)
	}

	return nil
}
