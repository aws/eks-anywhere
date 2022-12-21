package snow

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

type (
	// BaseWorkers represents the Snow specific CAPI spec for worker nodes.
	BaseWorkers     = clusterapi.Workers[*snowv1.AWSSnowMachineTemplate]
	baseWorkerGroup = clusterapi.WorkerGroup[*snowv1.AWSSnowMachineTemplate]
)

// Workers holds the Snow specific objects for CAPI snow worker groups.
type Workers struct {
	BaseWorkers
	CAPASIPPools CAPASIPPools
}

// Objects returns the worker nodes objects associated with the snow cluster.
func (w Workers) Objects() []kubernetes.Object {
	o := w.BaseWorkers.WorkerObjects()
	for _, p := range w.CAPASIPPools {
		o = append(o, p)
	}
	return o
}

// WorkersSpec generates a Snow specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, log logr.Logger, spec *cluster.Spec, client kubernetes.Client) (*Workers, error) {
	w := &Workers{
		BaseWorkers: BaseWorkers{
			Groups: make([]baseWorkerGroup, 0, len(spec.Cluster.Spec.WorkerNodeGroupConfigurations)),
		},
	}

	capasPools := CAPASIPPools{}

	for _, wc := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		machineConfig := spec.SnowMachineConfig(wc.MachineGroupRef.Name)

		capasPools.addPools(machineConfig.Spec.Network.DirectNetworkInterfaces, spec.SnowIPPools)

		machineTemplate := MachineTemplate(clusterapi.WorkerMachineTemplateName(spec, wc), spec.SnowMachineConfigs[wc.MachineGroupRef.Name], capasPools)

		kubeadmConfigTemplate, err := KubeadmConfigTemplate(log, spec, wc)
		if err != nil {
			return nil, err
		}

		machineDeployment := machineDeployment(spec, wc, kubeadmConfigTemplate, machineTemplate)

		w.Groups = append(w.Groups, baseWorkerGroup{
			MachineDeployment:       machineDeployment,
			KubeadmConfigTemplate:   kubeadmConfigTemplate,
			ProviderMachineTemplate: machineTemplate,
		})
	}

	if err := w.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, MachineTemplateDeepDerivative); err != nil {
		return nil, errors.Wrap(err, "updating snow worker immutable object names")
	}

	w.CAPASIPPools = capasPools

	return w, nil
}
