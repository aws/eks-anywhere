package framework

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
)

// WorkerNodeValidation should return an error if either an error is encountered during execution or the validation logically fails.
// This validation function will be executed by ValidateWorkerNodes with a worker node group configuration and a corresponding node
// which was created as a part of that worker node group configuration.
type WorkerNodeValidation func(configuration v1alpha1.WorkerNodeGroupConfiguration, node corev1.Node) (err error)

// ValidateWorkerNodes deduces the worker node group configuration to node mapping
// and for each configuration/node pair executes the provided validation functions.
func (e *ClusterE2ETest) ValidateWorkerNodes(workerNodeValidations ...WorkerNodeValidation) {
	ctx := context.Background()
	nodes, err := e.KubectlClient.GetNodes(ctx, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}

	c, err := v1alpha1.GetClusterConfigFromContent(e.ClusterConfigB)
	if err != nil {
		e.T.Fatal(err)
	}
	wn := c.Spec.WorkerNodeGroupConfigurations
	// deduce the worker node group configuration to node mapping via the machine deployment and machine set
	for _, w := range wn {
		mdName := fmt.Sprintf("%v-%v", e.ClusterName, w.Name)
		md, err := e.KubectlClient.GetMachineDeployment(ctx, mdName, executables.WithKubeconfig(e.cluster().KubeconfigFile), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get machine deployment for worker node %s when validating taints: %v", w.Name, err))
		}
		ms, err := e.KubectlClient.GetMachineSets(ctx, md.Name, e.cluster())
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get machine sets when validating taints: %v", err))
		}
		if len(ms) == 0 {
			e.T.Fatal(fmt.Errorf("invalid number of machine sets associated with worker node configuration %v", w.Name))
		}

		for _, node := range nodes {
			ownerName, ok := node.Annotations[ownerAnnotation]
			if ok {
				// there will be multiple machineSets present on a cluster following an upgrade.
				// find the one that is associated with this worker node, and execute the validations.
				for _, machineSet := range ms {
					if ownerName == machineSet.Name {
						for _, validation := range workerNodeValidations {
							err = validation(w, node)
							if err != nil {
								e.T.Errorf("Worker node %v, member of Worker Node Group Configuration %v, is not valid: %v", node.Name, w.Name, err)
							}
						}
					}
				}
			}
		}
		e.StopIfFailed()
	}
}
