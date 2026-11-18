package framework

import (
	_ "embed"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
	// PrometheusDefaultScrapeJob is the scrape job the Prometheus chart configures by default.
	PrometheusDefaultScrapeJob = "prometheus"
	// PrometheusExtraScrapeConfigsJob is the job added through extraScrapeConfigs by
	// InstallPrometheusPackageWithCustomScrapeConfigs.
	PrometheusExtraScrapeConfigsJob = "eksa-e2e-extra-scrape-configs"
	// PrometheusServerFilesScrapeJob is the job added through serverFiles."prometheus.yml".scrape_configs
	// by InstallPrometheusPackageWithCustomScrapeConfigs.
	PrometheusServerFilesScrapeJob = "eksa-e2e-server-files-scrape-configs"

	prometheusServerPort = 9090
)

//go:embed testdata/prometheus_scrape_pools_check.sh
var prometheusScrapePoolsCheckScript string

//go:embed testdata/prometheus_package_custom_scrape_configs.yaml
var prometheusPackageCustomScrapeConfigsTemplate string

// InstallPrometheusPackageWithCustomScrapeConfigs installs the Prometheus package with a scrape job from
// extraScrapeConfigs and one from serverFiles."prometheus.yml".scrape_configs.
func (e *ClusterE2ETest) InstallPrometheusPackageWithCustomScrapeConfigs(packageName, targetNamespace string) {
	pkg, err := templater.Execute(prometheusPackageCustomScrapeConfigsTemplate, map[string]interface{}{
		"name":                  packageName,
		"targetNamespace":       targetNamespace,
		"extraScrapeConfigsJob": PrometheusExtraScrapeConfigsJob,
		"serverFilesJob":        PrometheusServerFilesScrapeJob,
	})
	if err != nil {
		e.T.Fatalf("Rendering Prometheus package with custom scrape configs: %v", err)
	}
	e.ApplyPackageFile(packageName, targetNamespace, pkg)
}

// VerifyPrometheusScrapeJobsActive checks from inside the cluster that the package's Prometheus server is
// ready and that /api/v1/targets has an active scrape pool for every job.
// packageName must contain "prometheus" so the server service is named <packageName>-server.
func (e *ClusterE2ETest) VerifyPrometheusScrapeJobsActive(packageName, targetNamespace string, jobs ...string) {
	url := fmt.Sprintf("http://%s-server.%s.svc.cluster.local:%d", packageName, targetNamespace, prometheusServerPort)
	e.T.Logf("Verifying Prometheus %s has active scrape pools %v", url, jobs)
	container := corev1.Container{
		Name:    "prometheus-check",
		Image:   e.GetDiagnosticCollectorImage(),
		Command: append([]string{"bash", "-c", prometheusScrapePoolsCheckScript, "prometheus-check", url}, jobs...),
	}
	e.runCheckJob(targetNamespace, "prometheus-check", container, 12*time.Minute)
}
