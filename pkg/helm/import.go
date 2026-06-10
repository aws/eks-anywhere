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
	host, namespace := splitRegistryNamespace(i.registry)

	if err := i.client.RegistryLogin(ctx, host, i.username, i.password); err != nil {
		return fmt.Errorf("importing charts: %v", err)
	}

	for _, chart := range uniqueCharts(charts) {
		pushChartURL := oci.ChartPushURL(chart)
		pushChartURL = urls.ReplaceHost(pushChartURL, host)
		if namespace != "" {
			pushChartURL = insertNamespace(pushChartURL, namespace)
		}

		chartFilepath := filepath.Join(i.srcFolder, ChartFileName(chart))
		if err := i.client.PushChart(ctx, chartFilepath, pushChartURL); err != nil {
			return fmt.Errorf("pushing chart [%s] to registry [%s]: %v", chart, i.registry, err)
		}
	}

	return nil
}

// splitRegistryNamespace separates a registry string into a host:port and an
// optional namespace. For example "192.168.1.1:443/eks-a-test" returns
// ("192.168.1.1:443", "eks-a-test") and "192.168.1.1:443" returns
// ("192.168.1.1:443", "").
func splitRegistryNamespace(registry string) (host, namespace string) {
	slashIdx := strings.Index(registry, "/")
	if slashIdx == -1 {
		return registry, ""
	}
	return registry[:slashIdx], registry[slashIdx+1:]
}

// insertNamespace inserts a namespace path segment after the host in an OCI URL.
// For example, given "oci://192.168.1.1:443/project" and namespace "eks-a-test",
// it returns "oci://192.168.1.1:443/eks-a-test/project".
func insertNamespace(ociURL, namespace string) string {
	prefix := oci.OCIPrefix
	if !strings.HasPrefix(ociURL, prefix) {
		return namespace + "/" + ociURL
	}
	withoutScheme := strings.TrimPrefix(ociURL, prefix)
	slashIdx := strings.Index(withoutScheme, "/")
	if slashIdx == -1 {
		return prefix + withoutScheme + "/" + namespace
	}
	return prefix + withoutScheme[:slashIdx] + "/" + namespace + withoutScheme[slashIdx:]
}
