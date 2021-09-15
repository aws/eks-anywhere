package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PausedAnnotation is an annotation that can be applied to EKS-A cluster
	// object to prevent a controller from processing a resource.
	pausedAnnotation = "anywhere.eks.amazonaws.com/paused"

	// ControlPlaneAnnotation is an annotation that can be applied to EKS-A machineconfig
	// object to prevent a controller from making changes to that resource.
	controlPlaneAnnotation = "anywhere.eks.amazonaws.com/control-plane"

	clusterResourceType = "clusters.anywhere.eks.amazonaws.com"

	// etcdAnnotation can be applied to EKS-A machineconfig CR for etcd, to prevent controller from making changes to it
	etcdAnnotation = "anywhere.eks.amazonaws.com/etcd"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	KubernetesVersion             KubernetesVersion              `json:"kubernetesVersion,omitempty"`
	ControlPlaneConfiguration     ControlPlaneConfiguration      `json:"controlPlaneConfiguration,omitempty"`
	WorkerNodeGroupConfigurations []WorkerNodeGroupConfiguration `json:"workerNodeGroupConfigurations,omitempty"`
	DatacenterRef                 Ref                            `json:"datacenterRef,omitempty"`
	IdentityProviderRefs          []Ref                          `json:"identityProviderRefs,omitempty"`
	GitOpsRef                     *Ref                           `json:"gitOpsRef,omitempty"`
	// Deprecated: This field has no function and is going to be removed in a future release.
	OverrideClusterSpecFile string         `json:"overrideClusterSpecFile,omitempty"`
	ClusterNetwork          ClusterNetwork `json:"clusterNetwork,omitempty"`
	// +kubebuilder:validation:Optional
	ExternalEtcdConfiguration *ExternalEtcdConfiguration `json:"externalEtcdConfiguration,omitempty"`
	ProxyConfiguration        *ProxyConfiguration        `json:"proxyConfiguration,omitempty"`
}

type ProxyConfiguration struct {
	HttpProxy  string   `json:"httpProxy,omitempty"`
	HttpsProxy string   `json:"httpsProxy,omitempty"`
	NoProxy    []string `json:"noProxy,omitempty"`
}

func (n *ProxyConfiguration) Equal(o *ProxyConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.HttpProxy == o.HttpProxy && n.HttpsProxy == o.HttpsProxy && SliceEqual(n.NoProxy, o.NoProxy)
}

type ControlPlaneConfiguration struct {
	// Count defines the number of desired control plane nodes. Defaults to 1.
	Count int `json:"count,omitempty"`
	// Endpoint defines the host ip and port to use for the control plane.
	Endpoint *Endpoint `json:"endpoint,omitempty"`
	// MachineGroupRef defines the machine group configuration for the control plane.
	MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
}

type Endpoint struct {
	// Host defines the ip that you want to use to connect to the control plane
	Host string `json:"host"`
}

func (n *Endpoint) Equal(o *Endpoint) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Host == o.Host
}

type WorkerNodeGroupConfiguration struct {
	// Count defines the number of desired worker nodes. Defaults to 1.
	Count int `json:"count,omitempty"`
	// MachineGroupRef defines the machine group configuration for the worker nodes.
	MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
}

type ClusterNetwork struct {
	// Comma-separated list of CIDR blocks to use for pod and service subnets.
	// Defaults to 192.168.0.0/16 for pod subnet.
	Pods     Pods     `json:"pods,omitempty"`
	Services Services `json:"services,omitempty"`
	// CNI specifies the CNI plugin to be installed in the cluster
	CNI CNI `json:"cni,omitempty"`
}

func (n *ClusterNetwork) Equal(o *ClusterNetwork) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return SliceEqual(n.Pods.CidrBlocks, o.Pods.CidrBlocks) &&
		SliceEqual(n.Services.CidrBlocks, o.Services.CidrBlocks) &&
		n.CNI == o.CNI
}

func SliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		if _, ok := m[v]; !ok {
			return false
		}
		m[v] -= 1
		if m[v] == 0 {
			delete(m, v)
		}
	}
	return len(m) == 0
}

func RefSliceEqual(a, b []Ref) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v.Name+v.Kind]++
	}
	for _, v := range b {
		if _, ok := m[v.Name+v.Kind]; !ok {
			return false
		}
		m[v.Name+v.Kind] -= 1
		if m[v.Name+v.Kind] == 0 {
			delete(m, v.Name+v.Kind)
		}
	}
	return len(m) == 0
}

type Pods struct {
	CidrBlocks []string `json:"cidrBlocks,omitempty"`
}

type Services struct {
	CidrBlocks []string `json:"cidrBlocks,omitempty"`
}

type KubernetesVersion string

const (
	Kube118 KubernetesVersion = "1.18"
	Kube119 KubernetesVersion = "1.19"
	Kube120 KubernetesVersion = "1.20"
	Kube121 KubernetesVersion = "1.21"
)

type CNI string

const (
	Cilium           CNI = "cilium"
	CiliumEnterprise CNI = "cilium-enterprise"
)

var validCNIs = map[CNI]struct{}{
	Cilium: {},
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct{}

type Ref struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
}

func (n *Ref) Equal(o *Ref) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Kind == o.Kind && n.Name == o.Name
}

// +kubebuilder:object:generate=false
// Interface for getting DatacenterRef fields for Cluster type
type ProviderRefAccessor interface {
	Kind() string
	Name() string
}

// +kubebuilder:object:generate=false
// Interface for getting Kind field for Cluster type
type KindAccessor interface {
	Kind() string
	ExpectedKind() string
}

// ExternalEtcdConfiguration defines the configuration options for using unstacked etcd topology
type ExternalEtcdConfiguration struct {
	Count int `json:"count,omitempty"`
	// MachineGroupRef defines the machine group configuration for the etcd machines.
	MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
}

// +kubebuilder:object:root=true
// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as Cluster except stripped down for generation of yaml file during generate clusterconfig
type ClusterGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
}

func (c *Cluster) Kind() string {
	return c.TypeMeta.Kind
}

func (c *Cluster) ExpectedKind() string {
	return ClusterKind
}

func (c *Cluster) PausedAnnotation() string {
	return pausedAnnotation
}

func (c *Cluster) ControlPlaneAnnotation() string {
	return controlPlaneAnnotation
}

func (c *Cluster) ResourceType() string {
	return clusterResourceType
}

func (c *Cluster) EtcdAnnotation() string {
	return etcdAnnotation
}

// +kubebuilder:object:root=true
// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
