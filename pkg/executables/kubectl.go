package executables

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/pkg/errors"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	rufiov1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/rufiounreleased"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	kubectlPath        = "kubectl"
	timeoutPrecision   = 2
	minimumWaitTimeout = 0.01 // Smallest express-able timeout value given the precision

	networkFaultBaseRetryTime = 10 * time.Second
	networkFaultBackoffFactor = 1.5

	lastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"
)

var (
	capiClustersResourceType             = fmt.Sprintf("clusters.%s", clusterv1.GroupVersion.Group)
	capiProvidersResourceType            = fmt.Sprintf("providers.clusterctl.%s", clusterv1.GroupVersion.Group)
	capiMachinesType                     = fmt.Sprintf("machines.%s", clusterv1.GroupVersion.Group)
	capiMachineDeploymentsType           = fmt.Sprintf("machinedeployments.%s", clusterv1.GroupVersion.Group)
	capiMachineSetsType                  = fmt.Sprintf("machinesets.%s", clusterv1.GroupVersion.Group)
	eksaClusterResourceType              = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereDatacenterResourceType    = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType       = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
	vsphereMachineTemplatesType          = fmt.Sprintf("vspheremachinetemplates.infrastructure.%s", clusterv1.GroupVersion.Group)
	eksaTinkerbellDatacenterResourceType = fmt.Sprintf("tinkerbelldatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaTinkerbellMachineResourceType    = fmt.Sprintf("tinkerbellmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	TinkerbellHardwareResourceType       = fmt.Sprintf("hardware.%s", tinkv1alpha1.GroupVersion.Group)
	rufioMachineResourceType             = fmt.Sprintf("machines.%s", rufiov1alpha1.GroupVersion.Group)
	eksaCloudStackDatacenterResourceType = fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackMachineResourceType    = fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	cloudstackMachineTemplatesType       = fmt.Sprintf("cloudstackmachinetemplates.infrastructure.%s", clusterv1.GroupVersion.Group)
	eksaNutanixDatacenterResourceType    = fmt.Sprintf("nutanixdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaNutanixMachineResourceType       = fmt.Sprintf("nutanixmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaAwsResourceType                  = fmt.Sprintf("awsdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaGitOpsResourceType               = fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaFluxConfigResourceType           = fmt.Sprintf("fluxconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaOIDCResourceType                 = fmt.Sprintf("oidcconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaAwsIamResourceType               = fmt.Sprintf("awsiamconfigs.%s", v1alpha1.GroupVersion.Group)
	etcdadmClustersResourceType          = fmt.Sprintf("etcdadmclusters.%s", etcdv1.GroupVersion.Group)
	bundlesResourceType                  = fmt.Sprintf("bundles.%s", releasev1alpha1.GroupVersion.Group)
	clusterResourceSetResourceType       = fmt.Sprintf("clusterresourcesets.%s", addons.GroupVersion.Group)
	kubeadmControlPlaneResourceType      = fmt.Sprintf("kubeadmcontrolplanes.controlplane.%s", clusterv1.GroupVersion.Group)
	eksdReleaseType                      = fmt.Sprintf("releases.%s", eksdv1alpha1.GroupVersion.Group)
	eksaPackagesType                     = fmt.Sprintf("packages.%s", packagesv1.GroupVersion.Group)
	eksaPackagesBundleControllerType     = fmt.Sprintf("packagebundlecontroller.%s", packagesv1.GroupVersion.Group)
	eksaPackageBundlesType               = fmt.Sprintf("packagebundles.%s", packagesv1.GroupVersion.Group)
	kubectlConnectionRefusedRegex        = regexp.MustCompile("The connection to the server .* was refused")
	kubectlConnectionTimeoutRegex        = regexp.MustCompile("Unable to connect to the server.*timeout.*")
)

type Kubectl struct {
	Executable
	// networkFaultBackoffFactor drives the exponential backoff wait
	// for transient network failures during retry operations.
	networkFaultBackoffFactor float64

	// networkFaultBaseRetryTime drives the base time wait for the
	// exponential backoff for transient network failures during retry operations.
	networkFaultBaseRetryTime time.Duration
}

// KubectlConfigOpt configures Kubectl on construction.
type KubectlConfigOpt func(*Kubectl)

// NewKubectl builds a new Kubectl.
func NewKubectl(executable Executable, opts ...KubectlConfigOpt) *Kubectl {
	k := &Kubectl{
		Executable:                executable,
		networkFaultBackoffFactor: networkFaultBackoffFactor,
		networkFaultBaseRetryTime: networkFaultBaseRetryTime,
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
}

// WithKubectlNetworkFaultBaseRetryTime configures the base time wait for the
// exponential backoff for transient network failures during retry operations.
func WithKubectlNetworkFaultBaseRetryTime(wait time.Duration) KubectlConfigOpt {
	return func(k *Kubectl) {
		k.networkFaultBaseRetryTime = wait
	}
}

// WithNetworkFaultBackoffFactor configures the exponential backoff wait
// for transient network failures during retry operations.
func WithNetworkFaultBackoffFactor(factor float64) KubectlConfigOpt {
	return func(k *Kubectl) {
		k.networkFaultBackoffFactor = factor
	}
}

type capiMachinesResponse struct {
	Items []clusterv1.Machine
}

// GetCAPIMachines returns all the CAPI machines for the provided clusterName.
func (k *Kubectl) GetCAPIMachines(ctx context.Context, cluster *types.Cluster, clusterName string) ([]clusterv1.Machine, error) {
	params := []string{
		"get", capiMachinesType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
		"--selector=cluster.x-k8s.io/cluster-name=" + clusterName,
		"--namespace", constants.EksaSystemNamespace,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting machines: %v", err)
	}

	response := &capiMachinesResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get machines response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) SearchCloudStackMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackMachineConfig, error) {
	params := []string{
		"get", eksaCloudStackMachineResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa CloudStackMachineConfigResponse: %v", err)
	}

	response := &CloudStackMachineConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing CloudStackMachineConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) SearchCloudStackDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.CloudStackDatacenterConfig, error) {
	params := []string{
		"get", eksaCloudStackDatacenterResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa CloudStackDatacenterConfigResponse: %v", err)
	}

	response := &CloudStackDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing CloudStackDatacenterConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error) {
	response := &v1alpha1.CloudStackMachineConfig{}
	err := k.GetObject(ctx, eksaCloudStackMachineResourceType, cloudstackMachineConfigName, namespace, kubeconfigFile, response)
	if err != nil {
		return nil, fmt.Errorf("getting eksa cloudstack machineconfig: %v", err)
	}

	return response, nil
}

func (k *Kubectl) DeleteEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaCloudStackDatacenterResourceType, cloudstackDatacenterConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting cloudstackdatacenterconfig cluster %s apply: %v", cloudstackDatacenterConfigName, err)
	}
	return nil
}

func (k *Kubectl) GetEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDatacenterConfig, error) {
	response := &v1alpha1.CloudStackDatacenterConfig{}
	err := k.GetObject(ctx, eksaCloudStackDatacenterResourceType, cloudstackDatacenterConfigName, namespace, kubeconfigFile, response)
	if err != nil {
		return nil, fmt.Errorf("getting eksa cloudstack datacenterconfig: %v", err)
	}

	return response, nil
}

func (k *Kubectl) DeleteEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaCloudStackMachineResourceType, cloudstackMachineConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting cloudstackmachineconfig cluster %s apply: %v", cloudstackMachineConfigName, err)
	}
	return nil
}

type VersionResponse struct {
	ClientVersion version.Info `json:"clientVersion"`
	ServerVersion version.Info `json:"serverVersion"`
}

func (k *Kubectl) GetNamespace(ctx context.Context, kubeconfig string, namespace string) error {
	params := []string{"get", "namespace", namespace, "--kubeconfig", kubeconfig}
	_, err := k.Execute(ctx, params...)
	return err
}

func (k *Kubectl) CreateNamespace(ctx context.Context, kubeconfig string, namespace string) error {
	params := []string{"create", "namespace", namespace, "--kubeconfig", kubeconfig}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("creating namespace %v: %v", namespace, err)
	}
	return nil
}

func (k *Kubectl) CreateNamespaceIfNotPresent(ctx context.Context, kubeconfig string, namespace string) error {
	if err := k.GetNamespace(ctx, kubeconfig, namespace); err != nil {
		return k.CreateNamespace(ctx, kubeconfig, namespace)
	}
	return nil
}

func (k *Kubectl) DeleteNamespace(ctx context.Context, kubeconfig string, namespace string) error {
	params := []string{"delete", "namespace", namespace, "--kubeconfig", kubeconfig}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("creating namespace %v: %v", namespace, err)
	}
	return nil
}

func (k *Kubectl) LoadSecret(ctx context.Context, secretObject string, secretObjectType string, secretObjectName string, kubeConfFile string) error {
	params := []string{"create", "secret", "generic", secretObjectName, "--type", secretObjectType, "--from-literal", secretObject, "--kubeconfig", kubeConfFile, "--namespace", constants.EksaSystemNamespace}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("loading secret: %v", err)
	}
	return nil
}

// ApplyManifest uses client-side logic to create/update objects defined in a yaml manifest.
func (k *Kubectl) ApplyManifest(ctx context.Context, kubeconfigPath, manifestPath string) error {
	if _, err := k.Execute(ctx, "apply", "-f", manifestPath, "--kubeconfig", kubeconfigPath); err != nil {
		return fmt.Errorf("executing apply manifest: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyKubeSpecWithNamespace(ctx context.Context, cluster *types.Cluster, spec string, namespace string) error {
	params := []string{"apply", "-f", spec, "--namespace", namespace}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing apply: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	params := []string{"apply", "-f", "-"}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.ExecuteWithStdin(ctx, data, params...)
	if err != nil {
		return fmt.Errorf("executing apply: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error {
	if len(data) == 0 {
		logger.V(6).Info("Skipping applying empty kube spec from bytes")
		return nil
	}

	params := []string{"apply", "-f", "-", "--namespace", namespace}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.ExecuteWithStdin(ctx, data, params...)
	if err != nil {
		return fmt.Errorf("executing apply: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error {
	params := []string{"apply", "-f", "-", "--force"}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.ExecuteWithStdin(ctx, data, params...)
	if err != nil {
		return fmt.Errorf("executing apply --force: %v", err)
	}
	return nil
}

// DeleteManifest uses client-side logic to delete objects defined in a yaml manifest.
func (k *Kubectl) DeleteManifest(ctx context.Context, kubeconfigPath, manifestPath string, opts ...KubectlOpt) error {
	params := []string{
		"delete", "-f", manifestPath, "--kubeconfig", kubeconfigPath,
	}
	applyOpts(&params, opts...)
	if _, err := k.Execute(ctx, params...); err != nil {
		return fmt.Errorf("executing apply manifest: %v", err)
	}
	return nil
}

func (k *Kubectl) DeleteKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error {
	params := []string{"delete", "-f", "-"}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.ExecuteWithStdin(ctx, data, params...)
	if err != nil {
		return fmt.Errorf("executing apply: %v", err)
	}
	return nil
}

func (k *Kubectl) WaitForClusterReady(ctx context.Context, cluster *types.Cluster, timeout string, clusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "Ready", fmt.Sprintf("%s/%s", capiClustersResourceType, clusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForControlPlaneReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ControlPlaneReady", fmt.Sprintf("%s/%s", capiClustersResourceType, newClusterName), constants.EksaSystemNamespace)
}

// WaitForControlPlaneAvailable blocks until the first control plane is available.
func (k *Kubectl) WaitForControlPlaneAvailable(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ControlPlaneInitialized", fmt.Sprintf("%s/%s", capiClustersResourceType, newClusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForControlPlaneNotReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ControlPlaneReady=false", fmt.Sprintf("%s/%s", capiClustersResourceType, newClusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForManagedExternalEtcdReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ManagedEtcdReady", fmt.Sprintf("clusters.%s/%s", clusterv1.GroupVersion.Group, newClusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForManagedExternalEtcdNotReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ManagedEtcdReady=false", fmt.Sprintf("clusters.%s/%s", clusterv1.GroupVersion.Group, newClusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForMachineDeploymentReady(ctx context.Context, cluster *types.Cluster, timeout string, machineDeploymentName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "Ready=true", fmt.Sprintf("%s/%s", capiMachineDeploymentsType, machineDeploymentName), constants.EksaSystemNamespace)
}

// WaitForService blocks until an IP address is assigned.
//
// Until more generic status matching comes around (possibly in 1.23), poll
// the service, checking for an IP address. Would you like to know more?
// https://github.com/kubernetes/kubernetes/issues/83094
func (k *Kubectl) WaitForService(ctx context.Context, kubeconfig string, timeout string, target string, namespace string) error {
	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("parsing duration %q: %w", timeout, err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeoutDur)
	defer cancel()
	timedOut := timeoutCtx.Done()

	const pollInterval = time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	svc := &corev1.Service{}
	for {
		select {
		case <-timedOut:
			return timeoutCtx.Err()
		case <-ticker.C:
			err := k.GetObject(ctx, "service", target, namespace, kubeconfig, svc)
			if err != nil {
				logger.V(6).Info("failed to poll service", "target", target, "namespace", namespace, "error", err)
				continue
			}
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				if ingress.IP != "" {
					logger.V(5).Info("found a load balancer:", "IP", svc.Spec.ClusterIP)
					return nil
				}
			}
			if svc.Spec.ClusterIP != "" {
				logger.V(5).Info("found a ClusterIP:", "IP", svc.Spec.ClusterIP)
				return nil
			}
		}
	}
}

func (k *Kubectl) WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, condition, "deployments/"+target, namespace)
}

// WaitForResourceRolledout waits for a resource (deployment, daemonset, or statefulset) to be successfully rolled out before returning.
func (k *Kubectl) WaitForResourceRolledout(ctx context.Context, cluster *types.Cluster, timeout string, target string, namespace string, resource string) error {
	params := []string{"rollout", "status", resource, target, "--kubeconfig", cluster.KubeconfigFile, "--namespace", namespace, "--timeout", timeout}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("unable to finish %s roll out: %w", resource, err)
	}
	return nil
}

// WaitForPod waits for a pod resource to reach desired condition before returning.
func (k *Kubectl) WaitForPod(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, condition, "pod/"+target, namespace)
}

// WaitForRufioMachines blocks until all Rufio Machines have the desired condition.
func (k *Kubectl) WaitForRufioMachines(ctx context.Context, cluster *types.Cluster, timeout string, condition string, namespace string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, condition, rufioMachineResourceType, namespace, WithWaitAll())
}

// WaitForJobCompleted waits for a job resource to reach desired condition before returning.
func (k *Kubectl) WaitForJobCompleted(ctx context.Context, kubeconfig, timeout string, condition string, target string, namespace string) error {
	return k.Wait(ctx, kubeconfig, timeout, condition, "job/"+target, namespace)
}

// WaitForPackagesInstalled waits for a package resource to reach installed state before returning.
func (k *Kubectl) WaitForPackagesInstalled(ctx context.Context, cluster *types.Cluster, name string, timeout string, namespace string) error {
	return k.WaitJSONPathLoop(ctx, cluster.KubeconfigFile, timeout, "status.state", "installed", fmt.Sprintf("%s/%s", eksaPackagesType, name), namespace)
}

// WaitForPodCompleted waits for a pod to be terminated with a Completed state before returning.
func (k *Kubectl) WaitForPodCompleted(ctx context.Context, cluster *types.Cluster, name string, timeout string, namespace string) error {
	return k.WaitJSONPathLoop(ctx, cluster.KubeconfigFile, timeout, "status.containerStatuses[0].state.terminated.reason", "Completed", "pod/"+name, namespace)
}

func (k *Kubectl) Wait(ctx context.Context, kubeconfig string, timeout string, forCondition string, property string, namespace string, opts ...KubectlOpt) error {
	// On each retry kubectl wait timeout values will have to be adjusted to only wait for the remaining timeout duration.
	// Here we establish an absolute timeout time for this based on the caller-specified timeout.
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("unparsable timeout specified: %w", err)
	}
	if timeoutDuration < 0 {
		return fmt.Errorf("negative timeout specified: %w", err)
	}
	timeoutTime := time.Now().Add(timeoutDuration)

	retrier := retrier.New(timeoutDuration, retrier.WithRetryPolicy(k.kubectlWaitRetryPolicy))
	err = retrier.Retry(
		func() error {
			return k.wait(ctx, kubeconfig, timeoutTime, forCondition, property, namespace, opts...)
		},
	)
	if err != nil {
		return fmt.Errorf("executing wait: %w", err)
	}
	return nil
}

// WaitJSONPathLoop will wait for a given JSONPath to reach a required state similar to wait command for objects without conditions.
// This will be deprecated in favor of WaitJSONPath after version 1.23.
func (k *Kubectl) WaitJSONPathLoop(ctx context.Context, kubeconfig string, timeout string, jsonpath, forCondition string, property string, namespace string, opts ...KubectlOpt) error {
	// On each retry kubectl wait timeout values will have to be adjusted to only wait for the remaining timeout duration.
	//  Here we establish an absolute timeout time for this based on the caller-specified timeout.
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("unparsable timeout specified: %w", err)
	}
	if timeoutDuration < 0 {
		return fmt.Errorf("negative timeout specified: %w", err)
	}

	retrier := retrier.New(timeoutDuration, retrier.WithRetryPolicy(k.kubectlWaitRetryPolicy))
	err = retrier.Retry(
		func() error {
			return k.waitJSONPathLoop(ctx, kubeconfig, timeout, jsonpath, forCondition, property, namespace, opts...)
		},
	)
	if err != nil {
		return fmt.Errorf("executing wait: %w", err)
	}
	return nil
}

// WaitJSONPath will wait for a given JSONPath of a required state. Only compatible on K8s 1.23+.
func (k *Kubectl) WaitJSONPath(ctx context.Context, kubeconfig string, timeout string, jsonpath, forCondition string, property string, namespace string, opts ...KubectlOpt) error {
	// On each retry kubectl wait timeout values will have to be adjusted to only wait for the remaining timeout duration.
	//  Here we establish an absolute timeout time for this based on the caller-specified timeout.
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("unparsable timeout specified: %w", err)
	}
	if timeoutDuration < 0 {
		return fmt.Errorf("negative timeout specified: %w", err)
	}

	retrier := retrier.New(timeoutDuration, retrier.WithRetryPolicy(k.kubectlWaitRetryPolicy))
	err = retrier.Retry(
		func() error {
			return k.waitJSONPath(ctx, kubeconfig, timeout, jsonpath, forCondition, property, namespace, opts...)
		},
	)
	if err != nil {
		return fmt.Errorf("executing wait: %w", err)
	}
	return nil
}

func (k *Kubectl) kubectlWaitRetryPolicy(totalRetries int, err error) (retry bool, wait time.Duration) {
	// Exponential backoff on network errors.  Retrier built-in backoff is linear, so implementing here.

	// Retrier first calls the policy before retry #1.  We want it zero-based for exponentiation.
	if totalRetries < 1 {
		totalRetries = 1
	}

	waitTime := time.Duration(float64(k.networkFaultBaseRetryTime) * math.Pow(k.networkFaultBackoffFactor, float64(totalRetries-1)))

	if match := kubectlConnectionRefusedRegex.MatchString(err.Error()); match {
		return true, waitTime
	}
	if match := kubectlConnectionTimeoutRegex.MatchString(err.Error()); match {
		return true, waitTime
	}
	return false, 0
}

func (k *Kubectl) wait(ctx context.Context, kubeconfig string, timeoutTime time.Time, forCondition string, property string, namespace string, opts ...KubectlOpt) error {
	secondsRemainingUntilTimeout := time.Until(timeoutTime).Seconds()
	if secondsRemainingUntilTimeout <= minimumWaitTimeout {
		return fmt.Errorf("error: timed out waiting for condition %v on %v", forCondition, property)
	}
	kubectlTimeoutString := fmt.Sprintf("%.*fs", timeoutPrecision, secondsRemainingUntilTimeout)
	params := []string{
		"wait", "--timeout", kubectlTimeoutString,
		"--for=condition=" + forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace,
	}
	applyOpts(&params, opts...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing wait: %w", err)
	}
	return nil
}

func (k *Kubectl) waitJSONPath(ctx context.Context, kubeconfig, timeout string, jsonpath string, forCondition string, property string, namespace string, opts ...KubectlOpt) error {
	if jsonpath == "" || forCondition == "" {
		return fmt.Errorf("empty conditions params passed to waitJSONPath()")
	}
	params := []string{
		"wait", "--timeout", timeout, fmt.Sprintf("--for=jsonpath='{.%s}'=%s", jsonpath, forCondition), property, "--kubeconfig", kubeconfig, "-n", namespace,
	}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("executing wait: %w", err)
	}
	return nil
}

// waitJsonPathLoop will be deprecated in favor of waitJsonPath after version 1.23.
func (k *Kubectl) waitJSONPathLoop(ctx context.Context, kubeconfig string, timeout string, jsonpath string, forCondition string, property string, namespace string, opts ...KubectlOpt) error {
	if jsonpath == "" || forCondition == "" {
		return fmt.Errorf("empty conditions params passed to waitJSONPathLoop()")
	}
	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("parsing duration %q: %w", timeout, err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeoutDur)
	defer cancel()
	timedOut := timeoutCtx.Done()

	const pollInterval = time.Second * 5
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timedOut:
			return fmt.Errorf("waiting for %s %s on %s: timed out", jsonpath, forCondition, property)
		case <-ticker.C:
			params := []string{
				"get", property,
				"-o", fmt.Sprintf("jsonpath='{.%s}'", jsonpath),
				"--kubeconfig", kubeconfig,
				"-n", namespace,
			}
			stdout, err := k.Execute(ctx, params...)
			if err != nil {
				return fmt.Errorf("waiting for %s %s on %s: %w", jsonpath, forCondition, property, err)
			}
			if strings.Contains(stdout.String(), forCondition) {
				return nil
			}
			fmt.Printf("waiting 5 seconds.... current state=%v, desired state=%v\n", stdout.String(), fmt.Sprintf("'%s'", forCondition))
		}
	}
}

func (k *Kubectl) DeleteEksaDatacenterConfig(ctx context.Context, eksaDatacenterResourceType string, eksaDatacenterConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaDatacenterResourceType, eksaDatacenterConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting %s cluster %s apply: %v", eksaDatacenterResourceType, eksaDatacenterConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteEksaMachineConfig(ctx context.Context, eksaMachineConfigResourceType string, eksaMachineConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaMachineConfigResourceType, eksaMachineConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting %s cluster %s apply: %v", eksaMachineConfigResourceType, eksaMachineConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteEKSACluster(ctx context.Context, managementCluster *types.Cluster, eksaClusterName, eksaClusterNamespace string) error {
	params := []string{"delete", eksaClusterResourceType, eksaClusterName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", eksaClusterNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting eksa cluster %s apply: %v", eksaClusterName, err)
	}
	return nil
}

func (k *Kubectl) DeleteGitOpsConfig(ctx context.Context, managementCluster *types.Cluster, gitOpsConfigName, gitOpsConfigNamespace string) error {
	params := []string{"delete", eksaGitOpsResourceType, gitOpsConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", gitOpsConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting gitops config %s apply: %v", gitOpsConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteFluxConfig(ctx context.Context, managementCluster *types.Cluster, fluxConfigName, fluxConfigNamespace string) error {
	params := []string{"delete", eksaFluxConfigResourceType, fluxConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", fluxConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting gitops config %s apply: %v", fluxConfigName, err)
	}
	return nil
}

// GetPackageBundleController will retrieve the packagebundlecontroller from eksa-packages namespace and return the object.
func (k *Kubectl) GetPackageBundleController(ctx context.Context, kubeconfigFile, clusterName string) (packagesv1.PackageBundleController, error) {
	params := []string{"get", eksaPackagesBundleControllerType, clusterName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", constants.EksaPackagesName, "--ignore-not-found=true"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return packagesv1.PackageBundleController{}, fmt.Errorf("failed to execute cmd \"%s\": %w", strings.Join(params, " "), err)
	}
	pbc := &packagesv1.PackageBundleController{}
	err = json.Unmarshal(stdOut.Bytes(), pbc)
	if err != nil {
		return packagesv1.PackageBundleController{}, fmt.Errorf("unmarshalling kubectl response to GO struct %s: %v, response: %s", clusterName, err, stdOut.String())
	}
	return *pbc, nil
}

// GetPackageBundleList will retrieve the packagebundle list from eksa-packages namespace and return the list.
func (k *Kubectl) GetPackageBundleList(ctx context.Context, kubeconfigFile string) ([]packagesv1.PackageBundle, error) {
	err := k.WaitJSONPathLoop(ctx, kubeconfigFile, "5m", "items", "PackageBundle", eksaPackageBundlesType, constants.EksaPackagesName)
	if err != nil {
		return nil, fmt.Errorf("waiting on package bundle resource to exist %v", err)
	}
	params := []string{"get", eksaPackageBundlesType, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", constants.EksaPackagesName, "--ignore-not-found=true"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting package bundle resource %v", err)
	}
	response := &packagesv1.PackageBundleList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling kubectl response to GO struct %v", err)
	}
	return response.Items, nil
}

func (k *Kubectl) DeletePackageResources(ctx context.Context, managementCluster *types.Cluster, clusterName string) error {
	params := []string{"delete", eksaPackagesBundleControllerType, clusterName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaPackagesName, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting package resources for %s: %v", clusterName, err)
	}
	params = []string{"delete", "namespace", "eksa-packages-" + clusterName, "--kubeconfig", managementCluster.KubeconfigFile, "--ignore-not-found=true"}
	_, err = k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting package resources for %s: %v", clusterName, err)
	}
	return nil
}

func (k *Kubectl) DeleteSecret(ctx context.Context, managementCluster *types.Cluster, secretName, namespace string) error {
	params := []string{"delete", "secret", secretName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", namespace}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting secret %s in namespace %s: %v", secretName, namespace, err)
	}
	return nil
}

func (k *Kubectl) DeleteOIDCConfig(ctx context.Context, managementCluster *types.Cluster, oidcConfigName, oidcConfigNamespace string) error {
	params := []string{"delete", eksaOIDCResourceType, oidcConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", oidcConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting oidc config %s apply: %v", oidcConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteAWSIamConfig(ctx context.Context, managementCluster *types.Cluster, awsIamConfigName, awsIamConfigNamespace string) error {
	params := []string{"delete", eksaAwsIamResourceType, awsIamConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", awsIamConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting awsIam config %s apply: %v", awsIamConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster) error {
	params := []string{"delete", capiClustersResourceType, clusterToDelete.Name, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting cluster %s apply: %v", clusterToDelete.Name, err)
	}
	return nil
}

func (k *Kubectl) ListCluster(ctx context.Context) error {
	params := []string{"get", "pods", "-A", "-o", "jsonpath={..image}"}
	output, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("listing cluster versions: %v", err)
	}

	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strings.Fields(output.String()) {
		if _, found := keys[entry]; !found {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	sort.Strings(list)
	for _, value := range list {
		logger.Info(value)
	}
	return nil
}

func (k *Kubectl) GetNodes(ctx context.Context, kubeconfig string) ([]corev1.Node, error) {
	params := []string{"get", "nodes", "-o", "json", "--kubeconfig", kubeconfig}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting nodes: %v", err)
	}
	response := &corev1.NodeList{}
	err = json.Unmarshal(stdOut.Bytes(), response)

	return response.Items, err
}

func (k *Kubectl) GetControlPlaneNodes(ctx context.Context, kubeconfig string) ([]corev1.Node, error) {
	params := []string{"get", "nodes", "-o", "json", "--kubeconfig", kubeconfig, "--selector=node-role.kubernetes.io/control-plane"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting control plane nodes: %v", err)
	}
	response := &corev1.NodeList{}
	err = json.Unmarshal(stdOut.Bytes(), response)

	return response.Items, err
}

// GetVsphereMachine will return list of vSphere machines.
func (k *Kubectl) GetVsphereMachine(ctx context.Context, kubeconfig string, selector string) ([]vspherev1.VSphereMachine, error) {
	params := []string{"get", "vspheremachines", "-o", "json", "--namespace", constants.EksaSystemNamespace, "--kubeconfig", kubeconfig, "--selector=" + selector}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting VSphere machine: %v", err)
	}
	response := &vspherev1.VSphereMachineList{}
	err = json.Unmarshal(stdOut.Bytes(), response)

	return response.Items, err
}

func (k *Kubectl) ValidateNodes(ctx context.Context, kubeconfig string) error {
	template := "{{range .items}}{{.metadata.name}}\n{{end}}"
	params := []string{"get", "nodes", "-o", "go-template", "--template", template, "--kubeconfig", kubeconfig}
	buffer, err := k.Execute(ctx, params...)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(buffer.String()))
	for scanner.Scan() {
		node := scanner.Text()
		if len(node) != 0 {
			template = "{{range .status.conditions}}{{if eq .type \"Ready\"}}{{.reason}}{{end}}{{end}}"
			params = []string{"get", "node", node, "-o", "go-template", "--template", template, "--kubeconfig", kubeconfig}
			buffer, err = k.Execute(ctx, params...)
			if err != nil {
				return err
			}
			if buffer.String() != "KubeletReady" {
				return fmt.Errorf("node %s is not ready, currently in %s state", node, buffer.String())
			}
		}
	}
	return nil
}

func (k *Kubectl) DeleteOldWorkerNodeGroup(ctx context.Context, md *clusterv1.MachineDeployment, kubeconfig string) error {
	kubeadmConfigTemplateName := md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
	providerMachineTemplateName := md.Spec.Template.Spec.InfrastructureRef.Name
	params := []string{"delete", md.Kind, md.Name, "--kubeconfig", kubeconfig, "--namespace", constants.EksaSystemNamespace}
	if _, err := k.Execute(ctx, params...); err != nil {
		return err
	}
	params = []string{"delete", md.Spec.Template.Spec.Bootstrap.ConfigRef.Kind, kubeadmConfigTemplateName, "--kubeconfig", kubeconfig, "--namespace", constants.EksaSystemNamespace}
	if _, err := k.Execute(ctx, params...); err != nil {
		return err
	}
	params = []string{"delete", md.Spec.Template.Spec.InfrastructureRef.Kind, providerMachineTemplateName, "--kubeconfig", kubeconfig, "--namespace", constants.EksaSystemNamespace}
	if _, err := k.Execute(ctx, params...); err != nil {
		return err
	}
	return nil
}

func (k *Kubectl) ValidateControlPlaneNodes(ctx context.Context, cluster *types.Cluster, clusterName string) error {
	cp, err := k.GetKubeadmControlPlane(ctx, cluster, clusterName, WithCluster(cluster), WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return err
	}

	observedGeneration := cp.Status.ObservedGeneration
	generation := cp.Generation
	if observedGeneration != generation {
		return fmt.Errorf("kubeadm control plane %s status needs to be refreshed: observed generation is %d, want %d", cp.Name, observedGeneration, generation)
	}

	if !cp.Status.Ready {
		return errors.New("control plane is not ready")
	}

	if cp.Status.UnavailableReplicas != 0 {
		return fmt.Errorf("%v control plane replicas are unavailable", cp.Status.UnavailableReplicas)
	}

	if cp.Status.ReadyReplicas != cp.Status.Replicas {
		return fmt.Errorf("%v control plane replicas are not ready", cp.Status.Replicas-cp.Status.ReadyReplicas)
	}
	return nil
}

func (k *Kubectl) ValidateWorkerNodes(ctx context.Context, clusterName string, kubeconfig string) error {
	logger.V(6).Info("waiting for nodes", "cluster", clusterName)
	ready, total, err := k.CountMachineDeploymentReplicasReady(ctx, clusterName, kubeconfig)
	if err != nil {
		return err
	}
	if ready != total {
		return fmt.Errorf("%d machine deployment replicas are not ready", total-ready)
	}
	return nil
}

func (k *Kubectl) CountMachineDeploymentReplicasReady(ctx context.Context, clusterName string, kubeconfig string) (ready, total int, err error) {
	logger.V(6).Info("counting ready machine deployment replicas", "cluster", clusterName)
	deployments, err := k.GetMachineDeploymentsForCluster(ctx, clusterName, WithKubeconfig(kubeconfig), WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return 0, 0, err
	}
	for _, machineDeployment := range deployments {
		if machineDeployment.Status.Phase != "Running" {
			return 0, 0, fmt.Errorf("machine deployment is in %s phase", machineDeployment.Status.Phase)
		}

		if machineDeployment.Status.UnavailableReplicas != 0 {
			return 0, 0, fmt.Errorf("%d machine deployment replicas are unavailable", machineDeployment.Status.UnavailableReplicas)
		}

		ready += int(machineDeployment.Status.ReadyReplicas)
		total += int(machineDeployment.Status.Replicas)
	}
	return ready, total, nil
}

func (k *Kubectl) VsphereWorkerNodesMachineTemplate(ctx context.Context, clusterName string, kubeconfig string, namespace string) (*vspherev1.VSphereMachineTemplate, error) {
	machineTemplateName, err := k.MachineTemplateName(ctx, clusterName, kubeconfig, WithNamespace(namespace))
	if err != nil {
		return nil, err
	}

	params := []string{"get", vsphereMachineTemplatesType, machineTemplateName, "-o", "go-template", "--template", "{{.spec.template.spec}}", "-o", "yaml", "--kubeconfig", kubeconfig, "--namespace", namespace}
	buffer, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, err
	}
	machineTemplateSpec := &vspherev1.VSphereMachineTemplate{}
	if err := yaml.Unmarshal(buffer.Bytes(), machineTemplateSpec); err != nil {
		return nil, err
	}
	return machineTemplateSpec, nil
}

func (k *Kubectl) CloudstackWorkerNodesMachineTemplate(ctx context.Context, clusterName string, kubeconfig string, namespace string) (*cloudstackv1.CloudStackMachineTemplate, error) {
	machineTemplateName, err := k.MachineTemplateName(ctx, clusterName, kubeconfig, WithNamespace(namespace))
	if err != nil {
		return nil, err
	}

	params := []string{"get", cloudstackMachineTemplatesType, machineTemplateName, "-o", "go-template", "--template", "{{.spec.template.spec}}", "-o", "yaml", "--kubeconfig", kubeconfig, "--namespace", namespace}
	buffer, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, err
	}
	machineTemplateSpec := &cloudstackv1.CloudStackMachineTemplate{}
	if err := yaml.Unmarshal(buffer.Bytes(), machineTemplateSpec); err != nil {
		return nil, err
	}
	return machineTemplateSpec, nil
}

func (k *Kubectl) MachineTemplateName(ctx context.Context, clusterName string, kubeconfig string, opts ...KubectlOpt) (string, error) {
	template := "{{.spec.template.spec.infrastructureRef.name}}"
	params := []string{"get", capiMachineDeploymentsType, fmt.Sprintf("%s-md-0", clusterName), "-o", "go-template", "--template", template, "--kubeconfig", kubeconfig}
	applyOpts(&params, opts...)
	buffer, err := k.Execute(ctx, params...)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (k *Kubectl) ValidatePods(ctx context.Context, kubeconfig string) error {
	template := "{{range .items}}{{.metadata.name}},{{.status.phase}}\n{{end}}"
	params := []string{"get", "pods", "-A", "-o", "go-template", "--template", template, "--kubeconfig", kubeconfig}
	buffer, err := k.Execute(ctx, params...)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(buffer.String()))
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), ",")
		if len(data) == 2 {
			if data[1] != "Running" {
				return fmt.Errorf("pod %s is not running, currently in %s phase", data[0], data[1])
			}
		}
	}
	logger.Info("All pods are running")
	return nil
}

// RunCurlPod will run Kubectl with an image (with curl installed) and the command you pass in.
func (k *Kubectl) RunCurlPod(ctx context.Context, namespace, name, kubeconfig string, command []string) (string, error) {
	params := []string{"run", name, "--image=public.ecr.aws/eks-anywhere/diagnostic-collector:v0.16.2-eks-a-41", "-o", "json", "--kubeconfig", kubeconfig, "--namespace", namespace, "--restart=Never", "--"}
	params = append(params, command...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return "", err
	}
	return name, err
}

// GetPodNameByLabel will return the name of the first pod that matches the label.
func (k *Kubectl) GetPodNameByLabel(ctx context.Context, namespace, label, kubeconfig string) (string, error) {
	params := []string{"get", "pod", "-l=" + label, "-o=jsonpath='{.items[0].metadata.name}'", "--kubeconfig", kubeconfig, "--namespace", namespace}
	podName, err := k.Execute(ctx, params...)
	if err != nil {
		return "", err
	}
	return strings.Trim(podName.String(), `'"`), err
}

// GetPodIP will return the ip of the pod.
func (k *Kubectl) GetPodIP(ctx context.Context, namespace, podName, kubeconfig string) (string, error) {
	params := []string{"get", "pod", podName, "-o=jsonpath='{.status.podIP}'", "--kubeconfig", kubeconfig, "--namespace", namespace}
	ip, err := k.Execute(ctx, params...)
	if err != nil {
		return "", err
	}
	return strings.Trim(ip.String(), `'"`), err
}

// GetPodLogs returns the logs of the specified container (namespace/pod/container).
func (k *Kubectl) GetPodLogs(ctx context.Context, namespace, podName, containerName, kubeconfig string) (string, error) {
	return k.getPodLogs(ctx, namespace, podName, containerName, kubeconfig, nil, nil)
}

// GetPodLogsSince returns the logs of the specified container (namespace/pod/container) since a timestamp.
func (k *Kubectl) GetPodLogsSince(ctx context.Context, namespace, podName, containerName, kubeconfig string, since time.Time) (string, error) {
	sinceTime := metav1.NewTime(since)
	return k.getPodLogs(ctx, namespace, podName, containerName, kubeconfig, &sinceTime, nil)
}

func (k *Kubectl) getPodLogs(ctx context.Context, namespace, podName, containerName, kubeconfig string, sinceTime *metav1.Time, tailLines *int) (string, error) {
	params := []string{"logs", podName, containerName, "--kubeconfig", kubeconfig, "--namespace", namespace}
	if sinceTime != nil {
		params = append(params, "--since-time", sinceTime.Format(time.RFC3339))
	}
	if tailLines != nil {
		params = append(params, "--tail", strconv.Itoa(*tailLines))
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return "", err
	}
	logs := stdOut.String()
	if strings.Contains(logs, "Internal Error") {
		return "", fmt.Errorf("Fetched log contains \"Internal Error\": %q", logs)
	}
	return logs, err
}

func (k *Kubectl) SaveLog(ctx context.Context, cluster *types.Cluster, deployment *types.Deployment, fileName string, writer filewriter.FileWriter) error {
	params := []string{"--kubeconfig", cluster.KubeconfigFile}
	logParams := []string{
		"logs",
		fmt.Sprintf("deployment/%s", deployment.Name),
		"-n",
		deployment.Namespace,
	}
	if deployment.Container != "" {
		logParams = append(logParams, "-c", deployment.Container)
	}

	params = append(params, logParams...)

	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("saving logs: %v", err)
	}

	_, err = writer.Write(fileName, stdOut.Bytes())
	if err != nil {
		return err
	}
	return nil
}

type machinesResponse struct {
	Items []types.Machine `json:"items,omitempty"`
}

func (k *Kubectl) GetMachines(ctx context.Context, cluster *types.Cluster, clusterName string) ([]types.Machine, error) {
	params := []string{
		"get", capiMachinesType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
		"--selector=cluster.x-k8s.io/cluster-name=" + clusterName,
		"--namespace", constants.EksaSystemNamespace,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting machines: %v", err)
	}

	response := &machinesResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get machines response: %v", err)
	}

	return response.Items, nil
}

type machineSetResponse struct {
	Items []clusterv1.MachineSet `json:"items,omitempty"`
}

func (k *Kubectl) GetMachineSets(ctx context.Context, machineDeploymentName string, cluster *types.Cluster) ([]clusterv1.MachineSet, error) {
	params := []string{
		"get", capiMachineSetsType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
		"--selector=cluster.x-k8s.io/deployment-name=" + machineDeploymentName,
		"--namespace", constants.EksaSystemNamespace,
	}

	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting machineset associated with deployment %s: %v", machineDeploymentName, err)
	}

	response := &machineSetResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get machinesets response: %v", err)
	}

	return response.Items, nil
}

type ClustersResponse struct {
	Items []types.CAPICluster `json:"items,omitempty"`
}

type GitOpsConfigResponse struct {
	Items []*v1alpha1.GitOpsConfig `json:"items,omitempty"`
}

type VSphereDatacenterConfigResponse struct {
	Items []*v1alpha1.VSphereDatacenterConfig `json:"items,omitempty"`
}

type CloudStackDatacenterConfigResponse struct {
	Items []*v1alpha1.CloudStackDatacenterConfig `json:"items,omitempty"`
}

// TinkerbellDatacenterConfigResponse contains list of TinkerbellDatacenterConfig.
type TinkerbellDatacenterConfigResponse struct {
	Items []*v1alpha1.TinkerbellDatacenterConfig `json:"items,omitempty"`
}

type NutanixDatacenterConfigResponse struct {
	Items []*v1alpha1.NutanixDatacenterConfig `json:"items,omitempty"`
}

type IdentityProviderConfigResponse struct {
	Items []*v1alpha1.Ref `json:"items,omitempty"`
}

type VSphereMachineConfigResponse struct {
	Items []*v1alpha1.VSphereMachineConfig `json:"items,omitempty"`
}

type CloudStackMachineConfigResponse struct {
	Items []*v1alpha1.CloudStackMachineConfig `json:"items,omitempty"`
}

// TinkerbellMachineConfigResponse contains list of TinkerbellMachineConfig.
type TinkerbellMachineConfigResponse struct {
	Items []*v1alpha1.TinkerbellMachineConfig `json:"items,omitempty"`
}

type NutanixMachineConfigResponse struct {
	Items []*v1alpha1.NutanixMachineConfig `json:"items,omitempty"`
}

func (k *Kubectl) ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error {
	params := []string{"get", "customresourcedefinition", capiClustersResourceType, "--kubeconfig", cluster.KubeconfigFile}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("getting clusters crd: %v", err)
	}
	return nil
}

func (k *Kubectl) ValidateEKSAClustersCRD(ctx context.Context, cluster *types.Cluster) error {
	params := []string{"get", "customresourcedefinition", eksaClusterResourceType, "--kubeconfig", cluster.KubeconfigFile}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("getting eksa clusters crd: %v", err)
	}
	return nil
}

func (k *Kubectl) RolloutRestartDaemonSet(ctx context.Context, dsName, dsNamespace, kubeconfig string) error {
	params := []string{
		"rollout", "restart", "daemonset", dsName,
		"--kubeconfig", kubeconfig, "--namespace", dsNamespace,
	}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("restarting %s daemonset in namespace %s: %v", dsName, dsNamespace, err)
	}
	return nil
}

func (k *Kubectl) SetEksaControllerEnvVar(ctx context.Context, envVar, envVarVal, kubeconfig string) error {
	params := []string{
		"set", "env", "deployment/eksa-controller-manager", fmt.Sprintf("%s=%s", envVar, envVarVal),
		"--kubeconfig", kubeconfig, "--namespace", constants.EksaSystemNamespace,
	}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("setting %s=%s on eksa controller: %v", envVar, envVarVal, err)
	}
	return nil
}

func (k *Kubectl) GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error) {
	params := []string{"get", capiClustersResourceType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting clusters: %v", err)
	}

	response := &ClustersResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get clusters response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetApiServerUrl(ctx context.Context, cluster *types.Cluster) (string, error) {
	params := []string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.clusters[0].cluster.server}"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("getting api server url: %v", err)
	}

	return stdOut.String(), nil
}

func (k *Kubectl) Version(ctx context.Context, cluster *types.Cluster) (*VersionResponse, error) {
	params := []string{"version", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("executing kubectl version: %v", err)
	}
	response := &VersionResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling kubectl version response: %v", err)
	}
	return response, nil
}

type KubectlOpt func(*[]string)

// WithToken is a kubectl option to pass a token when making a kubectl call.
func WithToken(t string) KubectlOpt {
	return appendOpt("--token", t)
}

func WithServer(s string) KubectlOpt {
	return appendOpt("--server", s)
}

func WithCluster(c *types.Cluster) KubectlOpt {
	return WithKubeconfig(c.KubeconfigFile)
}

func WithKubeconfig(kubeconfigFile string) KubectlOpt {
	return appendOpt("--kubeconfig", kubeconfigFile)
}

func WithNamespace(n string) KubectlOpt {
	return appendOpt("--namespace", n)
}

// WithResourceName is a kubectl option to pass a resource name when making a kubectl call.
func WithResourceName(name string) KubectlOpt {
	return appendOpt(name)
}

// WithAllNamespaces is a kubectl option to add all namespaces when making a kubectl call.
func WithAllNamespaces() KubectlOpt {
	return appendOpt("-A")
}

func WithSkipTLSVerify() KubectlOpt {
	return appendOpt("--insecure-skip-tls-verify=true")
}

func WithOverwrite() KubectlOpt {
	return appendOpt("--overwrite")
}

func WithWaitAll() KubectlOpt {
	return appendOpt("--all")
}

// WithSelector is a kubectl option to pass a selector when making kubectl calls.
func WithSelector(selector string) KubectlOpt {
	return appendOpt("--selector=" + selector)
}

func appendOpt(new ...string) KubectlOpt {
	return func(args *[]string) {
		*args = append(*args, new...)
	}
}

func applyOpts(params *[]string, opts ...KubectlOpt) {
	for _, opt := range opts {
		opt(params)
	}
}

func (k *Kubectl) GetPods(ctx context.Context, opts ...KubectlOpt) ([]corev1.Pod, error) {
	params := []string{"get", "pods", "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting pods: %v", err)
	}

	response := &corev1.PodList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get pods response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetDeployments(ctx context.Context, opts ...KubectlOpt) ([]appsv1.Deployment, error) {
	params := []string{"get", "deployments", "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting deployments: %v", err)
	}

	response := &appsv1.DeploymentList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get deployments response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.Secret, error) {
	obj := &corev1.Secret{}
	if err := k.GetObject(ctx, "secret", name, namespace, kubeconfigFile, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) GetSecret(ctx context.Context, secretObjectName string, opts ...KubectlOpt) (*corev1.Secret, error) {
	params := []string{"get", "secret", secretObjectName, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting secret: %v", err)
	}
	response := &corev1.Secret{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get secret response: %v", err)
	}
	return response, err
}

func (k *Kubectl) GetKubeadmControlPlanes(ctx context.Context, opts ...KubectlOpt) ([]controlplanev1.KubeadmControlPlane, error) {
	params := []string{"get", kubeadmControlPlaneResourceType, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting kubeadmcontrolplanes: %v", err)
	}

	response := &controlplanev1.KubeadmControlPlaneList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get kubeadmcontrolplanes response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...KubectlOpt) (*controlplanev1.KubeadmControlPlane, error) {
	logger.V(6).Info("Getting KubeadmControlPlane CRDs", "cluster", clusterName)
	params := []string{"get", kubeadmControlPlaneResourceType, clusterName, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting kubeadmcontrolplane: %v", err)
	}

	response := &controlplanev1.KubeadmControlPlane{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get kubeadmcontrolplane response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...KubectlOpt) (*clusterv1.MachineDeployment, error) {
	params := []string{"get", capiMachineDeploymentsType, workerNodeGroupName, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting machine deployment: %v", err)
	}

	response := &clusterv1.MachineDeployment{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get machineDeployment response: %v", err)
	}

	return response, nil
}

// GetMachineDeployments retrieves all Machine Deployments.
func (k *Kubectl) GetMachineDeployments(ctx context.Context, opts ...KubectlOpt) ([]clusterv1.MachineDeployment, error) {
	params := []string{"get", capiMachineDeploymentsType, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting machine deployments: %v", err)
	}

	response := &clusterv1.MachineDeploymentList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get machineDeployments response: %v", err)
	}

	return response.Items, nil
}

// GetMachineDeploymentsForCluster retrieves all the Machine Deployments for a cluster with name "clusterName".
func (k *Kubectl) GetMachineDeploymentsForCluster(ctx context.Context, clusterName string, opts ...KubectlOpt) ([]clusterv1.MachineDeployment, error) {
	return k.GetMachineDeployments(ctx, append(opts, WithSelector(fmt.Sprintf("cluster.x-k8s.io/cluster-name=%s", clusterName)))...)
}

func (k *Kubectl) UpdateEnvironmentVariables(ctx context.Context, resourceType, resourceName string, envMap map[string]string, opts ...KubectlOpt) error {
	params := []string{"set", "env", resourceType, resourceName}
	for k, v := range envMap {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}
	applyOpts(&params, opts...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("setting the environment variables in %s %s: %v", resourceType, resourceName, err)
	}
	return nil
}

func (k *Kubectl) UpdateEnvironmentVariablesInNamespace(ctx context.Context, resourceType, resourceName string, envMap map[string]string, cluster *types.Cluster, namespace string) error {
	return k.UpdateEnvironmentVariables(ctx, resourceType, resourceName, envMap, WithCluster(cluster), WithNamespace(namespace))
}

func (k *Kubectl) UpdateAnnotation(ctx context.Context, resourceType, objectName string, annotations map[string]string, opts ...KubectlOpt) error {
	params := []string{"annotate", resourceType, objectName}
	for k, v := range annotations {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}
	applyOpts(&params, opts...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("updating annotation: %v", err)
	}
	return nil
}

func (k *Kubectl) UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error {
	return k.UpdateAnnotation(ctx, resourceType, objectName, annotations, WithOverwrite(), WithCluster(cluster), WithNamespace(namespace))
}

func (k *Kubectl) RemoveAnnotation(ctx context.Context, resourceType, objectName string, key string, opts ...KubectlOpt) error {
	params := []string{"annotate", resourceType, objectName, fmt.Sprintf("%s-", key)}
	applyOpts(&params, opts...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("removing annotation: %v", err)
	}
	return nil
}

func (k *Kubectl) RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error {
	return k.RemoveAnnotation(ctx, resourceType, objectName, key, WithCluster(cluster), WithNamespace(namespace))
}

func (k *Kubectl) GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error) {
	params := []string{"get", eksaClusterResourceType, "-A", "-o", "jsonpath={.items[0]}", "--kubeconfig", cluster.KubeconfigFile, "--field-selector=metadata.name=" + clusterName}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		params := []string{"get", eksaClusterResourceType, "-A", "--kubeconfig", cluster.KubeconfigFile, "--field-selector=metadata.name=" + clusterName}
		stdOut, err = k.Execute(ctx, params...)
		if err != nil {
			return nil, fmt.Errorf("getting eksa cluster: %v", err)
		}
		return nil, fmt.Errorf("cluster %s not found of custom resource type %s", clusterName, eksaClusterResourceType)
	}

	response := &v1alpha1.Cluster{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get eksa cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) SearchVsphereMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereMachineConfig, error) {
	params := []string{
		"get", eksaVSphereMachineResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa VSphereMachineConfigResponse: %v", err)
	}

	response := &VSphereMachineConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing VSphereMachineConfigResponse response: %v", err)
	}

	return response.Items, nil
}

// SearchTinkerbellMachineConfig returns the list of TinkerbellMachineConfig in the cluster.
func (k *Kubectl) SearchTinkerbellMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.TinkerbellMachineConfig, error) {
	params := []string{
		"get", eksaTinkerbellMachineResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa TinkerbellMachineConfigResponse: %v", err)
	}

	response := &TinkerbellMachineConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing TinkerbellMachineConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) SearchIdentityProviderConfig(ctx context.Context, ipName string, kind string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereDatacenterConfig, error) {
	var internalType string

	switch kind {
	case v1alpha1.OIDCConfigKind:
		internalType = fmt.Sprintf("oidcconfigs.%s", v1alpha1.GroupVersion.Group)
	case v1alpha1.AWSIamConfigKind:
		internalType = fmt.Sprintf("awsiamconfigs.%s", v1alpha1.GroupVersion.Group)
	default:
		return nil, fmt.Errorf("invalid identity provider %s", kind)
	}

	params := []string{
		"get", internalType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + ipName,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa IdentityProvider: %v", err)
	}

	response := &VSphereDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing IdentityProviderConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) SearchVsphereDatacenterConfig(ctx context.Context, datacenterName string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereDatacenterConfig, error) {
	params := []string{
		"get", eksaVSphereDatacenterResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + datacenterName,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa VSphereDatacenterConfigResponse: %v", err)
	}

	response := &VSphereDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing VSphereDatacenterConfigResponse response: %v", err)
	}

	return response.Items, nil
}

// SearchTinkerbellDatacenterConfig returns the list of TinkerbellDatacenterConfig in the cluster.
func (k *Kubectl) SearchTinkerbellDatacenterConfig(ctx context.Context, datacenterName string, kubeconfigFile string, namespace string) ([]*v1alpha1.TinkerbellDatacenterConfig, error) {
	params := []string{
		"get", eksaTinkerbellDatacenterResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + datacenterName,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa TinkerbellDatacenterConfigResponse: %v", err)
	}

	response := &TinkerbellDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing TinkerbellDatacenterConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetEksaFluxConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.FluxConfig, error) {
	params := []string{"get", eksaFluxConfigResourceType, gitOpsConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa FluxConfig: %v", err)
	}

	response := &v1alpha1.FluxConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing FluxConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error) {
	params := []string{"get", eksaGitOpsResourceType, gitOpsConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa GitOpsConfig: %v", err)
	}

	response := &v1alpha1.GitOpsConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing GitOpsConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error) {
	params := []string{"get", eksaOIDCResourceType, oidcConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa OIDCConfig: %v", err)
	}

	response := &v1alpha1.OIDCConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing OIDCConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error) {
	params := []string{"get", eksaAwsIamResourceType, awsIamConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa AWSIamConfig: %v", err)
	}

	response := &v1alpha1.AWSIamConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing AWSIamConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaTinkerbellDatacenterConfig(ctx context.Context, tinkerbellDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellDatacenterConfig, error) {
	params := []string{"get", eksaTinkerbellDatacenterResourceType, tinkerbellDatacenterConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa TinkerbellDatacenterConfig %v", err)
	}

	response := &v1alpha1.TinkerbellDatacenterConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get eksa TinkerbellDatacenterConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error) {
	params := []string{"get", eksaVSphereDatacenterResourceType, vsphereDatacenterConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa vsphere cluster %v", err)
	}

	response := &v1alpha1.VSphereDatacenterConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get eksa vsphere cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaTinkerbellMachineConfig(ctx context.Context, tinkerbellMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellMachineConfig, error) {
	params := []string{"get", eksaTinkerbellMachineResourceType, tinkerbellMachineConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa TinkerbellMachineConfig %v", err)
	}

	response := &v1alpha1.TinkerbellMachineConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get eksa TinkerbellMachineConfig response: %v", err)
	}

	return response, nil
}

// GetUnprovisionedTinkerbellHardware retrieves unprovisioned Tinkerbell Hardware objects.
// Unprovisioned objects are those without any owner reference information.
func (k *Kubectl) GetUnprovisionedTinkerbellHardware(ctx context.Context, kubeconfig, namespace string) ([]tinkv1alpha1.Hardware, error) {
	// Retrieve hardware resources that don't have the `v1alpha1.tinkerbell.org/ownerName` label.
	// This label is used to populate hardware when the CAPT controller acquires the Hardware
	// resource for provisioning.
	// See https://github.com/chrisdoherty4/cluster-api-provider-tinkerbell/blob/main/controllers/machine.go#L271
	params := []string{
		"get", TinkerbellHardwareResourceType,
		"-l", "!v1alpha1.tinkerbell.org/ownerName",
		"--kubeconfig", kubeconfig,
		"-o", "json",
		"--namespace", namespace,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, err
	}

	var list tinkv1alpha1.HardwareList
	if err := json.Unmarshal(stdOut.Bytes(), &list); err != nil {
		return nil, err
	}

	return list.Items, nil
}

// GetProvisionedTinkerbellHardware retrieves provisioned Tinkerbell Hardware objects.
// Provisioned objects are those with owner reference information.
func (k *Kubectl) GetProvisionedTinkerbellHardware(ctx context.Context, kubeconfig, namespace string) ([]tinkv1alpha1.Hardware, error) {
	// Retrieve hardware resources that have the `v1alpha1.tinkerbell.org/ownerName` label.
	// This label is used to populate hardware when the CAPT controller acquires the Hardware
	// resource for provisioning.
	params := []string{
		"get", TinkerbellHardwareResourceType,
		"-l", "v1alpha1.tinkerbell.org/ownerName",
		"--kubeconfig", kubeconfig,
		"-o", "json",
		"--namespace", namespace,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, err
	}

	var list tinkv1alpha1.HardwareList
	if err := json.Unmarshal(stdOut.Bytes(), &list); err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (k *Kubectl) GetEksaVSphereMachineConfig(ctx context.Context, vsphereMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereMachineConfig, error) {
	params := []string{"get", eksaVSphereMachineResourceType, vsphereMachineConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa vsphere cluster %v", err)
	}

	response := &v1alpha1.VSphereMachineConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get eksa vsphere cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaAWSDatacenterConfig(ctx context.Context, awsDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSDatacenterConfig, error) {
	params := []string{"get", eksaAwsResourceType, awsDatacenterConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting eksa aws cluster %v", err)
	}

	response := &v1alpha1.AWSDatacenterConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get eksa aws cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetCurrentClusterContext(ctx context.Context, cluster *types.Cluster) (string, error) {
	params := []string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.contexts[0].name}"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("getting current cluster context name: %v", err)
	}

	return stdOut.String(), nil
}

func (k *Kubectl) GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...KubectlOpt) (*etcdv1.EtcdadmCluster, error) {
	logger.V(6).Info("Getting EtcdadmCluster CRD", "cluster", clusterName)
	params := []string{"get", etcdadmClustersResourceType, fmt.Sprintf("%s-etcd", clusterName), "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting etcdadmCluster: %v", err)
	}

	response := &etcdv1.EtcdadmCluster{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing get etcdadmCluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) ValidateNodesVersion(ctx context.Context, kubeconfig string, kubeVersion v1alpha1.KubernetesVersion) error {
	template := "{{range .items}}{{.status.nodeInfo.kubeletVersion}}\n{{end}}"
	params := []string{"get", "nodes", "-o", "go-template", "--template", template, "--kubeconfig", kubeconfig}
	buffer, err := k.Execute(ctx, params...)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(strings.NewReader(buffer.String()))
	for scanner.Scan() {
		kubeletVersion := scanner.Text()
		if len(kubeletVersion) != 0 {
			if !strings.Contains(kubeletVersion, string(kubeVersion)) {
				return fmt.Errorf("validating node version: kubernetes version %s does not match expected version %s", kubeletVersion, kubeVersion)
			}
		}
	}
	return nil
}

func (k *Kubectl) GetBundles(ctx context.Context, kubeconfigFile, name, namespace string) (*releasev1alpha1.Bundles, error) {
	params := []string{"get", bundlesResourceType, name, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting Bundles with kubectl: %v", err)
	}

	response := &releasev1alpha1.Bundles{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing Bundles response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetClusterResourceSet(ctx context.Context, kubeconfigFile, name, namespace string) (*addons.ClusterResourceSet, error) {
	obj := &addons.ClusterResourceSet{}
	if err := k.GetObject(ctx, clusterResourceSetResourceType, name, namespace, kubeconfigFile, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) GetConfigMap(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.ConfigMap, error) {
	params := []string{"get", "configmap", name, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("getting ConfigMap with kubectl: %v", err)
	}

	response := &corev1.ConfigMap{}
	if err = json.Unmarshal(stdOut.Bytes(), response); err != nil {
		return nil, fmt.Errorf("parsing ConfigMap response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) SetDaemonSetImage(ctx context.Context, kubeconfigFile, name, namespace, container, image string) error {
	return k.setImage(ctx, "daemonset", name, container, image, WithNamespace(namespace), WithKubeconfig(kubeconfigFile))
}

func (k *Kubectl) setImage(ctx context.Context, kind, name, container, image string, opts ...KubectlOpt) error {
	params := []string{"set", "image", fmt.Sprintf("%s/%s", kind, name), fmt.Sprintf("%s=%s", container, image)}
	applyOpts(&params, opts...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("setting image for %s: %v", kind, err)
	}

	return nil
}

func (k *Kubectl) CheckProviderExists(ctx context.Context, kubeconfigFile, name, namespace string) (bool, error) {
	params := []string{"get", "namespace", fmt.Sprintf("--field-selector=metadata.name=%s", namespace), "--kubeconfig", kubeconfigFile}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return false, fmt.Errorf("checking whether provider namespace exists: %v", err)
	}
	if stdOut.Len() == 0 {
		return false, nil
	}

	params = []string{"get", capiProvidersResourceType, "--namespace", namespace, fmt.Sprintf("--field-selector=metadata.name=%s", name), "--kubeconfig", kubeconfigFile}
	stdOut, err = k.Execute(ctx, params...)
	if err != nil {
		return false, fmt.Errorf("checking whether provider exists: %v", err)
	}
	return stdOut.Len() != 0, nil
}

type Toleration struct {
	Effect            string      `json:"effect,omitempty"`
	Key               string      `json:"key,omitempty"`
	Operator          string      `json:"operator,omitempty"`
	Value             string      `json:"value,omitempty"`
	TolerationSeconds json.Number `json:"tolerationSeconds,omitempty"`
}

func (k *Kubectl) ApplyTolerationsFromTaintsToDaemonSet(ctx context.Context, oldTaints []corev1.Taint, newTaints []corev1.Taint, dsName string, kubeconfigFile string) error {
	return k.ApplyTolerationsFromTaints(ctx, oldTaints, newTaints, "ds", dsName, kubeconfigFile, "kube-system", "/spec/template/spec/tolerations")
}

func (k *Kubectl) ApplyTolerationsFromTaints(ctx context.Context, oldTaints []corev1.Taint, newTaints []corev1.Taint, resource string, name string, kubeconfigFile string, namespace string, path string) error {
	params := []string{
		"get", resource, name,
		"-o", "jsonpath={range .spec.template.spec}{.tolerations} {end}",
		"-n", namespace, "--kubeconfig", kubeconfigFile,
	}
	output, err := k.Execute(ctx, params...)
	if err != nil {
		return err
	}
	var appliedTolerations []Toleration
	if len(output.String()) > 0 {
		err = json.Unmarshal(output.Bytes(), &appliedTolerations)
		if err != nil {
			return fmt.Errorf("parsing toleration response: %v", err)
		}
	}

	oldTolerationSet := make(map[Toleration]bool)
	for _, taint := range oldTaints {
		var toleration Toleration
		toleration.Key = taint.Key
		toleration.Value = taint.Value
		toleration.Effect = string(taint.Effect)
		toleration.Operator = "Equal"
		oldTolerationSet[toleration] = true
	}

	var finalTolerations []string
	format := "{\"key\":\"%s\",\"operator\":\"%s\",\"value\":\"%s\",\"effect\":\"%s\",\"tolerationSeconds\":%s}"
	for _, toleration := range appliedTolerations {
		_, present := oldTolerationSet[toleration]
		if !present {
			finalTolerations = append(finalTolerations, fmt.Sprintf(format, toleration.Key, toleration.Operator, toleration.Value, toleration.Effect, string(toleration.TolerationSeconds)))
		}
	}
	for _, taint := range newTaints {
		finalTolerations = append(finalTolerations, fmt.Sprintf(format, taint.Key, "Equal", taint.Value, taint.Effect, ""))
	}

	if len(finalTolerations) > 0 {
		params := []string{
			"patch", resource, name,
			"--type=json", fmt.Sprintf("-p=[{\"op\": \"add\", \"path\": %s, \"value\":[%s]}]", path, strings.Join(finalTolerations, ", ")), "-n", namespace, "--kubeconfig", kubeconfigFile,
		}
		_, err = k.Execute(ctx, params...)
		if err != nil {
			return err
		}
	}
	return nil
}

// PauseCAPICluster adds a `spec.Paused: true` to the CAPI cluster resource. This will cause all
// downstream CAPI + provider controllers to skip reconciling on the paused cluster's objects.
func (k *Kubectl) PauseCAPICluster(ctx context.Context, cluster, kubeconfig string) error {
	patch := fmt.Sprintf("{\"spec\":{\"paused\":%t}}", true)
	return k.MergePatchResource(ctx, capiClustersResourceType, cluster, patch, kubeconfig, constants.EksaSystemNamespace)
}

// ResumeCAPICluster removes the `spec.Paused` on the CAPI cluster resource. This will cause all
// downstream CAPI + provider controllers to resume reconciling on the paused cluster's objects
// `spec.Paused` is set to `null` to drop the field instead of setting it to `false`.
func (k *Kubectl) ResumeCAPICluster(ctx context.Context, cluster, kubeconfig string) error {
	patch := "{\"spec\":{\"paused\":null}}"
	return k.MergePatchResource(ctx, capiClustersResourceType, cluster, patch, kubeconfig, constants.EksaSystemNamespace)
}

// MergePatchResource patches named resource using merge patch.
func (k *Kubectl) MergePatchResource(ctx context.Context, resource, name, patch, kubeconfig, namespace string) error {
	params := []string{
		"patch", resource, name, "--type=merge", "-p", patch, "--kubeconfig", kubeconfig, "--namespace", namespace,
	}

	_, err := k.Execute(ctx, params...)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kubectl) KubeconfigSecretAvailable(ctx context.Context, kubeconfig string, clusterName string, namespace string) (bool, error) {
	return k.HasResource(ctx, "secret", fmt.Sprintf("%s-kubeconfig", clusterName), kubeconfig, namespace)
}

// HasResource implements KubectlRunner.
func (k *Kubectl) HasResource(ctx context.Context, resourceType string, name string, kubeconfig string, namespace string) (bool, error) {
	throwaway := &unstructured.Unstructured{}
	err := k.Get(ctx, resourceType, kubeconfig, throwaway, withGetResourceName(name), withNamespaceOrDefaultForGet(namespace))
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetObject performs a GET call to the kube API server authenticating with a kubeconfig file
// and unmarshalls the response into the provided Object
// If the object is not found, it returns an error implementing apimachinery errors.APIStatus.
func (k *Kubectl) GetObject(ctx context.Context, resourceType, name, namespace, kubeconfig string, obj runtime.Object) error {
	return k.Get(ctx, resourceType, kubeconfig, obj, withGetResourceName(name), withNamespaceOrDefaultForGet(namespace))
}

// GetClusterObject performs a GET class like above except without namespace required.
func (k *Kubectl) GetClusterObject(ctx context.Context, resourceType, name, kubeconfig string, obj runtime.Object) error {
	return k.Get(ctx, resourceType, kubeconfig, obj, withGetResourceName(name), withClusterScope())
}

func (k *Kubectl) ListObjects(ctx context.Context, resourceType, namespace, kubeconfig string, list kubernetes.ObjectList) error {
	return k.Get(ctx, resourceType, kubeconfig, list, withNamespaceOrDefaultForGet(namespace))
}

func withGetResourceName(name string) kubernetes.KubectlGetOption {
	return &kubernetes.KubectlGetOptions{
		Name: name,
	}
}

// withNamespaceOrDefaultForGet returns an option for a get command to use the provided namespace
// or the default namespace if an empty string is provided.
// For backwards compatibility, we us the default namespace if this method is called explicitly
// with an empty namespace since some parts of the code rely on kubectl using the default namespace
// when no namespace argument is passed.
func withNamespaceOrDefaultForGet(namespace string) kubernetes.KubectlGetOption {
	if namespace == "" {
		namespace = "default"
	}
	return &kubernetes.KubectlGetOptions{
		Namespace: namespace,
	}
}

func withClusterScope() kubernetes.KubectlGetOption {
	return &kubernetes.KubectlGetOptions{
		ClusterScoped: ptr.Bool(true),
	}
}

// Get performs a kubectl get command.
func (k *Kubectl) Get(ctx context.Context, resourceType, kubeconfig string, obj runtime.Object, opts ...kubernetes.KubectlGetOption) error {
	o := &kubernetes.KubectlGetOptions{}
	for _, opt := range opts {
		opt.ApplyToGet(o)
	}

	clusterScoped := o.ClusterScoped != nil && *o.ClusterScoped

	if o.Name != "" && o.Namespace == "" && !clusterScoped {
		return errors.New("if Name is specified, Namespace is required")
	}

	params := getParams(resourceType, kubeconfig, o)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("getting %s with kubectl: %v", resourceType, err)
	}

	if stdOut.Len() == 0 {
		return newNotFoundErrorForTypeAndName(resourceType, o.Name)
	}

	if err = json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return fmt.Errorf("parsing get %s response: %v", resourceType, err)
	}

	return nil
}

func getParams(resourceType, kubeconfig string, o *kubernetes.KubectlGetOptions) []string {
	clusterScoped := o.ClusterScoped != nil && *o.ClusterScoped

	params := []string{"get", "--ignore-not-found", "-o", "json", "--kubeconfig", kubeconfig, resourceType}
	if o.Namespace != "" {
		params = append(params, "--namespace", o.Namespace)
	} else if !clusterScoped {
		params = append(params, "--all-namespaces")
	}

	if o.Name != "" {
		params = append(params, o.Name)
	}

	return params
}

// Create performs a kubectl create command.
func (k *Kubectl) Create(ctx context.Context, kubeconfig string, obj runtime.Object) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "marshalling object")
	}
	_, err = k.ExecuteWithStdin(ctx, b, "create", "-f", "-", "--kubeconfig", kubeconfig)
	if isKubectlAlreadyExistsError(err) {
		return newAlreadyExistsErrorForObj(obj)
	}

	if err != nil {
		return errors.Wrapf(err, "creating %s object with kubectl", obj.GetObjectKind().GroupVersionKind())
	}
	return nil
}

const alreadyExistsErrorMessageSubString = "AlreadyExists"

func isKubectlAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), alreadyExistsErrorMessageSubString)
}

const notFoundErrorMessageSubString = "NotFound"

// IsKubectlNotFoundError returns true if the kubectl call returned the NotFound error.
func IsKubectlNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), notFoundErrorMessageSubString)
}

func newAlreadyExistsErrorForObj(obj runtime.Object) error {
	return apierrors.NewAlreadyExists(
		groupResourceFromObj(obj),
		resourceNameFromObj(obj),
	)
}

func groupResourceFromObj(obj runtime.Object) schema.GroupResource {
	apiObj, ok := obj.(client.Object)
	if !ok {
		// If this doesn't implement the client object interface,
		// we don't know how to process it. This should never happen for
		// any of the known types.
		return schema.GroupResource{}
	}

	k := apiObj.GetObjectKind().GroupVersionKind()
	return schema.GroupResource{
		Group:    k.Group,
		Resource: k.Kind,
	}
}

func resourceNameFromObj(obj runtime.Object) string {
	apiObj, ok := obj.(client.Object)
	if !ok {
		// If this doesn't implement the client object interface,
		// we don't know how to process it. This should never happen for
		// any of the known types.
		return ""
	}

	return apiObj.GetName()
}

func newNotFoundErrorForTypeAndName(resourceType, name string) error {
	resourceTypeSplit := strings.SplitN(resourceType, ".", 2)
	gr := schema.GroupResource{Resource: resourceTypeSplit[0]}
	if len(resourceTypeSplit) == 2 {
		gr.Group = resourceTypeSplit[1]
	}
	return apierrors.NewNotFound(gr, name)
}

// Replace performs a kubectl replace command.
func (k *Kubectl) Replace(ctx context.Context, kubeconfig string, obj runtime.Object) error {
	// Even is --save-config=false is set (which is the default), kubectl replace will
	// not only respect the last-applied annotation if present in the object, but it will update
	// it with the provided state of the resource. This includes the metadata.resourceVersion. This
	// breaks future uses of kubectl apply. Since those commands' input never provide the resourceVersion,
	// kubectl will send a request trying to remove that field. That is obviously not a valid request, so
	// it gets rejected by the kube API server. To avoid this, we simply remove the annotation when passing
	// it to the replace command.
	// It's not recommended to use both imperative and "declarative" commands for the same resource. Unfortunately
	// our CLI makes extensive use of client side apply. Although not ideal, this mechanism allows us to perform
	// updates (using replace) where idempotency is necessary while maintaining the ability to continue to use apply.
	obj = removeLastAppliedAnnotation(obj)
	b, err := yaml.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "marshalling object")
	}
	if _, err := k.ExecuteWithStdin(ctx, b, "replace", "-f", "-", "--kubeconfig", kubeconfig); err != nil {
		return errors.Wrapf(err, "replacing %s object with kubectl", obj.GetObjectKind().GroupVersionKind())
	}
	return nil
}

// removeLastAppliedAnnotation deletes the kubectl last-applied annotation
// from the object if present.
func removeLastAppliedAnnotation(obj runtime.Object) runtime.Object {
	apiObj, ok := obj.(client.Object)
	// If this doesn't implement the client object interface,
	// we don't know how to access the annotations.
	// All the objects that we pass here do implement client.Client.
	if !ok {
		return obj
	}

	annotations := apiObj.GetAnnotations()
	delete(annotations, lastAppliedAnnotation)
	apiObj.SetAnnotations(annotations)

	return apiObj
}

// Delete performs a delete command authenticating with a kubeconfig file.
func (k *Kubectl) Delete(ctx context.Context, resourceType, kubeconfig string, opts ...kubernetes.KubectlDeleteOption) error {
	o := &kubernetes.KubectlDeleteOptions{}
	for _, opt := range opts {
		opt.ApplyToDelete(o)
	}

	if o.Name != "" && o.Namespace == "" {
		return errors.New("if Name is specified, Namespace is required")
	}

	if o.Name != "" && o.HasLabels != nil {
		return errors.New("options for HasLabels and Name are mutually exclusive")
	}

	params := deleteParams(resourceType, kubeconfig, o)
	_, err := k.Execute(ctx, params...)
	if IsKubectlNotFoundError(err) {
		return newNotFoundErrorForTypeAndName(resourceType, o.Name)
	}

	if err != nil {
		return errors.Wrapf(err, "deleting %s", resourceType)
	}
	return nil
}

func deleteParams(resourceType, kubeconfig string, o *kubernetes.KubectlDeleteOptions) []string {
	params := []string{"delete", "--kubeconfig", kubeconfig, resourceType}
	if o.Name != "" {
		params = append(params, o.Name)
	} else if o.HasLabels == nil {
		params = append(params, "--all")
	}

	if o.Namespace != "" {
		params = append(params, "--namespace", o.Namespace)
	} else {
		params = append(params, "--all-namespaces")
	}

	if len(o.HasLabels) > 0 {
		labelConstrains := make([]string, 0, len(o.HasLabels))
		for l, v := range o.HasLabels {
			labelConstrains = append(labelConstrains, l+"="+v)
		}

		sort.Strings(labelConstrains)

		params = append(params, "--selector", strings.Join(labelConstrains, ","))
	}

	return params
}

// Apply creates the resource or it updates if it already exists.
func (k *Kubectl) Apply(ctx context.Context, kubeconfig string, obj runtime.Object, opts ...kubernetes.KubectlApplyOption) error {
	o := &kubernetes.KubectlApplyOptions{}
	for _, opt := range opts {
		opt.ApplyToApply(o)
	}

	b, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshalling object: %v", err)
	}

	params := []string{"apply", "-f", "-", "--kubeconfig", kubeconfig}
	if o.FieldManager != "" {
		params = append(params, "--field-manager", o.FieldManager)
	}
	if o.ServerSide {
		params = append(params, "--server-side")
	}
	if o.ForceOwnership {
		params = append(params, "--force-conflicts")
	}

	if _, err := k.ExecuteWithStdin(ctx, b, params...); err != nil {
		return fmt.Errorf("applying object with kubectl: %v", err)
	}
	return nil
}

func (k *Kubectl) GetEksdRelease(ctx context.Context, name, namespace, kubeconfigFile string) (*eksdv1alpha1.Release, error) {
	obj := &eksdv1alpha1.Release{}
	if err := k.GetObject(ctx, eksdReleaseType, name, namespace, kubeconfigFile, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) GetDeployment(ctx context.Context, name, namespace, kubeconfig string) (*appsv1.Deployment, error) {
	obj := &appsv1.Deployment{}
	if err := k.GetObject(ctx, "deployment", name, namespace, kubeconfig, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) GetDaemonSet(ctx context.Context, name, namespace, kubeconfig string) (*appsv1.DaemonSet, error) {
	obj := &appsv1.DaemonSet{}
	if err := k.GetObject(ctx, "daemonset", name, namespace, kubeconfig, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) ExecuteCommand(ctx context.Context, opts ...string) (bytes.Buffer, error) {
	return k.Execute(ctx, opts...)
}

// DeleteClusterObject performs a DELETE call like above except without namespace required.
func (k *Kubectl) DeleteClusterObject(ctx context.Context, resourceType, name, kubeconfig string) error {
	if _, err := k.Execute(ctx, "delete", resourceType, name, "--kubeconfig", kubeconfig); err != nil {
		return fmt.Errorf("deleting %s %s: %v", name, resourceType, err)
	}
	return nil
}

func (k *Kubectl) ExecuteFromYaml(ctx context.Context, yaml []byte, opts ...string) (bytes.Buffer, error) {
	return k.ExecuteWithStdin(ctx, yaml, opts...)
}

func (k *Kubectl) SearchNutanixMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.NutanixMachineConfig, error) {
	params := []string{
		"get", eksaNutanixMachineResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa NutanixMachineConfigResponse: %v", err)
	}

	response := &NutanixMachineConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing NutanixMachineConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) SearchNutanixDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.NutanixDatacenterConfig, error) {
	params := []string{
		"get", eksaNutanixDatacenterResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + name,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("searching eksa NutanixDatacenterConfigResponse: %v", err)
	}

	response := &NutanixDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("parsing NutanixDatacenterConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetEksaNutanixDatacenterConfig(ctx context.Context, nutanixDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.NutanixDatacenterConfig, error) {
	response := &v1alpha1.NutanixDatacenterConfig{}
	err := k.GetObject(ctx, eksaNutanixDatacenterResourceType, nutanixDatacenterConfigName, namespace, kubeconfigFile, response)
	if err != nil {
		return nil, fmt.Errorf("getting eksa nutanix datacenterconfig: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaNutanixMachineConfig(ctx context.Context, nutanixMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.NutanixMachineConfig, error) {
	response := &v1alpha1.NutanixMachineConfig{}
	err := k.GetObject(ctx, eksaNutanixMachineResourceType, nutanixMachineConfigName, namespace, kubeconfigFile, response)
	if err != nil {
		return nil, fmt.Errorf("getting eksa nutanix machineconfig: %v", err)
	}

	return response, nil
}

func (k *Kubectl) DeleteEksaNutanixDatacenterConfig(ctx context.Context, nutanixDatacenterConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaNutanixDatacenterResourceType, nutanixDatacenterConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting nutanixdatacenterconfig cluster %s apply: %v", nutanixDatacenterConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteEksaNutanixMachineConfig(ctx context.Context, nutanixMachineConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaNutanixMachineResourceType, nutanixMachineConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("deleting nutanixmachineconfig cluster %s apply: %v", nutanixMachineConfigName, err)
	}
	return nil
}

// AllBaseboardManagements returns all the baseboard management resources in the cluster.
func (k *Kubectl) AllBaseboardManagements(ctx context.Context, kubeconfig string) ([]rufiounreleased.BaseboardManagement, error) {
	stdOut, err := k.Execute(ctx,
		"get", "baseboardmanagements.bmc.tinkerbell.org",
		"-o", "json",
		"--kubeconfig", kubeconfig,
		"--all-namespaces=true",
	)
	if err != nil {
		return nil, err
	}

	var list rufiounreleased.BaseboardManagementList
	if err := json.Unmarshal(stdOut.Bytes(), &list); err != nil {
		return nil, err
	}

	return list.Items, nil
}

// AllTinkerbellHardware returns all the hardware resources in the cluster.
func (k *Kubectl) AllTinkerbellHardware(ctx context.Context, kubeconfig string) ([]tinkv1alpha1.Hardware, error) {
	stdOut, err := k.Execute(ctx,
		"get", "hardware.tinkerbell.org",
		"-o", "json",
		"--kubeconfig", kubeconfig,
		"--all-namespaces=true",
	)
	if err != nil {
		return nil, err
	}

	var list tinkv1alpha1.HardwareList
	if err := json.Unmarshal(stdOut.Bytes(), &list); err != nil {
		return nil, err
	}

	return list.Items, nil
}

// HasCRD checks if the given CRD exists in the cluster specified by kubeconfig.
func (k *Kubectl) HasCRD(ctx context.Context, crd, kubeconfig string) (bool, error) {
	_, err := k.Execute(ctx, "get", "customresourcedefinition", crd, "--kubeconfig", kubeconfig)

	if err == nil {
		return true, nil
	}

	if strings.Contains(err.Error(), "NotFound") {
		return false, nil
	}

	return false, err
}

// DeleteCRD removes the given CRD from the cluster specified in kubeconfig.
func (k *Kubectl) DeleteCRD(ctx context.Context, crd, kubeconfig string) error {
	_, err := k.Execute(ctx, "delete", "customresourcedefinition", crd, "--kubeconfig", kubeconfig)
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return err
	}

	return nil
}
