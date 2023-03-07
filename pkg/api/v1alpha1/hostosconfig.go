package v1alpha1

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
)

func validateHostOSConfig(config *HostOSConfiguration, osFamily OSFamily) error {
	if config == nil {
		return nil
	}

	if err := validateNTPServers(config.NTPConfiguration); err != nil {
		return err
	}

	return validateBotterocketConfig(config.BottlerocketConfiguration, osFamily)
}

func validateNTPServers(config *NTPConfiguration) error {
	if config == nil {
		return nil
	}

	if len(config.Servers) == 0 {
		return errors.New("NTPConfiguration.Servers can not be empty")
	}
	var invalidServers []string
	for _, ntpServer := range config.Servers {
		// ParseRequestURI expects a scheme but ntp servers generally don't have one
		// Prepending a scheme here so it doesn't fail because of missing scheme
		if u, err := url.ParseRequestURI(addNTPScheme(ntpServer)); err != nil || u.Scheme == "" || u.Host == "" {
			invalidServers = append(invalidServers, ntpServer)
		}
	}
	if len(invalidServers) != 0 {
		return fmt.Errorf("ntp servers [%s] is not valid", strings.Join(invalidServers[:], ", "))
	}

	return nil
}

func addNTPScheme(server string) string {
	if strings.Contains(server, "://") {
		return server
	}
	return fmt.Sprintf("udp://%s", server)
}

func validateBotterocketConfig(config *BottlerocketConfiguration, osFamily OSFamily) error {
	if config == nil {
		return nil
	}

	if osFamily != Bottlerocket {
		return fmt.Errorf("BottlerocketConfiguration can only be used with osFamily: \"%s\"", Bottlerocket)
	}

	if err := validateBottlerocketKubernetesConfig(config.Kubernetes); err != nil {
		return err
	}

	return validateBottlerocketKernelConfiguration(config.Kernel)
}

func validateBottlerocketKubernetesConfig(config *v1beta1.BottlerocketKubernetesSettings) error {
	if config == nil {
		return nil
	}

	for _, val := range config.AllowedUnsafeSysctls {
		if val == "" {
			return errors.New("BottlerocketConfiguration.Kubernetes.AllowedUnsafeSysctls can not have an empty string (\"\")")
		}
	}

	for _, ip := range config.ClusterDNSIPs {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("IP address [%s] in BottlerocketConfiguration.Kubernetes.ClusterDNSIPs is not a valid IP", ip)
		}
	}

	if config.MaxPods < 0 {
		return errors.New("BottlerocketConfiguration.Kubernetes.MaxPods can not be negative")
	}

	return nil
}

func validateBottlerocketKernelConfiguration(config *v1beta1.BottlerocketKernelSettings) error {
	if config == nil {
		return nil
	}

	for key := range config.SysctlSettings {
		if key == "" {
			return errors.New("sysctlSettings key cannot be empty")
		}
	}
	return nil
}
