package snow

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/eks-anywhere/pkg/validations"
)

type bootstrapCreds struct {
	snowCredsB64 string
	snowCertsB64 string
}

func (p *snowProvider) setupBootstrapCreds() error {
	creds, err := encodeFileFromEnv(eksaSnowCredentialsFileKey)
	if err != nil {
		return fmt.Errorf("failed to set up snow credentials: %v", err)
	}
	p.bootstrapCreds.snowCredsB64 = creds

	certs, err := encodeFileFromEnv(eksaSnowCABundlesFileKey)
	if err != nil {
		return fmt.Errorf("failed to set up snow certificates: %v", err)
	}
	p.bootstrapCreds.snowCertsB64 = certs

	// TODO: add validation logic againts creds/certs
	return nil
}

func encodeFileFromEnv(envKey string) (string, error) {
	file, ok := os.LookupEnv(envKey)
	if !ok || len(file) <= 0 {
		return "", fmt.Errorf("%s is not set or is empty", envKey)
	}

	fileExists := validations.FileExists(file)
	if !fileExists {
		return "", fmt.Errorf("file %s does not exist", file)
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("unable to read file due to: %v", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
