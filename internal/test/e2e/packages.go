package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	packagesRegex         = `^.*CuratedPackages.*$`
	regionalPackagesRegex = `^.*RegionalCuratedPackages.*$`
	certManagerRegex      = "^.*CuratedPackagesCertManager.*$"
)

func (e *E2ESession) setupPackagesEnv(testRegex string) error {
	re := regexp.MustCompile(packagesRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredPackagesEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	// overwrite envs for regional curated packages test
	if regexp.MustCompile(regionalPackagesRegex).MatchString(testRegex) {
		for _, eVar := range requiredEnvVars {
			if val, ok := os.LookupEnv("REGIONAL_" + eVar); ok {
				e.testEnvVars[eVar] = val
			}
		}
	}
	return nil
}

func (e *E2ESession) setupCertManagerEnv(testRegex string) error {
	re := regexp.MustCompile(certManagerRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredCertManagerEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}
