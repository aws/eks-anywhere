package features

const (
	CloudStackProviderEnvVar        = "CLOUDSTACK_PROVIDER"
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	SnowProviderEnvVar              = "SNOW_PROVIDER"
	FullLifecycleAPIEnvVar          = "FULL_LIFECYCLE_API"
	FullLifecycleGate               = "FullLifecycleAPI"
	CheckpointEnabledEnvVar         = "CHECKPOINT_ENABLED"
	NutanixProviderEnvVar           = "NUTANIX_PROVIDER"
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

// ClearCache is mainly used for unit tests as of now
func ClearCache() {
	globalFeatures.clearCache()
}

func FullLifecycleAPI() Feature {
	return Feature{
		Name:     "Full lifecycle API support through the EKS-A controller",
		IsActive: globalFeatures.isActiveForEnvVarOrGate(FullLifecycleAPIEnvVar, FullLifecycleGate),
	}
}

func CloudStackProvider() Feature {
	return Feature{
		Name:     "CloudStack provider support",
		IsActive: globalFeatures.isActiveForEnvVar(CloudStackProviderEnvVar),
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

// NutanixProvider returns a feature that is active if the NUTANIX_PROVIDER environment variable is true
func NutanixProvider() Feature {
	return Feature{
		Name:     "Nutanix provider support",
		IsActive: globalFeatures.isActiveForEnvVar(NutanixProviderEnvVar),
	}
}
