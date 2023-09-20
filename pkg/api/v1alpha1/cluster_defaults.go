package v1alpha1

import (
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

var clusterDefaults = []func(*Cluster) error{
	setRegistryMirrorConfigDefaults,
	setWorkerNodeGroupDefaults,
	setCNIConfigDefault,
	setEtcdEncryptionConfigDefaults,
}

func setClusterDefaults(cluster *Cluster) error {
	for _, d := range clusterDefaults {
		if err := d(cluster); err != nil {
			return err
		}
	}
	return nil
}

func setRegistryMirrorConfigDefaults(clusterConfig *Cluster) error {
	if clusterConfig.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}
	if clusterConfig.Spec.RegistryMirrorConfiguration.Port == "" {
		logger.V(1).Info("RegistryMirrorConfiguration.Port is not specified, default port will be used", "Default Port", constants.DefaultHttpsPort)
		clusterConfig.Spec.RegistryMirrorConfiguration.Port = constants.DefaultHttpsPort
	}
	if clusterConfig.Spec.RegistryMirrorConfiguration.CACertContent == "" {
		if caCert, set := os.LookupEnv(RegistryMirrorCAKey); set && len(caCert) > 0 {
			content, err := os.ReadFile(caCert)
			if err != nil {
				return fmt.Errorf("reading the cert file %s: %v", caCert, err)
			}
			logger.V(4).Info(fmt.Sprintf("%s is set, using %s as ca cert for registry", RegistryMirrorCAKey, caCert))
			clusterConfig.Spec.RegistryMirrorConfiguration.CACertContent = string(content)
		}
	}
	return nil
}

func setWorkerNodeGroupDefaults(cluster *Cluster) error {
	if len(cluster.Spec.WorkerNodeGroupConfigurations) >= 1 && cluster.Spec.WorkerNodeGroupConfigurations[0].Name == "" {
		logger.V(1).Info("First worker node group name not specified. Defaulting name to md-0.")
		cluster.Spec.WorkerNodeGroupConfigurations[0].Name = "md-0"
	}

	for i := range cluster.Spec.WorkerNodeGroupConfigurations {
		w := &cluster.Spec.WorkerNodeGroupConfigurations[i]
		if w.Count == nil && w.AutoScalingConfiguration != nil {
			w.Count = &w.AutoScalingConfiguration.MinCount
		} else if w.Count == nil {
			w.Count = ptr.Int(1)
		}
	}

	return nil
}

func setCNIConfigDefault(cluster *Cluster) error {
	if cluster.Spec.ClusterNetwork.CNIConfig != nil {
		return nil
	}

	cluster.Spec.ClusterNetwork.CNIConfig = &CNIConfig{}
	switch cluster.Spec.ClusterNetwork.CNI {
	case Cilium, CiliumEnterprise:
		cluster.Spec.ClusterNetwork.CNIConfig.Cilium = &CiliumConfig{}
	case Kindnetd:
		cluster.Spec.ClusterNetwork.CNIConfig.Kindnetd = &KindnetdConfig{}
	}

	cluster.Spec.ClusterNetwork.CNI = ""
	return nil
}
