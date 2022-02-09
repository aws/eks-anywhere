package framework

type WorkloadCluster struct {
	*ClusterE2ETest
	managementClusterKubeconfigFile func() string
}

type WorkloadClusters map[string]*WorkloadCluster

func (w *WorkloadCluster) CreateCluster(opts ...CommandOpt) {
	opts = append(opts, withKubeconfig(w.managementClusterKubeconfigFile()))
	w.createCluster(opts...)
}

func (w *WorkloadCluster) UpgradeCluster(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	commandOpts = append(commandOpts, withKubeconfig(w.managementClusterKubeconfigFile()))
	w.upgradeCluster(clusterOpts, commandOpts...)
}

func (w *WorkloadCluster) DeleteCluster(opts ...CommandOpt) {
	opts = append(opts, withKubeconfig(w.managementClusterKubeconfigFile()))
	w.deleteCluster(opts...)
}

func (w *WorkloadCluster) UpgradeClusterWithGitOps(clusterOpts []ClusterE2ETestOpt) {
	w.UpgradeWithGitOps(clusterOpts)
}
