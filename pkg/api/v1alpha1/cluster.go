package v1alpha1

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/semver"
)

// constants defined for cluster.go.
const (
	ClusterKind              = "Cluster"
	RegistryMirrorCAKey      = "EKSA_REGISTRY_MIRROR_CA"
	podSubnetNodeMaskMaxDiff = 16
)

var re = regexp.MustCompile(constants.DefaultCuratedPackagesRegistryRegex)

// +kubebuilder:object:generate=false
type ClusterGenerateOpt func(config *ClusterGenerate)

// Used for generating yaml for generate clusterconfig command.
func NewClusterGenerate(clusterName string, opts ...ClusterGenerateOpt) *ClusterGenerate {
	clusterConfig := &ClusterGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       ClusterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: ClusterSpec{
			KubernetesVersion: GetClusterDefaultKubernetesVersion(),
			ClusterNetwork: ClusterNetwork{
				Pods: Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
			},
		},
	}
	clusterConfig.SetSelfManaged()

	for _, opt := range opts {
		opt(clusterConfig)
	}
	return clusterConfig
}

func ControlPlaneConfigCount(count int) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.ControlPlaneConfiguration.Count = count
	}
}

func ExternalETCDConfigCount(count int) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.ExternalEtcdConfiguration = &ExternalEtcdConfiguration{
			Count: count,
		}
	}
}

func WorkerNodeConfigCount(count int) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.WorkerNodeGroupConfigurations = []WorkerNodeGroupConfiguration{{Count: &count}}
	}
}

func WorkerNodeConfigName(name string) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.WorkerNodeGroupConfigurations[0].Name = name
	}
}

func WithClusterEndpoint() ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.ControlPlaneConfiguration.Endpoint = &Endpoint{Host: ""}
	}
}

// WithCPUpgradeRolloutStrategy allows add UpgradeRolloutStrategy option to cluster config under ControlPlaneConfiguration.
func WithCPUpgradeRolloutStrategy(maxSurge int, maxUnavailable int) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: ControlPlaneRollingUpdateParams{MaxSurge: maxSurge}}
	}
}

func WithDatacenterRef(ref ProviderRefAccessor) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.DatacenterRef = Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
	}
}

func WithSharedMachineGroupRef(ref ProviderRefAccessor) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.ControlPlaneConfiguration.MachineGroupRef = &Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
		c.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = &Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
	}
}

func WithCPMachineGroupRef(ref ProviderRefAccessor) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.ControlPlaneConfiguration.MachineGroupRef = &Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
	}
}

func WithWorkerMachineGroupRef(ref ProviderRefAccessor) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = &Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
	}
}

// WithWorkerMachineUpgradeRolloutStrategy allows add UpgradeRolloutStrategy option to cluster config under WorkerNodeGroupConfiguration.
func WithWorkerMachineUpgradeRolloutStrategy(maxSurge int, maxUnavailable int) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		c.Spec.WorkerNodeGroupConfigurations[0].UpgradeRolloutStrategy = &WorkerNodesUpgradeRolloutStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: maxSurge, MaxUnavailable: maxUnavailable},
		}
	}
}

func WithEtcdMachineGroupRef(ref ProviderRefAccessor) ClusterGenerateOpt {
	return func(c *ClusterGenerate) {
		if c.Spec.ExternalEtcdConfiguration != nil {
			c.Spec.ExternalEtcdConfiguration.MachineGroupRef = &Ref{
				Kind: ref.Kind(),
				Name: ref.Name(),
			}
		}
	}
}

var clusterConfigValidations = []func(*Cluster) error{
	validateClusterConfigName,
	validateControlPlaneEndpoint,
	validateExternalEtcdSupport,
	validateMachineGroupRefs,
	validateControlPlaneReplicas,
	validateWorkerNodeGroups,
	validateNetworking,
	validateGitOps,
	validateEtcdReplicas,
	validateIdentityProviderRefs,
	validateProxyConfig,
	validateMirrorConfig,
	validatePodIAMConfig,
	validateCPUpgradeRolloutStrategy,
	validateControlPlaneLabels,
	validatePackageControllerConfiguration,
	validateEksaVersion,
	validateControlPlaneCertSANs,
}

// GetClusterConfig parses a Cluster object from a multiobject yaml file in disk
// and sets defaults if necessary.
func GetClusterConfig(fileName string) (*Cluster, error) {
	clusterConfig := &Cluster{}
	err := ParseClusterConfig(fileName, clusterConfig)
	if err != nil {
		return clusterConfig, err
	}
	if err := setClusterDefaults(clusterConfig); err != nil {
		return clusterConfig, err
	}
	return clusterConfig, nil
}

// GetClusterConfig parses a Cluster object from a multiobject yaml file in disk
// sets defaults if necessary and validates the Cluster.
func GetAndValidateClusterConfig(fileName string) (*Cluster, error) {
	clusterConfig, err := GetClusterConfig(fileName)
	if err != nil {
		return nil, err
	}
	err = ValidateClusterConfigContent(clusterConfig)
	if err != nil {
		return nil, err
	}

	return clusterConfig, nil
}

// GetClusterDefaultKubernetesVersion returns the default kubernetes version for a Cluster.
func GetClusterDefaultKubernetesVersion() KubernetesVersion {
	return Kube128
}

// ValidateClusterConfigContent validates a Cluster object without modifying it
// Some of the validations are a bit heavy and need a network connection.
func ValidateClusterConfigContent(clusterConfig *Cluster) error {
	for _, v := range clusterConfigValidations {
		if err := v(clusterConfig); err != nil {
			return err
		}
	}
	return nil
}

// ParseClusterConfig unmarshalls an API object implementing the KindAccessor interface
// from a multiobject yaml file in disk. It doesn't set defaults nor validates the object.
func ParseClusterConfig(fileName string, clusterConfig KindAccessor) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read file due to: %v", err)
	}

	if err = ParseClusterConfigFromContent(content, clusterConfig); err != nil {
		return fmt.Errorf("unable to parse %s file: %v", fileName, err)
	}

	return nil
}

type kindObject struct {
	Kind string `json:"kind,omitempty"`
}

// ParseClusterConfigFromContent unmarshalls an API object implementing the KindAccessor interface
// from a multiobject yaml content. It doesn't set defaults nor validates the object.
func ParseClusterConfigFromContent(content []byte, clusterConfig KindAccessor) error {
	r := yamlutil.NewYAMLReader(bufio.NewReader(bytes.NewReader(content)))
	for {
		d, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		k := &kindObject{}

		if err := yaml.Unmarshal(d, k); err != nil {
			return err
		}

		if k.Kind == clusterConfig.ExpectedKind() {
			return yaml.UnmarshalStrict(d, clusterConfig)
		}
	}

	return fmt.Errorf("yamlop content is invalid or does not contain kind %s", clusterConfig.ExpectedKind())
}

func (c *Cluster) PauseReconcile() {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[pausedAnnotation] = "true"
}

func (c *Cluster) ClearPauseAnnotation() {
	if c.Annotations != nil {
		delete(c.Annotations, pausedAnnotation)
	}
}

// AddManagedByCLIAnnotation adds the managed-by-cli annotation to the cluster.
func (c *Cluster) AddManagedByCLIAnnotation() {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[ManagedByCLIAnnotation] = "true"
}

// ClearManagedByCLIAnnotation removes the managed-by-cli annotation from the cluster.
func (c *Cluster) ClearManagedByCLIAnnotation() {
	if c.Annotations != nil {
		delete(c.Annotations, ManagedByCLIAnnotation)
	}
}

// RegistryAuth returns whether registry requires authentication or not.
func (c *Cluster) RegistryAuth() bool {
	if c.Spec.RegistryMirrorConfiguration == nil {
		return false
	}
	return c.Spec.RegistryMirrorConfiguration.Authenticate
}

func (c *Cluster) ProxyConfiguration() map[string]string {
	if c.Spec.ProxyConfiguration == nil {
		return nil
	}
	noProxyList := append(c.Spec.ProxyConfiguration.NoProxy, c.Spec.ClusterNetwork.Pods.CidrBlocks...)
	noProxyList = append(noProxyList, c.Spec.ClusterNetwork.Services.CidrBlocks...)
	if c.Spec.ControlPlaneConfiguration.Endpoint != nil && c.Spec.ControlPlaneConfiguration.Endpoint.Host != "" {
		noProxyList = append(
			noProxyList,
			c.Spec.ControlPlaneConfiguration.Endpoint.Host,
		)
	}
	return map[string]string{
		"HTTP_PROXY":  c.Spec.ProxyConfiguration.HttpProxy,
		"HTTPS_PROXY": c.Spec.ProxyConfiguration.HttpsProxy,
		"NO_PROXY":    strings.Join(noProxyList[:], ","),
	}
}

func (c *Cluster) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func ValidateClusterName(clusterName string) error {
	// this regex will not work for AWS provider as CFN has restrictions with UPPERCASE chars;
	// if you are using AWS provider please use only lowercase chars
	allowedClusterNameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]+$`)
	if !allowedClusterNameRegex.MatchString(clusterName) {
		return fmt.Errorf("%v is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. ", clusterName)
	}
	return nil
}

func ValidateClusterNameLength(clusterName string) error {
	// vSphere has the maximum length for clusters to be 80 chars
	if len(clusterName) > 80 {
		return fmt.Errorf("number of characters in %v should be less than 81", clusterName)
	}
	return nil
}

func validateClusterConfigName(clusterConfig *Cluster) error {
	err := ValidateClusterName(clusterConfig.ObjectMeta.Name)
	if err != nil {
		return fmt.Errorf("failed to validate cluster config name: %v", err)
	}
	err = ValidateClusterNameLength(clusterConfig.ObjectMeta.Name)
	if err != nil {
		return fmt.Errorf("failed to validate cluster config name: %v", err)
	}
	return nil
}

func validateExternalEtcdSupport(cluster *Cluster) error {
	if cluster.Spec.DatacenterRef.Kind == TinkerbellDatacenterKind {
		if cluster.Spec.ExternalEtcdConfiguration != nil {
			return errors.New("tinkerbell external etcd configuration is unsupported")
		}
	}

	return nil
}

func validateMachineGroupRefs(cluster *Cluster) error {
	if cluster.Spec.DatacenterRef.Kind != DockerDatacenterKind {
		if cluster.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef control plane machines")
		}
		for _, workerNodeGroupConfiguration := range cluster.Spec.WorkerNodeGroupConfigurations {
			if workerNodeGroupConfiguration.MachineGroupRef == nil {
				return errors.New("must specify machineGroupRef for worker nodes")
			}
		}
		if cluster.Spec.ExternalEtcdConfiguration != nil && cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef for etcd machines")
		}
	}

	return nil
}

func validateControlPlaneReplicas(clusterConfig *Cluster) error {
	if clusterConfig.Spec.ControlPlaneConfiguration.Count <= 0 {
		return errors.New("control plane node count must be positive")
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration != nil {
		// For unstacked/external etcd, controlplane replicas can be any number including even numbers.
		return nil
	}
	if clusterConfig.Spec.ControlPlaneConfiguration.Count%2 == 0 {
		return errors.New("control plane node count cannot be an even number")
	}
	if clusterConfig.Spec.ControlPlaneConfiguration.Count != 3 && clusterConfig.Spec.ControlPlaneConfiguration.Count != 5 {
		if clusterConfig.Spec.DatacenterRef.Kind != DockerDatacenterKind {
			logger.Info("Warning: The recommended number of control plane nodes is 3 or 5")
		}
	}
	return nil
}

func validateControlPlaneLabels(clusterConfig *Cluster) error {
	if err := validateNodeLabels(clusterConfig.Spec.ControlPlaneConfiguration.Labels, field.NewPath("spec", "controlPlaneConfiguration", "labels")); err != nil {
		return fmt.Errorf("labels for control plane not valid: %v", err)
	}
	return nil
}

func validateControlPlaneEndpoint(clusterConfig *Cluster) error {
	if (clusterConfig.Spec.ControlPlaneConfiguration.Endpoint == nil || len(clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0) && clusterConfig.Spec.DatacenterRef.Kind != DockerDatacenterKind {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	return nil
}

var domainNameRegex = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)

func validateControlPlaneCertSANs(cfg *Cluster) error {
	var invalid []string
	for _, san := range cfg.Spec.ControlPlaneConfiguration.CertSANs {
		isDomain := domainNameRegex.MatchString(san)
		isIP := net.ParseIP(san)

		if !isDomain && isIP == nil {
			invalid = append(invalid, san)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("invalid ControlPlaneConfiguration.CertSANs; must be an IP or domain name: [%v]", strings.Join(invalid, ", "))
	}

	return nil
}

func validateWorkerNodeGroups(clusterConfig *Cluster) error {
	workerNodeGroupConfigs := clusterConfig.Spec.WorkerNodeGroupConfigurations
	if len(workerNodeGroupConfigs) <= 0 {
		if clusterConfig.Spec.DatacenterRef.Kind == TinkerbellDatacenterKind {
			logger.Info("Warning: No configurations provided for worker node groups, pods will be scheduled on control-plane nodes")
		} else {
			return fmt.Errorf("WorkerNodeGroupConfigs cannot be empty for %s", clusterConfig.Spec.DatacenterRef.Kind)
		}
	}

	workerNodeGroupNames := make(map[string]bool, len(workerNodeGroupConfigs))
	noExecuteNoScheduleTaintedNodeGroups := make(map[string]struct{})
	for i, workerNodeGroupConfig := range workerNodeGroupConfigs {
		if workerNodeGroupConfig.Name == "" {
			return errors.New("must specify name for worker nodes")
		}

		if workerNodeGroupConfig.Count == nil {
			// This block should never fire. If it does, it means we have a bug in how we set our defaults.
			// When Count == nil it should be set to 1 by SetDefaults method prior to reaching validation.
			return errors.New("worker node count must be >= 0")
		}

		if err := validateAutoscalingConfig(&workerNodeGroupConfig); err != nil {
			return fmt.Errorf("validating autoscaling configuration: %v", err)
		}

		if err := validateMDUpgradeRolloutStrategy(&workerNodeGroupConfig); err != nil {
			return fmt.Errorf("validating upgrade rollout strategy configuration: %v", err)
		}

		if workerNodeGroupNames[workerNodeGroupConfig.Name] {
			return errors.New("worker node group names must be unique")
		}

		if len(workerNodeGroupConfig.Taints) != 0 {
			for _, taint := range workerNodeGroupConfig.Taints {
				if taint.Effect == "NoExecute" || taint.Effect == "NoSchedule" {
					noExecuteNoScheduleTaintedNodeGroups[workerNodeGroupConfig.Name] = struct{}{}
				}
			}
		}

		workerNodeGroupField := fmt.Sprintf("workerNodeGroupConfigurations[%d]", i)
		if err := validateNodeLabels(workerNodeGroupConfig.Labels, field.NewPath("spec", workerNodeGroupField, "labels")); err != nil {
			return fmt.Errorf("labels for worker node group %v not valid: %v", workerNodeGroupConfig.Name, err)
		}

		workerNodeGroupNames[workerNodeGroupConfig.Name] = true
	}

	if len(workerNodeGroupConfigs) > 0 && len(noExecuteNoScheduleTaintedNodeGroups) == len(workerNodeGroupConfigs) {
		if clusterConfig.IsSelfManaged() {
			return errors.New("at least one WorkerNodeGroupConfiguration must not have NoExecute and/or NoSchedule taints")
		}
	}

	if len(workerNodeGroupConfigs) == 0 && len(clusterConfig.Spec.ControlPlaneConfiguration.Taints) != 0 {
		return errors.New("cannot taint control plane when there is no worker node")
	}

	if len(workerNodeGroupConfigs) == 0 && clusterConfig.Spec.KubernetesVersion <= Kube121 {
		return errors.New("Empty workerNodeGroupConfigs is not supported for kube version <= 1.21")
	}

	return nil
}

func validateAutoscalingConfig(w *WorkerNodeGroupConfiguration) error {
	if w == nil {
		return nil
	}
	if w.AutoScalingConfiguration == nil && *w.Count < 0 {
		return errors.New("worker node count must be zero or greater if autoscaling is not enabled")
	}
	if w.AutoScalingConfiguration == nil {
		return nil
	}
	if w.AutoScalingConfiguration.MinCount < 0 {
		return errors.New("min count must be non negative")
	}
	if w.AutoScalingConfiguration.MinCount > w.AutoScalingConfiguration.MaxCount {
		return errors.New("min count must be no greater than max count")
	}
	if w.AutoScalingConfiguration.MinCount > *w.Count {
		return errors.New("min count must be less than or equal to count")
	}
	if w.AutoScalingConfiguration.MaxCount < *w.Count {
		return errors.New("max count must be greater than or equal to count")
	}

	return nil
}

func validateNodeLabels(labels map[string]string, fldPath *field.Path) error {
	errList := validation.ValidateLabels(labels, fldPath)
	if len(errList) != 0 {
		return fmt.Errorf("found following errors with labels: %v", errList.ToAggregate().Error())
	}
	return nil
}

func validateEtcdReplicas(clusterConfig *Cluster) error {
	if clusterConfig.Spec.ExternalEtcdConfiguration == nil {
		return nil
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration.Count == 0 {
		return errors.New("no value set for etcd replicas")
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration.Count < 0 {
		return errors.New("etcd replicas cannot be a negative number")
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration.Count%2 == 0 {
		return errors.New("external etcd count cannot be an even number")
	}
	if clusterConfig.Spec.ExternalEtcdConfiguration.Count != 3 && clusterConfig.Spec.ExternalEtcdConfiguration.Count != 5 {
		if clusterConfig.Spec.DatacenterRef.Kind != DockerDatacenterKind {
			// only log warning about recommended etcd cluster size for providers other than docker
			logger.Info("Warning: The recommended size of an external etcd cluster is 3 or 5")
		}
	}
	return nil
}

func validateNetworking(clusterConfig *Cluster) error {
	clusterNetwork := clusterConfig.Spec.ClusterNetwork
	if clusterNetwork.CNI == Kindnetd || clusterNetwork.CNIConfig != nil && clusterNetwork.CNIConfig.Kindnetd != nil {
		if clusterConfig.Spec.DatacenterRef.Kind != DockerDatacenterKind {
			return errors.New("kindnetd is only supported on Docker provider for development and testing. For all other providers please use Cilium CNI")
		}
	}
	if len(clusterNetwork.Pods.CidrBlocks) <= 0 {
		return errors.New("pods CIDR block not specified or empty")
	}
	if len(clusterNetwork.Services.CidrBlocks) <= 0 {
		return errors.New("services CIDR block not specified or empty")
	}
	if len(clusterNetwork.Pods.CidrBlocks) > 1 {
		return fmt.Errorf("multiple CIDR blocks for Pods are not yet supported")
	}
	if len(clusterNetwork.Services.CidrBlocks) > 1 {
		return fmt.Errorf("multiple CIDR blocks for Services are not yet supported")
	}
	_, podCIDRIPNet, err := net.ParseCIDR(clusterNetwork.Pods.CidrBlocks[0])
	if err != nil {
		return fmt.Errorf("invalid CIDR block format for Pods: %s. Please specify a valid CIDR block for pod subnet", clusterNetwork.Pods)
	}
	_, serviceCIDRIPNet, err := net.ParseCIDR(clusterNetwork.Services.CidrBlocks[0])
	if err != nil {
		return fmt.Errorf("invalid CIDR block for Services: %s. Please specify a valid CIDR block for service subnet", clusterNetwork.Services)
	}

	if clusterConfig.Spec.DatacenterRef.Kind == SnowDatacenterKind {
		controlPlaneEndpoint := net.ParseIP(clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host)
		if controlPlaneEndpoint == nil {
			return fmt.Errorf("control plane endpoint %s is invalid", clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host)
		}
		if podCIDRIPNet.Contains(controlPlaneEndpoint) {
			return fmt.Errorf("control plane endpoint %s conflicts with pods CIDR block %s", clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, clusterNetwork.Pods.CidrBlocks[0])
		}
		if serviceCIDRIPNet.Contains(controlPlaneEndpoint) {
			return fmt.Errorf("control plane endpoint %s conflicts with services CIDR block %s", clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, clusterNetwork.Services.CidrBlocks[0])
		}
	}

	podMaskSize, _ := podCIDRIPNet.Mask.Size()
	nodeCidrMaskSize := constants.DefaultNodeCidrMaskSize

	if clusterNetwork.Nodes != nil && clusterNetwork.Nodes.CIDRMaskSize != nil {
		nodeCidrMaskSize = *clusterNetwork.Nodes.CIDRMaskSize
	}
	// the pod subnet mask needs to allow one or multiple node-masks
	// i.e. if it has a /24 the node mask must be between 24 and 32 for ipv4
	// the below validations are run by kubeadm and we are bubbling those up here for better customer experience
	if podMaskSize >= nodeCidrMaskSize {
		return fmt.Errorf("the size of pod subnet with mask %d is smaller than or equal to the size of node subnet with mask %d", podMaskSize, nodeCidrMaskSize)
	} else if (nodeCidrMaskSize - podMaskSize) > podSubnetNodeMaskMaxDiff {
		// PodSubnetNodeMaskMaxDiff is limited to 16 due to an issue with uncompressed IP bitmap in core
		// The node subnet mask size must be no more than the pod subnet mask size + 16
		return fmt.Errorf("pod subnet mask (%d) and node-mask (%d) difference is greater than %d", podMaskSize, nodeCidrMaskSize, podSubnetNodeMaskMaxDiff)
	}

	return validateCNIPlugin(clusterNetwork)
}

func validateCNIPlugin(network ClusterNetwork) error {
	if network.CNI != "" {
		if network.CNIConfig != nil {
			return fmt.Errorf("invalid format for cni plugin: both old and new formats used, use only the CNIConfig field")
		}
		logger.Info("Warning: CNI field is deprecated. Provide CNI information through CNIConfig")
		if _, ok := validCNIs[network.CNI]; !ok {
			return fmt.Errorf("cni %s not supported", network.CNI)
		}
		return nil
	}
	return validateCNIConfig(network.CNIConfig)
}

func validateCNIConfig(cniConfig *CNIConfig) error {
	if cniConfig == nil {
		return fmt.Errorf("cni not specified")
	}
	var cniPluginSpecified int
	var allErrs []error

	if cniConfig.Cilium != nil {
		cniPluginSpecified++
		if err := validateCiliumConfig(cniConfig.Cilium); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	if cniConfig.Kindnetd != nil {
		cniPluginSpecified++
	}

	if cniPluginSpecified == 0 {
		allErrs = append(allErrs, fmt.Errorf("no cni plugin specified"))
	} else if cniPluginSpecified > 1 {
		allErrs = append(allErrs, fmt.Errorf("cannot specify more than one cni plugins"))
	}

	if len(allErrs) > 0 {
		aggregate := utilerrors.NewAggregate(allErrs)
		return fmt.Errorf("validating cniConfig: %v", aggregate)
	}

	return nil
}

func validateCiliumConfig(cilium *CiliumConfig) error {
	if cilium == nil {
		return nil
	}

	if !cilium.IsManaged() {
		if cilium.PolicyEnforcementMode != "" {
			return errors.New("when using skipUpgrades for cilium all other fields must be empty")
		}
	}

	if cilium.PolicyEnforcementMode == "" {
		return nil
	}

	if !validCiliumPolicyEnforcementModes[cilium.PolicyEnforcementMode] {
		return fmt.Errorf("cilium policyEnforcementMode \"%s\" not supported", cilium.PolicyEnforcementMode)
	}

	return nil
}

func validateProxyConfig(clusterConfig *Cluster) error {
	if clusterConfig.Spec.ProxyConfiguration == nil {
		return nil
	}
	if clusterConfig.Spec.ProxyConfiguration.HttpProxy == "" {
		return errors.New("no value set for httpProxy")
	}
	if clusterConfig.Spec.ProxyConfiguration.HttpsProxy == "" {
		return errors.New("no value set for httpsProxy")
	}
	if err := validateProxyData(clusterConfig.Spec.ProxyConfiguration.HttpProxy); err != nil {
		return err
	}
	if err := validateProxyData(clusterConfig.Spec.ProxyConfiguration.HttpsProxy); err != nil {
		return err
	}
	return nil
}

func validateProxyData(proxy string) error {
	var proxyHost string
	if strings.HasPrefix(proxy, "http") {
		u, err := url.ParseRequestURI(proxy)
		if err != nil {
			return fmt.Errorf("proxy %s is invalid, please provide a valid URI", proxy)
		}
		proxyHost = u.Host
	} else {
		proxyHost = proxy
	}
	host, port, err := net.SplitHostPort(proxyHost)
	if err != nil {
		return fmt.Errorf("proxy endpoint %s is invalid (%s), please provide a valid proxy address", proxy, err)
	}
	_, err = net.DefaultResolver.LookupIPAddr(context.Background(), host)
	if err != nil && net.ParseIP(host) == nil {
		return fmt.Errorf("proxy endpoint %s is invalid, please provide a valid proxy domain name or ip: %v", host, err)
	}
	if p, err := strconv.Atoi(port); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("proxy port %s is invalid, please provide a valid proxy port", port)
	}
	return nil
}

func validateMirrorConfig(clusterConfig *Cluster) error {
	if clusterConfig.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}
	if clusterConfig.Spec.RegistryMirrorConfiguration.Endpoint == "" {
		return errors.New("no value set for RegistryMirrorConfiguration.Endpoint")
	}

	if !networkutils.IsPortValid(clusterConfig.Spec.RegistryMirrorConfiguration.Port) {
		return fmt.Errorf("registry mirror port %s is invalid, please provide a valid port", clusterConfig.Spec.RegistryMirrorConfiguration.Port)
	}

	mirrorCount := 0
	ociNamespaces := clusterConfig.Spec.RegistryMirrorConfiguration.OCINamespaces
	for _, ociNamespace := range ociNamespaces {
		if ociNamespace.Registry == "" {
			return errors.New("registry can't be set to empty in OCINamespaces")
		}
		if re.MatchString(ociNamespace.Registry) {
			mirrorCount++
			// More than one mirror for curated package would introduce ambiguity in the package controller
			if mirrorCount > 1 {
				return errors.New("only one registry mirror for curated packages is suppported")
			}
		}
	}

	return nil
}

func validateIdentityProviderRefs(clusterConfig *Cluster) error {
	refs := clusterConfig.Spec.IdentityProviderRefs
	if len(refs) == 0 {
		return nil
	}
	for _, ref := range refs {
		if ref.Kind != OIDCConfigKind && ref.Kind != AWSIamConfigKind {
			return fmt.Errorf("kind: %s for identityProviderRef is not supported", ref.Kind)
		}
		if ref.Name == "" {
			return errors.New("specify a valid name for identityProviderRef")
		}
	}
	return nil
}

func validateGitOps(clusterConfig *Cluster) error {
	gitOpsRef := clusterConfig.Spec.GitOpsRef
	if gitOpsRef == nil {
		return nil
	}

	gitOpsRefKind := gitOpsRef.Kind

	if gitOpsRefKind != GitOpsConfigKind && gitOpsRefKind != FluxConfigKind {
		return errors.New("only GitOpsConfig or FluxConfig Kind are supported at this time")
	}

	if gitOpsRef.Name == "" {
		return errors.New("GitOpsRef name can't be empty; specify a valid GitOpsConfig name")
	}
	return nil
}

func validatePodIAMConfig(clusterConfig *Cluster) error {
	if clusterConfig.Spec.PodIAMConfig == nil {
		return nil
	}
	if clusterConfig.Spec.PodIAMConfig.ServiceAccountIssuer == "" {
		return errors.New("ServiceAccount Issuer can't be empty while configuring IAM roles for pods")
	}
	return nil
}

func validateCPUpgradeRolloutStrategy(clusterConfig *Cluster) error {
	cpUpgradeRolloutStrategy := clusterConfig.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy
	if cpUpgradeRolloutStrategy == nil {
		return nil
	}

	if cpUpgradeRolloutStrategy.Type != "RollingUpdate" && cpUpgradeRolloutStrategy.Type != "InPlace" {
		return fmt.Errorf("ControlPlaneConfiguration: only 'RollingUpdate' and 'InPlace' are supported for upgrade rollout strategy type")
	}

	if cpUpgradeRolloutStrategy.RollingUpdate.MaxSurge < 0 {
		return fmt.Errorf("ControlPlaneConfiguration: maxSurge for control plane cannot be a negative value")
	}

	if cpUpgradeRolloutStrategy.RollingUpdate.MaxSurge > 1 {
		return fmt.Errorf("ControlPlaneConfiguration: maxSurge for control plane must be 0 or 1")
	}

	return nil
}

func validateMDUpgradeRolloutStrategy(w *WorkerNodeGroupConfiguration) error {
	if w.UpgradeRolloutStrategy == nil {
		return nil
	}

	if w.UpgradeRolloutStrategy.Type != "RollingUpdate" && w.UpgradeRolloutStrategy.Type != "InPlace" {
		return fmt.Errorf("WorkerNodeGroupConfiguration: only 'RollingUpdate' and 'InPlace' are supported for upgrade rollout strategy type")
	}

	if w.UpgradeRolloutStrategy.RollingUpdate.MaxSurge < 0 || w.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable < 0 {
		return fmt.Errorf("WorkerNodeGroupConfiguration: maxSurge and maxUnavailable values cannot be negative")
	}

	if w.UpgradeRolloutStrategy.RollingUpdate.MaxSurge == 0 && w.UpgradeRolloutStrategy.RollingUpdate.MaxUnavailable == 0 {
		return fmt.Errorf("WorkerNodeGroupConfiguration: maxSurge and maxUnavailable not specified or are 0. maxSurge and maxUnavailable cannot both be 0")
	}

	return nil
}

func validatePackageControllerConfiguration(clusterConfig *Cluster) error {
	if clusterConfig.IsManaged() {
		if clusterConfig.Spec.Packages != nil {
			if clusterConfig.Spec.Packages.Controller != nil {
				return fmt.Errorf("packages: controller should not be specified for a workload cluster")
			}
			if clusterConfig.Spec.Packages.CronJob != nil {
				return fmt.Errorf("packages: cronjob should not be specified for a workload cluster")
			}
		}
	}
	return nil
}

func validateEksaVersion(clusterConfig *Cluster) error {
	if clusterConfig.Spec.BundlesRef != nil && clusterConfig.Spec.EksaVersion != nil {
		return fmt.Errorf("cannot pass both bundlesRef and eksaVersion. New clusters should use eksaVersion instead of bundlesRef")
	}

	if clusterConfig.Spec.EksaVersion != nil {
		_, err := semver.New(string(*clusterConfig.Spec.EksaVersion))
		if err != nil {
			return fmt.Errorf("eksaVersion is not a valid semver")
		}
	}

	return nil
}
