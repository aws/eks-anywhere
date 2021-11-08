package framework

type WorkloadCluster struct {
	*ClusterE2ETest
	managementClusterKubeconfigFile func() string
}

type WorkloadClusters map[string]*WorkloadCluster

func (w *WorkloadCluster) CreateCluster() {
	w.createCluster(withKubeconfig(w.managementClusterKubeconfigFile()))
}

func (w *WorkloadCluster) UpgradeCluster(opts ...ClusterE2ETestOpt) {
	w.upgradeCluster([]commandOpt{withKubeconfig(w.managementClusterKubeconfigFile())}, opts...)
}

func (w *WorkloadCluster) DeleteCluster() {
	w.deleteCluster(withKubeconfig(w.managementClusterKubeconfigFile()))
}
