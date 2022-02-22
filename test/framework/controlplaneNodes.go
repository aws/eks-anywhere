package framework

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ControlPlaneNodeValidation should return an error if either an error is encountered during execution or the validation logically fails.
// This validation function will be executed by ValidateControlPlaneNodes with a Control Plane configuration and a corresponding node
// which was created as a part of that configuration.
type ControlPlaneNodeValidation func(configuration v1alpha1.ControlPlaneConfiguration, node corev1.Node) (err error)

// ValidateControlPlaneNodes deduces the control plane configuration to node mapping
// and for each configuration/node pair executes the provided validation functions.
func (e *ClusterE2ETest) ValidateControlPlaneNodes(validations ...ControlPlaneNodeValidation) {
	ctx := context.Background()
	c, err := v1alpha1.GetClusterConfigFromContent(e.ClusterConfigB)
	if err != nil {
		e.T.Fatal(err)
	}

	cpNodes, err := e.KubectlClient.GetControlPlaneNodes(ctx, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}

	for _, node := range cpNodes {
		for _, validation := range validations {
			err = validation(c.Spec.ControlPlaneConfiguration, node)
			if err != nil {
				e.T.Errorf("Control plane node %v is not valid: %v", node.Name, err)
			}
		}
	}
	e.StopIfFailed()
}
