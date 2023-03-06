package framework

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	ntpServersVar      = "T_NTP_SERVERS"
	defaultSSHUsername = "ec2-user"
	privateKeyFileName = "eks-a-id_rsa"
)

var ntpServersRequiredVar = []string{ntpServersVar}

// RequiredNTPServersEnvVars returns a slice of environment variables required for NTP tests.
func RequiredNTPServersEnvVars() []string {
	return ntpServersRequiredVar
}

// GetNTPServersFromEnv returns a slice of NTP servers read from the NTP environment veriables.
func GetNTPServersFromEnv() []string {
	serverFromEnv := os.Getenv(ntpServersVar)
	return strings.Split(serverFromEnv, ",")
}

// ValidateNTPConfig validates NTP servers are configured properly on all cluster nodes using SSH.
func (e *ClusterE2ETest) ValidateNTPConfig(osFamily v1alpha1.OSFamily) {
	ctx := context.Background()
	machines, err := e.KubectlClient.GetCAPIMachines(ctx, e.managementCluster(), e.ClusterName)
	if err != nil {
		e.T.Fatalf("Error getting machines: %v", err)
	}

	for _, machine := range machines {
		if len(machine.Status.Addresses) > 0 {
			e.T.Logf("Validating NTP servers for machine %s with IP %s", machine.Name, machine.Status.Addresses[0].Address)
			e.validateNTP(ctx, osFamily, machine.Status.Addresses[0].Address)
		}
	}
}

func (e *ClusterE2ETest) validateNTP(ctx context.Context, osFamily v1alpha1.OSFamily, IP string) {
	ssh := buildSSH(e.T)
	var command []string
	if osFamily == v1alpha1.Bottlerocket {
		command = []string{"apiclient", "get", "settings.ntp"}
	} else {
		command = []string{"chronyc", "sourcestats"}
	}

	out, err := ssh.RunCommand(ctx, filepath.Join(e.ClusterName, privateKeyFileName), defaultSSHUsername, IP, command...)
	if err != nil {
		e.T.Fatalf("failed to validate NTP server: %v", err)
	}

	for _, server := range GetNTPServersFromEnv() {
		if !strings.Contains(out, server) {
			e.T.Fatalf("NTP Server [%s] not configured on machine", server)
		}
		e.T.Logf("NTP server [%s] is configured", server)
	}
}
