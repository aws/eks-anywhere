package e2e

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupRegistryMirrorEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*RegistryMirror.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running RegistryMirror tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredRegistryMirrorEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	if e.testEnvVars[e2etests.RegistryCACertVar] != "" && e.testEnvVars[e2etests.RegistryEndpointVar] != "" && e.testEnvVars[e2etests.RegistryPortVar] != "" {
		return e.mountRegistryCert(e.testEnvVars[e2etests.RegistryCACertVar], net.JoinHostPort(e.testEnvVars[e2etests.RegistryEndpointVar], e.testEnvVars[e2etests.RegistryPortVar]))
	}

	return nil
}

func (e *E2ESession) mountRegistryCert(cert string, endpoint string) error {
	command := fmt.Sprintf("sudo mkdir -p /etc/docker/certs.d/%s", endpoint)

	if err := ssm.Run(e.session, e.instanceId, command); err != nil {
		return fmt.Errorf("creating directory in instance: %v", err)
	}
	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %v", err)
	}
	command = fmt.Sprintf("sudo cat <<EOF>> /etc/docker/certs.d/%s/ca.crt\n%s\nEOF", endpoint, string(decodedCert))

	if err := ssm.Run(e.session, e.instanceId, command); err != nil {
		return fmt.Errorf("mounting certificate in instance: %v", err)
	}

	return err
}
