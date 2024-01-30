package clusterapi_test

import (
	"testing"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestSetUbuntuConfigInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	eksaVersion := anywherev1.EksaVersion("v0.19.2")
	g.clusterSpec.Cluster.Spec.EksaVersion = &eksaVersion
	got := wantEtcdCluster()
	versionBundle := g.clusterSpec.VersionsBundles["1.21"]

	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format("cloud-config")
	want.Spec.EtcdadmConfigSpec.CloudInitConfig = &etcdbootstrapv1.CloudInitConfig{
		Version:        versionBundle.KubeDistro.EtcdVersion,
		InstallDir:     "/usr/bin",
		EtcdReleaseURL: versionBundle.KubeDistro.EtcdURL,
	}
	clusterapi.SetUbuntuConfigInEtcdCluster(got, versionBundle, string(eksaVersion))
	g.Expect(got).To(Equal(want))
}

func TestSetUbuntuConfigInEtcdClusterNoEtcdUrl(t *testing.T) {
	g := newApiBuilerTest(t)
	eksaVersion := anywherev1.EksaVersion("v0.18.2")
	g.clusterSpec.Cluster.Spec.EksaVersion = &eksaVersion
	got := wantEtcdCluster()
	versionBundle := g.clusterSpec.VersionsBundles["1.21"]

	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format("cloud-config")
	want.Spec.EtcdadmConfigSpec.CloudInitConfig = &etcdbootstrapv1.CloudInitConfig{
		Version:    versionBundle.KubeDistro.EtcdVersion,
		InstallDir: "/usr/bin",
	}
	clusterapi.SetUbuntuConfigInEtcdCluster(got, versionBundle, string(eksaVersion))
	g.Expect(got).To(Equal(want))
}

func TestClusterUnstackedEtcd(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{
		Count: 3,
	}
	got := clusterapi.Cluster(tt.clusterSpec, tt.providerCluster, tt.controlPlane, tt.unstackedEtcdCluster)
	want := wantCluster()
	want.Spec.ManagedExternalEtcdRef = &v1.ObjectReference{
		APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
		Kind:       "UnstackedEtcdCluster",
		Name:       "unstacked-etcd-cluster",
		Namespace:  "eksa-system",
	}
	tt.Expect(got).To(Equal(want))
}

func TestSetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket(t *testing.T) {
	tt := newApiBuilerTest(t)
	etcdConfig := &anywherev1.ExternalEtcdConfiguration{
		Count: 3,
	}
	got := wantKubeadmControlPlane()
	got.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = &bootstrapv1.ExternalEtcd{
		Endpoints: []string{},
		CAFile:    "/var/lib/kubeadm/pki/etcd/ca.crt",
		CertFile:  "/var/lib/kubeadm/pki/server-etcd-client.crt",
		KeyFile:   "/var/lib/kubeadm/pki/apiserver-etcd-client.key",
	}
	want := got.DeepCopy()
	clusterapi.SetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket(got, etcdConfig)
	tt.Expect(got).To(Equal(want))
}

func TestSetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu(t *testing.T) {
	tt := newApiBuilerTest(t)
	etcdConfig := &anywherev1.ExternalEtcdConfiguration{
		Count: 3,
	}
	got := wantKubeadmControlPlane()
	got.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = &bootstrapv1.ExternalEtcd{
		Endpoints: []string{},
		CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
		CertFile:  "/etc/kubernetes/pki/apiserver-etcd-client.crt",
		KeyFile:   "/etc/kubernetes/pki/apiserver-etcd-client.key",
	}
	want := got.DeepCopy()
	clusterapi.SetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu(got, etcdConfig)
	tt.Expect(got).To(Equal(want))
}

func TestSetStackedEtcdConfigInKubeadmControlPlane(t *testing.T) {
	tt := newApiBuilerTest(t)
	want := wantKubeadmControlPlane()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local = &bootstrapv1.LocalEtcd{
		ImageMeta: bootstrapv1.ImageMeta{
			ImageRepository: "public.ecr.aws/eks-distro/etcd-io",
			ImageTag:        "v3.4.16-eks-1-21-9",
		},
		ExtraArgs: map[string]string{
			"cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		},
	}
	got, err := clusterapi.KubeadmControlPlane(tt.clusterSpec, tt.providerMachineTemplate)
	tt.Expect(err).To(Succeed())
	tt.Expect(got).To(Equal(want))
}
