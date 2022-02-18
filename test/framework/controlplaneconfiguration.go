package framework

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type ControlPlaneValidation func(configuration v1alpha1.ControlPlaneConfiguration, node corev1.Node) (err error)

func (e *ClusterE2ETest) ValidateControlPlaneNodes(validations ...ControlPlaneValidation) {
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
