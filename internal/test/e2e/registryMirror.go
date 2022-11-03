package e2e

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupRegistryMirrorEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*RegistryMirror.*$`)
	if !re.MatchString(testRegex) {
		e.logger.V(2).Info("Not running RegistryMirror tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredRegistryMirrorEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	endpoint := e.testEnvVars[e2etests.RegistryEndpointVar]
	port := e.testEnvVars[e2etests.RegistryPortVar]
	caCert := e.testEnvVars[e2etests.RegistryCACertVar]

	// Since Tinkerbell uses a separate harbor registry,
	// we need to setup cert for that registry for Tinkerbell tests.
	re = regexp.MustCompile(`^.*Tinkerbell.*$`)
	if re.MatchString(testRegex) {
		endpoint = e.testEnvVars[e2etests.RegistryEndpointTinkerbellVar]
		port = e.testEnvVars[e2etests.RegistryPortTinkerbellVar]
		caCert = e.testEnvVars[e2etests.RegistryCACertTinkerbellVar]
	}

	if endpoint != "" && port != "" && caCert != "" {
		return e.mountRegistryCert(caCert, net.JoinHostPort(endpoint, port))
	}

	return nil
}

func (e *E2ESession) mountRegistryCert(cert string, endpoint string) error {
	command := fmt.Sprintf("sudo mkdir -p /etc/docker/certs.d/%s", endpoint)

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command); err != nil {
		return fmt.Errorf("creating directory in instance: %v", err)
	}
	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %v", err)
	}
	command = fmt.Sprintf("sudo cat <<EOF>> /etc/docker/certs.d/%s/ca.crt\n%s\nEOF", endpoint, string(decodedCert))

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command); err != nil {
		return fmt.Errorf("mounting certificate in instance: %v", err)
	}

	return err
}
