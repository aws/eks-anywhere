package features

// These are environment variables used as flags to enable/disable features.
const (
	CloudStackKubeVipDisabledEnvVar = "CLOUDSTACK_KUBE_VIP_DISABLED"
	CheckpointEnabledEnvVar         = "CHECKPOINT_ENABLED"
	UseControllerForCli             = "USE_CONTROLLER_FOR_CLI"
	VSphereInPlaceEnvVar            = "VSPHERE_IN_PLACE_UPGRADE"
	APIServerExtraArgsEnabledEnvVar = "API_SERVER_EXTRA_ARGS_ENABLED"
	K8s135SupportEnvVar             = "K8S_1_35_SUPPORT"
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

// VSphereInPlaceUpgradeEnabled is the feature flag for performing in-place upgrades with the vSphere provider.
func VSphereInPlaceUpgradeEnabled() Feature {
	return Feature{
		Name:     "Perform in-place upgrades with the vSphere provider",
		IsActive: globalFeatures.isActiveForEnvVar(VSphereInPlaceEnvVar),
	}
}

// APIServerExtraArgsEnabled is the feature flag for configuring api server extra args.
func APIServerExtraArgsEnabled() Feature {
	return Feature{
		Name:     "Configure api server extra args",
		IsActive: globalFeatures.isActiveForEnvVar(APIServerExtraArgsEnabledEnvVar),
	}
}

// K8s135Support is the feature flag for Kubernetes 1.35 support.
func K8s135Support() Feature {
	return Feature{
		Name:     "Kubernetes version 1.35 support",
		IsActive: globalFeatures.isActiveForEnvVar(K8s135SupportEnvVar),
	}
}
