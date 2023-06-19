package v1alpha1

// Conditions and condition Reasons for the Cluster object.
const (
	// ReadyCondition reports a summary of other conditions, indicating an overall operational
	// state of the cluster: all control plane and worker nodes are the right version,
	// all nodes are ready not including old nodes.
	ReadyCondition ConditionType = "Ready"

	// OutdatedInformationReason reports the system is waiting for stale cluster information to be refreshed.
	OutdatedInformationReason = "OutdatedInformation"

	// ControlPlaneInitializedCondition reports that all the control plane nodes are
	// in a fully ready and running state.
	ControlPlaneReadyCondition ConditionType = "ControlPlaneReady"

	// ControlPlaneInitializedCondition reports that the first control plane instance has been initialized
	// and so the control plane is available and an API server instance is ready for processing requests.
	ControlPlaneInitializedCondition ConditionType = "ControlPlaneInitialized"

	// ControlPlaneInitializationInProgressReason used when cluster the control plane initilization is in progress.
	ControlPlaneInitializationInProgressReason = "ControlPlaneInitializationInProgress"

	// FirstControlPlaneUnavailableMessage used when waiting for the first control plane instance to become
	// initialized and available.
	FirstControlPlaneUnavailableMessage = "The first control plane instance is not available yet"

	// WorkersReadyConditon used reports the status on the worker nodes and is true when the expected
	// number of workers are in a fully ready and running state.
	WorkersReadyConditon ConditionType = "WorkersReady"

	// DefaultCNIConfiguredCondition reports the default cni cluster has been configured successfully.
	DefaultCNIConfiguredCondition ConditionType = "DefaultCNIConfigured"
)
