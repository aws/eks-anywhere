package urls

import (
	"net/url"
	"strings"
)

// ReplaceHost replaces the host in a url
// It supports full URLs and container image URLs
// If the provided original url is malformed, the are no guarantees
// that the returned value will be valid
// If host is empty, it will return the original URL.
func ReplaceHost(orgURL, host string) string {
	if host == "" {
		return orgURL
	}

	u, _ := url.Parse(orgURL)
	if u.Scheme == "" {
		u, _ = url.Parse("oci://" + orgURL)
		u.Scheme = ""
	}
	u.Host = host
	return strings.ReplaceAll(strings.TrimPrefix(u.String(), "//"), "%2F", "/")
}

// ToAPIEndpoint turns URL to a valid API endpoint used in
// a containerd config file for a local registry.
func ToAPIEndpoint(URL string) string {
	index := strings.Index(URL, "/")
	if index == -1 {
		return URL
	}
	return URL[:index] + "/v2" + URL[index:]
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
