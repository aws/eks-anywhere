package features

const (
	AwsIamAuthenticatorEnvVar = "AWS_IAM_AUTHENTICATOR"
	TaintsSupportEnvVar       = "TAINTS_SUPPORT"
	TinkerbellProviderEnvVar  = "TINKERBELL_PROVIDER"
	FullLifecycleAPIEnvVar    = "FULL_LIFECYCLE_API"
	FullLifecycleGate         = "FullLifecycleAPI"
	NetworkingUpgradeEnvVar   = "NETWORKING_UPGRADE"
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

func AwsIamAuthenticator() Feature {
	return Feature{
		Name:     "aws-iam-authenticator identity provider",
		IsActive: globalFeatures.isActiveForEnvVar(AwsIamAuthenticatorEnvVar),
	}
}

func TaintsSupport() Feature {
	return Feature{
		Name:     "Taints support",
		IsActive: globalFeatures.isActiveForEnvVar(TaintsSupportEnvVar),
	}
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

func NetworkingUpgrade() Feature {
	return Feature{
		Name:     "Upgrade CNI during cluster upgrades",
		IsActive: globalFeatures.isActiveForEnvVar(NetworkingUpgradeEnvVar),
	}
}
