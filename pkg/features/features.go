package features

// These are environment variables used as flags to enable/disable features.
const (
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	FullLifecycleAPIEnvVar          = "FULL_LIFECYCLE_API"
	FullLifecycleGate               = "FullLifecycleAPI"
	CheckpointEnabledEnvVar         = "CHECKPOINT_ENABLED"
	UseNewWorkflowsEnvVar           = "USE_NEW_WORKFLOWS"
	K8s126SupportEnvVar             = "K8S_1_26_SUPPORT"
)

func FeedGates(featureGates []string) {
	globalFeatures.feedGates(featureGates)
}

type Feature struct {
	Name     string
	IsActive func() bool
}

func IsActive(feature Feature) bool {
	return feature.IsActive()
}

// ClearCache is mainly used for unit tests as of now.
func ClearCache() {
	globalFeatures.clearCache()
}

func FullLifecycleAPI() Feature {
	return Feature{
		Name:     "Full lifecycle API support through the EKS-A controller",
		IsActive: globalFeatures.isActiveForEnvVarOrGate(FullLifecycleAPIEnvVar, FullLifecycleGate),
	}
}

func CloudStackKubeVipDisabled() Feature {
	return Feature{
		Name:     "Kube-vip support disabled in CloudStack provider",
		IsActive: globalFeatures.isActiveForEnvVar(CloudStackKubeVipDisabledEnvVar),
	}
}

func CheckpointEnabled() Feature {
	return Feature{
		Name:     "Checkpoint to rerun commands enabled",
		IsActive: globalFeatures.isActiveForEnvVar(CheckpointEnabledEnvVar),
	}
}

func UseNewWorkflows() Feature {
	return Feature{
		Name:     "Use new workflow logic for cluster management operations",
		IsActive: globalFeatures.isActiveForEnvVar(UseNewWorkflowsEnvVar),
	}
}

// K8s126Support is the feature flag for Kubernetes 1.26 support.
func K8s126Support() Feature {
	return Feature{
		Name:     "Kubernetes version 1.26 support",
		IsActive: globalFeatures.isActiveForEnvVar(K8s126SupportEnvVar),
	}
}
