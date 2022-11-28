package snow_test

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

type apiBuilerTest struct {
	*WithT
	clusterSpec    *cluster.Spec
	machineConfigs map[string]*v1alpha1.SnowMachineConfig
	logger         logr.Logger
}

func newApiBuilerTest(t *testing.T) apiBuilerTest {
	format.MaxLength = 0
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

func TestCAPICluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	snowCluster := snow.SnowCluster(tt.clusterSpec, wantSnowCredentialsSecret())
	controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
	kubeadmControlPlane, err := snow.KubeadmControlPlane(tt.logger, tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	got := snow.CAPICluster(tt.clusterSpec, snowCluster, kubeadmControlPlane)
	tt.Expect(got).To(Equal(wantCAPICluster()))
}

func wantKubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	wantReplicas := int32(3)
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
					ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
					DNS: bootstrapv1.DNS{
						ImageMeta: bootstrapv1.ImageMeta{
							ImageRepository: "public.ecr.aws/eks-distro/coredns",
							ImageTag:        "v1.8.4-eks-1-21-9",
						},
					},
					Etcd: bootstrapv1.Etcd{
						Local: &bootstrapv1.LocalEtcd{
							ImageMeta: bootstrapv1.ImageMeta{
								ImageRepository: "public.ecr.aws/eks-distro/etcd-io",
								ImageTag:        "v3.4.16-eks-1-21-9",
							},
							ExtraArgs: map[string]string{
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
						ExtraArgs: tlsCipherSuitesArgs(),
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
					"/etc/eks/bootstrap.sh public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433 1.2.3.4",
				},
				PostKubeadmCommands: []string{
					"/etc/eks/bootstrap-after.sh public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433 1.2.3.4",
				},
				Files: []bootstrapv1.File{
					{
						Path:    "/etc/kubernetes/manifests/kube-vip.yaml",
						Owner:   "root:root",
						Content: test.KubeVipTemplate,
					},
				},
			},
			Replicas: &wantReplicas,
			Version:  "v1.21.5-eks-1-21-9",
		},
	}
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
	controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
	got, err := snow.KubeadmControlPlane(tt.logger, tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	want := wantKubeadmControlPlane()
	tt.Expect(got).To(Equal(want))
}

var registryMirrorTests = []struct {
	name                 string
	registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration
	wantFiles            []bootstrapv1.File
	wantRegistryConfig   bootstrapv1.RegistryMirrorConfiguration
}{
	{
		name: "with ca cert",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:      "1.2.3.4",
			Port:          "443",
			CACertContent: "xyz",
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]
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
			Endpoint: "1.2.3.4:443",
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
			controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", g.machineConfigs[g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Files = append(want.Spec.KubeadmConfigSpec.Files, tt.wantFiles...)
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(want.Spec.KubeadmConfigSpec.PreKubeadmCommands, wantRegistryMirrorCommands()...)
			g.Expect(got).To(Equal(want))
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
			controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", g.machineConfigs["test-cp"])
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
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
			g.Expect(got).To(Equal(want))
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
			controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", g.machineConfigs[g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Files = append(want.Spec.KubeadmConfigSpec.Files, tt.wantFiles...)
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(want.Spec.KubeadmConfigSpec.PreKubeadmCommands, wantProxyConfigCommands()...)
			g.Expect(got).To(Equal(want))
		})
	}
}

func TestKubeadmControlPlaneWithProxyConfigBottlerocket(t *testing.T) {
	for _, tt := range proxyTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.Cluster.Spec.ProxyConfiguration = tt.proxy
			g.clusterSpec.SnowMachineConfig("test-cp").Spec.OSFamily = v1alpha1.Bottlerocket
			controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", g.machineConfigs["test-cp"])
			got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
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
			g.Expect(got).To(Equal(want))
		})
	}
}

func wantContainersVolumeCommands() []string {
	return []string{
		"/etc/eks/bootstrap-volume.sh",
	}
}

func TestKubeadmControlPlaneWithContainersVolume(t *testing.T) {
	g := newApiBuilerTest(t)
	cpMachineConfig := g.machineConfigs[g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	cpMachineConfig.Spec.ContainersVolume = &snowv1.Volume{Size: 8}
	controlPlaneMachineTemplate := snow.SnowMachineTemplate("snow-test-control-plane-1", cpMachineConfig)
	got, err := snow.KubeadmControlPlane(g.logger, g.clusterSpec, controlPlaneMachineTemplate)
	g.Expect(err).To(Succeed())
	want := wantKubeadmControlPlane()
	want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(want.Spec.KubeadmConfigSpec.PreKubeadmCommands, wantContainersVolumeCommands()...)
	g.Expect(got).To(Equal(want))
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

func TestKubeadmConfigTemplateWithContainersVolume(t *testing.T) {
	g := newApiBuilerTest(t)
	workerNodeGroupConfig := g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]
	g.clusterSpec.SnowMachineConfigs["test-wn"].Spec.ContainersVolume = &snowv1.Volume{Size: 8}
	got, err := snow.KubeadmConfigTemplate(g.logger, g.clusterSpec, workerNodeGroupConfig)
	g.Expect(err).To(Succeed())
	want := wantKubeadmConfigTemplate()
	want.Spec.Template.Spec.PreKubeadmCommands = append(want.Spec.Template.Spec.PreKubeadmCommands, wantContainersVolumeCommands()...)
	g.Expect(got).To(Equal(want))
}

func wantMachineDeployment() *clusterv1.MachineDeployment {
	wantVersion := "v1.21.5-eks-1-21-9"
	wantReplicas := int32(3)
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
				},
			},
		},
	}
}

func TestSnowMachineTemplate(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := snow.SnowMachineTemplate("snow-test-control-plane-1", tt.machineConfigs["test-cp"])
	want := wantSnowMachineTemplate()
	want.SetName("snow-test-control-plane-1")
	want.Spec.Template.Spec.InstanceType = "sbe-c.large"
	tt.Expect(got).To(Equal(want))
}

func tlsCipherSuitesArgs() map[string]string {
	return map[string]string{"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}
}
