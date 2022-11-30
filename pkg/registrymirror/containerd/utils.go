package containerd

import (
	"net/url"
	"strings"
)

// ToAPIEndpoint turns URL to a valid API endpoint used in
// a containerd config file for a local registry.
func ToAPIEndpoint(URL string) string {
	cnt := strings.Count(URL, "/")
	index := -1
	if cnt > 1 {
		u, _ := url.Parse(URL)
		if u.Scheme == "" {
			u, _ = url.Parse("oci://" + URL)
			u.Scheme = ""
		}
		if u.Path != "" {
			index = strings.Index(URL, u.Path)
		}
	} else {
		index = strings.Index(URL, "/")
	}
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
