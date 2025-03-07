package framework

var requiredCommonEnvVars = []string{
	LicenseTokenEnvVar,
	LicenseToken2EnvVar,
	StagingLicenseTokenEnvVar,
}

// RequiredCommonEnvVars returns the list of env variables required for all tests.
func RequiredCommonEnvVars() []string {
	return requiredCommonEnvVars
}
