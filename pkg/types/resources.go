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

func (m *Machine) HasAnyLabel(labels []string) bool {
	for _, label := range labels {
		if _, ok := m.Metadata.Labels[label]; ok {
			return true
		}
	}
	return false
}

type MachineStatus struct {
	NodeRef    *ResourceRef `json:"nodeRef,omitempty"`
	Conditions Conditions
}

type MachineMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

type ResourceRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"Name"`
}

type Conditions []Condition

type ConditionType string

type ConditionStatus string

type Condition struct {
	Type   ConditionType   `json:"type"`
	Status ConditionStatus `json:"status"`
}

type CAPICluster struct {
	Metadata Metadata
	Status   ClusterStatus
}

type ClusterStatus struct {
	Phase string
}

type CmkTemplate struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Zonename string `json:"zonename"`
}

type CmkServiceOffering struct {
	CpuNumber int    `json:"cpunumber"`
	CpuSpeed  int    `json:"cpuspeed""`
	Memory    int    `json:"memory"`
	Id        string `json:"id"`
	Name      string `json:"name"`
}

type CmkDiskOffering struct {
	DiskSize int    `json:"disksize"`
	Id       string `json:"id"`
	Name     string `json:"name"`
}

type CmkAffinityGroup struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Name string `json:"name"`
}

type CmkZone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type CmkAccount struct {
	RoleType string `json:"roletype"`
	Domain   string `json:"domain"`
	Id       string `json:"id"`
	Name     string `json:"name"`
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

type NodeReadyChecker func(status MachineStatus) bool

func WithNodeRef() NodeReadyChecker {
	return func(status MachineStatus) bool {
		return status.NodeRef != nil
	}
}

func WithNodeHealthy() NodeReadyChecker {
	return func(status MachineStatus) bool {
		for _, c := range status.Conditions {
			if c.Type == "NodeHealthy" {
				return c.Status == "True"
			}
		}
		return false
	}
}
