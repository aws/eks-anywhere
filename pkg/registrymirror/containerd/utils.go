package containerd

import (
	"net/url"
	"path/filepath"
	"strings"
)

// ToAPIEndpoint turns URL to a valid API endpoint used in
// a containerd config file for a local registry.
// Original input is returned in case of malformed inputs.
func ToAPIEndpoint(url string) string {
	u, err := parseURL(url)
	if err != nil {
		return url
	}
	if u.Path != "" {
		u.Path = filepath.Join("v2", u.Path)
	}
	return strings.TrimPrefix(u.String(), "//")
}

func parseURL(in string) (*url.URL, error) {
	urlIn := in
	if !strings.Contains(in, "//") {
		urlIn = "//" + in
	}
	return url.Parse(urlIn)
}

// ToAPIEndpoints utilizes ToAPIEndpoint to turn all URLs from a
// map to valid API endpoints for a local registry.
func ToAPIEndpoints(URLs map[string]string) map[string]string {
	endpoints := make(map[string]string)
	for key, url := range URLs {
		endpoints[key] = ToAPIEndpoint(url)
	}
	return endpoints
}
