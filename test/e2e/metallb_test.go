//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	cluster *framework.ClusterE2ETest
}

func TestPackagesMetalLB(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (suite *Suite) SetupSuite() {
	t := suite.T()
	suite.cluster = framework.NewClusterE2ETest(t, framework.NewDocker(t),
		framework.WithPackageConfig(t, EksaPackageBundleURI, EksaPackageControllerHelmChartName,
			EksaPackageControllerHelmURI, EksaPackageControllerHelmVersion,
			EksaPackageControllerHelmValues),
	)
}

func (suite *Suite) TestPackagesMetalLB() {
	// This should be split into multiple tests with a cluster setup in `SetupSuite`.
	// This however requires the creation of utilites managing cluster creation.
	t := suite.T()
	suite.cluster.WithPersistentCluster(func(test *framework.ClusterE2ETest) {
		test.InstallCuratedPackagesController()
		ctx := context.Background()
		namespace := "metallb-system"
		t.Cleanup(func() { test.DeleteNamespace(namespace) })
		test.CreateNamespace(namespace)
		packageName := "metallb"
		packagePrefix := "test"
		err := WaitForLatestBundleToBeAvailable(test, ctx, 2*time.Minute)
		if err != nil {
			t.Fatalf("waiting for latest bundle: %s", err)
		}
		bundle, err := GetLatestBundleFromCluster(test)
		if err != nil {
			t.Fatal(err)
		}
		UpgradePackages(test, bundle)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("Basic installation", func(t *testing.T) {
			t.Cleanup(func() { test.UninstallCuratedPackage(packagePrefix) })
			test.InstallCuratedPackage(packageName, packagePrefix, "--kube-version=1.22")
			err = WaitForPackageToBeInstalled(test, ctx, packagePrefix, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb package to be installed: %s", err)
			}
			err = test.KubectlClient.WaitForDeployment(context.Background(),
				kubeconfig.FromClusterName(test.ClusterName), "1m", "Available", "test-metallb-controller", namespace)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
			err = WaitForDaemonset(test, ctx, "test-metallb-speaker", namespace, 2, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
		})

		t.Run("Address pool configuration", func(t *testing.T) {
			address := "10.100.100.1"
			t.Cleanup(func() { test.UninstallCuratedPackage(packagePrefix) })
			test.CreateResource(ctx, fmt.Sprintf(
				`
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: test
  namespace: eksa-packages
spec:
  packageName: metallb
  config: |
    IPAddressPools:
      - name: default
        addresses:
          - %s/32
    L2Advertisements:
      - IPAddressPools:
        - default
`, address))
			err = WaitForPackageToBeInstalled(test, ctx, packagePrefix, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb package to be installed: %s", err)
			}
			err = test.KubectlClient.WaitForDeployment(context.Background(),
				kubeconfig.FromClusterName(test.ClusterName), "1m", "Available", "test-metallb-controller", namespace)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
			err = WaitForDaemonset(test, ctx, "test-metallb-speaker", namespace, 2, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb speaker deployment to be available: %s", err)
			}

			expectedConfig := fmt.Sprintf(
				`address-pools:
  - name: default
    addresses:
      - %s/32
    protocol: layer2
`, address)
			err = WaitForResource(
				test,
				ctx,
				"configmap/test-metallb",
				namespace,
				"{.data.config}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedConfig),
			)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				test.KubectlClient.Delete(ctx, "service", "my-service", "default", kubeconfig.FromClusterName(test.ClusterName))
			})
			test.CreateResource(ctx, `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
`)
			err = WaitForResource(
				test,
				ctx,
				"service/my-service",
				"default",
				"{.status.loadBalancer.ingress[0].ip}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(address),
			)
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("BGP configuration", func(t *testing.T) {
			address := "10.100.100.2"
			t.Cleanup(func() { test.UninstallCuratedPackage(packagePrefix) })
			test.CreateResource(ctx, fmt.Sprintf(
				`
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: test
  namespace: eksa-packages
spec:
  packageName: metallb
  config: |
    IPAddressPools:
      - name: default
        addresses:
          - 10.100.0.1/32
        autoAssign: false
      - name: bgp
        addresses:
          - %s/32
    L2Advertisements:
      - IPAddressPools:
        - default
    BGPAdvertisements:
      - IPAddressPools:
          - bgp
        localPref: 123
        aggregationLength: 32
        aggregationLengthV6: 32
    BGPPeers:
      - myASN: 123
        peerASN: 55001
        peerAddress: 12.2.4.2
        keepaliveTime: 30s
`, address))
			err = WaitForPackageToBeInstalled(test, ctx, packagePrefix, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb package to be installed: %s", err)
			}
			err = test.KubectlClient.WaitForDeployment(context.Background(),
				kubeconfig.FromClusterName(test.ClusterName), "1m", "Available", "test-metallb-controller", namespace)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
			err = WaitForDaemonset(test, ctx, "test-metallb-speaker", namespace, 2, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb speaker deployment to be available: %s", err)
			}

			expectedConfig := fmt.Sprintf(
				`address-pools:
  - name: default
    addresses:
      - 10.100.0.1/32
    protocol: layer2
    auto-assign: false
  - name: bgp
    addresses:
      - %s/32
    protocol: bgp
    bgp-advertisements:
      - aggregation-length: 32
        aggregation-length-v6: 32
        localpref: 123
peers:
  - peer-address: 12.2.4.2
    peer-asn: 55001
    my-asn: 123
    keepalive-time: 30s
`, address)
			err = WaitForResource(
				test,
				ctx,
				"configmap/test-metallb",
				namespace,
				"{.data.config}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedConfig),
			)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				test.KubectlClient.Delete(ctx, "service", "my-service", "default", kubeconfig.FromClusterName(test.ClusterName))
			})
			test.CreateResource(ctx, `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
`)
			err = WaitForResource(
				test,
				ctx,
				"service/my-service",
				"default",
				"{.status.loadBalancer.ingress[0].ip}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(address),
			)
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}
