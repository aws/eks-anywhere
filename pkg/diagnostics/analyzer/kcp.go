package analyzer

import (
	"context"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

func newKCP(name, namespace string) kcp {
	return kcp{
		name:      name,
		namespace: namespace,
	}
}

type kcp struct {
	name, namespace string
}

func (k kcp) analyze(ctx context.Context, readers readers) (*analyzeResults, error) {
	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := readers.client.Get(ctx, k.name, k.namespace, kcp); err != nil {
		return nil, err
	}

	readyCondition := conditions.Get(kcp, clusterv1.ReadyCondition)

	if isTrue(readyCondition) {
		return nil, nil
	}

	clusterName := kcp.Labels[clusterv1.ClusterNameLabel]
	machines := &clusterv1.MachineList{}
	cpMachinesFilter := kubernetes.ListOptions{LabelSelector: collections.ControlPlaneSelectorForCluster(clusterName)}
	if err := readers.client.List(ctx, machines, cpMachinesFilter); err != nil {
		return nil, err
	}

	r := &analyzeResults{
		Finding: Finding{
			Severity: SeverityWarning,
			Message:  resourceStatusMessage(SeverityWarning, "KubeadmControlPlane", k.name, k.namespace, "not ready", readyCondition.Message),
		},
	}

	for _, machine := range machines.Items {
		r.nextAnalyzers = append(r.nextAnalyzers, newMachine(machine.Name, machine.Namespace))
	}

	return r, nil
}
