package framework

type WorkloadCluster struct {
	*ClusterE2ETest
	managementClusterKubeconfigFile func() string
}

type WorkloadClusters map[string]*WorkloadCluster

func (w *WorkloadCluster) CreateCluster() {
	w.createCluster(nil, withKubeconfig(w.managementClusterKubeconfigFile()))
}

func (w *WorkloadCluster) UpgradeCluster(opts ...ClusterE2ETestOpt) {
	w.upgradeCluster(opts, []CommandOpt{withKubeconfig(w.managementClusterKubeconfigFile())}...)
}

func (w *WorkloadCluster) DeleteCluster() {
	w.deleteCluster(nil, withKubeconfig(w.managementClusterKubeconfigFile()))
}
