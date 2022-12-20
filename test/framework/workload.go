package framework

type WorkloadCluster struct {
	*ClusterE2ETest
	ManagementClusterKubeconfigFile func() string
}

type WorkloadClusters map[string]*WorkloadCluster

func (w *WorkloadCluster) CreateCluster(opts ...CommandOpt) {
	opts = append(opts, withKubeconfig(w.ManagementClusterKubeconfigFile()))
	w.createCluster(opts...)
}

func (w *WorkloadCluster) UpgradeCluster(clusterOpts []ClusterE2ETestOpt, commandOpts ...CommandOpt) {
	commandOpts = append(commandOpts, withKubeconfig(w.ManagementClusterKubeconfigFile()))
	w.upgradeCluster(clusterOpts, commandOpts...)
}

func (w *WorkloadCluster) DeleteCluster(opts ...CommandOpt) {
	opts = append(opts, withKubeconfig(w.ManagementClusterKubeconfigFile()))
	w.deleteCluster(opts...)
}
