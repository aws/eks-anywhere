package e2e

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/eks-anywhere/internal/pkg/s3"
)

const (
	snowCredentialsS3Path  = "T_SNOW_CREDENTIALS_S3_PATH"
	snowCertificatesS3Path = "T_SNOW_CERTIFICATES_S3_PATH"
	snowDevices            = "T_SNOW_DEVICES"
	snowCPCidr             = "T_SNOW_CONTROL_PLANE_CIDR"
	snowCPCidrs            = "T_SNOW_CONTROL_PLANE_CIDRS"
	snowCredsFile          = "EKSA_AWS_CREDENTIALS_FILE"
	snowCertsFile          = "EKSA_AWS_CA_BUNDLES_FILE"

	snowTestsRe       = `^.*Snow.*$`
	snowCredsFilename = "snow_creds"
	snowCertsFilename = "snow_certs"
)

var (
	snowCPCidrArray  []string
	snowCPCidrArrayM sync.Mutex
)

func init() {
	snowCPCidrArray = strings.Split(os.Getenv(snowCPCidrs), ",")
}

// Note that this function cannot be called more than the the number of cidrs in the list.
func getSnowCPCidr() (string, error) {
	snowCPCidrArrayM.Lock()
	defer snowCPCidrArrayM.Unlock()

	if len(snowCPCidrArray) == 0 {
		return "", fmt.Errorf("no more snow control plane cidrs available")
	}
	var r string
	r, snowCPCidrArray = snowCPCidrArray[0], snowCPCidrArray[1:]
	return r, nil
}

func (e *E2ESession) setupSnowEnv(testRegex string) error {
	re := regexp.MustCompile(snowTestsRe)
	if !re.MatchString(testRegex) {
		return nil
	}

	e.testEnvVars[snowDevices] = os.Getenv(snowDevices)
	cpCidr, err := getSnowCPCidr()
	if err != nil {
		return err
	}
	e.testEnvVars[snowCPCidr] = cpCidr
	e.logger.V(1).Info("Assigned control plane CIDR to admin instance", "cidr", cpCidr, "instanceId", e.instanceId)

	if err := sendFileViaS3(e, os.Getenv(snowCredentialsS3Path), snowCredsFilename); err != nil {
		return err
	}
	if err := sendFileViaS3(e, os.Getenv(snowCertificatesS3Path), snowCertsFilename); err != nil {
		return err
	}
	e.testEnvVars[snowCredsFile] = "bin/" + snowCredsFilename
	e.testEnvVars[snowCertsFile] = "bin/" + snowCertsFilename

	return nil
}

func sendFileViaS3(e *E2ESession, s3Path string, filename string) error {
	if err := s3.DownloadToDisk(e.session, s3Path, e.storageBucket, "bin/"+filename); err != nil {
		return err
	}

	err := e.uploadRequiredFile(filename)
	if err != nil {
		return fmt.Errorf("failed to upload file (%s) : %v", filename, err)
	}

	err = e.downloadRequiredFileInInstance(filename)
	if err != nil {
		return fmt.Errorf("failed to download file (%s) in admin instance : %v", filename, err)
	}
	return nil
}
