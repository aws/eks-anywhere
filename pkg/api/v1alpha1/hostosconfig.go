package v1alpha1

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func validateHostOSConfig(config *HostOSConfiguration) error {
	if config == nil {
		return nil
	}

	return validateNTPServers(config)
}

func validateNTPServers(config *HostOSConfiguration) error {
	if config.NTPConfiguration == nil {
		return nil
	}

	if len(config.NTPConfiguration.Servers) == 0 {
		return errors.New("NTPConfiguration.Servers can not be empty")
	}
	var invalidServers []string
	for _, ntpServer := range config.NTPConfiguration.Servers {
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
