package features

const (
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	SnowProviderEnvVar              = "SNOW_PROVIDER"
	FullLifecycleAPIEnvVar          = "FULL_LIFECYCLE_API"
	FullLifecycleGate               = "FullLifecycleAPI"
	CheckpointEnabledEnvVar         = "CHECKPOINT_ENABLED"
	NutanixProviderEnvVar           = "NUTANIX_PROVIDER"
	UseNewWorkflowsEnvVar           = "USE_NEW_WORKFLOWS"
	K8s124SupportEnvVar             = "K8S_1_24_SUPPORT"
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

func K8s124Support() Feature {
	return Feature{
		Name:     "Kubernetes version 1.24 support",
		IsActive: globalFeatures.isActiveForEnvVar(K8s124SupportEnvVar),
	}
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

func SnowProvider() Feature {
	return Feature{
		Name:     "Snow provider support",
		IsActive: globalFeatures.isActiveForEnvVar(SnowProviderEnvVar),
	}
}

func CheckpointEnabled() Feature {
	return Feature{
		Name:     "Checkpoint to rerun commands enabled",
		IsActive: globalFeatures.isActiveForEnvVar(CheckpointEnabledEnvVar),
	}
}

// NutanixProvider returns a feature that is active if the NUTANIX_PROVIDER environment variable is true.
func NutanixProvider() Feature {
	return Feature{
		Name:     "Nutanix provider support",
		IsActive: globalFeatures.isActiveForEnvVar(NutanixProviderEnvVar),
	}
}

func UseNewWorkflows() Feature {
	return Feature{
		Name:     "Use new workflow logic for cluster management operations",
		IsActive: globalFeatures.isActiveForEnvVar(UseNewWorkflowsEnvVar),
	}
}
