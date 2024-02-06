package snow_test

import (
	"testing"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

type apiBuilerTest struct {
	*WithT
	clusterSpec    *cluster.Spec
	machineConfigs map[string]*v1alpha1.SnowMachineConfig
	logger         logr.Logger
}

func newApiBuilerTest(t *testing.T) apiBuilerTest {
	return apiBuilerTest{
		WithT:          NewWithT(t),
		clusterSpec:    givenClusterSpec(),
		machineConfigs: givenMachineConfigs(),
		logger:         test.NewNullLogger(),
	}
}

func wantCAPICluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name":                        "snow-test",
				"cluster.anywhere.eks.amazonaws.com/cluster-name":      "snow-test",
				"cluster.anywhere.eks.amazonaws.com/cluster-namespace": "test-namespace",
			},
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						"10.1.0.0/16",
					},
				},
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						"10.96.0.0/12",
					},
				},
			},
			ControlPlaneRef: &v1.ObjectReference{
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
				Kind:       "KubeadmControlPlane",
				Name:       "snow-test",
			},
			InfrastructureRef: &v1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "AWSSnowCluster",
				Name:       "snow-test",
			},
		},
	}
}

func wantCAPIClusterUnstackedEtcd() *clusterv1.Cluster {
	cluster := wantCAPICluster()
	cluster.Spec.ManagedExternalEtcdRef = &v1.ObjectReference{
		Kind:       "EtcdadmCluster",
		APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
		Namespace:  "eksa-system",
		Name:       "snow-test-etcd",
	}
	return cluster
}

func TestCAPICluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	snowCluster := snow.SnowCluster(tt.clusterSpec, wantSnowCredentialsSecret())
	controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], nil)
	kubeadmControlPlane, err := snow.KubeadmControlPlane(tt.logger, tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	got := snow.CAPICluster(tt.clusterSpec, snowCluster, kubeadmControlPlane, nil)
	tt.Expect(got).To(Equal(wantCAPICluster()))
}

func wantKubeadmControlPlane(kubeVersion v1alpha1.KubernetesVersion) *controlplanev1.KubeadmControlPlane {
	wantReplicas := int32(3)
	wantMaxSurge := intstr.FromInt(1)
	versionBundles := givenVersionsBundle(kubeVersion)
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			Kind:       "KubeadmControlPlane",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test",
			Namespace: "eksa-system",
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					Kind:       "AWSSnowMachineTemplate",
					Name:       "snow-test-control-plane-1",
				},
			},
			KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
					ImageRepository: versionBundles.KubeDistro.Kubernetes.Repository,
					DNS: bootstrapv1.DNS{
						ImageMeta: bootstrapv1.ImageMeta{
							ImageRepository: versionBundles.KubeDistro.CoreDNS.Repository,
							ImageTag:        versionBundles.KubeDistro.CoreDNS.Tag,
						},
					},
					Etcd: bootstrapv1.Etcd{
						Local: &bootstrapv1.LocalEtcd{
							ImageMeta: bootstrapv1.ImageMeta{
								ImageRepository: versionBundles.KubeDistro.Etcd.Repository,
								ImageTag:        versionBundles.KubeDistro.Etcd.Tag,
							},
							ExtraArgs: map[string]string{
								"cipher-suites":      "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"listen-peer-urls":   "https://0.0.0.0:2380",
								"listen-client-urls": "https://0.0.0.0:2379",
							},
						},
					},
					APIServer: bootstrapv1.APIServer{
						ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
							ExtraArgs:    map[string]string{},
							ExtraVolumes: []bootstrapv1.HostPathMount{},
						},
					},
					ControllerManager: bootstrapv1.ControlPlaneComponent{
						ExtraArgs:    tlsCipherSuitesArgs(),
						ExtraVolumes: []bootstrapv1.HostPathMount{},
					},
					Scheduler: bootstrapv1.ControlPlaneComponent{
						ExtraArgs:    map[string]string{},
						ExtraVolumes: []bootstrapv1.HostPathMount{},
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"provider-id":       "aws-snow:////'{{ ds.meta_data.instance_id }}'",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						},
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"provider-id":       "aws-snow:////'{{ ds.meta_data.instance_id }}'",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						},
					},
				},
				PreKubeadmCommands: []string{
					"/etc/eks/bootstrap.sh",
				},
				PostKubeadmCommands: []string{},
				Files: []bootstrapv1.File{
					{
						Path:    "/etc/kubernetes/manifests/kube-vip.yaml",
						Owner:   "root:root",
						Content: test.KubeVipTemplate,
					},
				},
			},
			RolloutStrategy: &controlplanev1.RolloutStrategy{
				Type: controlplanev1.RollingUpdateStrategyType,
				RollingUpdate: &controlplanev1.RollingUpdate{
					MaxSurge: &wantMaxSurge,
				},
			},
			Replicas: &wantReplicas,
			Version:  versionBundles.KubeDistro.Kubernetes.Tag,
		},
	}
}

func wantKubeadmControlPlaneUnstackedEtcd() *controlplanev1.KubeadmControlPlane {
	kcp := wantKubeadmControlPlane("1.21")
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd = bootstrapv1.Etcd{
		External: &bootstrapv1.ExternalEtcd{
			Endpoints: []string{},
			CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
			CertFile:  "/etc/kubernetes/pki/apiserver-etcd-client.crt",
			KeyFile:   "/etc/kubernetes/pki/apiserver-etcd-client.key",
		},
	}
	return kcp
}

func wantRegistryMirrorCommands() []string {
	return []string{
		"cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml",
		"sudo systemctl daemon-reload",
		"sudo systemctl restart containerd",
	}
}

func TestKubeadmControlPlane(t *testing.T) {
	tt := newApiBuilerTest(t)
	controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], nil)
	got, err := snow.KubeadmControlPlane(tt.logger, tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	want := wantKubeadmControlPlane("1.21")
	want.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.IgnorePreflightErrors = []string{"DirAvailable--etc-kubernetes-manifests"}
	tt.Expect(got).To(BeComparableTo(want))
}

var registryMirrorTests = []struct {
	name                 string
	registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration
	wantFiles            []bootstrapv1.File
	wantRegistryConfig   bootstrapv1.RegistryMirrorConfiguration
}{
	{
		name: "with namespace",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4",
			Port:     "443",
			OCINamespaces: []v1alpha1.OCINamespace{
				{
					Registry:  "public.ecr.aws",
					Namespace: "eks-anywhere",
				},
			},
			CACertContent: "xyz",
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443/v2/eks-anywhere"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    ca_file = "/etc/containerd/certs.d/1.2.3.4:443/ca.crt"`,
			},
			{
				Path:    "/etc/containerd/certs.d/1.2.3.4:443/ca.crt",
				Owner:   "root:root",
				Content: "xyz",
			},
		},
		wantRegistryConfig: bootstrapv1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4:443/v2/eks-anywhere",
			CACert:   "xyz",
		},
	},
	{
		name: "with ca cert",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4",
			Port:     "443",
			OCINamespaces: []v1alpha1.OCINamespace{
				{
					Registry:  "public.ecr.aws",
					Namespace: "eks-anywhere",
				},
				{
					Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
					Namespace: "curated-packages",
				},
			},
			CACertContent: "xyz",
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."783794618700.dkr.ecr.*.amazonaws.com"]
    endpoint = ["https://1.2.3.4:443/v2/curated-packages"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443/v2/eks-anywhere"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    ca_file = "/etc/containerd/certs.d/1.2.3.4:443/ca.crt"`,
			},
			{
				Path:    "/etc/containerd/certs.d/1.2.3.4:443/ca.crt",
				Owner:   "root:root",
				Content: "xyz",
			},
		},
		wantRegistryConfig: bootstrapv1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4:443/v2/eks-anywhere",
			CACert:   "xyz",
		},
	},
	{
		name: "with insecure skip",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:           "1.2.3.4",
			Port:               "443",
			InsecureSkipVerify: true,
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    insecure_skip_verify = true`,
			},
		},
		wantRegistryConfig: bootstrapv1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4:443",
		},
	},
	{
		name: "without ca cert",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4",
			Port:     "443",
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]`,
			},
		},
		wantRegistryConfig: bootstrapv1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4:443",
		},
	},
	{
		name: "with ca cert and insecure skip",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:           "1.2.3.4",
			Port:               "443",
			CACertContent:      "xyz",
			InsecureSkipVerify: true,
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    ca_file = "/etc/containerd/certs.d/1.2.3.4:443/ca.crt"
    insecure_skip_verify = true`,
			},
			{
				Path:    "/etc/containerd/certs.d/1.2.3.4:443/ca.crt",
				Owner:   "root:root",
				Content: "xyz",
			},
		},
		wantRegistryConfig: bootstrapv1.RegistryMirrorConfiguration{
			Endpoint: "1.2.3.4:443",
			CACert:   "xyz",
		},
	},
}

func TestKubeadmControlPlaneWithRegistryMirrorUbuntu(t *testing.T) {
	for _, tt := range registryMirrorTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = tt.registryMirrorConfig
			controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", g.machineConfigs[g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], nil)
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane("1.21")
			want.Spec.KubeadmConfigSpec.Files = append(want.Spec.KubeadmConfigSpec.Files, tt.wantFiles...)
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(want.Spec.KubeadmConfigSpec.PreKubeadmCommands, wantRegistryMirrorCommands()...)
			want.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.IgnorePreflightErrors = []string{"DirAvailable--etc-kubernetes-manifests"}
			g.Expect(got).To(BeComparableTo(want))
		})
	}
}

var pause = bootstrapv1.Pause{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-distro/kubernetes/pause",
		ImageTag:        "0.0.1",
	},
}

var bootstrap = bootstrapv1.BottlerocketBootstrap{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap",
		ImageTag:        "0.0.1",
	},
}

var admin = bootstrapv1.BottlerocketAdmin{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-admin",
		ImageTag:        "0.0.1",
	},
}

var control = bootstrapv1.BottlerocketControl{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-control",
		ImageTag:        "0.0.1",
	},
}

var bootstrapCustom = []bootstrapv1.BottlerocketBootstrapContainer{
	{
		Name: "bottlerocket-bootstrap-snow",
		ImageMeta: bootstrapv1.ImageMeta{
			ImageRepository: "public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap-snow",
			ImageTag:        "v1-20-22-eks-a-v0.0.0-dev-build.4984",
		},
		Mode: "always",
	},
}

func TestKubeadmControlPlaneWithRegistryMirrorBottlerocket(t *testing.T) {
	for _, tt := range registryMirrorTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = tt.registryMirrorConfig
			g.clusterSpec.SnowMachineConfig("test-cp").Spec.OSFamily = v1alpha1.Bottlerocket
			controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", g.machineConfigs["test-cp"], nil)
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane("1.21")
			want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = []string{}
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = bootstrap
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketAdmin = admin
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketControl = control
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketCustomBootstrapContainers = bootstrapCustom
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = pause
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = bootstrap
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketAdmin = admin
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketControl = control
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketCustomBootstrapContainers = bootstrapCustom
			want.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = pause
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.RegistryMirror = tt.wantRegistryConfig
			want.Spec.KubeadmConfigSpec.JoinConfiguration.RegistryMirror = tt.wantRegistryConfig
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes,
				bootstrapv1.HostPathMount{
					HostPath:  "/var/lib/kubeadm/controller-manager.conf",
					MountPath: "/etc/kubernetes/controller-manager.conf",
					Name:      "kubeconfig",
					PathType:  "File",
					ReadOnly:  true,
				},
			)
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes,
				bootstrapv1.HostPathMount{
					HostPath:  "/var/lib/kubeadm/scheduler.conf",
					MountPath: "/etc/kubernetes/scheduler.conf",
					Name:      "kubeconfig",
					PathType:  "File",
					ReadOnly:  true,
				},
			)
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.CertificatesDir = "/var/lib/kubeadm/pki"

			g.Expect(got).To(BeComparableTo(want))
		})
	}
}

func wantProxyConfigCommands() []string {
	return []string{
		"sudo systemctl daemon-reload",
		"sudo systemctl restart containerd",
	}
}

var proxyTests = []struct {
	name            string
	proxy           *v1alpha1.ProxyConfiguration
	wantFiles       []bootstrapv1.File
	wantProxyConfig bootstrapv1.ProxyConfiguration
}{
	{
		name: "with proxy, pods cidr, service cidr, cp endpoint",
		proxy: &v1alpha1.ProxyConfiguration{
			HttpProxy:  "1.2.3.4:8888",
			HttpsProxy: "1.2.3.4:8888",
			NoProxy: []string{
				"1.2.3.4/0",
				"1.2.3.5/0",
			},
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/systemd/system/containerd.service.d/http-proxy.conf",
				Owner: "root:root",
				Content: `[Service]
Environment="HTTP_PROXY=1.2.3.4:8888"
Environment="HTTPS_PROXY=1.2.3.4:8888"
Environment="NO_PROXY=10.1.0.0/16,10.96.0.0/12,1.2.3.4/0,1.2.3.5/0,localhost,127.0.0.1,.svc,1.2.3.4"`,
			},
		},
		wantProxyConfig: bootstrapv1.ProxyConfiguration{
			HTTPSProxy: "1.2.3.4:8888",
			NoProxy: []string{
				"10.1.0.0/16",
				"10.96.0.0/12",
				"1.2.3.4/0",
				"1.2.3.5/0",
				"localhost",
				"127.0.0.1",
				".svc",
				"1.2.3.4",
			},
		},
	},
}

func TestKubeadmControlPlaneWithProxyConfigUbuntu(t *testing.T) {
	for _, tt := range proxyTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.Cluster.Spec.ProxyConfiguration = tt.proxy
			controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", g.machineConfigs[g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], nil)
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane("1.21")
			want.Spec.KubeadmConfigSpec.Files = append(want.Spec.KubeadmConfigSpec.Files, tt.wantFiles...)
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(want.Spec.KubeadmConfigSpec.PreKubeadmCommands, wantProxyConfigCommands()...)
			want.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.IgnorePreflightErrors = []string{"DirAvailable--etc-kubernetes-manifests"}
			g.Expect(got).To(BeComparableTo(want))
		})
	}
}

func TestKubeadmControlPlaneUbuntuKubernetes129(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.KubernetesVersion = "1.29"
	controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name], nil)
	got, err := snow.KubeadmControlPlane(tt.logger, tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	want := wantKubeadmControlPlane("1.29")
	want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(want.Spec.KubeadmConfigSpec.PreKubeadmCommands, "if [ -f /run/kubeadm/kubeadm.yaml ]; then sed -i 's#path: /etc/kubernetes/admin.conf#path: /etc/kubernetes/super-admin.conf#' /etc/kubernetes/manifests/kube-vip.yaml; fi")
	want.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.IgnorePreflightErrors = []string{"DirAvailable--etc-kubernetes-manifests"}
	tt.Expect(got).To(BeComparableTo(want))
}

func TestKubeadmControlPlaneWithProxyConfigBottlerocket(t *testing.T) {
	for _, tt := range proxyTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.Cluster.Spec.ProxyConfiguration = tt.proxy
			g.clusterSpec.SnowMachineConfig("test-cp").Spec.OSFamily = v1alpha1.Bottlerocket
			controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", g.machineConfigs["test-cp"], nil)
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane("1.21")
			want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = []string{}
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = bootstrap
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketAdmin = admin
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketControl = control
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketCustomBootstrapContainers = bootstrapCustom
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = pause
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = bootstrap
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketAdmin = admin
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketControl = control
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketCustomBootstrapContainers = bootstrapCustom
			want.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = pause
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Proxy = tt.wantProxyConfig
			want.Spec.KubeadmConfigSpec.JoinConfiguration.Proxy = tt.wantProxyConfig
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes,
				bootstrapv1.HostPathMount{
					HostPath:  "/var/lib/kubeadm/controller-manager.conf",
					MountPath: "/etc/kubernetes/controller-manager.conf",
					Name:      "kubeconfig",
					PathType:  "File",
					ReadOnly:  true,
				},
			)
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes,
				bootstrapv1.HostPathMount{
					HostPath:  "/var/lib/kubeadm/scheduler.conf",
					MountPath: "/etc/kubernetes/scheduler.conf",
					Name:      "kubeconfig",
					PathType:  "File",
					ReadOnly:  true,
				},
			)
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.CertificatesDir = "/var/lib/kubeadm/pki"

			g.Expect(got).To(Equal(want))
		})
	}
}

var bottlerocketAdditionalSettingsTests = []struct {
	name       string
	settings   *v1alpha1.HostOSConfiguration
	wantConfig *bootstrapv1.BottlerocketSettings
}{
	{
		name: "with kernel sysctl settings",
		settings: &v1alpha1.HostOSConfiguration{
			BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{
				Kernel: &bootstrapv1.BottlerocketKernelSettings{
					SysctlSettings: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		wantConfig: &bootstrapv1.BottlerocketSettings{
			Kernel: &bootstrapv1.BottlerocketKernelSettings{
				SysctlSettings: map[string]string{
					"foo": "bar",
				},
			},
		},
	},
	{
		name: "with boot kernel settings",
		settings: &v1alpha1.HostOSConfiguration{
			BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{
				Boot: &bootstrapv1.BottlerocketBootSettings{
					BootKernelParameters: map[string][]string{
						"foo": {
							"abc",
							"def",
						},
					},
				},
			},
		},
		wantConfig: &bootstrapv1.BottlerocketSettings{
			Boot: &bootstrapv1.BottlerocketBootSettings{
				BootKernelParameters: map[string][]string{
					"foo": {
						"abc",
						"def",
					},
				},
			},
		},
	},
	{
		name: "with both empty",
		settings: &v1alpha1.HostOSConfiguration{
			BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{
				Boot:   &bootstrapv1.BottlerocketBootSettings{},
				Kernel: &bootstrapv1.BottlerocketKernelSettings{},
			},
		},
		wantConfig: &bootstrapv1.BottlerocketSettings{
			Boot:   &bootstrapv1.BottlerocketBootSettings{},
			Kernel: &bootstrapv1.BottlerocketKernelSettings{},
		},
	},
}

func TestKubeadmControlPlaneWithBottlerocketAdditionalSettings(t *testing.T) {
	for _, tt := range bottlerocketAdditionalSettingsTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.SnowMachineConfig("test-cp").Spec.HostOSConfiguration = tt.settings
			g.clusterSpec.SnowMachineConfig("test-cp").Spec.OSFamily = v1alpha1.Bottlerocket
			controlPlaneMachineTemplate := snow.MachineTemplate("snow-test-control-plane-1", g.machineConfigs["test-cp"], nil)
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane("1.21")
			want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = []string{}
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = bootstrap
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketAdmin = admin
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketControl = control
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketCustomBootstrapContainers = bootstrapCustom
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = pause
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = bootstrap
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketAdmin = admin
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketControl = control
			want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketCustomBootstrapContainers = bootstrapCustom
			want.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = pause
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Bottlerocket = tt.wantConfig
			want.Spec.KubeadmConfigSpec.JoinConfiguration.Bottlerocket = tt.wantConfig
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes,
				bootstrapv1.HostPathMount{
					HostPath:  "/var/lib/kubeadm/controller-manager.conf",
					MountPath: "/etc/kubernetes/controller-manager.conf",
					Name:      "kubeconfig",
					PathType:  "File",
					ReadOnly:  true,
				},
			)
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes = append(want.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes,
				bootstrapv1.HostPathMount{
					HostPath:  "/var/lib/kubeadm/scheduler.conf",
					MountPath: "/etc/kubernetes/scheduler.conf",
					Name:      "kubeconfig",
					PathType:  "File",
					ReadOnly:  true,
				},
			)
			want.Spec.KubeadmConfigSpec.ClusterConfiguration.CertificatesDir = "/var/lib/kubeadm/pki"

			g.Expect(got).To(Equal(want))
		})
	}
}

func wantKubeadmConfigTemplate() *bootstrapv1.KubeadmConfigTemplate {
	return &bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
			Kind:       "KubeadmConfigTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test-md-0-1",
			Namespace: "eksa-system",
		},
		Spec: bootstrapv1.KubeadmConfigTemplateSpec{
			Template: bootstrapv1.KubeadmConfigTemplateResource{
				Spec: bootstrapv1.KubeadmConfigSpec{
					ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
						ControllerManager: bootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{},
						},
						APIServer: bootstrapv1.APIServer{
							ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
								ExtraArgs: map[string]string{},
							},
						},
					},
					JoinConfiguration: &bootstrapv1.JoinConfiguration{
						NodeRegistration: bootstrapv1.NodeRegistrationOptions{
							KubeletExtraArgs: map[string]string{
								"provider-id": "aws-snow:////'{{ ds.meta_data.instance_id }}'",
							},
						},
					},
					PreKubeadmCommands: []string{
						"/etc/eks/bootstrap.sh",
					},
					PostKubeadmCommands: []string{},
					Files:               []bootstrapv1.File{},
				},
			},
		},
	}
}

func TestKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	workerNodeGroupConfig := g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]
	g.clusterSpec.SnowMachineConfigs["test-wn"].Spec.ContainersVolume = &snowv1.Volume{Size: 8}
	got, err := snow.KubeadmConfigTemplate(g.logger, g.clusterSpec, workerNodeGroupConfig)
	g.Expect(err).To(Succeed())
	want := wantKubeadmConfigTemplate()
	g.Expect(got).To(Equal(want))
}

func wantMachineDeployment() *clusterv1.MachineDeployment {
	wantVersion := "v1.21.5-eks-1-21-9"
	wantReplicas := int32(3)
	wantMaxUnavailable := intstr.FromInt(0)
	wantMaxSurge := intstr.FromInt(1)
	return &clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "MachineDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test-md-0",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name":                        "snow-test",
				"cluster.anywhere.eks.amazonaws.com/cluster-name":      "snow-test",
				"cluster.anywhere.eks.amazonaws.com/cluster-namespace": "test-namespace",
			},
			Annotations: map[string]string{},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: "snow-test",
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels: map[string]string{
						"cluster.x-k8s.io/cluster-name": "snow-test",
					},
				},
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
							Kind:       "KubeadmConfigTemplate",
							Name:       "snow-test-md-0-1",
						},
					},
					ClusterName: "snow-test",
					InfrastructureRef: v1.ObjectReference{
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
						Kind:       "AWSSnowMachineTemplate",
						Name:       "snow-test-md-0-1",
					},
					Version: &wantVersion,
				},
			},
			Replicas: &wantReplicas,
			Strategy: &clusterv1.MachineDeploymentStrategy{
				RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
					MaxUnavailable: &wantMaxUnavailable,
					MaxSurge:       &wantMaxSurge,
				},
				Type: clusterv1.RollingUpdateMachineDeploymentStrategyType,
			},
		},
	}
}

func wantSnowCluster() *snowv1.AWSSnowCluster {
	return &snowv1.AWSSnowCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSSnowCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test",
			Namespace: "eksa-system",
		},
		Spec: snowv1.AWSSnowClusterSpec{
			Region: "snow",
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "1.2.3.4",
				Port: 6443,
			},
			IdentityRef: &snowv1.AWSSnowIdentityReference{
				Name: "snow-test-snow-credentials",
				Kind: "Secret",
			},
		},
	}
}

func TestSnowCluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := snow.SnowCluster(tt.clusterSpec, wantSnowCredentialsSecret())
	tt.Expect(got).To(Equal(wantSnowCluster()))
}

func TestSnowCredentialsSecret(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := snow.CAPASCredentialsSecret(tt.clusterSpec, []byte("creds"), []byte("certs"))
	want := wantSnowCredentialsSecret()
	tt.Expect(got).To(Equal(want))
}

func wantSnowCredentialsSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test-snow-credentials",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
		},
		Data: map[string][]byte{
			"credentials": []byte("creds"),
			"ca-bundle":   []byte("certs"),
		},
		Type: "Opaque",
	}
}

func wantEksaCredentialsSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-snow-credentials",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"credentials": []byte("creds"),
			"ca-bundle":   []byte("certs"),
		},
		Type: "Opaque",
	}
}

func wantSnowMachineTemplate() *snowv1.AWSSnowMachineTemplate {
	wantAMIID := "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea"
	wantSSHKey := "default"
	wantPhysicalNetworkConnector := "SFP_PLUS"
	osFamily := snowv1.Ubuntu
	return &snowv1.AWSSnowMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			Kind:       "AWSSnowMachineTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test-md-0-1",
			Namespace: "eksa-system",
		},
		Spec: snowv1.AWSSnowMachineTemplateSpec{
			Template: snowv1.AWSSnowMachineTemplateResource{
				Spec: snowv1.AWSSnowMachineSpec{
					IAMInstanceProfile: "control-plane.cluster-api-provider-aws.sigs.k8s.io",
					InstanceType:       "sbe-c.xlarge",
					SSHKeyName:         &wantSSHKey,
					AMI: snowv1.AWSResourceReference{
						ID: &wantAMIID,
					},
					CloudInit: snowv1.CloudInit{
						InsecureSkipSecretsManager: true,
					},
					PhysicalNetworkConnectorType: &wantPhysicalNetworkConnector,
					Devices: []string{
						"1.2.3.4",
						"1.2.3.5",
					},
					OSFamily: &osFamily,
					Network: snowv1.AWSSnowNetwork{
						DirectNetworkInterfaces: []snowv1.AWSSnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
				},
			},
		},
	}
}

func wantSnowIPPool() *snowv1.AWSSnowIPPool {
	return &snowv1.AWSSnowIPPool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion(),
			Kind:       snow.SnowIPPoolKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ip-pool-1",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: snowv1.AWSSnowIPPoolSpec{
			IPPools: []snowv1.IPPool{
				{
					IPStart: ptr.String("start"),
					IPEnd:   ptr.String("end"),
					Gateway: ptr.String("gateway"),
					Subnet:  ptr.String("subnet"),
				},
			},
		},
	}
}

func TestSnowMachineTemplate(t *testing.T) {
	tt := newApiBuilerTest(t)
	mc := tt.machineConfigs["test-cp"]
	mc.Spec.NonRootVolumes = []*snowv1.Volume{
		{
			DeviceName: "/dev/sdc",
			Size:       10,
		},
	}
	got := snow.MachineTemplate("snow-test-control-plane-1", mc, nil)
	want := wantSnowMachineTemplate()
	want.SetName("snow-test-control-plane-1")
	want.Spec.Template.Spec.InstanceType = "sbe-c.large"
	want.Spec.Template.Spec.NonRootVolumes = mc.Spec.NonRootVolumes
	tt.Expect(got).To(Equal(want))
}

func TestSnowMachineTemplateWithNetwork(t *testing.T) {
	tt := newApiBuilerTest(t)
	network := snowv1.AWSSnowNetwork{
		DirectNetworkInterfaces: []snowv1.AWSSnowDirectNetworkInterface{
			{
				Index: 1,
				DHCP:  false,
				IPPool: &v1.ObjectReference{
					Kind: "AWSSnowIPPool",
					Name: "ip-pool",
				},
				Primary: true,
			},
		},
	}
	tt.machineConfigs["test-cp"].Spec.Network = v1alpha1.SnowNetwork{
		DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
			{
				Index: 1,
				DHCP:  false,
				IPPoolRef: &v1alpha1.Ref{
					Kind: "SnowIPPool",
					Name: "ip-pool",
				},
				Primary: true,
			},
		},
	}
	capasPools := snow.CAPASIPPools{
		"ip-pool": &snowv1.AWSSnowIPPool{
			TypeMeta: metav1.TypeMeta{
				Kind: "AWSSnowIPPool",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ip-pool",
			},
		},
	}
	got := snow.MachineTemplate("snow-test-control-plane-1", tt.machineConfigs["test-cp"], capasPools)
	want := wantSnowMachineTemplate()
	want.SetName("snow-test-control-plane-1")
	want.Spec.Template.Spec.InstanceType = "sbe-c.large"
	want.Spec.Template.Spec.Network = network
	tt.Expect(got).To(BeComparableTo(want))
}

func tlsCipherSuitesArgs() map[string]string {
	return map[string]string{"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}
}

func wantEtcdCluster() *etcdv1.EtcdadmCluster {
	replicas := int32(3)
	return &etcdv1.EtcdadmCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			Kind:       "EtcdadmCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test-etcd",
			Namespace: "eksa-system",
		},
		Spec: etcdv1.EtcdadmClusterSpec{
			Replicas: &replicas,
			EtcdadmConfigSpec: etcdbootstrapv1.EtcdadmConfigSpec{
				EtcdadmBuiltin:     true,
				CipherSuites:       "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				PreEtcdadmCommands: []string{},
			},
			InfrastructureTemplate: v1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "AWSSnowMachineTemplate",
				Name:       "test-etcd",
			},
		},
	}
}

func wantEtcdClusterUbuntu() *etcdv1.EtcdadmCluster {
	etcd := wantEtcdCluster()
	etcd.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format("cloud-config")
	etcd.Spec.EtcdadmConfigSpec.CloudInitConfig = &etcdbootstrapv1.CloudInitConfig{
		Version:        "3.4.16",
		InstallDir:     "/usr/bin",
		EtcdReleaseURL: "https://distro.eks.amazonaws.com/kubernetes-1-21/releases/4/artifacts/etcd/v3.4.16/etcd-linux-amd64-v3.4.16.tar.gz",
	}
	etcd.Spec.EtcdadmConfigSpec.PreEtcdadmCommands = []string{
		"/etc/eks/bootstrap.sh",
	}
	return etcd
}

func wantEtcdClusterBottlerocket() *etcdv1.EtcdadmCluster {
	etcd := wantEtcdCluster()
	etcd.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format("bottlerocket")
	etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
		AdminImage:     "public.ecr.aws/eks-anywhere/bottlerocket-admin:0.0.1",
		ControlImage:   "public.ecr.aws/eks-anywhere/bottlerocket-control:0.0.1",
		CustomBootstrapContainers: []etcdbootstrapv1.BottlerocketBootstrapContainer{
			{
				Name:      "bottlerocket-bootstrap-snow",
				Image:     "public.ecr.aws/l0g8r8j6/bottlerocket-bootstrap-snow:v1-20-22-eks-a-v0.0.0-dev-build.4984",
				Essential: false,
				Mode:      "always",
			},
		},
		Kernel: &bootstrapv1.BottlerocketKernelSettings{
			SysctlSettings: map[string]string{
				"foo": "bar",
			},
		},
		Boot: &bootstrapv1.BottlerocketBootSettings{
			BootKernelParameters: map[string][]string{
				"foo": {
					"abc",
					"def",
				},
			},
		},
	}
	return etcd
}

func TestEtcdadmClusterUbuntu(t *testing.T) {
	tt := newApiBuilerTest(t)
	eksaVersion := v1alpha1.EksaVersion("v0.19.0")
	tt.clusterSpec.Cluster.Spec.EksaVersion = &eksaVersion
	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
		MachineGroupRef: &v1alpha1.Ref{
			Kind: v1alpha1.SnowMachineConfigKind,
			Name: "test-etcd",
		},
	}
	tt.clusterSpec.SnowMachineConfigs["test-etcd"] = &v1alpha1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-etcd",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.SnowMachineConfigSpec{
			OSFamily: "ubuntu",
		},
	}
	tt.machineConfigs["test-etcd"] = tt.clusterSpec.SnowMachineConfigs["test-etcd"]
	etcdMachineTemplates := snow.MachineTemplate("test-etcd", tt.machineConfigs["test-etcd"], nil)
	got := snow.EtcdadmCluster(tt.logger, tt.clusterSpec, etcdMachineTemplates)
	want := wantEtcdClusterUbuntu()
	tt.Expect(got).To(Equal(want))
}

func TestEtcdadmClusterBottlerocket(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
		MachineGroupRef: &v1alpha1.Ref{
			Kind: v1alpha1.SnowMachineConfigKind,
			Name: "test-etcd",
		},
	}
	tt.clusterSpec.SnowMachineConfigs["test-etcd"] = &v1alpha1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-etcd",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.SnowMachineConfigSpec{
			OSFamily: "bottlerocket",
			HostOSConfiguration: &v1alpha1.HostOSConfiguration{
				BottlerocketConfiguration: &v1alpha1.BottlerocketConfiguration{
					Kernel: &bootstrapv1.BottlerocketKernelSettings{
						SysctlSettings: map[string]string{
							"foo": "bar",
						},
					},
					Boot: &bootstrapv1.BottlerocketBootSettings{
						BootKernelParameters: map[string][]string{
							"foo": {
								"abc",
								"def",
							},
						},
					},
				},
			},
		},
	}
	tt.machineConfigs["test-etcd"] = tt.clusterSpec.SnowMachineConfigs["test-etcd"]
	etcdMachineTemplates := snow.MachineTemplate("test-etcd", tt.machineConfigs["test-etcd"], nil)
	got := snow.EtcdadmCluster(tt.logger, tt.clusterSpec, etcdMachineTemplates)
	want := wantEtcdClusterBottlerocket()
	tt.Expect(got).To(Equal(want))
}

func TestEtcdadmClusterUnsupportedOS(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		Count: 3,
		MachineGroupRef: &v1alpha1.Ref{
			Kind: v1alpha1.SnowMachineConfigKind,
			Name: "test-etcd",
		},
	}
	tt.clusterSpec.SnowMachineConfigs["test-etcd"] = &v1alpha1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-etcd",
			Namespace: "test-namespace",
		},
		Spec: v1alpha1.SnowMachineConfigSpec{
			OSFamily: "unsupported",
		},
	}
	tt.machineConfigs["test-etcd"] = tt.clusterSpec.SnowMachineConfigs["test-etcd"]
	etcdMachineTemplates := snow.MachineTemplate("test-etcd", tt.machineConfigs["test-etcd"], nil)
	got := snow.EtcdadmCluster(tt.logger, tt.clusterSpec, etcdMachineTemplates)
	want := wantEtcdCluster()
	tt.Expect(got).To(Equal(want))
}
