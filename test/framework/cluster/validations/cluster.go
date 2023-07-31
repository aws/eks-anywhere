package validations

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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

// ValidateClusterReady gets the CAPICluster from the client then validates that it is in a ready state.
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
		ms, err := getWorkerNodeMachineSets(ctx, vc, w)
		if err != nil {
			return fmt.Errorf("failed to get machine sets when validating worker node: %v", err)
		}
		workerNodes := filterWorkerNodes(nodes.Items, ms, w)
		workerGroupCount += len(workerNodes)
		for _, node := range workerNodes {
			if err := validateNodeReady(node, vc.ClusterSpec.Cluster.Spec.KubernetesVersion); err != nil {
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
