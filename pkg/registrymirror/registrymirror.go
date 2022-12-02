package registrymirror

import (
	"net/url"
	"regexp"
	"strings"
)

// RegistryMirror constructs a registry mirror mapping
// from Cluster.Spec.RegistryMirrorConfiguration.
type RegistryMirror struct {
	BaseRegistry          string
	NamespacedRegistryMap map[string]string
	Auth                  bool
}

// RegistryMirrorWithOCINamespace returns the mirror for public.ecr.aws.
func (r *RegistryMirror) RegistryMirrorWithOCINamespace() string {
	if v, ok := r.NamespacedRegistryMap[DefaultRegistry]; ok {
		return v
	}
	return r.BaseRegistry
}

// RegistryMirrorWithGatedOCINamespace returns the mirror for curated packages.
func (r *RegistryMirror) RegistryMirrorWithGatedOCINamespace() string {
	if v, ok := r.NamespacedRegistryMap[DefaultPackageRegistryRegex]; ok {
		return v
	}
	return r.BaseRegistry
}

// ReplaceRegistry replaces the host in a url with corresponding registry mirror
// It supports full URLs and container image URLs
// If the provided original url is malformed, there are no guarantees
// that the returned value will be valid
// If no corresponding registry mirror, it will return the original URL.
func (r *RegistryMirror) ReplaceRegistry(URL string) string {
	if r == nil {
		return URL
	}

	u, _ := url.Parse(URL)
	if u.Scheme == "" {
		u, _ = url.Parse("oci://" + URL)
		u.Scheme = ""
	}
	key := u.Host
	if regexp.MustCompile(DefaultPackageRegistryRegex).MatchString(key) {
		key = DefaultPackageRegistryRegex
	}
	if v, ok := r.NamespacedRegistryMap[key]; ok {
		return strings.ReplaceAll(URL, u.Host, v)
	}
	return URL
}
