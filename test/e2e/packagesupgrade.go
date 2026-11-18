//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"time"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	bundleUpgradeHarborPrefix           = "upgrade"
	bundleUpgradeHarborNamespace        = "harbor"
	bundleUpgradeHarborRepository       = "library/eksa-e2e-bundle-upgrade"
	bundleUpgradeObservabilityNamespace = "observability"
	bundleUpgradePrometheusName         = "upgrade-prometheus"
	bundleUpgradePrometheusScrapeJob    = "eksa-e2e-bundle-upgrade-scrape"
	bundleUpgradeAdotName               = "upgrade-adot"
	bundleUpgradeMetalLBName            = "upgrade-metallb"
	bundleUpgradeMetalLBNamespace       = "metallb-system"
	bundleUpgradeMetalLBPool            = "eksa-e2e-bundle-upgrade"
	bundleUpgradeMetalLBAddresses       = "10.100.100.8/29"
	bundleUpgradeLoadBalancerService    = "eksa-e2e-bundle-upgrade-lb"
	bundleUpgradeLoadBalancerNamespace  = "default"
	// hello-eks-anywhere, Emissary, cluster-autoscaler and metrics-server go in eksa-packages, where
	// their verify helpers look for them. The autoscaler and metrics-server helpers also expect these
	// Package names.
	bundleUpgradeHelloName         = "upgrade-hello-eks-anywhere"
	bundleUpgradeHelloSubtitle     = "eksa-e2e-bundle-upgrade-subtitle"
	bundleUpgradeEmissaryName      = "upgrade-emissary"
	bundleUpgradeEmissaryRoute     = "eksa-e2e-bundle-upgrade"
	bundleUpgradeAutoscalerName    = "cluster-autoscaler"
	bundleUpgradeMetricsServerName = "metrics-server"
)

// runCuratedPackagesBundleUpgradeFlow installs every curated package except cert-manager from an older
// package bundle, creates state in each, activates the newest bundle, and verifies that every package is
// upgraded to the newest bundle's version and keeps its state.
//
// The newest bundle is the one the controller activated at cluster creation. T_PACKAGES_UPGRADE_FROM_BUNDLE picks
// the older bundle, usually "prod" for the current prod bundle; see framework.PackagesUpgradeFromBundleVar.
// The cluster needs worker node autoscaling from 1 to 2 nodes for the cluster-autoscaler check.
func runCuratedPackagesBundleUpgradeFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(func(test *framework.ClusterE2ETest) {
		ctx := context.Background()
		if err := WaitForPackageToBeInstalled(test, ctx, "eks-anywhere-packages", 15*time.Minute); err != nil {
			test.T.Fatalf("packages controller not in installed state: %s", err)
		}
		test.ValidatePackageBundleControllerRegistry()

		// Package object names, keyed by package name in the bundle.
		packages := map[string]string{
			"harbor":             bundleUpgradeHarborPrefix,
			"prometheus":         bundleUpgradePrometheusName,
			"metallb":            bundleUpgradeMetalLBName,
			"hello-eks-anywhere": bundleUpgradeHelloName,
			"emissary":           bundleUpgradeEmissaryName,
			"adot":               bundleUpgradeAdotName,
			"cluster-autoscaler": bundleUpgradeAutoscalerName,
			"metrics-server":     bundleUpgradeMetricsServerName,
		}

		newest := test.WaitForActivePackageBundle()
		previous := test.PackageBundleToUpgradeFrom(newest)
		test.ActivatePackageBundle(previous)
		test.WaitForPackageBundleActive(previous)
		test.WaitForPackagesControllerAtBundle(previous)
		fromVersions := map[string]packagesv1.SourceVersion{}
		for pkg := range packages {
			fromVersions[pkg] = test.PackageBundleVersion(previous, pkg)
		}

		test.InstallLocalStorageProvisioner()
		for _, ns := range []string{bundleUpgradeHarborNamespace, bundleUpgradeObservabilityNamespace, bundleUpgradeMetalLBNamespace} {
			test.CreateNamespace(ns)
		}
		installHarborWithLocalStorage(test, bundleUpgradeHarborPrefix, bundleUpgradeHarborNamespace)
		test.ApplyBundleUpgradeTestPackages(framework.BundleUpgradeTestPackages{
			PrometheusName:      bundleUpgradePrometheusName,
			PrometheusNamespace: bundleUpgradeObservabilityNamespace,
			PrometheusScrapeJob: bundleUpgradePrometheusScrapeJob,
			MetalLBName:         bundleUpgradeMetalLBName,
			MetalLBNamespace:    bundleUpgradeMetalLBNamespace,
			MetalLBPool:         bundleUpgradeMetalLBPool,
			MetalLBAddresses:    bundleUpgradeMetalLBAddresses,
			HelloName:           bundleUpgradeHelloName,
			HelloNamespace:      EksaPackagesNamespace,
			HelloSubtitle:       bundleUpgradeHelloSubtitle,
			EmissaryName:        bundleUpgradeEmissaryName,
			EmissaryNamespace:   EksaPackagesNamespace,
			AdotName:            bundleUpgradeAdotName,
			AdotNamespace:       bundleUpgradeObservabilityNamespace,
		})
		test.InstallAutoScalerWithMetricServer(EksaPackagesNamespace)
		for pkg, name := range packages {
			test.WaitForPackageInstalledAtVersion(name, fromVersions[pkg], 20*time.Minute)
		}

		test.T.Logf("Creating package state with bundle %s", previous)
		test.VerifyHarborPackageInstalled(bundleUpgradeHarborPrefix, bundleUpgradeHarborNamespace)
		test.PushHarborTestImage(bundleUpgradeHarborPrefix, bundleUpgradeHarborNamespace, bundleUpgradeHarborRepository)
		test.VerifyPrometheusPackageInstalled(bundleUpgradePrometheusName, bundleUpgradeObservabilityNamespace)
		test.VerifyPrometheusScrapeJobsActive(bundleUpgradePrometheusName, bundleUpgradeObservabilityNamespace,
			bundleUpgradePrometheusScrapeJob, framework.PrometheusDefaultScrapeJob)
		test.VerifyPackageImagesFromConfiguredRegistry(bundleUpgradeMetalLBName, bundleUpgradeMetalLBNamespace, test.Cluster())
		lbIP := createBundleUpgradeLoadBalancerService(test, ctx)
		test.VerifyHelloPackageInstalled(bundleUpgradeHelloName, test.Cluster())
		test.VerifyEmissaryPackageInstalled(bundleUpgradeEmissaryName, test.Cluster())
		test.ApplyEmissaryRoute(bundleUpgradeEmissaryRoute, EksaPackagesNamespace, bundleUpgradeHelloName)
		verifyBundleUpgradeEmissaryRoute(test)
		test.VerifyAdotPackageInstalled(bundleUpgradeAdotName, bundleUpgradeObservabilityNamespace)
		test.VerifyMetricServerPackageInstalled(bundleUpgradeMetricsServerName, EksaPackagesNamespace, test.Cluster())
		test.VerifyMetricsAPIServing()
		test.VerifyAutoScalerPackageInstalled(bundleUpgradeAutoscalerName, EksaPackagesNamespace, test.Cluster())

		test.ActivatePackageBundle(newest)
		test.WaitForPackagesControllerAtBundle(newest)
		for pkg, name := range packages {
			from, to := fromVersions[pkg], test.PackageBundleVersion(newest, pkg)
			if from.Digest == to.Digest {
				test.T.Logf("Package %s is the same chart (%s) in %s and %s, only checking its state survives the bundle switch", pkg, to.Name, previous, newest)
			} else {
				test.T.Logf("Package %s should upgrade from %s (%s) to %s (%s)", pkg, from.Name, from.Digest, to.Name, to.Digest)
			}
			test.WaitForPackageInstalledAtVersion(name, to, 30*time.Minute)
		}

		test.T.Logf("Verifying package state after upgrading to bundle %s", newest)
		test.VerifyHarborPackageInstalled(bundleUpgradeHarborPrefix, bundleUpgradeHarborNamespace)
		test.VerifyHarborTestImagePullable(bundleUpgradeHarborPrefix, bundleUpgradeHarborNamespace, bundleUpgradeHarborRepository)

		test.VerifyPrometheusPrometheusServerStates(bundleUpgradePrometheusName, bundleUpgradeObservabilityNamespace, "deployment")
		test.VerifyPrometheusPackageInstalled(bundleUpgradePrometheusName, bundleUpgradeObservabilityNamespace)
		test.VerifyPrometheusScrapeJobsActive(bundleUpgradePrometheusName, bundleUpgradeObservabilityNamespace,
			bundleUpgradePrometheusScrapeJob, framework.PrometheusDefaultScrapeJob)

		verifyBundleUpgradeMetalLBState(test, ctx, lbIP)

		test.VerifyHelloPackageInstalled(bundleUpgradeHelloName, test.Cluster())
		test.VerifyEmissaryPackageInstalled(bundleUpgradeEmissaryName, test.Cluster())
		verifyBundleUpgradeEmissaryRoute(test)

		// The ADOT check reads the logs of any collector pod, so wait for the old pods to go away.
		if err := test.KubectlClient.WaitForResourceRolledout(ctx, test.Cluster(), "10m",
			bundleUpgradeAdotName+"-aws-otel-collector", bundleUpgradeObservabilityNamespace, "deployment"); err != nil {
			test.T.Fatalf("waiting for the ADOT collector rollout after the upgrade: %s", err)
		}
		test.VerifyAdotPackageInstalled(bundleUpgradeAdotName, bundleUpgradeObservabilityNamespace)

		test.VerifyMetricsAPIServing()
		// Last, because the scale-up load stays on the cluster. This also checks that metrics-server and
		// cluster-autoscaler are installed.
		test.CombinedAutoScalerMetricServerTest(bundleUpgradeAutoscalerName, bundleUpgradeMetricsServerName,
			EksaPackagesNamespace, test.Cluster())
	})
}

// verifyBundleUpgradeEmissaryRoute checks that Emissary routes to the hello-eks-anywhere package and that
// hello still serves its configured subtitle.
func verifyBundleUpgradeEmissaryRoute(test *framework.ClusterE2ETest) {
	test.VerifyEndpointServesContent(EksaPackagesNamespace,
		framework.EmissaryRouteURL(bundleUpgradeEmissaryName, EksaPackagesNamespace, bundleUpgradeEmissaryRoute),
		bundleUpgradeHelloSubtitle)
}

// createBundleUpgradeLoadBalancerService creates a LoadBalancer Service and returns the IP MetalLB assigns it.
func createBundleUpgradeLoadBalancerService(test *framework.ClusterE2ETest, ctx context.Context) string {
	test.CreateResource(ctx, fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
`, bundleUpgradeLoadBalancerService, bundleUpgradeLoadBalancerNamespace))

	var ip string
	err := WaitForResource(test, ctx, "service/"+bundleUpgradeLoadBalancerService, bundleUpgradeLoadBalancerNamespace,
		"{.status.loadBalancer.ingress[0].ip}", 5*time.Minute, func(out string, err error) bool {
			ip = out
			return err == nil && out != ""
		})
	if err != nil {
		test.T.Fatalf("waiting for MetalLB to assign an IP to service %s: %s", bundleUpgradeLoadBalancerService, err)
	}
	test.T.Logf("MetalLB assigned %s to service %s", ip, bundleUpgradeLoadBalancerService)
	return ip
}

// verifyBundleUpgradeMetalLBState checks that the upgraded MetalLB still has the IPAddressPool from the
// package config, that the LoadBalancer Service kept lbIP, and that its images come from the registry.
func verifyBundleUpgradeMetalLBState(test *framework.ClusterE2ETest, ctx context.Context, lbIP string) {
	if err := test.KubectlClient.WaitForDeployment(ctx, test.Cluster(), "10m", "Available",
		bundleUpgradeMetalLBName+"-controller", bundleUpgradeMetalLBNamespace); err != nil {
		test.T.Fatalf("waiting for the MetalLB controller after the upgrade: %s", err)
	}
	if err := WaitForResource(test, ctx, "ipaddresspools.metallb.io/"+bundleUpgradeMetalLBPool, bundleUpgradeMetalLBNamespace,
		"{.spec.addresses[0]}", 5*time.Minute, StringMatchPredicate(bundleUpgradeMetalLBAddresses)); err != nil {
		test.T.Fatalf("IPAddressPool %s missing after the upgrade: %s", bundleUpgradeMetalLBPool, err)
	}
	// Give the new controller time to reconcile services before checking the IP didn't change.
	time.Sleep(30 * time.Second)
	if err := WaitForResource(test, ctx, "service/"+bundleUpgradeLoadBalancerService, bundleUpgradeLoadBalancerNamespace,
		"{.status.loadBalancer.ingress[0].ip}", 2*time.Minute, StringMatchPredicate(lbIP)); err != nil {
		test.T.Fatalf("service %s did not keep IP %s after the upgrade: %s", bundleUpgradeLoadBalancerService, lbIP, err)
	}
	test.VerifyPackageImagesFromConfiguredRegistry(bundleUpgradeMetalLBName, bundleUpgradeMetalLBNamespace, test.Cluster())
}
