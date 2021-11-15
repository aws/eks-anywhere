package controllers

import (
	"context"
	"fmt"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	MachineGroupLabelName        = "cluster.anywhere.eks.amazonaws.com/machinegroup-name"
	MachineDeploymentLabelType   = "cluster.anywhere.eks.amazonaws.com/machinedeployment-type"
	MachineDeploymentWorkersType = "workers"
)

func (r *ClusterReconcilerV2) reconcileWorkers(ctx context.Context, cluster *anywhere.Cluster) error {
	// TODO: make this support upgrades, this only supports creation, once
	//  Most of the comments in control plane reconcile method apply here as well
	// Compare objects one by one, individual structs, caching phase in the status ...

	// Check if there any MachineDeployment maching the worker node config and the cluster
	machineGroups := 0
	for _, w := range cluster.Spec.WorkerNodeGroupConfigurations {
		// TODO: this is terrible. hacking it this way bc docker doesn't have machine group configs so we can't use the name here
		machineGroupName := ""
		if w.MachineGroupRef == nil {
			machineGroupName = fmt.Sprintf("%s-machine-group-%d", cluster.Name, machineGroups)
		} else {
			machineGroupName = w.MachineGroupRef.Name
		}

		machineDeployments, err := r.Client.GetMachineDeploymentsForMachineGroup(ctx, cluster, machineGroupName)
		if err != nil {
			return err
		}

		if machineDeployments != nil && len(machineDeployments.Items) > 0 {
			// There is already at least one machine deployment, we already created the workers objects
			r.Log.Info("Found at least one MachineDeployment maching cluster and worker node group config, skipping reconcile")
			return nil
		}
	}

	return r.buildProviderReconciler(cluster).ReconcileWorkers(ctx, cluster)
}

func machineDeploymentReady(machineDeployment *capi.MachineDeployment) bool {
	status := machineDeployment.Status
	return status.Phase == "Running" && status.UnavailableReplicas == 0 && status.ReadyReplicas == status.Replicas
}
