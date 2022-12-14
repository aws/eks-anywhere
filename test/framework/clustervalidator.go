package framework

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterValidation = func(client client.Client, config *cluster.Config) error

type validationOpt struct {
	timeout  time.Duration
	validate clusterValidation
}

// ClusterValidator is responsible for checking if a cluster is valid against the cluster config that is provided.
type ClusterValidator struct {
	clusterConfig *cluster.Config
	client        client.Client
	validations   []validationOpt
}

// NewDefaultClusterValidator returns a cluster validator pre-loaded with validations by default.
func NewDefaultClusterValidator(client client.Client, config *cluster.Config) ClusterValidator {
	cv := ClusterValidator{
		clusterConfig: config,
		client:        client,
	}

	cv.withValidation(validateObjects, 5*time.Minute)
	cv.withValidation(validateControlPlaneNodes, 5*time.Minute)
	return cv
}

func validateObjects(c client.Client, clusterConfig *cluster.Config) error {
	ctx := context.Background()

	for _, obj := range clusterConfig.ChildObjects() {
		u := &unstructured.Unstructured{}
		err := c.Get(ctx, client.ObjectKeyFromObject(obj), u)

		if err != nil {
			return fmt.Errorf("cluster object does not exist %s", err)
		}
	}

	return nil
}

func validateControlPlaneNodes(c client.Client, clusterConfig *cluster.Config) error {
	ctx := context.Background()
	cpNodes := &corev1.NodeList{}
	_ = c.List(ctx, cpNodes, client.MatchingLabels{"node-role.kubernetes.io/control-plane": ""})

	cpConfig := clusterConfig.Cluster.Spec.ControlPlaneConfiguration
	if len(cpNodes.Items) != cpConfig.Count {
		return fmt.Errorf("control plane node count does not match expected: %v of %v", len(cpNodes.Items), cpConfig.Count)
	}

	// TODO: Validate contents of each node.
	return nil
}

func (c *ClusterValidator) withValidation(validate clusterValidation, timeout time.Duration) {
	c.validations = append(c.validations, validationOpt{
		timeout:  timeout,
		validate: validate,
	})
}

// ValidateCluster runs through the set clusterValidations returns an error if any of them fail after a number of retries.
func (c *ClusterValidator) ValidateCluster() error {
	for _, validation := range c.validations {
		err := retrier.Retry(10, validation.timeout, func() error {
			return validation.validate(c.client, c.clusterConfig)
		})

		if err != nil {
			return fmt.Errorf("validation faild %v", err)
		}

	}

	return nil
}
