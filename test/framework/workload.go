package framework

type WorkloadCluster struct {
	*ClusterE2ETest
	managementClusterKubeconfigFile func() string
}

type WorkloadClusters map[string]*WorkloadCluster

func (w *WorkloadCluster) CreateCluster() {
	w.createCluster(nil, withKubeconfig(w.managementClusterKubeconfigFile()))
}

func (w *WorkloadCluster) CreateClusterWithVersion(opt VersionOpt) {
	w.createCluster(opt, withKubeconfig(w.managementClusterKubeconfigFile()))
}

func (w *WorkloadCluster) UpgradeCluster(opts ...ClusterE2ETestOpt) {
	w.upgradeCluster(nil, []commandOpt{withKubeconfig(w.managementClusterKubeconfigFile())}, opts...)
}

func (w *WorkloadCluster) UpgradeClusterWithVersion(opt VersionOpt, opts ...ClusterE2ETestOpt) {
	w.upgradeCluster(opt, []commandOpt{withKubeconfig(w.managementClusterKubeconfigFile())}, opts...)
}

func (w *WorkloadCluster) DeleteCluster() {
	w.deleteCluster(nil, withKubeconfig(w.managementClusterKubeconfigFile()))
}

func (w *WorkloadCluster) DeleteClusterWithVersion(opt VersionOpt) {
	w.deleteCluster(opt, withKubeconfig(w.managementClusterKubeconfigFile()))
}
