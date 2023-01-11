package framework

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

// TODO: Move ClusterValidator to a separate package

// ClusterValidation defines a validation that can be registered to the ClusterValidator.
type ClusterValidation = func(ctx context.Context, vc ClusterValidatorConfig) error

type retriableValidation struct {
	validation    ClusterValidation // the validation to run against the cluster.
	backoffPeriod time.Duration     // the duration to wait another validation attempt.
	maxRetries    int               // the maximum number of retries to validate.
}

// ClusterValidator is responsible for checking if a cluster is valid against the spec that is provided.
type ClusterValidator struct {
	Config      ClusterValidatorConfig
	validations []retriableValidation
}

// WithValidation registers a validation to the ClusterValidator that will be run when Validate is called.
func (c *ClusterValidator) WithValidation(validation ClusterValidation, backoffPeriod time.Duration, maxRetries int) {
	c.validations = append(c.validations, retriableValidation{validation, backoffPeriod, maxRetries})
}

// WithWorkloadClusterValidations registers a validation set for a workload cluster.
func (c *ClusterValidator) WithWorkloadClusterValidations() {}

// WithExpectedObjectsExist registers a set of validations for the existence of various cluster objects.
func (c *ClusterValidator) WithExpectedObjectsExist() {
	c.WithValidation(validateClusterReady, 5*time.Second, 60)
	c.WithValidation(validateEKSAObjects, 5*time.Second, 60)
	c.WithValidation(validateControlPlaneNodes, 5*time.Second, 120)
	c.WithValidation(validateWorkerNodes, 5*time.Second, 120)
}

// WithClusterDoesNotExist registers a validation to check that a cluster does not exist or has been deleted.
func (c *ClusterValidator) WithClusterDoesNotExist() {
	c.WithValidation(validateClusterDoesNotExist, 5*time.Second, 60)
}

// Validate runs through the set registered validations and returns an error if any of them fail after a number of retries.
func (c *ClusterValidator) Validate(ctx context.Context) error {
	for _, v := range c.validations {
		err := retrier.Retry(v.maxRetries, v.backoffPeriod, func() error {
			return v.validation(ctx, c.Config)
		})
		if err != nil {
			return fmt.Errorf("validation failed %v", err)
		}
	}

	return nil
}

// Reset clears the registered validations.
func (c *ClusterValidator) Reset() {
	c.validations = []retriableValidation{}
}

// ClusterValidatorOpt is a function that receives a ClusterValidator to be customized.
type ClusterValidatorOpt = func(cv *ClusterValidator)

// NewClusterValidator returns a cluster validator which can be configured by passing ClusterValidatorOpt arguments.
func NewClusterValidator(opts ...ClusterValidatorOpt) *ClusterValidator {
	cv := ClusterValidator{
		Config:      ClusterValidatorConfig{},
		validations: []retriableValidation{},
	}

	for _, opt := range opts {
		opt(&cv)
	}
	return &cv
}

// ClusterValidatorConfig holds the data required to perform validations on a cluster.
type ClusterValidatorConfig struct {
	ClusterClient           client.Client // the client for the cluster
	ManagementClusterClient client.Client // the client for the management cluster
	ClusterSpec             *cluster.Spec // the cluster spec
}

// TODO: Move Validations to separate package

func validateEKSAObjects(ctx context.Context, vc ClusterValidatorConfig) error {
	clus := vc.ClusterSpec.Cluster
	mgmtClusterClient := vc.ManagementClusterClient
	for _, obj := range vc.ClusterSpec.ChildObjects() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		key := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}

		if key.Namespace == "" {
			key.Namespace = "default"
		}

		if err := mgmtClusterClient.Get(ctx, key, u); err != nil {
			return fmt.Errorf("cluster object does not exist %s", err)
		}
	}
	logger.V(4).Info("EKSA objects exists validated", "cluster", clus.Name)
	return nil
}

func validateClusterReady(ctx context.Context, vc ClusterValidatorConfig) error {
	clus := vc.ClusterSpec.Cluster
	mgmtClusterClient := vc.ManagementClusterClient
	capiCluster, err := controller.GetCAPICluster(ctx, mgmtClusterClient, clus)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}
	if capiCluster == nil {
		return fmt.Errorf("cluster %s does not exist", clus.Name)
	}
	for _, condition := range capiCluster.GetConditions() {
		if condition.Type == "Ready" && condition.Status != "True" {
			return fmt.Errorf("node %s not ready yet. %s", capiCluster.GetName(), condition.Reason)
		}
	}
	logger.V(4).Info("CAPI cluster ready validated", "cluster", clus.Name)
	return nil
}

func validateControlPlaneNodes(ctx context.Context, vc ClusterValidatorConfig) error {
	clus := vc.ClusterSpec.Cluster
	cpNodes := &corev1.NodeList{}
	if err := vc.ClusterClient.List(ctx, cpNodes, client.MatchingLabels{"node-role.kubernetes.io/control-plane": ""}); err != nil {
		return fmt.Errorf("failed to list controlplane nodes %s", err)
	}

	cpConfig := clus.Spec.ControlPlaneConfiguration
	if len(cpNodes.Items) != cpConfig.Count {
		return fmt.Errorf("control plane node count does not match expected: %v of %v", len(cpNodes.Items), cpConfig.Count)
	}

	for _, node := range cpNodes.Items {
		if err := validateNodeReady(node, clus.Spec.KubernetesVersion); err != nil {
			return fmt.Errorf("failed to validate controlplane %s", err)
		}
	}
	logger.V(4).Info("Control plane nodes validated", "cluster", clus.Name)
	return nil
}

func validateWorkerNodes(ctx context.Context, vc ClusterValidatorConfig) error {
	clus := vc.ClusterSpec.Cluster
	clusterName := clus.Name
	nodes := &corev1.NodeList{}
	if err := vc.ClusterClient.List(ctx, nodes); err != nil {
		return fmt.Errorf("failed to list nodes %s", err)
	}
	wn := clus.Spec.WorkerNodeGroupConfigurations
	// deduce the worker node group configuration to node mapping via the machine deployment and machine set
	for _, w := range wn {
		workerGroupCount := 0
		ms, err := getWorkerNodeMachineSets(ctx, vc, w)
		if err != nil {
			return fmt.Errorf("failed to get machine sets when validating worker node: %v", err)
		}
		for _, node := range nodes.Items {
			ownerName, ok := node.Annotations["cluster.x-k8s.io/owner-name"]
			if ok {
				// there will be multiple machineSets present on a cluster following an upgrade.
				// find the one that is associated with this worker node, and execute the validations.
				for _, machineSet := range ms {
					if ownerName == machineSet.Name {
						workerGroupCount++
						if err := validateNodeReady(node, vc.ClusterSpec.Cluster.Spec.KubernetesVersion); err != nil {
							return fmt.Errorf("failed to validate worker node ready %v", err)
						}
					}
				}
			}
		}
		if workerGroupCount != *w.Count {
			return fmt.Errorf("worker node group %s count does not match expected: %d of %d", w.Name, workerGroupCount, *w.Count)
		}
	}
	logger.V(4).Info("Worker nodes validated", "cluster", clusterName)
	return nil
}

func validateNodeReady(node corev1.Node, kubeVersion v1alpha1.KubernetesVersion) error {
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" && condition.Status != "True" {
			return fmt.Errorf("node %s not ready yet. %s", node.GetName(), condition.Reason)
		}
	}
	kubeletVersion := node.Status.NodeInfo.KubeletVersion
	if !strings.Contains(kubeletVersion, string(kubeVersion)) {
		return fmt.Errorf("validating node version: kubernetes version %s does not match expected version %s", kubeletVersion, kubeVersion)
	}
	return nil
}

func validateClusterDoesNotExist(ctx context.Context, vc ClusterValidatorConfig) error {
	clus := vc.ClusterSpec.Cluster
	capiCluster, err := controller.GetCAPICluster(ctx, vc.ManagementClusterClient, clus)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}
	if capiCluster != nil {
		return fmt.Errorf("cluster %s exists", capiCluster.Name)
	}
	logger.V(4).Info("Cluster does not exist validated", "cluster", clus.Name)
	return nil
}

// getWorkerNodeMachineSets gets a list of MachineSets corresponding the provided WorkerNodeGroupConfiguration from the management cluster.
func getWorkerNodeMachineSets(ctx context.Context, vc ClusterValidatorConfig, w v1alpha1.WorkerNodeGroupConfiguration) ([]v1beta1.MachineSet, error) {
	md := &v1beta1.MachineDeployment{}
	mdName := fmt.Sprintf("%s-%s", vc.ClusterSpec.Cluster.Name, w.Name)
	key := types.NamespacedName{Name: mdName, Namespace: constants.EksaSystemNamespace}
	if err := vc.ManagementClusterClient.Get(ctx, key, md); err != nil {
		return nil, fmt.Errorf("failed to get machine deployment %s when validating worker node: %v", w.Name, err)
	}
	ms := &v1beta1.MachineSetList{}
	err := vc.ManagementClusterClient.List(ctx, ms, client.InNamespace(constants.EksaSystemNamespace), client.MatchingLabels{
		"cluster.x-k8s.io/deployment-name": md.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get machine sets for deployment %s: %v", md.Name, err)
	}
	if len(ms.Items) == 0 {
		return nil, fmt.Errorf("invalid number of machine sets associated with worker node configuration %s", w.Name)
	}
	return ms.Items, nil
}
