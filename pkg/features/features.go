package features

const (
	TinkerbellProviderEnvVar        = "TINKERBELL_PROVIDER"
	CloudStackProviderEnvVar        = "CLOUDSTACK_PROVIDER"
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	SnowProviderEnvVar              = "SNOW_PROVIDER"
	FullLifecycleAPIEnvVar          = "FULL_LIFECYCLE_API"
	FullLifecycleGate               = "FullLifecycleAPI"
	CuratedPackagesEnvVar           = "CURATED_PACKAGES_SUPPORT"
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

func FullLifecycleAPI() Feature {
	return Feature{
		Name:     "Full lifecycle API support through the EKS-A controller",
		IsActive: globalFeatures.isActiveForEnvVarOrGate(FullLifecycleAPIEnvVar, FullLifecycleGate),
	}
}

func TinkerbellProvider() Feature {
	return Feature{
		Name:     "Tinkerbell provider support",
		IsActive: globalFeatures.isActiveForEnvVar(TinkerbellProviderEnvVar),
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

func NutanixProvider() Feature {
	return Feature{
		Name:     "Nutanix provider support",
		IsActive: globalFeatures.isActiveForEnvVar(NutanixProviderEnvVar),
	}
}
