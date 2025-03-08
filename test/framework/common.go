package framework

var requiredCommonEnvVars = []string{
	LicenseTokenEnvVar,
	LicenseToken2EnvVar,
	StagingLicenseTokenEnvVar,
	StagingLicenseToken2EnvVar,
}

// RequiredCommonEnvVars returns the list of env variables required for all tests.
func RequiredCommonEnvVars() []string {
	return requiredCommonEnvVars
}
