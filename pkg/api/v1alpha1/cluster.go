package v1alpha1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	_ "regexp"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	ClusterKind         = "Cluster"
	YamlSeparator       = "---"
	RegistryMirrorCAKey = "EKSA_REGISTRY_MIRROR_CA"
)

// +kubebuilder:object:generate=false
type ClusterGenerateOpt func(config *ClusterGenerate)

// Used for generating yaml for generate clusterconfig command
func NewClusterGenerate(clusterName string, opts ...ClusterGenerateOpt) *ClusterGenerate {
	config := &ClusterGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       ClusterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: ClusterSpec{
			KubernetesVersion: Kube121,
			ClusterNetwork: ClusterNetwork{
				Pods: Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
				CNI: Cilium,
			},
		},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
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

	return c
}

var clusterConfigValidations = []func(*Cluster) error{
	validateControlPlaneReplicas,
	validateWorkerNodeGroups,
	validateNetworking,
	validateGitOps,
	validateEtcdReplicas,
	validateIdentityProviderRefs,
	validateProxyConfig,
	validateMirrorConfig,
}

func GetClusterConfig(fileName string) (*Cluster, error) {
	clusterConfig := &Cluster{}
	err := ParseClusterConfig(fileName, clusterConfig)
	if err != nil {
		return clusterConfig, err
	}

	return clusterConfig, nil
}

func GetAndValidateClusterConfig(fileName string) (*Cluster, error) {
	clusterConfig, err := ValidateClusterConfig(fileName)
	if err != nil {
		return clusterConfig, err
	}
	return GetClusterConfig(fileName)
}

func ValidateClusterConfig(fileName string) (*Cluster, error) {
	clusterConfig := &Cluster{}
	err := ParseClusterConfig(fileName, clusterConfig)
	if err != nil {
		return nil, err
	}
	err = ValidateClusterConfigContent(clusterConfig)
	if err != nil {
		return nil, err
	}
	return clusterConfig, err
}

func ValidateClusterConfigContent(clusterConfig *Cluster) error {
	for _, v := range clusterConfigValidations {
		if err := v(clusterConfig); err != nil {
			return err
		}
	}
	return nil
}

func ParseClusterConfig(fileName string, clusterConfig KindAccessor) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		if err = yaml.UnmarshalStrict([]byte(c), clusterConfig); err == nil {
			if clusterConfig.Kind() == clusterConfig.ExpectedKind() {
				return nil
			}
		}
		_ = yaml.Unmarshal([]byte(c), clusterConfig) // this is to check if there is a bad spec in the file
		if clusterConfig.Kind() == clusterConfig.ExpectedKind() {
			return fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	return fmt.Errorf("unable to find kind %v in file", clusterConfig.ExpectedKind())
}

func (c *Cluster) HasOverrideClusterSpecFile() bool {
	return c.Spec.OverrideClusterSpecFile != ""
}

func (c *Cluster) ReadOverrideClusterSpecFile() (string, error) {
	content, err := ioutil.ReadFile(c.Spec.OverrideClusterSpecFile)
	if err != nil {
		return "", fmt.Errorf("unable to read override configuration: %v", err)
	}
	return string(content), nil
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
	return c.Spec.RegistryMirrorConfiguration.Endpoint + imageUrl.Path
}

func (c *Cluster) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
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
	return nil
}

func validateWorkerNodeGroups(clusterConfig *Cluster) error {
	if len(clusterConfig.Spec.WorkerNodeGroupConfigurations) <= 0 {
		return errors.New("worker node group must be specified")
	}
	if len(clusterConfig.Spec.WorkerNodeGroupConfigurations) > 1 {
		return errors.New("only one worker node group is supported at this time")
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
	if clusterConfig.Spec.ClusterNetwork.CNI == "" {
		return errors.New("cni not specified or empty")
	}
	if _, ok := validCNIs[clusterConfig.Spec.ClusterNetwork.CNI]; !ok {
		return fmt.Errorf("cni %s not supported", clusterConfig.Spec.ClusterNetwork.CNI)
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
	ip, port, err := net.SplitHostPort(proxy)
	if err != nil {
		return fmt.Errorf("proxy %s is invalid, please provide a valid proxy in the format proxy_ip:port", proxy)
	}
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("proxy ip %s is invalid, please provide a valid proxy ip", ip)
	}
	if p, err := strconv.Atoi(port); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("proxy port %s is invalid, please provide a valid proxy ip", port)
	}
	return nil
}

func validateMirrorConfig(clusterConfig *Cluster) error {
	if clusterConfig.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}
	if clusterConfig.Spec.RegistryMirrorConfiguration.Endpoint == "" {
		return errors.New("no value set for ECRMirror.Endpoint")
	}
	if caCert, set := os.LookupEnv(RegistryMirrorCAKey); set && len(caCert) > 0 {
		content, err := ioutil.ReadFile(caCert)
		if err != nil {
			return fmt.Errorf("error reading the ca cert file %s: %v", caCert, err)
		}
		clusterConfig.Spec.RegistryMirrorConfiguration.CACertContent = string(content)
	} else {
		logger.Info(fmt.Sprintf("Warning: %s environment variable is not set, TLS verification will be disabled", RegistryMirrorCAKey))
	}
	return nil
}

func validateIdentityProviderRefs(clusterConfig *Cluster) error {
	refs := clusterConfig.Spec.IdentityProviderRefs
	if len(refs) == 0 {
		return nil
	}
	// Only 1 ref of type OIDCConfig is supported as of now
	if len(refs) > 1 {
		return errors.New("multiple identityProviderRefs not supported at this time")
	}
	if refs[0].Kind != OIDCConfigKind {
		return errors.New("only OIDCConfig Kind is supported at this time")
	}
	if refs[0].Name == "" {
		return errors.New("specify a valid name for OIDCConfig identityProviderRef")
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
