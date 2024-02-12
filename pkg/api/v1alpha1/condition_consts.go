package v1alpha1

// Conditions, condition reasons, and messages for the Cluster object.
const (
	// ReadyCondition reports a summary of other conditions, indicating an overall operational
	// state of the cluster: all control plane and worker nodes are the right version,
	// all nodes are ready, not including old nodes.
	ReadyCondition ConditionType = "Ready"

	// OutdatedInformationReason reports the system is waiting for stale cluster information to be refreshed.
	OutdatedInformationReason = "OutdatedInformation"

	// ControlPlaneReadyCondition reports the status on the control plane nodes, indicating all those control plane
	// nodes are the right version and are ready, not including the old nodes.
	ControlPlaneReadyCondition ConditionType = "ControlPlaneReady"

	// ControlPlaneInitializedCondition reports that the first control plane instance has been initialized
	// and so the control plane is available and an API server instance is ready for processing requests.
	ControlPlaneInitializedCondition ConditionType = "ControlPlaneInitialized"

	// ControlPlaneInitializationInProgressReason reports that the control plane initilization is in progress.
	ControlPlaneInitializationInProgressReason = "ControlPlaneInitializationInProgress"

	// ControlPlaneNotReadyReason reports that the control plane is not ready.
	ControlPlaneNotReadyReason = "ControlPlaneNotReady"

	// ControlPlaneNotInitializedReason reports that the control plane is not initialized.
	ControlPlaneNotInitializedReason = "ControlPlaneNotInitialized"

	// WorkersReadyCondition reports the status on the worker nodes, indicating all those worker nodes
	// are the right version and are ready, not including the old nodes.
	WorkersReadyCondition ConditionType = "WorkersReady"
)

const (
	// NodesNotReadyReason reports the Cluster has some nodes that are not ready.
	NodesNotReadyReason = "NodesNotReady"

	// ControlPlaneComponentsUnhealthyReason reports that the Cluster control plane components are unhealthy.
	ControlPlaneComponentsUnhealthyReason = "ControlPlaneComponentsUnhealthy"

	// ScalingUpReason reports the Cluster is increasing the number of replicas for a set of nodes.
	ScalingUpReason = "ScalingUp"

	// ScalingDownReason reports the Cluster is decreasing the number of replicas for a set of nodes.
	ScalingDownReason = "ScalingDown"

	// RollingUpgradeInProgress reports the Cluster is executing a rolling upgrading to align the nodes to
	// a new desired machine spec.
	RollingUpgradeInProgress = "RollingUpgradeInProgress"

	// InPlaceUpgradeInProgress reports the Cluster is executing a in place upgrade to align the nodes to
	// a new desired machine specs.
	InPlaceUpgradeInProgress = "InPlaceUpgradeInProgress"

	// ExternalEtcdNotAvailable reports the Cluster status is waiting for Etcd to be available.
	ExternalEtcdNotAvailable = "ExternalEtcdNotAvailable"
)

const (
	// DefaultCNIConfiguredCondition reports the default cni cluster has been configured successfully.
	DefaultCNIConfiguredCondition ConditionType = "DefaultCNIConfigured"

	// DefaultCNIUpgradeInProgressReason used when cluster is upgrading the default CNI.
	DefaultCNIUpgradeInProgressReason = "DefaultCNIUpgradeInProgressReason"

	// SkipUpgradesForDefaultCNIConfiguredReason used to indicate the custer has been configured to skip
	// upgrades for the default cni. The default cni may still be installed, for example to successfully
	// create a cluster.
	SkipUpgradesForDefaultCNIConfiguredReason = "SkipUpgradesForDefaultCNIConfigured"
)
