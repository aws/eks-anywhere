package netest

import "fmt"

// GetGitOpsEndpoints returns the endpoints the system must be able to connect to for GitOps features.
func GetGitOpsEndpoints() []string {
	return []string{
		"api.github.com",
	}
}

// GetPackagesEndpoint creates an endpoint using region that the system should be able to connect to.
func GetPackagesEndpoint(region string) string {
	return fmt.Sprintf("api.ecr.%s.amazonaws.com", region)
}
