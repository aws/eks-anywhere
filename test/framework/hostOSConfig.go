package framework

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	// NTP configuration environment variables.
	ntpServersVar = "T_NTP_SERVERS"

	// Bottlerocket configuration environment variables.
	maxPodsVar              = "T_BR_K8S_SETTINGS_MAX_PODS"
	clusterDNSIPSVar        = "T_BR_K8S_SETTINGS_CLUSTER_DNS_IPS"
	allowedUnsafeSysctlsVar = "T_BR_K8S_SETTINGS_ALLOWED_UNSAFE_SYSCTLS"

	// other constants.
	defaultSSHUsername = "ec2-user"
	privateKeyFileName = "eks-a-id_rsa"
)

var (
	ntpServersRequiredVar   = []string{ntpServersVar}
	brKubernetesRequiredVar = []string{maxPodsVar, clusterDNSIPSVar, allowedUnsafeSysctlsVar}
)

// RequiredNTPServersEnvVars returns a slice of environment variables required for NTP tests.
func RequiredNTPServersEnvVars() []string {
	return ntpServersRequiredVar
}

// RequiredBottlerocketKubernetesSettingsEnvVars returns a slice of environment variables required for Bottlerocket Kubernetes tests.
func RequiredBottlerocketKubernetesSettingsEnvVars() []string {
	return brKubernetesRequiredVar
}

// GetNTPServersFromEnv returns a slice of NTP servers read from the NTP environment veriables.
func GetNTPServersFromEnv() []string {
	serverFromEnv := os.Getenv(ntpServersVar)
	return strings.Split(serverFromEnv, ",")
}

// GetBottlerocketKubernetesSettingsFromEnv returns a Bottlerocket Kubernetes settings read from the environment variables.
func GetBottlerocketKubernetesSettingsFromEnv() (allowedUnsafeSysclts, clusterDNSIPS []string, maxPods int, err error) {
	allowedUnsafeSysclts = strings.Split(os.Getenv(allowedUnsafeSysctlsVar), ",")
	clusterDNSIPS = strings.Split(os.Getenv(clusterDNSIPSVar), ",")
	maxPods, err = strconv.Atoi(os.Getenv(maxPodsVar))
	return allowedUnsafeSysclts, clusterDNSIPS, maxPods, err
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

// ValidateBottlerocketKubernetesSettings validates Bottlerocket Kubernetes settings are configured properly on all cluster nodes using SSH.
func (e *ClusterE2ETest) ValidateBottlerocketKubernetesSettings() {
	ctx := context.Background()
	machines, err := e.KubectlClient.GetCAPIMachines(ctx, e.managementCluster(), e.ClusterName)
	if err != nil {
		e.T.Fatalf("Error getting machines: %v", err)
	}

	for _, machine := range machines {
		if len(machine.Status.Addresses) > 0 {
			e.T.Logf("Validating Bottlerocket Kubernetes settings for machine %s with IP %s", machine.Name, machine.Status.Addresses[0].Address)
			e.validateBottlerocketKubernetesSettings(ctx, machine.Status.Addresses[0].Address)
		}
	}
}

// nolint:gocyclo
func (e *ClusterE2ETest) validateBottlerocketKubernetesSettings(ctx context.Context, IP string) {
	ssh := buildSSH(e.T)

	command := []string{"apiclient", "get", "settings.network.hostname"}
	gotHostname, err := ssh.RunCommand(ctx, filepath.Join(e.ClusterName, privateKeyFileName), defaultSSHUsername, IP, command...)
	if err != nil {
		e.T.Errorf("failed to validate Bottlerocket Kubernetes settings: %v", err)
	}

	if strings.Contains(gotHostname, "etcd") {
		e.T.Log("Skipping Bottlerocket Kubernetes settings validation for etcd node")
		return
	}

	command = []string{"apiclient", "get", "settings.kubernetes.allowed-unsafe-sysctls"}
	gotAllowedUnsafeSysctls, err := ssh.RunCommand(ctx, filepath.Join(e.ClusterName, privateKeyFileName), defaultSSHUsername, IP, command...)
	if err != nil {
		e.T.Errorf("failed to validate Bottlerocket Kubernetes settings: %v", err)
	}

	command = []string{"apiclient", "get", "settings.kubernetes.cluster-dns-ip"}
	gotClusterDNSIPs, err := ssh.RunCommand(ctx, filepath.Join(e.ClusterName, privateKeyFileName), defaultSSHUsername, IP, command...)
	if err != nil {
		e.T.Errorf("failed to validate Bottlerocket Kubernetes settings: %v", err)
	}

	command = []string{"apiclient", "get", "settings.kubernetes.max-pods"}
	gotMaxPods, err := ssh.RunCommand(ctx, filepath.Join(e.ClusterName, privateKeyFileName), defaultSSHUsername, IP, command...)
	if err != nil {
		e.T.Errorf("failed to validate Bottlerocket Kubernetes settings: %v", err)
	}

	expectedAllowedUnsafeSysctls, expectedClusterDNSIPs, expectedMaxPods, err := GetBottlerocketKubernetesSettingsFromEnv()
	if err != nil {
		e.T.Errorf("failed to get Bottlerocket Kubernetes settings from environment variables: %v", err)
	}

	for _, sysctl := range expectedAllowedUnsafeSysctls {
		if !strings.Contains(gotAllowedUnsafeSysctls, sysctl) {
			e.T.Errorf("Bottlerocket Kubernetes setting [allowed-unsafe-sysctls: %s] not configured on machine", sysctl)
		}
		e.T.Logf("Bottlerocket Kubernetes setting [allowed-unsafe-sysctls: %s] is configured", sysctl)
	}

	for _, ip := range expectedClusterDNSIPs {
		if !strings.Contains(gotClusterDNSIPs, ip) {
			e.T.Errorf("Bottlerocket Kubernetes setting [cluster-dns-ips: %s] not configured on machine", ip)
		}
		e.T.Logf("Bottlerocket Kubernetes setting [cluster-dns-ips: %s] is configured", ip)
	}

	if !strings.Contains(gotMaxPods, strconv.Itoa(expectedMaxPods)) {
		e.T.Errorf("Bottlerocket Kubernetes setting [max-pods: %d] not configured on machine", expectedMaxPods)
	}

	e.T.Logf("Bottlerocket Kubernetes setting [max-pods: %d] is configured", expectedMaxPods)
}
