package e2e

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aws/eks-anywhere/internal/pkg/oidc"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	openIDConfPath = "oidc/.well-known/openid-configuration"
	keysPath       = "oidc/keys.json"
	saSignerPath   = "oidc/sa-signer.key"
)

func (e *E2ESession) setupOIDC(testRegex string) error {
	re := regexp.MustCompile(`^.*OIDC.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running OIDC tests, skipping setup")
		return nil
	}

	folder := e.jobId
	bucketUrl := s3.GetBucketPublicURL(e.session, e.storageBucket)
	issuerURL := fmt.Sprintf("%s/%s/%s", bucketUrl, folder, "oidc")

	logger.V(1).Info("OIDC test found. Checking if OIDC folder present in bucket")
	oidcPresent, err := s3.ObjectPresent(e.session, filepath.Join(e.jobId, keysPath), e.storageBucket)
	if err != nil {
		return fmt.Errorf("checking if oidc is present in bucket: %v", err)
	}

	if !oidcPresent {
		logger.V(1).Info("OIDC not present in bucket, creating necessary files")
		err = e.createOIDCFiles(issuerURL, folder)
		if err != nil {
			return err
		}
	} else {
		logger.V(1).Info("OIDC already present in bucket, skipping setup")
	}

	logger.V(1).Info("Getting key id from s3 bucket")
	keyID, err := e.getKeyID(folder)
	if err != nil {
		return err
	}

	keyPath, err := e.downloadSignerKeyInInstance(folder)
	if err != nil {
		return err
	}

	e.testEnvVars[e2etests.OIDCIssuerUrlVar] = issuerURL
	e.testEnvVars[e2etests.OIDCClientIdVar] = keyID
	e.testEnvVars[e2etests.OIDCKidVar] = keyID
	e.testEnvVars[e2etests.OIDCKeyFileVar] = keyPath

	return nil
}

func (e *E2ESession) createOIDCFiles(issuerURL, folder string) error {
	provider, err := oidc.GenerateMinimalProvider(issuerURL)
	if err != nil {
		return fmt.Errorf("setting up generating oidc provider for s3: %v", err)
	}

	logger.V(2).Info("Uploading OIDC discovery file to S3")
	discoveryKey := filepath.Join(folder, openIDConfPath)
	err = s3.Upload(e.session, provider.Discovery, discoveryKey, e.storageBucket, s3.WithPublicRead())
	if err != nil {
		return fmt.Errorf("uploading oidc openid-configuration to s3: %v", err)
	}

	logger.V(2).Info("Uploading OIDC keys file to S3")
	keysKey := filepath.Join(folder, keysPath)
	err = s3.Upload(e.session, provider.Keys, keysKey, e.storageBucket, s3.WithPublicRead())
	if err != nil {
		return fmt.Errorf("uploading oidc keys.json to s3: %v", err)
	}

	logger.V(2).Info("Uploading OIDC signer key to S3")
	saSignerKey := filepath.Join(folder, saSignerPath)
	err = s3.Upload(e.session, provider.PrivateKey, saSignerKey, e.storageBucket)
	if err != nil {
		return fmt.Errorf("uploading oidc sa-signer.key to s3: %v", err)
	}

	return nil
}

func (e *E2ESession) getKeyID(folder string) (string, error) {
	keysKey := filepath.Join(folder, keysPath)
	logger.V(2).Info("Downloading keys.json file from s3")
	keysBytes, err := s3.Download(e.session, keysKey, e.storageBucket)
	if err != nil {
		return "", fmt.Errorf("downloading keys.json to get kid: %v", err)
	}

	keyID, err := oidc.GetKeyID(keysBytes)
	if err != nil {
		return "", fmt.Errorf("getting kid from s3: %v", err)
	}

	return keyID, nil
}

func (e *E2ESession) downloadSignerKeyInInstance(folder string) (pathInInstance string, err error) {
	saSignerKey := filepath.Join(folder, saSignerPath)
	logger.V(1).Info("Downloading from s3 in instance", "key", saSignerKey)
	command := fmt.Sprintf("aws s3 cp s3://%s/%s ./%s", e.storageBucket, saSignerKey, saSignerPath)

	if err = ssm.Run(e.session, e.instanceId, command); err != nil {
		return "", fmt.Errorf("downloading signer key in instance: %v", err)
	}
	logger.V(1).Info("Successfully downloaded signer key")

	return saSignerPath, nil
}
