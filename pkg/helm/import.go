package helm

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/utils/oci"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

type ChartRegistryImporter struct {
	client             Client
	registry           string
	username, password string
	srcFolder          string
}

func NewChartRegistryImporter(client Client, srcFolder, registry, username, password string) *ChartRegistryImporter {
	return &ChartRegistryImporter{
		client:    client,
		srcFolder: srcFolder,
		registry:  registry,
		username:  username,
		password:  password,
	}
}

func (i *ChartRegistryImporter) Import(ctx context.Context, charts ...string) error {
	// Extract host:port for registry login. Newer Helm versions do not accept
	// a path component in the registry argument for login or when validating
	// the --insecure-skip-tls-verify flag. The path/namespace is only used when
	// constructing the OCI push URL.
	registryHost, registryPath := splitRegistryHostAndPath(i.registry)

	if err := i.client.RegistryLogin(ctx, registryHost, i.username, i.password); err != nil {
		return fmt.Errorf("importing charts: %v", err)
	}

	for _, chart := range uniqueCharts(charts) {
		pushChartURL := oci.ChartPushURL(chart)
		pushChartURL = urls.ReplaceHost(pushChartURL, registryHost)

		// If the registry includes a namespace path (e.g., "host:port/namespace"),
		// inject the namespace into the OCI URL path after the host.
		if registryPath != "" {
			pushChartURL = injectNamespace(pushChartURL, registryPath)
		}

		chartFilepath := filepath.Join(i.srcFolder, ChartFileName(chart))
		if err := i.client.PushChart(ctx, chartFilepath, pushChartURL); err != nil {
			return fmt.Errorf("pushing chart [%s] to registry [%s]: %v", chart, i.registry, err)
		}
	}

	return nil
}

// splitRegistryHostAndPath splits a registry string into host:port and path components.
// For example:
//
//	"192.168.1.1:443/my-namespace" -> ("192.168.1.1:443", "my-namespace")
//	"192.168.1.1:443"             -> ("192.168.1.1:443", "")
//	"myregistry.com/namespace"    -> ("myregistry.com", "namespace")
func splitRegistryHostAndPath(registry string) (host, path string) {
	return SplitRegistryHostAndPath(registry)
}

// SplitRegistryHostAndPath splits a registry string into host:port and path components.
// For example:
//
//	"192.168.1.1:443/my-namespace" -> ("192.168.1.1:443", "my-namespace")
//	"192.168.1.1:443"             -> ("192.168.1.1:443", "")
//	"myregistry.com/namespace"    -> ("myregistry.com", "namespace")
func SplitRegistryHostAndPath(registry string) (host, path string) {
	// Find the first slash that is not part of a scheme (no "://" present in registry values)
	slashIdx := strings.Index(registry, "/")
	if slashIdx == -1 {
		return registry, ""
	}
	return registry[:slashIdx], registry[slashIdx+1:]
}

// injectNamespace inserts a namespace path segment into an OCI URL right after the host.
// For example:
//
//	injectNamespace("oci://192.168.1.1:443/eks/cilium", "my-namespace")
//	-> "oci://192.168.1.1:443/my-namespace/eks/cilium"
func injectNamespace(ociURL, namespace string) string {
	return InjectNamespace(ociURL, namespace)
}

// InjectNamespace inserts a namespace path segment into an OCI URL right after the host.
// For example:
//
//	InjectNamespace("oci://192.168.1.1:443/eks/cilium", "my-namespace")
//	-> "oci://192.168.1.1:443/my-namespace/eks/cilium"
func InjectNamespace(ociURL, namespace string) string {
	if namespace == "" {
		return ociURL
	}
	const prefix = oci.OCIPrefix
	if strings.HasPrefix(ociURL, prefix) {
		withoutPrefix := strings.TrimPrefix(ociURL, prefix)
		// Find the first slash separating host from path
		slashIdx := strings.Index(withoutPrefix, "/")
		if slashIdx == -1 {
			// No path, just append namespace
			return prefix + withoutPrefix + "/" + namespace
		}
		host := withoutPrefix[:slashIdx]
		path := withoutPrefix[slashIdx+1:]
		return prefix + host + "/" + namespace + "/" + path
	}
	return ociURL
}
