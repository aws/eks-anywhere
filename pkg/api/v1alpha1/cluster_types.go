package v1alpha1

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
)

const (
	// PausedAnnotation is an annotation that can be applied to EKS-A cluster
	// object to prevent a controller from processing a resource.
	pausedAnnotation = "anywhere.eks.amazonaws.com/paused"

	// ManagedByCLIAnnotation can be applied to an EKS-A Cluster to signal when the CLI is currently
	// performing an operation so the controller should not take any action. When marked for deletion,
	// the controller will remove the finalizer and let the cluster be deleted.
	ManagedByCLIAnnotation = "anywhere.eks.amazonaws.com/managed-by-cli"

	// ControlPlaneAnnotation is an annotation that can be applied to EKS-A machineconfig
	// object to prevent a controller from making changes to that resource.
	controlPlaneAnnotation = "anywhere.eks.amazonaws.com/control-plane"

	clusterResourceType = "clusters.anywhere.eks.amazonaws.com"

	// etcdAnnotation can be applied to EKS-A machineconfig CR for etcd, to prevent controller from making changes to it.
	etcdAnnotation = "anywhere.eks.amazonaws.com/etcd"

	// managementAnnotation points to the name of a management cluster
	// cluster object.
	managementAnnotation = "anywhere.eks.amazonaws.com/managed-by"

	// defaultEksaNamespace is the default namespace for EKS-A resources when not specified.
	defaultEksaNamespace = "default"

	// ControlEndpointDefaultPort defaults cluster control plane endpoint port if not specified.
	ControlEndpointDefaultPort = "6443"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster.
type ClusterSpec struct {
	KubernetesVersion             KubernetesVersion              `json:"kubernetesVersion,omitempty"`
	ControlPlaneConfiguration     ControlPlaneConfiguration      `json:"controlPlaneConfiguration,omitempty"`
	WorkerNodeGroupConfigurations []WorkerNodeGroupConfiguration `json:"workerNodeGroupConfigurations,omitempty"`
	DatacenterRef                 Ref                            `json:"datacenterRef,omitempty"`
	IdentityProviderRefs          []Ref                          `json:"identityProviderRefs,omitempty"`
	GitOpsRef                     *Ref                           `json:"gitOpsRef,omitempty"`
	ClusterNetwork                ClusterNetwork                 `json:"clusterNetwork,omitempty"`
	// +kubebuilder:validation:Optional
	ExternalEtcdConfiguration   *ExternalEtcdConfiguration   `json:"externalEtcdConfiguration,omitempty"`
	ProxyConfiguration          *ProxyConfiguration          `json:"proxyConfiguration,omitempty"`
	RegistryMirrorConfiguration *RegistryMirrorConfiguration `json:"registryMirrorConfiguration,omitempty"`
	ManagementCluster           ManagementCluster            `json:"managementCluster,omitempty"`
	PodIAMConfig                *PodIAMConfig                `json:"podIamConfig,omitempty"`
	Packages                    *PackageConfiguration        `json:"packages,omitempty"`
	// BundlesRef contains a reference to the Bundles containing the desired dependencies for the cluster.
	// DEPRECATED: Use EksaVersion instead.
	BundlesRef  *BundlesRef  `json:"bundlesRef,omitempty"`
	EksaVersion *EksaVersion `json:"eksaVersion,omitempty"`
}

// EksaVersion is the semver identifying the release of eks-a used to populate the cluster components.
type EksaVersion string

// Equal checks if two EksaVersions are equal.
func (n *EksaVersion) Equal(o *EksaVersion) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return *n == *o
}

// HasAWSIamConfig checks if AWSIamConfig is configured for the cluster.
func (c *Cluster) HasAWSIamConfig() bool {
	for _, identityProvider := range c.Spec.IdentityProviderRefs {
		if identityProvider.Kind == AWSIamConfigKind {
			return true
		}
	}

	return false
}

// IsPackagesEnabled checks if the user has opted out of curated packages
// installation.
func (c *Cluster) IsPackagesEnabled() bool {
	return c.Spec.Packages == nil || !c.Spec.Packages.Disable
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

	if !n.Spec.DatacenterRef.Equal(&o.Spec.DatacenterRef) {
		return false
	}
	if !n.Spec.ControlPlaneConfiguration.Endpoint.Equal(o.Spec.ControlPlaneConfiguration.Endpoint, n.Spec.DatacenterRef.Kind) {
		return false
	}
	if !n.Spec.ControlPlaneConfiguration.Equal(&o.Spec.ControlPlaneConfiguration) {
		return false
	}
	if !WorkerNodeGroupConfigurationsSliceEqual(n.Spec.WorkerNodeGroupConfigurations, o.Spec.WorkerNodeGroupConfigurations) {
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
	if !n.Spec.Packages.Equal(o.Spec.Packages) {
		return false
	}
	if !n.ManagementClusterEqual(o) {
		return false
	}
	if !n.Spec.BundlesRef.Equal(o.Spec.BundlesRef) {
		return false
	}
	if !n.Spec.EksaVersion.Equal(o.Spec.EksaVersion) {
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

// RegistryMirrorConfiguration defines the settings for image registry mirror.
type RegistryMirrorConfiguration struct {
	// Endpoint defines the registry mirror endpoint to use for pulling images
	Endpoint string `json:"endpoint,omitempty"`

	// Port defines the port exposed for registry mirror endpoint
	Port string `json:"port,omitempty"`

	// OCINamespaces defines the mapping from an upstream registry to a local namespace where upstream
	// artifacts are placed into
	OCINamespaces []OCINamespace `json:"ociNamespaces,omitempty"`

	// CACertContent defines the contents registry mirror CA certificate
	CACertContent string `json:"caCertContent,omitempty"`

	// Authenticate defines if registry requires authentication
	Authenticate bool `json:"authenticate,omitempty"`

	// InsecureSkipVerify skips the registry certificate verification.
	// Only use this solution for isolated testing or in a tightly controlled, air-gapped environment.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// OCINamespace represents an entity in a local reigstry to group related images.
type OCINamespace struct {
	// Name refers to the name of the upstream registry
	Registry string `json:"registry"`
	// Namespace refers to the name of a namespace in the local registry
	Namespace string `json:"namespace"`
}

func (n *RegistryMirrorConfiguration) Equal(o *RegistryMirrorConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Endpoint == o.Endpoint && n.Port == o.Port && n.CACertContent == o.CACertContent &&
		n.InsecureSkipVerify == o.InsecureSkipVerify && n.Authenticate == o.Authenticate &&
		OCINamespacesSliceEqual(n.OCINamespaces, o.OCINamespaces)
}

// OCINamespacesSliceEqual is used to check equality of the OCINamespaces fields of two RegistryMirrorConfiguration.
func OCINamespacesSliceEqual(a, b []OCINamespace) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, v := range a {
		m[generateOCINamespaceKey(v)]++
	}
	for _, v := range b {
		k := generateOCINamespaceKey(v)
		if _, ok := m[k]; !ok {
			return false
		}
		m[k]--
		if m[k] == 0 {
			delete(m, k)
		}
	}
	return len(m) == 0
}

func generateOCINamespaceKey(n OCINamespace) (key string) {
	return n.Registry + n.Namespace
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
	// UpgradeRolloutStrategy determines the rollout strategy to use for rolling upgrades
	// and related parameters/knobs
	UpgradeRolloutStrategy *ControlPlaneUpgradeRolloutStrategy `json:"upgradeRolloutStrategy,omitempty"`
	// SkipLoadBalancerDeployment skip deploying control plane load balancer.
	// Make sure your infrastructure can handle control plane load balancing when you set this field to true.
	SkipLoadBalancerDeployment bool `json:"skipLoadBalancerDeployment,omitempty"`
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

// MapEqual compares two maps to check whether or not they are equal.
func MapEqual(s1, s2 map[string]string) bool {
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
	return n.Count == o.Count && n.MachineGroupRef.Equal(o.MachineGroupRef) &&
		TaintsSliceEqual(n.Taints, o.Taints) && MapEqual(n.Labels, o.Labels)
}

type Endpoint struct {
	// Host defines the ip that you want to use to connect to the control plane
	Host string `json:"host"`
}

// Equal compares if expected endpoint and existing endpoint are equal for non CloudStack clusters.
func (n *Endpoint) Equal(o *Endpoint, kind string) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	if kind == CloudStackDatacenterKind {
		return n.CloudStackEqual(o)
	}
	return n.Host == o.Host
}

// CloudStackEqual makes CloudStack cluster upgrade to new release backward compatible by striping CloudStack cluster existing endpoint default port
// and comparing if expected endpoint and existing endpoint are equal.
// Cloudstack CLI used to add default port to cluster object.
// Now cluster object stays the same with customer input and port is defaulted only in CAPI template.
func (n *Endpoint) CloudStackEqual(o *Endpoint) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	if n.Host == o.Host {
		return true
	}
	nhost, nport, err := GetControlPlaneHostPort(n.Host, "")
	if err != nil {
		return false
	}
	ohost, oport, _ := GetControlPlaneHostPort(o.Host, "")
	if oport == ControlEndpointDefaultPort {
		switch nport {
		case ControlEndpointDefaultPort, "":
			return nhost == ohost
		default:
			return false
		}
	}

	if nport == ControlEndpointDefaultPort && oport == "" {
		return nhost == ohost
	}

	return n.Host == o.Host
}

// GetControlPlaneHostPort retrieves the ControlPlaneConfiguration host and port split defined in the cluster.Spec.
func GetControlPlaneHostPort(pHost string, defaultPort string) (string, string, error) {
	host, port, err := net.SplitHostPort(pHost)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			host = pHost
			port = defaultPort
			err = nil
		} else {
			return "", "", fmt.Errorf("host %s is invalid: %v", pHost, err.Error())
		}
	}
	return host, port, err
}

type WorkerNodeGroupConfiguration struct {
	// Name refers to the name of the worker node group
	Name string `json:"name,omitempty"`
	// Count defines the number of desired worker nodes. Defaults to 1.
	Count *int `json:"count,omitempty"`
	// AutoScalingConfiguration defines the auto scaling configuration
	AutoScalingConfiguration *AutoScalingConfiguration `json:"autoscalingConfiguration,omitempty"`
	// MachineGroupRef defines the machine group configuration for the worker nodes.
	MachineGroupRef *Ref `json:"machineGroupRef,omitempty"`
	// Taints define the set of taints to be applied on worker nodes
	Taints []corev1.Taint `json:"taints,omitempty"`
	// Labels define the labels to assign to the node
	Labels map[string]string `json:"labels,omitempty"`
	// UpgradeRolloutStrategy determines the rollout strategy to use for rolling upgrades
	// and related parameters/knobs
	UpgradeRolloutStrategy *WorkerNodesUpgradeRolloutStrategy `json:"upgradeRolloutStrategy,omitempty"`
}

func generateWorkerNodeGroupKey(c WorkerNodeGroupConfiguration) (key string) {
	key = c.Name
	if c.MachineGroupRef != nil {
		key += c.MachineGroupRef.Kind + c.MachineGroupRef.Name
	}
	if c.AutoScalingConfiguration != nil {
		key += "autoscaling" + strconv.Itoa(c.AutoScalingConfiguration.MaxCount) + strconv.Itoa(c.AutoScalingConfiguration.MinCount)
	}
	if c.Count == nil {
		return "nil" + key
	}
	return strconv.Itoa(*c.Count) + key
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
			if !MapEqual(m[nodeGroup.Name], nodeGroup.Labels) {
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
	Nodes     *Nodes     `json:"nodes,omitempty"`
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

	return n.Pods.Equal(&o.Pods) &&
		n.Services.Equal(&o.Services) &&
		n.DNS.Equal(&o.DNS) &&
		n.Nodes.Equal(o.Nodes)
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

	if n.PolicyEnforcementMode != o.PolicyEnforcementMode {
		return false
	}

	if n.EgressMasqueradeInterfaces != o.EgressMasqueradeInterfaces {
		return false
	}

	oSkipUpgradeIsFalse := o.SkipUpgrade == nil || !*o.SkipUpgrade
	nSkipUpgradeIsFalse := n.SkipUpgrade == nil || !*n.SkipUpgrade

	// We consider nil to be false in equality checks. Here we're checking if o is false then
	// n must be false and vice-versa. If neither of these are true, then both o and n must be
	// true so we don't need an explicit check.
	if oSkipUpgradeIsFalse && !nSkipUpgradeIsFalse || !oSkipUpgradeIsFalse && nSkipUpgradeIsFalse {
		return false
	}

	return true
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

func UsersSliceEqual(a, b []UserConfiguration) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string][]string, len(a))
	for _, v := range a {
		m[v.Name] = v.SshAuthorizedKeys
	}
	for _, v := range b {
		if _, ok := m[v.Name]; !ok {
			return false
		}
		if !SliceEqual(v.SshAuthorizedKeys, m[v.Name]) {
			return false
		}
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

type Nodes struct {
	// CIDRMaskSize defines the mask size for node cidr in the cluster, default for ipv4 is 24. This is an optional field
	CIDRMaskSize *int `json:"cidrMaskSize,omitempty"`
}

// Equal compares two Nodes definitions and return true if the are equivalent.
func (n *Nodes) Equal(o *Nodes) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}

	if n.CIDRMaskSize == o.CIDRMaskSize {
		return true
	}
	if n.CIDRMaskSize == nil || o.CIDRMaskSize == nil {
		return false
	}

	return *n.CIDRMaskSize == *o.CIDRMaskSize
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
	Kube123 KubernetesVersion = "1.23"
	Kube124 KubernetesVersion = "1.24"
	Kube125 KubernetesVersion = "1.25"
	Kube126 KubernetesVersion = "1.26"
	Kube127 KubernetesVersion = "1.27"
)

// KubeVersionToSemver converts kube version to semver for comparisons.
func KubeVersionToSemver(kubeVersion KubernetesVersion) (*semver.Version, error) {
	// appending the ".0" as the patch version to have a valid semver string and use those semvers for comparison
	return semver.New(string(kubeVersion) + ".0")
}

type CNI string

type CiliumPolicyEnforcementMode string

type CNIConfig struct {
	Cilium   *CiliumConfig   `json:"cilium,omitempty"`
	Kindnetd *KindnetdConfig `json:"kindnetd,omitempty"`
}

type CiliumConfig struct {
	// PolicyEnforcementMode determines communication allowed between pods. Accepted values are default, always, never.
	PolicyEnforcementMode CiliumPolicyEnforcementMode `json:"policyEnforcementMode,omitempty"`

	// EgressMasquaradeInterfaces determines which network interfaces are used for masquerading. Accepted values are a valid interface name or interface prefix.
	// +optional
	EgressMasqueradeInterfaces string `json:"egressMasqueradeInterfaces,omitempty"`

	// SkipUpgrade indicicates that Cilium maintenance should be skipped during upgrades. This can
	// be used when operators wish to self manage the Cilium installation.
	// +optional
	SkipUpgrade *bool `json:"skipUpgrade,omitempty"`
}

// IsManaged returns true if SkipUpgrade is nil or false indicating EKS-A is responsible for
// the Cilium installation.
func (n *CiliumConfig) IsManaged() bool {
	return n.SkipUpgrade == nil || !*n.SkipUpgrade
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

// ClusterStatus defines the observed state of Cluster.
type ClusterStatus struct {
	// Descriptive message about a fatal problem while reconciling a cluster
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
	// EksdReleaseRef defines the properties of the EKS-D object on the cluster
	EksdReleaseRef *EksdReleaseRef `json:"eksdReleaseRef,omitempty"`
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// ReconciledGeneration represents the .metadata.generation the last time the
	// cluster was successfully reconciled. It is the latest generation observed
	// by the controller.
	// NOTE: This field was added for internal use and we do not provide guarantees
	// to its behavior if changed externally. Its meaning and implementation are
	// subject to change in the future.
	ReconciledGeneration int64 `json:"reconciledGeneration,omitempty"`

	// ChildrenReconciledGeneration represents the sum of the .metadata.generation
	// for all the linked objects for the cluster, observed the last time the
	// cluster was successfully reconciled.
	// NOTE: This field was added for internal use and we do not provide guarantees
	// to its behavior if changed externally. Its meaning and implementation are
	// subject to change in the future.
	ChildrenReconciledGeneration int64 `json:"childrenReconciledGeneration,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
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

type BundlesRef struct {
	// APIVersion refers to the Bundles APIVersion
	APIVersion string `json:"apiVersion"`
	// Name refers to the name of the Bundles object in the cluster
	Name string `json:"name"`
	// Namespace refers to the Bundles's namespace
	Namespace string `json:"namespace"`
}

func (b *BundlesRef) Equal(o *BundlesRef) bool {
	if b == nil || o == nil {
		return b == o
	}

	return b.APIVersion == o.APIVersion && b.Name == o.Name && b.Namespace == o.Namespace
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
// Interface for getting DatacenterRef fields for Cluster type.
type ProviderRefAccessor interface {
	Kind() string
	Name() string
}

// +kubebuilder:object:generate=false
// Interface for getting Kind field for Cluster type.
type KindAccessor interface {
	Kind() string
	ExpectedKind() string
}

// PackageConfiguration for installing EKS Anywhere curated packages.
type PackageConfiguration struct {
	// Disable package controller on cluster
	Disable bool `json:"disable,omitempty"`

	// Controller package controller configuration
	Controller *PackageControllerConfiguration `json:"controller,omitempty"`

	// Cronjob for ecr token refresher
	CronJob *PackageControllerCronJob `json:"cronjob,omitempty"`
}

// Equal for PackageConfiguration.
func (n *PackageConfiguration) Equal(o *PackageConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Disable == o.Disable && n.Controller.Equal(o.Controller) && n.CronJob.Equal(o.CronJob)
}

// PackageControllerConfiguration configure aspects of package controller.
type PackageControllerConfiguration struct {
	// Repository package controller repository
	Repository string `json:"repository,omitempty"`

	// Tag package controller tag
	Tag string `json:"tag,omitempty"`

	// Digest package controller digest
	Digest string `json:"digest,omitempty"`

	// DisableWebhooks on package controller
	DisableWebhooks bool `json:"disableWebhooks,omitempty"`

	// Env of package controller in the format `key=value`
	Env []string `json:"env,omitempty"`

	// Resources of package controller
	Resources PackageControllerResources `json:"resources,omitempty"`
}

// Equal for PackageControllerConfiguration.
func (n *PackageControllerConfiguration) Equal(o *PackageControllerConfiguration) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Repository == o.Repository && n.Tag == o.Tag && n.Digest == o.Digest &&
		n.DisableWebhooks == o.DisableWebhooks && SliceEqual(n.Env, o.Env) && n.Resources.Equal(&o.Resources)
}

// PackageControllerResources resource aspects of package controller.
type PackageControllerResources struct {
	// Requests for image resources
	Requests ImageResource `json:"requests,omitempty"`
	Limits   ImageResource `json:"limits,omitempty"`
}

// Equal for PackageControllerResources.
func (n *PackageControllerResources) Equal(o *PackageControllerResources) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Requests.Equal(&o.Requests) && n.Limits.Equal(&o.Limits)
}

// ImageResource resources for container image.
type ImageResource struct {
	// CPU image cpu
	CPU string `json:"cpu,omitempty"`

	// Memory image memory
	Memory string `json:"memory,omitempty"`
}

// Equal for ImageResource.
func (n *ImageResource) Equal(o *ImageResource) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.CPU == o.CPU && n.Memory == o.Memory
}

// PackageControllerCronJob configure aspects of package controller.
type PackageControllerCronJob struct {
	// Repository ecr token refresher repository
	Repository string `json:"repository,omitempty"`

	// Tag ecr token refresher tag
	Tag string `json:"tag,omitempty"`

	// Digest ecr token refresher digest
	Digest string `json:"digest,omitempty"`

	// Disable on cron job
	Disable bool `json:"disable,omitempty"`
}

// Equal for PackageControllerCronJob.
func (n *PackageControllerCronJob) Equal(o *PackageControllerCronJob) bool {
	if n == o {
		return true
	}
	if n == nil || o == nil {
		return false
	}
	return n.Repository == o.Repository && n.Tag == o.Tag && n.Digest == o.Digest && n.Disable == o.Disable
}

// ExternalEtcdConfiguration defines the configuration options for using unstacked etcd topology.
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

// AutoScalingConfiguration defines the configuration for the node autoscaling feature.
type AutoScalingConfiguration struct {
	// MinCount defines the minimum number of nodes for the associated resource group.
	// +optional
	MinCount int `json:"minCount,omitempty"`

	// MaxCount defines the maximum number of nodes for the associated resource group.
	// +optional
	MaxCount int `json:"maxCount,omitempty"`
}

// ControlPlaneUpgradeRolloutStrategy indicates rollout strategy for cluster.
type ControlPlaneUpgradeRolloutStrategy struct {
	Type          string                          `json:"type,omitempty"`
	RollingUpdate ControlPlaneRollingUpdateParams `json:"rollingUpdate,omitempty"`
}

// ControlPlaneRollingUpdateParams is API for rolling update strategy knobs.
type ControlPlaneRollingUpdateParams struct {
	MaxSurge int `json:"maxSurge"`
}

// WorkerNodesUpgradeRolloutStrategy indicates rollout strategy for cluster.
type WorkerNodesUpgradeRolloutStrategy struct {
	Type          string                         `json:"type,omitempty"`
	RollingUpdate WorkerNodesRollingUpdateParams `json:"rollingUpdate,omitempty"`
}

// WorkerNodesRollingUpdateParams is API for rolling update strategy knobs.
type WorkerNodesRollingUpdateParams struct {
	MaxSurge       int `json:"maxSurge"`
	MaxUnavailable int `json:"maxUnavailable"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Cluster is the Schema for the clusters API.
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
// Same as Cluster except stripped down for generation of yaml file during generate clusterconfig.
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

func (c *Cluster) IsSelfManaged() bool {
	return c.Spec.ManagementCluster.Name == "" || c.Spec.ManagementCluster.Name == c.Name
}

func (c *Cluster) SetManagedBy(managementClusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}

	c.Annotations[managementAnnotation] = managementClusterName
	c.Spec.ManagementCluster.Name = managementClusterName
}

func (c *Cluster) SetSelfManaged() {
	c.Spec.ManagementCluster.Name = c.Name
}

func (c *ClusterGenerate) SetSelfManaged() {
	c.Spec.ManagementCluster.Name = c.Name
}

func (c *Cluster) ManagementClusterEqual(s2 *Cluster) bool {
	return c.IsSelfManaged() && s2.IsSelfManaged() || c.Spec.ManagementCluster.Equal(s2.Spec.ManagementCluster)
}

// IsSingleNode checks if the cluster has only a single node specified between the controlplane and worker nodes.
func (c *Cluster) IsSingleNode() bool {
	return c.Spec.ControlPlaneConfiguration.Count == 1 &&
		len(c.Spec.WorkerNodeGroupConfigurations) <= 0
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

// IsManaged returns true if the Cluster is not self managed.
func (c *Cluster) IsManaged() bool {
	return !c.IsSelfManaged()
}

// ManagedBy returns the Cluster's management cluster's name.
func (c *Cluster) ManagedBy() string {
	return c.Spec.ManagementCluster.Name
}

// +kubebuilder:object:root=true
// ClusterList contains a list of Cluster.
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
