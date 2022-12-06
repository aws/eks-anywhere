package registrymirror

import (
	"net"
	urllib "net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// RegistryMirror configures mirror mappings for artifact registries.
type RegistryMirror struct {
	// the address of registry mirror without namespace
	BaseRegistry string
	// it stores mirror mappings for artifact registries
	NamespacedRegistryMap map[string]string
	// if authentication is required for the registry mirror
	Auth bool
}

var re = regexp.MustCompile(constants.DefaultCuratedPackagesRegistryRegex)

// FromCluster is a constructor for RegistryMirror from a cluster schema.
func FromCluster(cluster *v1alpha1.Cluster) *RegistryMirror {
	return FromClusterRegistryMirrorConfiguration(cluster.Spec.RegistryMirrorConfiguration)
}

// FromClusterRegistryMirrorConfiguration is a constructor for RegistryMirror from a RegistryMirrorConfiguration schema.
func FromClusterRegistryMirrorConfiguration(config *v1alpha1.RegistryMirrorConfiguration) *RegistryMirror {
	if config == nil {
		return nil
	}
	registryMap := make(map[string]string)
	base := net.JoinHostPort(config.Endpoint, config.Port)
	// add registry mirror base address
	// for each namespace, add corresponding endpoint
	for _, ociNamespace := range config.OCINamespaces {
		mirror := filepath.Join(base, ociNamespace.Namespace)
		if re.MatchString(ociNamespace.Registry) {
			// handle curated packages in all regions
			// static key makes it easier for mirror lookup
			registryMap[constants.DefaultCuratedPackagesRegistryRegex] = mirror
		} else {
			registryMap[ociNamespace.Registry] = mirror
		}
	}
	if len(registryMap) == 0 {
		// for backward compatibility, default mapping for public.ecr.aws is added
		// when no namespace mapping is specified
		registryMap[constants.DefaultCoreEKSARegistry] = base
	}
	return &RegistryMirror{
		BaseRegistry:          base,
		NamespacedRegistryMap: registryMap,
		Auth:                  config.Authenticate,
	}
}

// CoreEKSAMirror returns the configured mirror for public.ecr.aws.
func (r *RegistryMirror) CoreEKSAMirror() string {
	if v, ok := r.NamespacedRegistryMap[constants.DefaultCoreEKSARegistry]; ok {
		return v
	}
	// TODO: handle this case after BottleRocket supports multiple registry mirros
	return ""
}

// CuratedPackagesMirror returns the mirror for curated packages.
func (r *RegistryMirror) CuratedPackagesMirror() string {
	if v, ok := r.NamespacedRegistryMap[constants.DefaultCuratedPackagesRegistryRegex]; ok {
		return v
	}
	return ""
}

// ReplaceRegistry replaces the host in a url with corresponding registry mirror
// It supports full URLs and container image URLs
// If the provided original url is malformed, there are no guarantees
// that the returned value will be valid
// If no corresponding registry mirror, it will return the original URL.
func (r *RegistryMirror) ReplaceRegistry(url string) string {
	if r == nil {
		return url
	}

	u, _ := urllib.Parse(url)
	if u.Scheme == "" {
		u, _ = urllib.Parse("oci://" + url)
		u.Scheme = ""
	}
	key := u.Host
	if re.MatchString(key) {
		key = constants.DefaultCuratedPackagesRegistryRegex
	}
	if v, ok := r.NamespacedRegistryMap[key]; ok {
		return strings.Replace(url, u.Host, v, 1)
	}
	return url
}
