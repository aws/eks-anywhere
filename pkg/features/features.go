package features

const (
	AwsIamAuthenticatorEnvVar = "AWS_IAM_AUTHENTICATOR"
	TaintsSupportEnvVar       = "TAINTS_SUPPORT"
	TinkerbellProviderEnvVar  = "TINKERBELL_PROVIDER"
	FullLifecycleAPIEnvVar    = "FULL_LIFECYCLE_API"
	FullLifecycleGate         = "FullLifecycleAPI"
	V1beta1BundleRelease      = "V1BETA1_BUNDLE"
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

func UseV1beta1BundleRelease() Feature {
	return Feature{
		Name:     "Use tags from v1beta1 bundle-release.yaml",
		IsActive: globalFeatures.isActiveForEnvVar(V1beta1BundleRelease),
	}
}
