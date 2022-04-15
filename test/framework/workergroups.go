package framework

import "github.com/aws/eks-anywhere/internal/pkg/api"

type WorkerNodeGroup struct {
	Name                                 string
	Fillers                              []api.WorkerNodeGroupFiller
	MachineConfigKind, MachineConfigName string
}

func WithWorkerNodeGroup(name string, fillers ...api.WorkerNodeGroupFiller) *WorkerNodeGroup {
	return &WorkerNodeGroup{
		Name:    name,
		Fillers: fillers,
	}
}

func (w *WorkerNodeGroup) ClusterFiller() api.ClusterFiller {
	wf := make([]api.WorkerNodeGroupFiller, 0, len(w.Fillers)+1)
	wf = append(wf, api.WithMachineGroupRef(w.MachineConfigName, w.MachineConfigKind))
	wf = append(wf, w.Fillers...)

	return api.WithWorkerNodeGroup(w.Name, wf...)
}
