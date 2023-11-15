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
	// BaseRegistry is the address of the registry mirror without namespace. Just the host and the port.
	BaseRegistry string
	// NamespacedRegistryMap stores mirror mappings for artifact registries
	NamespacedRegistryMap map[string]string
	// Auth should be marked as true if authentication is required for the registry mirror
	Auth bool
	// CACertContent defines the contents registry mirror CA certificate
	CACertContent string
	// InsecureSkipVerify skips the registry certificate verification.
	// Only use this solution for isolated testing or in a tightly controlled, air-gapped environment.
	InsecureSkipVerify bool
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
			registryMap[constants.DefaultCuratedPackagesRegistry] = mirror
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
		CACertContent:         config.CACertContent,
		InsecureSkipVerify:    config.InsecureSkipVerify,
	}
}

// CoreEKSAMirror returns the configured mirror for public.ecr.aws.
func (r *RegistryMirror) CoreEKSAMirror() string {
	return r.NamespacedRegistryMap[constants.DefaultCoreEKSARegistry]
}

// CuratedPackagesMirror returns the mirror for curated packages.
func (r *RegistryMirror) CuratedPackagesMirror() string {
	return r.NamespacedRegistryMap[constants.DefaultCuratedPackagesRegistry]
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
		key = constants.DefaultCuratedPackagesRegistry
	}
	if v, ok := r.NamespacedRegistryMap[key]; ok {
		return strings.Replace(url, u.Host, v, 1)
	}
	return url
}
