package v1alpha1

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
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

	// managementAnnotation points to the name of a management cluster
	// cluster object
	managementAnnotation = "anywhere.eks.amazonaws.com/managed-by"

	// defaultEksaNamespace is the default namespace for EKS-A resources when not specified.
	defaultEksaNamespace = "default"
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
	ExternalEtcdConfiguration   *ExternalEtcdConfiguration   `json:"externalEtcdConfiguration,omitempty"`
	ProxyConfiguration          *ProxyConfiguration          `json:"proxyConfiguration,omitempty"`
	RegistryMirrorConfiguration *RegistryMirrorConfiguration `json:"registryMirrorConfiguration,omitempty"`
	ManagementCluster           ManagementCluster            `json:"managementCluster,omitempty"`
	PodIAMConfig                *PodIAMConfig                `json:"podIamConfig,omitempty"`
}

func (n *Cluster) Equal(o *Cluster) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	if n.Spec.KubernetesVersion != o.Spec.KubernetesVersion {
		return false
	}
	if !n.Spec.ControlPlaneConfiguration.Equal(&o.Spec.ControlPlaneConfiguration) {
		return false
	}
	if !WorkerNodeGroupConfigurationsSliceEqual(n.Spec.WorkerNodeGroupConfigurations, o.Spec.WorkerNodeGroupConfigurations) {
		return false
	}
	if !n.Spec.DatacenterRef.Equal(&o.Spec.DatacenterRef) {
		return false
	}
	if !RefSliceEqual(n.Spec.IdentityProviderRefs, o.Spec.IdentityProviderRefs) {
		return false
	}
	if !n.Spec.GitOpsRef.Equal(o.Spec.GitOpsRef) {
		return false
	}
	if !n.Spec.ClusterNetwork.Equal(&o.Spec.ClusterNetwork) {
		return false
	}
	if !n.Spec.ExternalEtcdConfiguration.Equal(o.Spec.ExternalEtcdConfiguration) {
		return false
	}
	if !n.Spec.ProxyConfiguration.Equal(o.Spec.ProxyConfiguration) {
		return false
	}
	if !n.Spec.RegistryMirrorConfiguration.Equal(o.Spec.RegistryMirrorConfiguration) {
		return false
	}
	if !n.ManagementClusterEqual(o) {
		return false
	}
	return true
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

// RegistryMirrorConfiguration defines the settings for image registry mirror
type RegistryMirrorConfiguration struct {
	// Endpoint defines the registry mirror endpoint to use for pulling images
	Endpoint string `json:"endpoint,omitempty"`

	// Port defines the port exposed for registry mirror endpoint
	Port string `json:"port,omitempty"`

	// CACertContent defines the contents registry mirror CA certificate
	CACertContent string `json:"caCertContent,omitempty"`
}

func (n *RegistryMirrorConfiguration) Equal(o *RegistryMirrorConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Endpoint == o.Endpoint && n.Port == o.Port && n.CACertContent == o.CACertContent
}

type ControlPlaneConfiguration struct {
	// Count defines the number of desired control plane nodes. Defaults to 1.
	Count int `json:"count,omitempty"`
	// Endpoint defines the host ip and port to use for the control plane.
	Endpoint       *Endpoint `json:"endpoint,omitempty"`
	ExtraEndpoints []string  `json:"extraEndpoints,omitempty"`
	// MachineGroupRef defines the machine group configuration for the control plane.
	MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
	// Taints define the set of taints to be applied on control plane nodes
	Taints []corev1.Taint `json:"taints,omitempty"`
	// Labels define the labels to assign to the node
	Labels map[string]string `json:"labels,omitempty"`
}

func TaintsSliceEqual(s1, s2 []corev1.Taint) bool {
	if len(s1) != len(s2) {
		return false
	}
	taints := make(map[corev1.Taint]bool)
	for _, taint := range s1 {
		taints[taint] = true
	}
	for _, taint := range s2 {
		_, ok := taints[taint]
		if !ok {
			return false
		}
	}
	return true
}

func (n *ControlPlaneConfiguration) Equal(o *ControlPlaneConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Count == o.Count && n.Endpoint.Equal(o.Endpoint) && n.MachineGroupRef.Equal(o.MachineGroupRef) && TaintsSliceEqual(n.Taints, o.Taints)
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
	// Name refers to the name of the worker node group
	Name string `json:"name,omitempty"`
	// Count defines the number of desired worker nodes. Defaults to 1.
	Count int `json:"count,omitempty"`
	// MachineGroupRef defines the machine group configuration for the worker nodes.
	MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
	// Labels define the labels to assign to the node
	Labels map[string]string `json:"labels,omitempty"`
}

func generateWorkerNodeGroupKey(c WorkerNodeGroupConfiguration) (key string) {
	if c.MachineGroupRef != nil {
		key = c.MachineGroupRef.Kind + c.MachineGroupRef.Name
	}
	return strconv.Itoa(c.Count) + key
}

func WorkerNodeGroupConfigurationsSliceEqual(a, b []WorkerNodeGroupConfiguration) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, v := range a {
		m[generateWorkerNodeGroupKey(v)]++
	}
	for _, v := range b {
		k := generateWorkerNodeGroupKey(v)
		if _, ok := m[k]; !ok {
			return false
		}
		m[k] -= 1
		if m[k] == 0 {
			delete(m, k)
		}
	}
	return len(m) == 0
}

type ClusterNetwork struct {
	// Comma-separated list of CIDR blocks to use for pod and service subnets.
	// Defaults to 192.168.0.0/16 for pod subnet.
	Pods     Pods     `json:"pods,omitempty"`
	Services Services `json:"services,omitempty"`
	// CNI specifies the CNI plugin to be installed in the cluster
	CNI CNI `json:"cni,omitempty"`
	DNS DNS `json:"dns,omitempty"`
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
		n.CNI == o.CNI && n.DNS.ResolvConf.Equal(o.DNS.ResolvConf)
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

type DNS struct {
	// ResolvConf refers to the DNS resolver configuration
	ResolvConf *ResolvConf `json:"resolvConf,omitempty"`
}

type ResolvConf struct {
	// Path defines the path to the file that contains the DNS resolver configuration
	Path string `json:"path,omitempty"`
}

func (n *ResolvConf) Equal(o *ResolvConf) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Path == o.Path
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
	Kindnetd         CNI = "kindnetd"
)

var validCNIs = map[CNI]struct{}{
	Cilium:   {},
	Kindnetd: {},
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// Descriptive message about a fatal problem while reconciling a cluster
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

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

func (n *ExternalEtcdConfiguration) Equal(o *ExternalEtcdConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Count == o.Count && n.MachineGroupRef.Equal(o.MachineGroupRef)
}

type ManagementCluster struct {
	Name string `json:"name,omitempty"`
}

func (n *ManagementCluster) Equal(o ManagementCluster) bool {
	return n.Name == o.Name
}

type PodIAMConfig struct {
	ServiceAccountIssuer string `json:"serviceAccountIssuer"`
}

func (n *PodIAMConfig) Equal(o *PodIAMConfig) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.ServiceAccountIssuer == o.ServiceAccountIssuer
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

func (s *Cluster) IsSelfManaged() bool {
	return s.Spec.ManagementCluster.Name == "" || s.Spec.ManagementCluster.Name == s.Name
}

func (s *Cluster) SetManagedBy(managementClusterName string) {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	s.Annotations[managementAnnotation] = managementClusterName
	s.Spec.ManagementCluster.Name = managementClusterName
}

func (s *Cluster) SetSelfManaged() {
	s.Spec.ManagementCluster.Name = s.Name
}

func (c *ClusterGenerate) SetSelfManaged() {
	c.Spec.ManagementCluster.Name = c.Name
}

func (s *Cluster) ManagementClusterEqual(s2 *Cluster) bool {
	return s.IsSelfManaged() && s2.IsSelfManaged() || s.Spec.ManagementCluster.Equal(s2.Spec.ManagementCluster)
}

func (c *Cluster) MachineConfigRefs() []Ref {
	machineConfigRefMap := make(refSet, 1)

	machineConfigRefMap.addIfNotNil(c.Spec.ControlPlaneConfiguration.MachineGroupRef)

	for _, m := range c.Spec.WorkerNodeGroupConfigurations {
		machineConfigRefMap.addIfNotNil(m.MachineGroupRef)
	}

	if c.Spec.ExternalEtcdConfiguration != nil {
		machineConfigRefMap.addIfNotNil(c.Spec.ExternalEtcdConfiguration.MachineGroupRef)
	}

	return machineConfigRefMap.toSlice()
}

type refSet map[Ref]struct{}

func (r refSet) addIfNotNil(ref *Ref) bool {
	if ref != nil {
		return r.add(*ref)
	}

	return false
}

func (r refSet) add(ref Ref) bool {
	if _, present := r[ref]; !present {
		r[ref] = struct{}{}
		return true
	} else {
		return false
	}
}

func (r refSet) toSlice() []Ref {
	refs := make([]Ref, 0, len(r))
	for ref := range r {
		refs = append(refs, ref)
	}

	return refs
}

func (c *Cluster) ConvertConfigToConfigGenerateStruct() *ClusterGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &ClusterGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   namespace,
		},
		Spec: c.Spec,
	}

	return config
}

func (c *Cluster) IsManaged() bool {
	return !c.IsSelfManaged()
}

func (c *Cluster) ManagedBy() string {
	return c.Spec.ManagementCluster.Name
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
