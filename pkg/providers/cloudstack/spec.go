package cloudstack

import (
	"fmt"
	"net"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

func etcdMachineConfig(s *cluster.Spec) *anywherev1.CloudStackMachineConfig {
	if s.Cluster.Spec.ExternalEtcdConfiguration == nil || s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
		return nil
	}
	return s.CloudStackMachineConfigs[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

func controlPlaneMachineConfig(s *cluster.Spec) *anywherev1.CloudStackMachineConfig {
	return s.CloudStackMachineConfigs[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

func workerMachineConfig(s *cluster.Spec, workers anywherev1.WorkerNodeGroupConfiguration) *anywherev1.CloudStackMachineConfig {
	return s.CloudStackMachineConfigs[workers.MachineGroupRef.Name]
}

func controlPlaneEndpointHost(clusterSpec *cluster.Spec) (string, error) {
	host, port, err := getControlPlaneHostPort(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(host, port), nil
}

// getControlPlaneHostPort retrieves the ControlPlaneConfiguration host and port split defined in the cluster.Spec. If it's valid, it checks the port
// to see if the default port should be used and returns it.
func getControlPlaneHostPort(pHost string) (string, string, error) {
	host, port, err := net.SplitHostPort(pHost)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			host = pHost
			port = controlEndpointDefaultPort
		} else {
			return "", "", fmt.Errorf("host %s is invalid: %v", pHost, err.Error())
		}
	}
	if !networkutils.IsPortValid(port) {
		return "", "", fmt.Errorf("host %s has an invalid port", pHost)
	}
	return host, port, nil
}
