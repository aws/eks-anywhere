package executables

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/version"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	kubectlPath = "kubectl"
)

var (
	capiClustersResourceType             = fmt.Sprintf("clusters.%s", clusterv1.GroupVersion.Group)
	eksaClusterResourceType              = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereDatacenterResourceType    = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaVSphereMachineResourceType       = fmt.Sprintf("vspheremachineconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackDatacenterResourceType = fmt.Sprintf("cloudstackdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaCloudStackMachineResourceType    = fmt.Sprintf("cloudstackmachineconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaAwsResourceType                  = fmt.Sprintf("awsdatacenterconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaGitOpsResourceType               = fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaOIDCResourceType                 = fmt.Sprintf("oidcconfigs.%s", v1alpha1.GroupVersion.Group)
	eksaAwsIamResourceType               = fmt.Sprintf("awsiamconfigs.%s", v1alpha1.GroupVersion.Group)
	etcdadmClustersResourceType          = fmt.Sprintf("etcdadmclusters.%s", etcdv1.GroupVersion.Group)
	bundlesResourceType                  = fmt.Sprintf("bundles.%s", releasev1alpha1.GroupVersion.Group)
	clusterResourceSetResourceType       = fmt.Sprintf("clusterresourcesets.%s", addons.GroupVersion.Group)
	kubeadmControlPlaneResourceType      = fmt.Sprintf("kubeadmcontrolplanes.controlplane.%s", clusterv1.GroupVersion.Group)
)

type Kubectl struct {
	Executable
}

func (k *Kubectl) GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error) {
	response := &v1alpha1.CloudStackMachineConfig{}
	err := k.getObject(ctx, eksaCloudStackMachineResourceType, cloudstackMachineConfigName, namespace, kubeconfigFile, response)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa cloudstack machineconfig: %v", err)
	}

	return response, nil
}

func (k *Kubectl) DeleteEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaCloudStackDatacenterResourceType, cloudstackDatacenterConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting cloudstackdatacenterconfig cluster %s apply: %v", cloudstackDatacenterConfigName, err)
	}
	return nil
}

func (k *Kubectl) GetEksaCloudStackDatacenterConfig(ctx context.Context, cloudstackDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackDatacenterConfig, error) {
	response := &v1alpha1.CloudStackDatacenterConfig{}
	err := k.getObject(ctx, eksaCloudStackDatacenterResourceType, cloudstackDatacenterConfigName, namespace, kubeconfigFile, response)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa cloudstack machineconfig: %v", err)
	}

	return response, nil
}

func (k *Kubectl) DeleteEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaCloudStackMachineResourceType, cloudstackMachineConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting cloudstackmachineconfig cluster %s apply: %v", cloudstackMachineConfigName, err)
	}
	return nil
}

type VersionResponse struct {
	ClientVersion version.Info `json:"clientVersion"`
	ServerVersion version.Info `json:"serverVersion"`
}

func NewKubectl(executable Executable) *Kubectl {
	return &Kubectl{
		Executable: executable,
	}
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
		return fmt.Errorf("error creating namespace %v: %v", namespace, err)
	}
	return nil
}

func (k *Kubectl) DeleteNamespace(ctx context.Context, kubeconfig string, namespace string) error {
	params := []string{"delete", "namespace", namespace, "--kubeconfig", kubeconfig}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error creating namespace %v: %v", namespace, err)
	}
	return nil
}

func (k *Kubectl) LoadSecret(ctx context.Context, secretObject string, secretObjectType string, secretObjectName string, kubeConfFile string) error {
	params := []string{"create", "secret", "generic", secretObjectName, "--type", secretObjectType, "--from-literal", secretObject, "--kubeconfig", kubeConfFile, "--namespace", constants.EksaSystemNamespace}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error loading secret: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyHardware(ctx context.Context, hardwareYaml string, kubeConfFile string) error {
	params := []string{"apply", "-f", hardwareYaml}
	params = append(params, "--kubeconfig", kubeConfFile)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error executing hardware yaml apply: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyKubeSpec(ctx context.Context, cluster *types.Cluster, spec string) error {
	params := []string{"apply", "-f", spec}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error executing apply: %v", err)
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
		return fmt.Errorf("error executing apply: %v", err)
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
		return fmt.Errorf("error executing apply: %v", err)
	}
	return nil
}

func (k *Kubectl) ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error {
	params := []string{"apply", "-f", "-", "--namespace", namespace}
	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}
	_, err := k.ExecuteWithStdin(ctx, data, params...)
	if err != nil {
		return fmt.Errorf("error executing apply: %v", err)
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
		return fmt.Errorf("error executing apply --force: %v", err)
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
		return fmt.Errorf("error executing apply: %v", err)
	}
	return nil
}

func (k *Kubectl) WaitForControlPlaneReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ControlPlaneReady", fmt.Sprintf("%s/%s", capiClustersResourceType, newClusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForManagedExternalEtcdReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, "ManagedEtcdReady", fmt.Sprintf("clusters.%s/%s", clusterv1.GroupVersion.Group, newClusterName), constants.EksaSystemNamespace)
}

func (k *Kubectl) WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error {
	return k.Wait(ctx, cluster.KubeconfigFile, timeout, condition, "deployments/"+target, namespace)
}

func (k *Kubectl) Wait(ctx context.Context, kubeconfig string, timeout string, forCondition string, property string, namespace string) error {
	_, err := k.Execute(ctx, "wait", "--timeout", timeout,
		"--for=condition="+forCondition, property, "--kubeconfig", kubeconfig, "-n", namespace)
	if err != nil {
		return fmt.Errorf("error executing wait: %v", err)
	}
	return nil
}

func (k *Kubectl) DeleteEksaDatacenterConfig(ctx context.Context, eksaDatacenterResourceType string, eksaDatacenterConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaDatacenterResourceType, eksaDatacenterConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting %s cluster %s apply: %v", eksaDatacenterResourceType, eksaDatacenterConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteEksaMachineConfig(ctx context.Context, eksaMachineConfigResourceType string, eksaMachineConfigName string, kubeconfigFile string, namespace string) error {
	params := []string{"delete", eksaMachineConfigResourceType, eksaMachineConfigName, "--kubeconfig", kubeconfigFile, "--namespace", namespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting %s cluster %s apply: %v", eksaMachineConfigResourceType, eksaMachineConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteEKSACluster(ctx context.Context, managementCluster *types.Cluster, eksaClusterName, eksaClusterNamespace string) error {
	params := []string{"delete", eksaClusterResourceType, eksaClusterName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", eksaClusterNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting eksa cluster %s apply: %v", eksaClusterName, err)
	}
	return nil
}

func (k *Kubectl) DeleteGitOpsConfig(ctx context.Context, managementCluster *types.Cluster, gitOpsConfigName, gitOpsConfigNamespace string) error {
	params := []string{"delete", eksaGitOpsResourceType, gitOpsConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", gitOpsConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting gitops config %s apply: %v", gitOpsConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteSecret(ctx context.Context, managementCluster *types.Cluster, secretName, namespace string) error {
	params := []string{"delete", "secret", secretName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", namespace}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting secret %s in namespace %s: %v", secretName, namespace, err)
	}
	return nil
}

func (k *Kubectl) DeleteOIDCConfig(ctx context.Context, managementCluster *types.Cluster, oidcConfigName, oidcConfigNamespace string) error {
	params := []string{"delete", eksaOIDCResourceType, oidcConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", oidcConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting oidc config %s apply: %v", oidcConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteAWSIamConfig(ctx context.Context, managementCluster *types.Cluster, awsIamConfigName, awsIamConfigNamespace string) error {
	params := []string{"delete", eksaAwsIamResourceType, awsIamConfigName, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", awsIamConfigNamespace, "--ignore-not-found=true"}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting awsIam config %s apply: %v", awsIamConfigName, err)
	}
	return nil
}

func (k *Kubectl) DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster) error {
	params := []string{"delete", capiClustersResourceType, clusterToDelete.Name, "--kubeconfig", managementCluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error deleting cluster %s apply: %v", clusterToDelete.Name, err)
	}
	return nil
}

func (k *Kubectl) ListCluster(ctx context.Context) error {
	params := []string{"get", "pods", "-A", "-o", "jsonpath={..image}"}
	output, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error listing cluster versions: %v", err)
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
		return nil, fmt.Errorf("error getting nodes: %v", err)
	}
	response := &corev1.NodeList{}
	err = json.Unmarshal(stdOut.Bytes(), response)

	return response.Items, err
}

func (k *Kubectl) GetControlPlaneNodes(ctx context.Context, kubeconfig string) ([]corev1.Node, error) {
	params := []string{"get", "nodes", "-o", "json", "--kubeconfig", kubeconfig, "--selector=node-role.kubernetes.io/control-plane"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting control plane nodes: %v", err)
	}
	response := &corev1.NodeList{}
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
	deployments, err := k.GetMachineDeployments(ctx, WithKubeconfig(kubeconfig), WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return err
	}
	for _, machineDeployment := range deployments {
		if machineDeployment.Status.Phase != "Running" {
			return fmt.Errorf("machine deployment is in %s phase", machineDeployment.Status.Phase)
		}

		if machineDeployment.Status.UnavailableReplicas != 0 {
			return fmt.Errorf("%v machine deployment replicas are unavailable", machineDeployment.Status.UnavailableReplicas)
		}

		if machineDeployment.Status.ReadyReplicas != machineDeployment.Status.Replicas {
			return fmt.Errorf("%v machine deployment replicas are not ready", machineDeployment.Status.Replicas-machineDeployment.Status.ReadyReplicas)
		}
	}
	return nil
}

func (k *Kubectl) VsphereWorkerNodesMachineTemplate(ctx context.Context, clusterName string, kubeconfig string, namespace string) (*vspherev1.VSphereMachineTemplate, error) {
	machineTemplateName, err := k.MachineTemplateName(ctx, clusterName, kubeconfig, WithNamespace(namespace))
	if err != nil {
		return nil, err
	}

	params := []string{"get", "vspheremachinetemplates", machineTemplateName, "-o", "go-template", "--template", "{{.spec.template.spec}}", "-o", "yaml", "--kubeconfig", kubeconfig, "--namespace", namespace}
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

func (k *Kubectl) MachineTemplateName(ctx context.Context, clusterName string, kubeconfig string, opts ...KubectlOpt) (string, error) {
	template := "{{.spec.template.spec.infrastructureRef.name}}"
	params := []string{"get", "MachineDeployment", fmt.Sprintf("%s-md-0", clusterName), "-o", "go-template", "--template", template, "--kubeconfig", kubeconfig}
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
		return fmt.Errorf("error saving logs: %v", err)
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
		"get", "machines", "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
		"--selector=cluster.x-k8s.io/cluster-name=" + clusterName,
		"--namespace", constants.EksaSystemNamespace,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting machines: %v", err)
	}

	response := &machinesResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get machines response: %v", err)
	}

	return response.Items, nil
}

type machineSetResponse struct {
	Items []clusterv1.MachineSet `json:"items,omitempty"`
}

func (k *Kubectl) GetMachineSets(ctx context.Context, machineDeploymentName string, cluster *types.Cluster) ([]clusterv1.MachineSet, error) {
	params := []string{
		"get", "machinesets", "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
		"--selector=cluster.x-k8s.io/deployment-name=" + machineDeploymentName,
		"--namespace", constants.EksaSystemNamespace,
	}

	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting machineset associated with deployment %s: %v", machineDeploymentName, err)
	}

	response := &machineSetResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get machinesets response: %v", err)
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

type IdentityProviderConfigResponse struct {
	Items []*v1alpha1.Ref `json:"items,omitempty"`
}

type VSphereMachineConfigResponse struct {
	Items []*v1alpha1.VSphereMachineConfig `json:"items,omitempty"`
}

func (k *Kubectl) ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error {
	params := []string{"get", "crd", capiClustersResourceType, "--kubeconfig", cluster.KubeconfigFile}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error getting clusters crd: %v", err)
	}
	return nil
}

func (k *Kubectl) ValidateEKSAClustersCRD(ctx context.Context, cluster *types.Cluster) error {
	params := []string{"get", "crd", eksaClusterResourceType, "--kubeconfig", cluster.KubeconfigFile}
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error getting eksa clusters crd: %v", err)
	}
	return nil
}

func (k *Kubectl) GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error) {
	params := []string{"get", capiClustersResourceType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting clusters: %v", err)
	}

	response := &ClustersResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get clusters response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetApiServerUrl(ctx context.Context, cluster *types.Cluster) (string, error) {
	params := []string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.clusters[0].cluster.server}"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("error getting api server url: %v", err)
	}

	return stdOut.String(), nil
}

func (k *Kubectl) GetClusterCATlsCert(ctx context.Context, clusterName string, cluster *types.Cluster, namespace string) ([]byte, error) {
	secretName := fmt.Sprintf("%s-ca", clusterName)
	params := []string{"get", "secret", secretName, "--kubeconfig", cluster.KubeconfigFile, "-o", `jsonpath={.data.tls\.crt}`, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting cluster ca tls cert: %v", err)
	}

	return stdOut.Bytes(), nil
}

func (k *Kubectl) Version(ctx context.Context, cluster *types.Cluster) (*VersionResponse, error) {
	params := []string{"version", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error executing kubectl version: %v", err)
	}
	response := &VersionResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling kubectl version response: %v", err)
	}
	return response, nil
}

type KubectlOpt func(*[]string)

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

func WithAllNamespaces() KubectlOpt {
	return appendOpt("-A")
}

func WithSkipTLSVerify() KubectlOpt {
	return appendOpt("--insecure-skip-tls-verify=true")
}

func WithOverwrite() KubectlOpt {
	return appendOpt("--overwrite")
}

func WithOutput(output string) KubectlOpt {
	return appendOpt("-o", output)
}

func WithArgs(args []string) KubectlOpt {
	return appendOpt(args...)
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
		return nil, fmt.Errorf("error getting pods: %v", err)
	}

	response := &corev1.PodList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get pods response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetDeployments(ctx context.Context, opts ...KubectlOpt) ([]appsv1.Deployment, error) {
	params := []string{"get", "deployments", "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting deployments: %v", err)
	}

	response := &appsv1.DeploymentList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get deployments response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.Secret, error) {
	return k.GetSecret(ctx, name, WithKubeconfig(kubeconfigFile), WithNamespace(namespace))
}

func (k *Kubectl) GetSecret(ctx context.Context, secretObjectName string, opts ...KubectlOpt) (*corev1.Secret, error) {
	params := []string{"get", "secret", secretObjectName, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting secret: %v", err)
	}
	response := &corev1.Secret{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get secret response: %v", err)
	}
	return response, err
}

func (k *Kubectl) GetKubeadmControlPlanes(ctx context.Context, opts ...KubectlOpt) ([]controlplanev1.KubeadmControlPlane, error) {
	params := []string{"get", kubeadmControlPlaneResourceType, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting kubeadmcontrolplanes: %v", err)
	}

	response := &controlplanev1.KubeadmControlPlaneList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get kubeadmcontrolplanes response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...KubectlOpt) (*controlplanev1.KubeadmControlPlane, error) {
	logger.V(6).Info("Getting KubeadmControlPlane CRDs", "cluster", clusterName)
	params := []string{"get", kubeadmControlPlaneResourceType, clusterName, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting kubeadmcontrolplane: %v", err)
	}

	response := &controlplanev1.KubeadmControlPlane{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get kubeadmcontrolplane response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...KubectlOpt) (*clusterv1.MachineDeployment, error) {
	params := []string{"get", fmt.Sprintf("machinedeployments.%s", clusterv1.GroupVersion.Group), workerNodeGroupName, "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting machine deployment: %v", err)
	}

	response := &clusterv1.MachineDeployment{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get machineDeployment response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetMachineDeployments(ctx context.Context, opts ...KubectlOpt) ([]clusterv1.MachineDeployment, error) {
	params := []string{"get", fmt.Sprintf("machinedeployments.%s", clusterv1.GroupVersion.Group), "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting machine deployments: %v", err)
	}

	response := &clusterv1.MachineDeploymentList{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get machineDeployments response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) UpdateEnvironmentVariables(ctx context.Context, resourceType, resourceName string, envMap map[string]string, opts ...KubectlOpt) error {
	params := []string{"set", "env", resourceType, resourceName}
	for k, v := range envMap {
		params = append(params, fmt.Sprintf("%s=%s", k, v))
	}
	applyOpts(&params, opts...)
	_, err := k.Execute(ctx, params...)
	if err != nil {
		return fmt.Errorf("error setting the environment variables in %s %s: %v", resourceType, resourceName, err)
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
		return fmt.Errorf("error updating annotation: %v", err)
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
		return fmt.Errorf("error removing annotation: %v", err)
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
		return nil, fmt.Errorf("error getting eksa cluster: %v", err)
	}

	response := &v1alpha1.Cluster{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get eksa cluster response: %v", err)
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
		return nil, fmt.Errorf("error searching eksa VSphereMachineConfigResponse: %v", err)
	}

	response := &VSphereMachineConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing VSphereMachineConfigResponse response: %v", err)
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
		return nil, fmt.Errorf("error searching eksa IdentityProvider: %v", err)
	}

	response := &VSphereDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing IdentityProviderConfigResponse response: %v", err)
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
		return nil, fmt.Errorf("error searching eksa VSphereDatacenterConfigResponse: %v", err)
	}

	response := &VSphereDatacenterConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing VSphereDatacenterConfigResponse response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) SearchEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) ([]*v1alpha1.GitOpsConfig, error) {
	params := []string{
		"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
		kubeconfigFile, "--namespace", namespace, "--field-selector=metadata.name=" + gitOpsConfigName,
	}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error searching eksa GitOpsConfig: %v", err)
	}

	response := &GitOpsConfigResponse{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing GitOpsConfig response: %v", err)
	}

	return response.Items, nil
}

func (k *Kubectl) GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error) {
	params := []string{"get", eksaGitOpsResourceType, gitOpsConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa GitOpsConfig: %v", err)
	}

	response := &v1alpha1.GitOpsConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing GitOpsConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error) {
	params := []string{"get", eksaOIDCResourceType, oidcConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa OIDCConfig: %v", err)
	}

	response := &v1alpha1.OIDCConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing OIDCConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error) {
	params := []string{"get", eksaAwsIamResourceType, awsIamConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa AWSIamConfig: %v", err)
	}

	response := &v1alpha1.AWSIamConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing AWSIamConfig response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error) {
	params := []string{"get", eksaVSphereDatacenterResourceType, vsphereDatacenterConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa vsphere cluster %v", err)
	}

	response := &v1alpha1.VSphereDatacenterConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get eksa vsphere cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaVSphereMachineConfig(ctx context.Context, vsphereMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereMachineConfig, error) {
	params := []string{"get", eksaVSphereMachineResourceType, vsphereMachineConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa vsphere cluster %v", err)
	}

	response := &v1alpha1.VSphereMachineConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get eksa vsphere cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetEksaAWSDatacenterConfig(ctx context.Context, awsDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSDatacenterConfig, error) {
	params := []string{"get", eksaAwsResourceType, awsDatacenterConfigName, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting eksa aws cluster %v", err)
	}

	response := &v1alpha1.AWSDatacenterConfig{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get eksa aws cluster response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetCurrentClusterContext(ctx context.Context, cluster *types.Cluster) (string, error) {
	params := []string{"config", "view", "--kubeconfig", cluster.KubeconfigFile, "--minify", "--raw", "-o", "jsonpath={.contexts[0].name}"}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("error getting current cluster context name: %v", err)
	}

	return stdOut.String(), nil
}

func (k *Kubectl) GetEtcdadmCluster(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...KubectlOpt) (*etcdv1.EtcdadmCluster, error) {
	logger.V(6).Info("Getting EtcdadmCluster CRD", "cluster", clusterName)
	params := []string{"get", etcdadmClustersResourceType, fmt.Sprintf("%s-etcd", clusterName), "-o", "json"}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting etcdadmCluster: %v", err)
	}

	response := &etcdv1.EtcdadmCluster{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get etcdadmCluster response: %v", err)
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
				return fmt.Errorf("error validating node version: kubernetes version %s does not match expected version %s", kubeletVersion, kubeVersion)
			}
		}
	}
	return nil
}

func (k *Kubectl) GetBundles(ctx context.Context, kubeconfigFile, name, namespace string) (*releasev1alpha1.Bundles, error) {
	params := []string{"get", bundlesResourceType, name, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting Bundles with kubectl: %v", err)
	}

	response := &releasev1alpha1.Bundles{}
	err = json.Unmarshal(stdOut.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing Bundles response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetClusterResourceSet(ctx context.Context, kubeconfigFile, name, namespace string) (*addons.ClusterResourceSet, error) {
	params := []string{"get", clusterResourceSetResourceType, name, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting ClusterResourceSet with kubectl: %v", err)
	}

	response := &addons.ClusterResourceSet{}
	if err = json.Unmarshal(stdOut.Bytes(), response); err != nil {
		return nil, fmt.Errorf("error parsing ClusterResourceSet response: %v", err)
	}

	return response, nil
}

func (k *Kubectl) GetConfigMap(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.ConfigMap, error) {
	params := []string{"get", "configmap", name, "-o", "json", "--kubeconfig", kubeconfigFile, "--namespace", namespace}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting ConfigMap with kubectl: %v", err)
	}

	response := &corev1.ConfigMap{}
	if err = json.Unmarshal(stdOut.Bytes(), response); err != nil {
		return nil, fmt.Errorf("error parsing ConfigMap response: %v", err)
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
		return fmt.Errorf("error setting image for %s: %v", kind, err)
	}

	return nil
}

func (k *Kubectl) CheckProviderExists(ctx context.Context, kubeconfigFile, name, namespace string) (bool, error) {
	params := []string{"get", "namespace", fmt.Sprintf("--field-selector=metadata.name=%s", namespace), "--kubeconfig", kubeconfigFile}
	stdOut, err := k.Execute(ctx, params...)
	if err != nil {
		return false, fmt.Errorf("error checking whether provider namespace exists: %v", err)
	}
	if stdOut.Len() == 0 {
		return false, nil
	}

	params = []string{"get", "provider", "--namespace", namespace, fmt.Sprintf("--field-selector=metadata.name=%s", name), "--kubeconfig", kubeconfigFile}
	stdOut, err = k.Execute(ctx, params...)
	if err != nil {
		return false, fmt.Errorf("error checking whether provider exists: %v", err)
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
			return fmt.Errorf("error parsing toleration response: %v", err)
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

func (k *Kubectl) KubeconfigSecretAvailable(ctx context.Context, kubeconfig string, clusterName string, namespace string) (bool, error) {
	return k.GetResource(ctx, "secret", fmt.Sprintf("%s-kubeconfig", clusterName), kubeconfig, namespace)
}

func (k *Kubectl) GetResource(ctx context.Context, resourceType string, name string, kubeconfig string, namespace string) (bool, error) {
	params := []string{"get", resourceType, name, "--ignore-not-found", "-n", namespace, "--kubeconfig", kubeconfig}
	output, err := k.Execute(ctx, params...)
	var found bool
	if err == nil && len(output.String()) > 0 {
		found = true
	}
	return found, err
}

func (k *Kubectl) getObject(ctx context.Context, resourceType, name, namespace, kubeconfig string, obj client.Object) error {
	stdOut, err := k.Execute(ctx, "get", "--namespace", namespace, resourceType, name, "-o", "json", "--kubeconfig", kubeconfig)
	if err != nil {
		return fmt.Errorf("error getting %s with kubectl: %v", resourceType, err)
	}

	if err = json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return fmt.Errorf("error parsing %s response: %v", resourceType, err)
	}

	return nil
}

func (k *Kubectl) GetDeployment(ctx context.Context, name, namespace, kubeconfig string) (*appsv1.Deployment, error) {
	obj := &appsv1.Deployment{}
	if err := k.getObject(ctx, "deployment", name, namespace, kubeconfig, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) GetDaemonSet(ctx context.Context, name, namespace, kubeconfig string) (*appsv1.DaemonSet, error) {
	obj := &appsv1.DaemonSet{}
	if err := k.getObject(ctx, "daemonset", name, namespace, kubeconfig, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (k *Kubectl) GetPackages(ctx context.Context, opts ...KubectlOpt) (string, error) {
	params := []string{
		"get", "packages",
	}
	applyOpts(&params, opts...)
	stdOut, err := k.Execute(ctx, params...)
	return stdOut.String(), err
}
