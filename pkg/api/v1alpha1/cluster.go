package v1alpha1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const (
	ClusterKind         = "Cluster"
	YamlSeparator       = "\n---\n"
	RegistryMirrorCAKey = "EKSA_REGISTRY_MIRROR_CA"
)

// +kubebuilder:object:generate=false
type ClusterGenerateOpt func(config *ClusterGenerate)

// Used for generating yaml for generate clusterconfig command
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
		c.Spec.WorkerNodeGroupConfigurations = []WorkerNodeGroupConfiguration{{Count: count}}
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

func NewCluster(clusterName string) *Cluster {
	c := &Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       ClusterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: ClusterSpec{
			KubernetesVersion: Kube119,
		},
		Status: ClusterStatus{},
	}
	c.SetSelfManaged()

	return c
}

var clusterConfigValidations = []func(*Cluster) error{
	validateClusterConfigName,
	validateControlPlaneReplicas,
	validateWorkerNodeGroups,
	validateNetworking,
	validateGitOps,
	validateEtcdReplicas,
	validateIdentityProviderRefs,
	validateProxyConfig,
	validateMirrorConfig,
	validatePodIAMConfig,
	validateControlPlaneLabels,
}

// GetClusterConfig parses a Cluster object from a multiobject yaml file in disk
// and sets defaults if necessary
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

// GetClusterConfigFromContent parses a Cluster object from a multiobject yaml content
func GetClusterConfigFromContent(content []byte) (*Cluster, error) {
	clusterConfig := &Cluster{}
	err := ParseClusterConfigFromContent(content, clusterConfig)
	if err != nil {
		return clusterConfig, err
	}
	return clusterConfig, nil
}

// GetClusterConfig parses a Cluster object from a multiobject yaml file in disk
// sets defaults if necessary and validates the Cluster
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

// GetClusterDefaultKubernetesVersion returns the default kubernetes version for a Cluster
func GetClusterDefaultKubernetesVersion() KubernetesVersion {
	return Kube122
}

// ValidateClusterConfigContent validates a Cluster object without modifying it
// Some of the validations are a bit heavy and need a network connection
func ValidateClusterConfigContent(clusterConfig *Cluster) error {
	for _, v := range clusterConfigValidations {
		if err := v(clusterConfig); err != nil {
			return err
		}
	}
	return nil
}

// ParseClusterConfig unmarshalls an API object implementing the KindAccessor interface
// from a multiobject yaml file in disk. It doesn't set defaults nor validates the object
func ParseClusterConfig(fileName string, clusterConfig KindAccessor) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read file due to: %v", err)
	}

	if err = ParseClusterConfigFromContent(content, clusterConfig); err != nil {
		return fmt.Errorf("unable to parse %s file: %v", fileName, err)
	}

	return nil
}

// ParseClusterConfigFromContent unmarshalls an API object implementing the KindAccessor interface
// from a multiobject yaml content. It doesn't set defaults nor validates the object
func ParseClusterConfigFromContent(content []byte, clusterConfig KindAccessor) error {
	for _, c := range strings.Split(string(content), YamlSeparator) {
		if err := yaml.Unmarshal([]byte(c), clusterConfig); err != nil {
			return err
		}

		if clusterConfig.Kind() == clusterConfig.ExpectedKind() {
			return yaml.UnmarshalStrict([]byte(c), clusterConfig)
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

func (c *Cluster) UseImageMirror(defaultImage string) string {
	if c.Spec.RegistryMirrorConfiguration == nil {
		return defaultImage
	}
	imageUrl, _ := url.Parse("https://" + defaultImage)
	return net.JoinHostPort(c.Spec.RegistryMirrorConfiguration.Endpoint, c.Spec.RegistryMirrorConfiguration.Port) + imageUrl.Path
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

func validateWorkerNodeGroups(clusterConfig *Cluster) error {
	workerNodeGroupConfigs := clusterConfig.Spec.WorkerNodeGroupConfigurations
	if len(workerNodeGroupConfigs) <= 0 {
		return errors.New("worker node group must be specified")
	}
	workerNodeGroupNames := make(map[string]bool, len(workerNodeGroupConfigs))
	noExecuteNoScheduleTaintedNodeGroups := make(map[string]struct{})
	for i, workerNodeGroupConfig := range workerNodeGroupConfigs {
		if workerNodeGroupConfig.Name == "" {
			return errors.New("must specify name for worker nodes")
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
	if len(noExecuteNoScheduleTaintedNodeGroups) == len(workerNodeGroupConfigs) {
		return errors.New("at least one WorkerNodeGroupConfiguration must not have NoExecute and/or NoSchedule taints")
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
	if len(clusterConfig.Spec.ClusterNetwork.Pods.CidrBlocks) <= 0 {
		return errors.New("pods CIDR block not specified or empty")
	}
	if len(clusterConfig.Spec.ClusterNetwork.Services.CidrBlocks) <= 0 {
		return errors.New("services CIDR block not specified or empty")
	}
	if len(clusterConfig.Spec.ClusterNetwork.Pods.CidrBlocks) > 1 {
		return fmt.Errorf("multiple CIDR blocks for Pods are not yet supported")
	}
	if len(clusterConfig.Spec.ClusterNetwork.Services.CidrBlocks) > 1 {
		return fmt.Errorf("multiple CIDR blocks for Services are not yet supported")
	}
	_, _, err := net.ParseCIDR(clusterConfig.Spec.ClusterNetwork.Pods.CidrBlocks[0])
	if err != nil {
		return fmt.Errorf("invalid CIDR block format for Pods: %s. Please specify a valid CIDR block for pod subnet", clusterConfig.Spec.ClusterNetwork.Pods)
	}
	_, _, err = net.ParseCIDR(clusterConfig.Spec.ClusterNetwork.Services.CidrBlocks[0])
	if err != nil {
		return fmt.Errorf("invalid CIDR block for Services: %s. Please specify a valid CIDR block for service subnet", clusterConfig.Spec.ClusterNetwork.Services)
	}

	return validateCNIPlugin(clusterConfig.Spec.ClusterNetwork)
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
	ip, port, err := net.SplitHostPort(proxyHost)
	if err != nil {
		return fmt.Errorf("proxy %s is invalid, please provide a valid proxy in the format proxy_ip:port", proxy)
	}
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("proxy ip %s is invalid, please provide a valid proxy ip", ip)
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

	if clusterConfig.Spec.RegistryMirrorConfiguration.InsecureSkipVerify && clusterConfig.Spec.DatacenterRef.Kind != constants.SnowProviderName {
		return errors.New("insecureSkipVerify is only supported for snow provider")
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
	if gitOpsRef.Kind != GitOpsConfigKind {
		return errors.New("only GitOpsConfig Kind is supported at this time")
	}
	if gitOpsRef.Name == "" {
		return errors.New("GitOpsConfig name can't be empty; specify a valid name for GitOpsConfig")
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
