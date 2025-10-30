package e2e

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	packagesRegex            = `^.*CuratedPackages.*$`
	nonRegionalPackagesRegex = `^.*NonRegionalCuratedPackages.*$`
	certManagerRegex         = "^.*CuratedPackagesCertManager.*$"
)

// assumeRoleAndGetCredentials assumes an IAM role using the role ARN from env and
// returns the temporary credentials (access key, secret key, session token) or an error.
func assumeRoleAndGetCredentials(roleArnEnvVar, sessionName string) (accessKey, secretKey, sessionToken string, err error) {
	roleArn := os.Getenv(roleArnEnvVar)
	if roleArn == "" {
		return "", "", "", fmt.Errorf("%s environment variable not set", roleArnEnvVar)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	duration := int32(3600) // (max for role chaining)

	result, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: &sessionName,
		DurationSeconds: &duration,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to assume role %s: %w", roleArn, err)
	}

	return *result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, nil
}

func (e *E2ESession) setupPackagesEnv(testRegex string) error {
	re := regexp.MustCompile(packagesRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	isNonRegional := regexp.MustCompile(nonRegionalPackagesRegex).MatchString(testRegex)

	var roleArnEnvVar, sessionName string
	if isNonRegional {
		roleArnEnvVar = "NON_REGIONAL_PACKAGES_ROLE_ARN"
		sessionName = "test-packages-nonregional"
	} else {
		roleArnEnvVar = "PACKAGES_ROLE_ARN"
		sessionName = "test-packages"
	}

	accessKey, secretKey, sessionToken, err := assumeRoleAndGetCredentials(roleArnEnvVar, sessionName)
	if err != nil {
		return fmt.Errorf("failed to get packages credentials: %w", err)
	}

	e.testEnvVars["EKSA_AWS_ACCESS_KEY_ID"] = accessKey
	e.testEnvVars["EKSA_AWS_SECRET_ACCESS_KEY"] = secretKey
	e.testEnvVars["EKSA_AWS_SESSION_TOKEN"] = sessionToken

	if region, ok := os.LookupEnv("EKSA_AWS_REGION"); ok {
		e.testEnvVars["EKSA_AWS_REGION"] = region
	}

	return nil
}

func (e *E2ESession) setupCertManagerEnv(testRegex string) error {
	re := regexp.MustCompile(certManagerRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	accessKey, secretKey, sessionToken, err := assumeRoleAndGetCredentials("CERT_MANAGER_ROLE_ARN", "test-certmanager")
	if err != nil {
		return fmt.Errorf("failed to get cert manager credentials: %w", err)
	}

	e.testEnvVars["ROUTE53_ACCESS_KEY_ID"] = accessKey
	e.testEnvVars["ROUTE53_SECRET_ACCESS_KEY"] = secretKey
	e.testEnvVars["ROUTE53_SESSION_TOKEN"] = sessionToken

	if region, ok := os.LookupEnv("ROUTE53_REGION"); ok {
		e.testEnvVars["ROUTE53_REGION"] = region
	}
	if zoneId, ok := os.LookupEnv("ROUTE53_ZONEID"); ok {
		e.testEnvVars["ROUTE53_ZONEID"] = zoneId
	}

	return nil
}
