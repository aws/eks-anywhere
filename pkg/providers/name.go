package providers

import (
	"fmt"
)

const (
	EtcdNodeNameSuffix         = "etcd"
	ControlPlaneNodeNameSuffix = "cp"
)

func GetControlPlaneNodeName(clusterName string) string {
	return fmt.Sprintf("%s-%s", clusterName, ControlPlaneNodeNameSuffix)
}

func GetEtcdNodeName(clusterName string) string {
	return fmt.Sprintf("%s-%s", clusterName, EtcdNodeNameSuffix)
}
