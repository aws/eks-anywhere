package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

var clusterDefaults = []func(*Cluster) error{
	setRegistryMirrorConfigDefaults,
	setWorkerNodeGroupDefaults,
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
			content, err := ioutil.ReadFile(caCert)
			if err != nil {
				return fmt.Errorf("error reading the cert file %s: %v", caCert, err)
			}
			logger.V(4).Info(fmt.Sprintf("%s is set, using %s as ca cert for registry", RegistryMirrorCAKey, caCert))
			clusterConfig.Spec.RegistryMirrorConfiguration.CACertContent = string(content)
		}
	}
	return nil
}

func setWorkerNodeGroupDefaults(cluster *Cluster) error {
	// TODO: add default worker node group name when it becomes available
	return nil
}
