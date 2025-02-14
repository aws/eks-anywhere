package framework

var requiredCommonEnvVars = []string{
	LicenseTokenEnvVar,
}

// RequiredCommonEnvVars returns the list of env variables required for all tests.
func RequiredCommonEnvVars() []string {
	return requiredCommonEnvVars
}
