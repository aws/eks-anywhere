package framework

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type clusterValidation = func(ctx context.Context, client client.Client, spec *cluster.Spec) error

type retriableValidation struct {
	validation    clusterValidation // the validation to run against the cluster.
	backoffPeriod time.Duration     // the duration to wait another validation attempt.
	maxRetries    int               // the maximum number of retries to validate.
}

// ClusterValidator is responsible for checking if a cluster is valid against the clusterSpec that is provided.
type ClusterValidator struct {
	clusterSpec *cluster.Spec
	client      client.Client
	validations []retriableValidation
}

// WithValidation registers a validation to the ClusterValidator that will be run when Validate is called.
func (c *ClusterValidator) WithValidation(validation clusterValidation, backoffPeriod time.Duration, maxRetries int) {
	c.validations = append(c.validations, retriableValidation{validation, backoffPeriod, maxRetries})
}

// Validate runs through the set clusterValidations returns an error if any of them fail after a number of retries.
func (c *ClusterValidator) Validate(ctx context.Context) error {
	for _, v := range c.validations {
		err := retrier.Retry(v.maxRetries, v.backoffPeriod, func() error {
			return v.validation(ctx, c.client, c.clusterSpec)
		})
		if err != nil {
			return fmt.Errorf("validation faild %v", err)
		}

	}

	return nil
}

// NewClusterValidator returns a cluster validator.
func NewClusterValidator(client client.Client, spec *cluster.Spec) ClusterValidator {
	cv := ClusterValidator{
		clusterSpec: spec,
		client:      client,
	}

	return cv
}

// NewFullClusterValidator returns a cluster with a complete validation set pre-loaded.
func NewFullClusterValidator(client client.Client, spec *cluster.Spec) ClusterValidator {
	cv := NewClusterValidator(client, spec)
	cv.WithValidation(validateObjects, 5*time.Second, 60)
	cv.WithValidation(validateControlPlaneNodes, 5*time.Second, 60)
	return cv
}

func validateObjects(ctx context.Context, c client.Client, clusterSpec *cluster.Spec) error {
	for _, obj := range clusterSpec.ChildObjects() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		err := c.Get(ctx, client.ObjectKeyFromObject(obj), u)
		if err != nil {
			return fmt.Errorf("cluster object does not exist %s", err)
		}
	}

	return nil
}

func validateControlPlaneNodes(ctx context.Context, c client.Client, clusterSpec *cluster.Spec) error {
	cpNodes := &corev1.NodeList{}
	_ = c.List(ctx, cpNodes, client.MatchingLabels{"node-role.kubernetes.io/control-plane": ""})

	cpConfig := clusterSpec.Cluster.Spec.ControlPlaneConfiguration
	if len(cpNodes.Items) != cpConfig.Count {
		return fmt.Errorf("control plane node count does not match expected: %v of %v", len(cpNodes.Items), cpConfig.Count)
	}

	// TODO: Validate each node according to different criteria.
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
