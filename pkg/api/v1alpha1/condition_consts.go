package v1alpha1

// Conditions and condition Reasons for the Cluster object.
const (
	// ReadyCondition reports a summary of other conditions, indicating an overall operational
	// state of the cluster.
	ReadyCondition ConditionType = "Ready"

	// PendingUpdateReason reports when an update has been scheduled but has yet to happen.
	PendingUpdateReason = "PendingUpdate"

	// ControlPlaneInitializedCondition reports that all the control plane nodes are
	// in a fully ready and running state.
	ControlPlaneReadyCondition = "ControlPlaneReady"

	// ControlPlaneInitializedCondition reports that the first control plane instance has been initialized
	// and so the control plane is available and an API server instance is ready for processing requests.
	ControlPlaneInitializedCondition = "ControlPlaneInitialized"

	// WaitingForControlPlaneInitializedReason used when cluster is waiting for control plane to be initialized.
	WaitingForControlPlaneInitializedReason = "WaitingForControlPlaneInitialized"

	// FirstControlPlaneUnavailableMessage used when waiting for the first control plane instance to become
	// initialized and available.
	FirstControlPlaneUnavailableMessage = "The first control plane instance is not available yet"

	// WaitingForControlPlaneReadyReason used when cluster is waiting for control plane nodes to transition to a
	// fully ready and running state.
	WaitingForControlPlaneReadyReason = "WaitingForControlPlaneReady"

	// WorkersReadyConditon used reports the status on the worker nodes and is true when the expected
	// number of workers are in a fully ready and running state.
	WorkersReadyConditon ConditionType = "WorkersReady"

	// WaitingForWorkersReadyReason used when cluster is waiting for control plane nodes to transition to a
	// fully ready and running state.
	WaitingForWorkersReadyReason = "WaitingForWorkersReady"

	// DefaultCNIConfiguredCondition reports the default cni cluster has been configured successfully.
	DefaultCNIConfiguredCondition ConditionType = "DefaultCNIConfigured"

	// WaitingForDefaultCNIConfiguredReason used when cluster is waiting default cni to be installed.
	WaitingForDefaultCNIConfiguredReason = "WaitingForDefaultCNIConfigured"

	// SkipUpgradesForDefaultCNIConfiguredReason used to indicate the custer has been configured to skip
	// upgrades for the default cni. The default cni may still be installed, for example to successfully
	// create a cluster.
	SkipUpgradesForDefaultCNIConfiguredReason = "SkipUpgradesForDefaultCNIConfigured"
)
