package types

import "time"

type Deployment struct {
	Namespace string
	Name      string
	Container string
}

type Machine struct {
	Metadata MachineMetadata `json:"metadata"`
	Status   MachineStatus   `json:"status"`
}

type MachineStatus struct {
	NodeRef *ResourceRef `json:"nodeRef,omitempty"`
}

type MachineMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

type ResourceRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"Name"`
}

type CAPICluster struct {
	Metadata Metadata
	Status   ClusterStatus
}

type ClusterStatus struct {
	Phase string
}

type Metadata struct {
	Name string
}

type Datastores struct {
	Info Info `json:"Info"`
}

type Info struct {
	FreeSpace float64 `json:"FreeSpace"`
}

type NowFunc func() time.Time
