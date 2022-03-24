package v1alpha1

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/logger"
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

func (n *Cluster) Validate() error {
	return ValidateClusterConfigContent(n)
}

func (n *Cluster) SetDefaults() {
	// TODO: move any defaults that can return error out of this package
	// All the defaults here should be context unaware
	if err := setClusterDefaults(n); err != nil {
		logger.Error(err, "Failed to validate Cluster")
	}
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
	Endpoint *Endpoint `json:"endpoint,omitempty"`
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
	taints := make(map[corev1.Taint]struct{})
	for _, taint := range s1 {
		taints[taint] = struct{}{}
	}
	for _, taint := range s2 {
		_, ok := taints[taint]
		if !ok {
			return false
		}
	}
	return true
}

func LabelsMapEqual(s1, s2 map[string]string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for key, val := range s2 {
		v, ok := s1[key]
		if !ok {
			return false
		}
		if val != v {
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
	return n.Count == o.Count && n.Endpoint.Equal(o.Endpoint) && n.MachineGroupRef.Equal(o.MachineGroupRef) &&
		TaintsSliceEqual(n.Taints, o.Taints) && LabelsMapEqual(n.Labels, o.Labels)
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
	// Taints define the set of taints to be applied on worker nodes
	Taints []corev1.Taint `json:"taints,omitempty"`
	// Labels define the labels to assign to the node
	Labels map[string]string `json:"labels,omitempty"`
}

func generateWorkerNodeGroupKey(c WorkerNodeGroupConfiguration) (key string) {
	key = c.Name
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
	if len(m) != 0 {
		return false
	}

	return WorkerNodeGroupConfigurationSliceTaintsEqual(a, b) && WorkerNodeGroupConfigurationsLabelsMapEqual(a, b)
}

func WorkerNodeGroupConfigurationSliceTaintsEqual(a, b []WorkerNodeGroupConfiguration) bool {
	m := make(map[string][]corev1.Taint, len(a))
	for _, nodeGroup := range a {
		m[nodeGroup.Name] = nodeGroup.Taints
	}

	for _, nodeGroup := range b {
		if _, ok := m[nodeGroup.Name]; !ok {
			// this method is not concerned with added/removed node groups,
			// only with the comparison of taints on existing node groups
			// if a node group is present in a but not b, or vise versa, it's immaterial
			continue
		} else {
			if !TaintsSliceEqual(m[nodeGroup.Name], nodeGroup.Taints) {
				return false
			}
		}
	}
	return true
}

func WorkerNodeGroupConfigurationsLabelsMapEqual(a, b []WorkerNodeGroupConfiguration) bool {
	m := make(map[string]map[string]string, len(a))
	for _, nodeGroup := range a {
		m[nodeGroup.Name] = nodeGroup.Labels
	}

	for _, nodeGroup := range b {
		if _, ok := m[nodeGroup.Name]; !ok {
			// this method is not concerned with added/removed node groups,
			// only with the comparison of labels on existing node groups
			// if a node group is present in a but not b, or vise versa, it's immaterial
			continue
		} else {
			if !LabelsMapEqual(m[nodeGroup.Name], nodeGroup.Labels) {
				return false
			}
		}
	}
	return true
}

type ClusterNetwork struct {
	// Comma-separated list of CIDR blocks to use for pod and service subnets.
	// Defaults to 192.168.0.0/16 for pod subnet.
	Pods     Pods     `json:"pods,omitempty"`
	Services Services `json:"services,omitempty"`
	// Deprecated. Use CNIConfig
	CNI CNI `json:"cni,omitempty"`
	// CNIConfig specifies the CNI plugin to be installed in the cluster
	CNIConfig *CNIConfig `json:"cniConfig,omitempty"`
	DNS       DNS        `json:"dns,omitempty"`
}

func (n *ClusterNetwork) Equal(o *ClusterNetwork) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}

	if !CNIPluginSame(*n, *o) {
		return false
	}

	oldCNIConfig := getCNIConfig(o)
	newCNIConfig := getCNIConfig(n)
	if !newCNIConfig.Equal(oldCNIConfig) {
		return false
	}

	return n.Pods.Equal(&o.Pods) && n.Services.Equal(&o.Services) && n.DNS.Equal(&o.DNS)
}

func getCNIConfig(cn *ClusterNetwork) *CNIConfig {
	/* Only needed since we're introducing CNIConfig to replace the deprecated CNI field. This way we can compare the individual fields
	for the CNI plugin configuration*/
	var tempCNIConfig *CNIConfig
	if cn.CNIConfig == nil {
		// This is for upgrading from release-0.7, to ensure that all oCNIConfig fields, such as policyEnforcementMode have the default values
		switch cn.CNI {
		case Cilium, CiliumEnterprise:
			tempCNIConfig = &CNIConfig{Cilium: &CiliumConfig{}}
		case Kindnetd:
			tempCNIConfig = &CNIConfig{Kindnetd: &KindnetdConfig{}}
		}
	} else {
		tempCNIConfig = cn.CNIConfig
	}
	return tempCNIConfig
}

func (n *Pods) Equal(o *Pods) bool {
	return SliceEqual(n.CidrBlocks, o.CidrBlocks)
}

func (n *Services) Equal(o *Services) bool {
	return SliceEqual(n.CidrBlocks, o.CidrBlocks)
}

func (n *DNS) Equal(o *DNS) bool {
	return n.ResolvConf.Equal(o.ResolvConf)
}

func (n *CNIConfig) Equal(o *CNIConfig) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	if !n.Cilium.Equal(o.Cilium) {
		return false
	}
	if !n.Kindnetd.Equal(o.Kindnetd) {
		return false
	}
	return true
}

func (n *CiliumConfig) Equal(o *CiliumConfig) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.PolicyEnforcementMode == o.PolicyEnforcementMode
}

func (n *KindnetdConfig) Equal(o *KindnetdConfig) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return true
}

func CNIPluginSame(n ClusterNetwork, o ClusterNetwork) bool {
	if n.CNI != "" {
		/*This shouldn't be required since we set CNIConfig and unset CNI as part of cluster_defaults. However, while upgrading an existing cluster, the eks-a controller
		does not set any defaults (no mutating webhook), so it gets stuck in an error loop. Adding these checks to avoid that. We can remove it when removing the CNI field
		in a later release*/
		return o.CNI == n.CNI
	}

	if n.CNIConfig != nil {
		if o.CNI != "" {
			switch o.CNI {
			case Cilium, CiliumEnterprise:
				if n.CNIConfig.Cilium == nil {
					return false
				}
			case Kindnetd:
				if n.CNIConfig.Kindnetd == nil {
					return false
				}
			default:
				return false
			}
			return true
		}
		if o.CNIConfig != nil {
			if (n.CNIConfig.Cilium != nil && o.CNIConfig.Cilium == nil) || (n.CNIConfig.Cilium == nil && o.CNIConfig.Cilium != nil) {
				return false
			}
			if (n.CNIConfig.Kindnetd != nil && o.CNIConfig.Kindnetd == nil) || (n.CNIConfig.Kindnetd == nil && o.CNIConfig.Kindnetd != nil) {
				return false
			}
		}
	}

	return true
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
	Kube122 KubernetesVersion = "1.22"
)

type CNI string

type CiliumPolicyEnforcementMode string

type CNIConfig struct {
	Cilium   *CiliumConfig   `json:"cilium,omitempty"`
	Kindnetd *KindnetdConfig `json:"kindnetd,omitempty"`
}

type CiliumConfig struct {
	// PolicyEnforcementMode determines communication allowed between pods. Accepted values are default, always, never.
	PolicyEnforcementMode CiliumPolicyEnforcementMode `json:"policyEnforcementMode,omitempty"`
}

type KindnetdConfig struct{}

const (
	Cilium           CNI = "cilium"
	CiliumEnterprise CNI = "cilium-enterprise"
	Kindnetd         CNI = "kindnetd"
)

var validCNIs = map[CNI]struct{}{
	Cilium:   {},
	Kindnetd: {},
}

const (
	CiliumPolicyModeDefault CiliumPolicyEnforcementMode = "default"
	CiliumPolicyModeAlways  CiliumPolicyEnforcementMode = "always"
	CiliumPolicyModeNever   CiliumPolicyEnforcementMode = "never"
)

var validCiliumPolicyEnforcementModes = map[CiliumPolicyEnforcementMode]bool{
	CiliumPolicyModeAlways:  true,
	CiliumPolicyModeDefault: true,
	CiliumPolicyModeNever:   true,
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// Descriptive message about a fatal problem while reconciling a cluster
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
	// EksdReleaseRef defines the properties of the EKS-D object on the cluster
	EksdReleaseRef *EksdReleaseRef `json:"eksdReleaseRef,omitempty"`
	// +optional
	Conditions []clusterv1.Condition `json:"conditions,omitempty"`
}

type EksdReleaseRef struct {
	// ApiVersion refers to the EKS-D API version
	ApiVersion string `json:"apiVersion"`
	// Kind refers to the Release kind for the EKS-D object
	Kind string `json:"kind"`
	// Name refers to the name of the EKS-D object on the cluster
	Name string `json:"name"`
	// Namespace refers to the namespace for the EKS-D release resources
	Namespace string `json:"namespace"`
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
// +kubebuilder:subresource:status
// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (c *Cluster) GetConditions() clusterv1.Conditions {
	return c.Status.Conditions
}

func (c *Cluster) SetConditions(conditions clusterv1.Conditions) {
	c.Status.Conditions = conditions
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
