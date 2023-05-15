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

		if err := e.mountRegistryCert(caCert, net.JoinHostPort(endpoint, port)); err != nil {
			return err
		}
	}

	// Since Authenticated tests needs to use separate harbor registries.
	re = regexp.MustCompile(`^.*(VSphere|CloudStack).*Authenticated.*$`)
	if re.MatchString(testRegex) {
		endpoint = e.testEnvVars[e2etests.PrivateRegistryEndpointVar]
		port = e.testEnvVars[e2etests.PrivateRegistryPortVar]
		caCert = e.testEnvVars[e2etests.PrivateRegistryCACertVar]
	} else if re = regexp.MustCompile(`^.*Tinkerbell.*Authenticated.*$`); re.MatchString(testRegex) {
		endpoint = e.testEnvVars[e2etests.PrivateRegistryEndpointTinkerbellVar]
		port = e.testEnvVars[e2etests.PrivateRegistryPortTinkerbellVar]
		caCert = e.testEnvVars[e2etests.PrivateRegistryCACertTinkerbellVar]
	}

	if endpoint != "" && port != "" && caCert != "" {
		return e.mountRegistryCert(caCert, net.JoinHostPort(endpoint, port))
	}

	re = regexp.MustCompile(`^.*Docker.*Airgapped.*$`)
	if re.MatchString(testRegex) {
		err := os.Setenv("DEFAULT_SECURITY_GROUP", e.testEnvVars[e2etests.RegistryMirrorDefaultSecurityGroup])
		if err != nil {
			return fmt.Errorf("unable to set DEFAULT_SECURITY_GROUP: %v", err)
		}
		err = os.Setenv("AIRGAPPED_SECURITY_GROUP", e.testEnvVars[e2etests.RegistryMirrorAirgappedSecurityGroup])
		if err != nil {
			return fmt.Errorf("unable to set AIRGAPPED_SECURITY_GROUP: %v", err)
		}
	}

	return nil
}

func (e *E2ESession) mountRegistryCert(cert string, endpoint string) error {
	command := fmt.Sprintf("sudo mkdir -p /etc/docker/certs.d/%s", endpoint)

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("creating directory in instance: %v", err)
	}
	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return fmt.Errorf("failed to decode certificate: %v", err)
	}
	command = fmt.Sprintf("sudo cat <<EOF>> /etc/docker/certs.d/%s/ca.crt\n%s\nEOF", endpoint, string(decodedCert))

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("mounting certificate in instance: %v", err)
	}

	return err
}
