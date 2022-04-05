package snow

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type bootstrapCreds struct {
	snowCredsB64 string
	snowCertsB64 string
}

func (p *snowProvider) setupBootstrapCreds() error {
	creds, err := aws.EncodeFileFromEnv(aws.EksaAwsCredentialsFileKey)
	if err != nil {
		return fmt.Errorf("failed to set up snow credentials: %v", err)
	}
	p.bootstrapCreds.snowCredsB64 = creds

	certs, err := aws.EncodeFileFromEnv(aws.EksaAwsCABundlesFileKey)
	if err != nil {
		return fmt.Errorf("failed to set up snow certificates: %v", err)
	}
	p.bootstrapCreds.snowCertsB64 = certs

	// TODO: add validation logic againts creds/certs
	return nil
}
