package v1alpha1

import (
	"crypto/x509"
	"encoding/pem"
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

	for _, certBundle := range config.CertBundles {
		if err := validateCertBundles(&certBundle, osFamily); err != nil {
			return err
		}
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

func validateCertBundles(config *certBundle, osFamily OSFamily) error {
	if config == nil {
		return nil
	}

	if osFamily != Bottlerocket {
		return fmt.Errorf("CertBundles can only be used with osFamily: \"%s\"", Bottlerocket)
	}

	if config.Name == "" {
		return errors.New("certBundles name cannot be empty")
	}
	if err := validateTrustedCertBundle(config.Data); err != nil {
		return err
	}
	return nil
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

	if err := validateBottlerocketKernelConfiguration(config.Kernel); err != nil {
		return err
	}

	return validateBottlerocketBootSettingsConfiguration(config.Boot)
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

func validateBottlerocketBootSettingsConfiguration(config *v1beta1.BottlerocketBootSettings) error {
	if config == nil {
		return nil
	}

	for key := range config.BootKernelParameters {
		if key == "" {
			return fmt.Errorf("bootKernelParameters key cannot be empty")
		}
	}

	return nil
}

// validateTrustedCertBundle validates that the cert is valid.
func validateTrustedCertBundle(certBundle string) error {
	var blocks []byte
	rest := []byte(certBundle)

	// cert bundles could contain more than one certificate
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)

		// no more PEM structed objects
		if block == nil {
			break
		}
		blocks = append(blocks, block.Bytes...)
		if len(rest) == 0 {
			break
		}
	}

	if len(blocks) == 0 {
		return fmt.Errorf("failed to parse certificate PEM")
	}

	_, err := x509.ParseCertificates(blocks)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	return nil
}
