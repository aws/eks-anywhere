package framework

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type clusterValidation = func(ctx context.Context, validateOpts clusterOpts) error

type retriableValidation struct {
	validation    clusterValidation // the validation to run against the cluster.
	backoffPeriod time.Duration     // the duration to wait another validation attempt.
	maxRetries    int               // the maximum number of retries to validate.
}

// ClusterValidator is responsible for checking if a cluster is valid against the spec that is provided.
type ClusterValidator struct {
	ClusterOpts clusterOpts
	validations []retriableValidation
}

// WithValidation registers a validation to the ClusterValidator that will be run when Validate is called.
func (c *ClusterValidator) WithValidation(validation clusterValidation, backoffPeriod time.Duration, maxRetries int) {
	c.validations = append(c.validations, retriableValidation{validation, backoffPeriod, maxRetries})
}

// WithWorkloadClusterValidations registers a validation set for a management cluster.
func (c *ClusterValidator) WithWorkloadClusterValidations() {
	c.WithValidation(validateClusterReady, 5*time.Second, 60)
	c.WithValidation(validateEKSAObjects, 5*time.Second, 60)
	c.WithValidation(validateControlPlaneNodes, 5*time.Second, 60)
}

// WithClusterDoesNotExist registers a validation to check that a cluster does not exist or has been deleted.
func (c *ClusterValidator) WithClusterDoesNotExist() {
	c.WithValidation(validateClusterDoesNotExist, 5*time.Second, 60)
}

// Validate runs through the set registered validations and returns an error if any of them fail after a number of retries.
func (c *ClusterValidator) Validate(ctx context.Context) error {
	for _, v := range c.validations {
		err := retrier.Retry(v.maxRetries, v.backoffPeriod, func() error {
			return v.validation(ctx, c.ClusterOpts)
		})
		if err != nil {
			return fmt.Errorf("validation faild %v", err)
		}
	}

	return nil
}

// Reset clears the registered validations.
func (c *ClusterValidator) Reset() {
	c.validations = []retriableValidation{}
}

// ClusterValidatorOpts represent the data to be passed as an argument to the Validate method.
type ClusterValidatorOpts = func(cv *ClusterValidator)

// NewClusterValidator returns a cluster validator.
func NewClusterValidator(opts ...ClusterValidatorOpts) *ClusterValidator {
	cv := ClusterValidator{
		ClusterOpts: clusterOpts{},
		validations: []retriableValidation{},
	}

	for _, opt := range opts {
		opt(&cv)
	}
	return &cv
}

type clusterOpts struct {
	ClusterClient          client.Client // the client for the cluster
	ManagmentClusterClient client.Client // the client for the management cluster
	ClusterSpec            *cluster.Spec // the cluster spec
}

func validateEKSAObjects(ctx context.Context, validateOpts clusterOpts) error {
	for _, obj := range validateOpts.ClusterSpec.ChildObjects() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		key := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}

		if key.Namespace == "" {
			key.Namespace = "default"
		}

		err := validateOpts.ManagmentClusterClient.Get(ctx, key, u)
		if err != nil {
			return fmt.Errorf("cluster object does not exist %s", err)
		}
	}

	return nil
}

func validateClusterReady(ctx context.Context, validateOpts clusterOpts) error {
	capiCluster, err := controller.GetCAPICluster(ctx, validateOpts.ManagmentClusterClient, validateOpts.ClusterSpec.Cluster)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}

	if capiCluster == nil && err == nil {
		return fmt.Errorf("cluster %s does not exist", capiCluster.Name)
	}

	if err := checkClusterReady(capiCluster); err != nil {
		return err
	}
	return nil
}

func validateControlPlaneNodes(ctx context.Context, validateOpts clusterOpts) error {
	cpNodes := &corev1.NodeList{}
	err := validateOpts.ClusterClient.List(ctx, cpNodes, client.MatchingLabels{"node-role.kubernetes.io/control-plane": ""})
	if err != nil {
		return fmt.Errorf("failed to list controlplane nodes %s", err)
	}

	cpConfig := validateOpts.ClusterSpec.Cluster.Spec.ControlPlaneConfiguration
	if len(cpNodes.Items) != cpConfig.Count {
		return fmt.Errorf("control plane node count does not match expected: %v of %v", len(cpNodes.Items), cpConfig.Count)
	}

	for _, node := range cpNodes.Items {
		err := validateNodeReady(node)
		if err != nil {
			return fmt.Errorf("failed to validate controlplane %s", err)
		}
	}
	return nil
}

func validateNodeReady(node corev1.Node) error {
	for _, condition := range node.Status.Conditions {
		if condition.Type == "Ready" && condition.Status != "True" {
			return fmt.Errorf("node %s not ready yet. %s", node.GetName(), condition.Reason)
		}
	}

	return nil
}

func validateClusterDoesNotExist(ctx context.Context, validateOpts clusterOpts) error {
	capiCluster, err := controller.GetCAPICluster(ctx, validateOpts.ManagmentClusterClient, validateOpts.ClusterSpec.Cluster)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster %s", err)
	}

	if capiCluster != nil {
		return fmt.Errorf("cluster %s exists", capiCluster.Name)
	}

	return nil
}

func checkClusterReady(cluster *v1beta1.Cluster) error {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == "Ready" && condition.Status != "True" {
			return fmt.Errorf("node %s not ready yet. %s", cluster.GetName(), condition.Reason)
		}
	}

	return nil
}
