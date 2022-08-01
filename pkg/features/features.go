package features

const (
	CloudStackProviderEnvVar        = "CLOUDSTACK_PROVIDER"
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	SnowProviderEnvVar              = "SNOW_PROVIDER"
	FullLifecycleAPIEnvVar          = "FULL_LIFECYCLE_API"
	FullLifecycleGate               = "FullLifecycleAPI"
	CuratedPackagesEnvVar           = "CURATED_PACKAGES_SUPPORT"
	K8s123SupportEnvVar             = "K8S_1_23_SUPPORT"
	CheckpointEnabledEnvVar         = "CHECKPOINT_ENABLED"
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

func CuratedPackagesSupport() Feature {
	return Feature{
		Name:     "Curated Packages Support",
		IsActive: globalFeatures.isActiveForEnvVar(CuratedPackagesEnvVar),
	}
}

func K8s123Support() Feature {
	return Feature{
		Name:     "Kubernetes version 1.23 support",
		IsActive: globalFeatures.isActiveForEnvVar(K8s123SupportEnvVar),
	}
}

func CheckpointEnabled() Feature {
	return Feature{
		Name:     "Checkpoint to rerun commands enabled",
		IsActive: globalFeatures.isActiveForEnvVar(CheckpointEnabledEnvVar),
	}
}
