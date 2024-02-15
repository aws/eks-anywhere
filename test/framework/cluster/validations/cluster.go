package validations

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors1 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	apierrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

// ValidateClusterReady gets the CAPICluster from the client then validates that it is in a ready state. Also check if CAPI objects are in expected state for InPlace Upgrades.
func ValidateClusterReady(ctx context.Context, vc clusterf.StateValidationConfig) error {
	clus := vc.ClusterSpec.Cluster
	mgmtClusterClient := vc.ManagementClusterClient
	capiCluster, err := controller.GetCAPICluster(ctx, mgmtClusterClient, clus)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}
	if capiCluster == nil {
		return fmt.Errorf("cluster %s does not exist", clus.Name)
	}
	if conditions.IsFalse(capiCluster, v1beta1.ReadyCondition) {
		return fmt.Errorf("CAPI cluster %s not ready yet. %s", capiCluster.GetName(), conditions.GetReason(capiCluster, v1beta1.ReadyCondition))
	}
	if clus.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil && clus.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type == v1alpha1.InPlaceStrategyType {
		return validateCAPIobjectsForInPlace(ctx, vc)
	}
	return nil
}

// ValidateEKSAObjects retrieves all the child objects from the cluster.Spec and validates that they exist in the clusterf.
func ValidateEKSAObjects(ctx context.Context, vc clusterf.StateValidationConfig) error {
	mgmtClusterClient := vc.ManagementClusterClient
	errorList := make([]error, 0)
	for _, obj := range vc.ClusterSpec.ChildObjects() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		key := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
		if key.Namespace == "" {
			key.Namespace = "default"
		}
		if err := mgmtClusterClient.Get(ctx, key, u); err != nil {
			errorList = append(errorList, errors.Wrap(err, "reading eks-a cluster's child object"))
		}
	}
	if len(errorList) > 0 {
		return apierrors.NewAggregate(errorList)
	}
	return nil
}

func validateCAPIobjectsForInPlace(ctx context.Context, vc clusterf.StateValidationConfig) error {
	if err := validateKCP(ctx, vc); err != nil {
		return fmt.Errorf("failed to validate KubeadmControlPlane: %v", err)
	}
	if err := validateMDs(ctx, vc); err != nil {
		return fmt.Errorf("failed to validate MachineDeployment: %v", err)
	}
	if err := validateInPlaceCRsDoesNotExist(ctx, vc); err != nil {
		return fmt.Errorf("failed to validate InPlace CRDs: %v", err)
	}
	return nil
}

func validateKCP(ctx context.Context, vc clusterf.StateValidationConfig) error {
	kcp, err := controller.GetKubeadmControlPlane(ctx, vc.ManagementClusterClient, vc.ClusterSpec.Cluster)
	if err != nil {
		return fmt.Errorf("failed to retrieve kcp: %s", err)
	}
	if kcp == nil {
		return errors.New("KubeadmControlPlane object not found")
	}
	if conditions.IsFalse(kcp, v1beta1.ReadyCondition) {
		return errors.New("kcp ready condition is not true")
	} else if kcp.Status.UpdatedReplicas != kcp.Status.ReadyReplicas || *kcp.Spec.Replicas != kcp.Status.UpdatedReplicas {
		return fmt.Errorf("kcp replicas count %d, updated replicas count %d and ready replicas count %d are not in sync", *kcp.Spec.Replicas, kcp.Status.UpdatedReplicas, kcp.Status.ReadyReplicas)
	}
	return nil
}

func validateMDs(ctx context.Context, vc clusterf.StateValidationConfig) error {
	mds, err := controller.GetMachineDeployments(ctx, vc.ManagementClusterClient, vc.ClusterSpec.Cluster)
	if err != nil {
		return fmt.Errorf("failed to retrieve machinedeployments: %s", err)
	}
	if len(mds) == 0 && len(vc.ClusterSpec.Config.Cluster.Spec.WorkerNodeGroupConfigurations) != 0 {
		return errors.New("machinedeployment object not found")
	}
	for _, md := range mds {
		if conditions.IsFalse(&md, v1beta1.ReadyCondition) {
			return fmt.Errorf("md ready condition is not true for md %s", md.Name)
		} else if md.Status.UpdatedReplicas != md.Status.ReadyReplicas || *md.Spec.Replicas != md.Status.UpdatedReplicas {
			return fmt.Errorf("md replicas count %d, updated replicas count %d and ready replicas count %d for md %s are not in sync", *md.Spec.Replicas, md.Status.UpdatedReplicas, md.Status.ReadyReplicas, md.Name)
		}
	}
	return nil
}

func validateInPlaceCRsDoesNotExist(ctx context.Context, vc clusterf.StateValidationConfig) error {
	if err := validateCPUDeleted(ctx, vc); err != nil {
		return err
	}
	if err := validateMDUsDeleted(ctx, vc); err != nil {
		return err
	}
	if err := validateNUsAndPodsDeleted(ctx, vc); err != nil {
		return err
	}
	return nil
}

func validateCPUDeleted(ctx context.Context, vc clusterf.StateValidationConfig) error {
	clusterName := vc.ClusterSpec.Cluster.Name
	client := vc.ManagementClusterClient
	cpu := &v1alpha1.ControlPlaneUpgrade{}
	key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: clusterName + "-cp-upgrade"}
	if err := client.Get(ctx, key, cpu); err != nil {
		if !apierrors1.IsNotFound(err) {
			return fmt.Errorf("failed to get ControlPlaneUpgrade: %s", err)
		}
	}
	if cpu.Name != "" {
		return errors.New("CPUpgrade object not expected but still exists on the cluster")
	}
	return nil
}

func validateMDUsDeleted(ctx context.Context, vc clusterf.StateValidationConfig) error {
	mds, err := controller.GetMachineDeployments(ctx, vc.ManagementClusterClient, vc.ClusterSpec.Cluster)
	if err != nil {
		return fmt.Errorf("failed to retrieve machinedeployments: %s", err)
	}
	client := vc.ManagementClusterClient
	for _, md := range mds {
		mdu := &v1alpha1.MachineDeploymentUpgrade{}
		key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: md.Name + "-md-upgrade"}
		if err := client.Get(ctx, key, mdu); err != nil {
			if !apierrors1.IsNotFound(err) {
				return fmt.Errorf("failed to get MachineDeploymentUpgrade: %s", err)
			}
		}
		if mdu.Name != "" {
			return errors.New("MDUpgrade object not expected but still exists on the cluster")
		}
	}
	return nil
}

func validateNUsAndPodsDeleted(ctx context.Context, vc clusterf.StateValidationConfig) error {
	machines := &v1beta1.MachineList{}
	if err := vc.ManagementClusterClient.List(ctx, machines); err != nil {
		return fmt.Errorf("failed to list machines: %s", err)
	}
	client := vc.ManagementClusterClient
	clusterClient := vc.ClusterClient
	for _, machine := range machines.Items {
		nu := &v1alpha1.NodeUpgrade{}
		po := &corev1.Pod{}
		key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: machine.Name + "-node-upgrader"}
		if err := client.Get(ctx, key, nu); err != nil {
			if !apierrors1.IsNotFound(err) {
				return fmt.Errorf("failed to get NodeUpgrade: %s", err)
			}
		}
		if nu.Name != "" {
			return errors.New("NodeUpgrade object not expected, but still exists on the cluster")
		}
		if err := clusterClient.Get(ctx, key, po); err != nil {
			if !apierrors1.IsNotFound(err) {
				return fmt.Errorf("failed to get Upgrader Pod: %s", err)
			}
		}
		if po.Name != "" {
			return errors.New("Upgrader pod object not expected, but still exists on the cluster")
		}
	}
	return nil
}

// ValidateControlPlaneNodes retrieves the control plane nodes from the cluster and checks them against the cluster.Spec.
func ValidateControlPlaneNodes(ctx context.Context, vc clusterf.StateValidationConfig) error {
	clus := vc.ClusterSpec.Cluster
	cpNodes := &corev1.NodeList{}
	if err := vc.ClusterClient.List(ctx, cpNodes, client.MatchingLabels{"node-role.kubernetes.io/control-plane": ""}); err != nil {
		return fmt.Errorf("failed to list controlplane nodes %s", err)
	}
	cpConfig := clus.Spec.ControlPlaneConfiguration
	if len(cpNodes.Items) != cpConfig.Count {
		return fmt.Errorf("control plane node count does not match expected: %v of %v", len(cpNodes.Items), cpConfig.Count)
	}
	errorList := make([]error, 0)
	for _, node := range cpNodes.Items {
		if err := validateNodeReady(node, clus.Spec.KubernetesVersion); err != nil {
			errorList = append(errorList, fmt.Errorf("failed to validate controlplane node ready: %v", err))
		}
		if err := validateControlPlaneTaints(clus, node); err != nil {
			errorList = append(errorList, fmt.Errorf("failed to validate controlplane node taints: %v", err))
		}
	}
	if len(errorList) > 0 {
		return apierrors.NewAggregate(errorList)
	}
	return nil
}

// ValidateWorkerNodes retries the worker nodes from the cluster and checks them against the cluster.Spec.
func ValidateWorkerNodes(ctx context.Context, vc clusterf.StateValidationConfig) error {
	clus := vc.ClusterSpec.Cluster
	nodes := &corev1.NodeList{}
	if err := vc.ClusterClient.List(ctx, nodes); err != nil {
		return fmt.Errorf("failed to list nodes %s", err)
	}
	errorList := make([]error, 0)
	wn := clus.Spec.WorkerNodeGroupConfigurations
	// deduce the worker node group configuration to node mapping via the machine deployment and machine set
	for _, w := range wn {
		workerGroupCount := 0
		k8sVersion := vc.ClusterSpec.Cluster.Spec.KubernetesVersion
		if w.KubernetesVersion != nil {
			k8sVersion = *w.KubernetesVersion
		}
		ms, err := getWorkerNodeMachineSets(ctx, vc, w)
		if err != nil {
			return fmt.Errorf("failed to get machine sets when validating worker node: %v", err)
		}
		workerNodes := filterWorkerNodes(nodes.Items, ms, w)
		workerGroupCount += len(workerNodes)
		for _, node := range workerNodes {
			if err := validateNodeReady(node, k8sVersion); err != nil {
				errorList = append(errorList, fmt.Errorf("failed to validate worker node ready %v", err))
			}
			if err := api.ValidateWorkerNodeTaints(w, node); err != nil {
				errorList = append(errorList, fmt.Errorf("failed to validate worker node taints %v", err))
			}
		}
		if workerGroupCount != *w.Count {
			errorList = append(errorList, fmt.Errorf("worker node group %s count does not match expected: %d of %d", w.Name, workerGroupCount, *w.Count))
		}
	}

	if len(errorList) > 0 {
		return apierrors.NewAggregate(errorList)
	}
	return nil
}

// ValidateClusterDoesNotExist checks that the cluster does not exist by attempting to retrieve the CAPI cluster.
func ValidateClusterDoesNotExist(ctx context.Context, vc clusterf.StateValidationConfig) error {
	clus := vc.ClusterSpec.Cluster
	capiCluster, err := controller.GetCAPICluster(ctx, vc.ManagementClusterClient, clus)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}
	if capiCluster != nil {
		return fmt.Errorf("cluster %s exists", capiCluster.Name)
	}
	return nil
}

// ValidateCilium gets the cilium-config from the cluster and checks that the cilium
// policy in cluster.Spec matches the enabled policy in the config.
func ValidateCilium(ctx context.Context, vc clusterf.StateValidationConfig) error {
	cniConfig := vc.ClusterSpec.Cluster.Spec.ClusterNetwork.CNIConfig

	if cniConfig == nil || cniConfig.Cilium == nil {
		return errors.New("Cilium configuration missing from cluster spec")
	}

	if !cniConfig.Cilium.IsManaged() {
		// It would be nice if we could log something here given we're skipping the validation.
		return nil
	}

	clusterClient := vc.ClusterClient

	yaml := vc.ClusterSpec.Cluster
	cm := &corev1.ConfigMap{}
	key := types.NamespacedName{Namespace: "kube-system", Name: "cilium-config"}
	err := clusterClient.Get(ctx, key, cm)
	if err != nil {
		return fmt.Errorf("failed to retrieve configmap: %s", err)
	}

	clusterCilium := cm.Data["enable-policy"]
	yamlCilium := string(yaml.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode)
	if yamlCilium == "" && clusterCilium == "default" {
		return nil
	}
	if clusterCilium != yamlCilium {
		return fmt.Errorf("cilium policy does not match. ConfigMap: %s, YAML: %s", clusterCilium, yamlCilium)
	}

	return nil
}

func validateNodeReady(node corev1.Node, kubeVersion v1alpha1.KubernetesVersion) error {
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" && condition.Status != corev1.ConditionTrue {
			return fmt.Errorf("node %s not ready yet. %s", node.GetName(), condition.Reason)
		}
	}
	kubeletVersion := node.Status.NodeInfo.KubeletVersion
	if !strings.Contains(kubeletVersion, string(kubeVersion)) {
		return fmt.Errorf("validating node version: kubernetes version %s does not match expected version %s", kubeletVersion, kubeVersion)
	}
	return nil
}

func validateControlPlaneTaints(cluster *v1alpha1.Cluster, node corev1.Node) error {
	if cluster.IsSingleNode() {
		return api.ValidateControlPlaneNoTaints(cluster.Spec.ControlPlaneConfiguration, node)
	}

	return api.ValidateControlPlaneTaints(cluster.Spec.ControlPlaneConfiguration, node)
}

func filterWorkerNodes(nodes []corev1.Node, ms []v1beta1.MachineSet, w v1alpha1.WorkerNodeGroupConfiguration) []corev1.Node {
	wNodes := make([]corev1.Node, 0)
	for _, node := range nodes {
		ownerName, ok := node.Annotations["cluster.x-k8s.io/owner-name"]
		if ok {
			// there will be multiple machineSets present on a cluster following an upgrade.
			// find the one that is associated with this worker node, and execute the validations.
			for _, machineSet := range ms {
				if ownerName == machineSet.Name {
					wNodes = append(wNodes, node)
				}
			}
		}
	}
	return wNodes
}

func getWorkerNodeMachineSets(ctx context.Context, vc clusterf.StateValidationConfig, w v1alpha1.WorkerNodeGroupConfiguration) ([]v1beta1.MachineSet, error) {
	mdName := clusterapi.MachineDeploymentName(vc.ClusterSpec.Cluster, w)
	ms := &v1beta1.MachineSetList{}
	err := vc.ManagementClusterClient.List(ctx, ms, client.InNamespace(constants.EksaSystemNamespace), client.MatchingLabels{
		"cluster.x-k8s.io/deployment-name": mdName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get machine sets for deployment %s: %v", mdName, err)
	}
	if len(ms.Items) == 0 {
		return nil, fmt.Errorf("invalid number of machine sets associated with worker node configuration %s", w.Name)
	}
	return ms.Items, nil
}
