package docker_test

import (
	"context"
	"testing"
	"time"

	etcdadmbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestControlPlaneObjects(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane *docker.ControlPlane
		want         []kubernetes.Object
	}{
		{
			name: "stacked etcd",
			controlPlane: &docker.ControlPlane{
				Cluster:                     capiCluster(),
				ProviderCluster:             dockerCluster(),
				KubeadmControlPlane:         kubeadmControlPlane(),
				ControlPlaneMachineTemplate: dockerMachineTemplate("cp-mt"),
			},
			want: []kubernetes.Object{
				capiCluster(),
				dockerCluster(),
				kubeadmControlPlane(),
				dockerMachineTemplate("cp-mt"),
			},
		},
		{
			name: "unstacked etcd",
			controlPlane: &docker.ControlPlane{
				Cluster:                     capiCluster(),
				ProviderCluster:             dockerCluster(),
				KubeadmControlPlane:         kubeadmControlPlane(),
				ControlPlaneMachineTemplate: dockerMachineTemplate("cp-mt"),
				EtcdCluster:                 etcdCluster(),
				EtcdMachineTemplate:         dockerMachineTemplate("etcd-mt"),
			},
			want: []kubernetes.Object{
				capiCluster(),
				dockerCluster(),
				kubeadmControlPlane(),
				dockerMachineTemplate("cp-mt"),
				etcdCluster(),
				dockerMachineTemplate("etcd-mt"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.controlPlane.Objects()).To(ConsistOf(tt.want))
		})
	}
}

func TestControlPlaneSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := testClusterSpec()
	wantCPMachineTemplate := dockerMachineTemplate("test-control-plane-1")
	wantEtcdMachineTemplate := dockerMachineTemplate("test-etcd-1")

	cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane()))
	g.Expect(cp.EtcdCluster).To(Equal(etcdCluster()))
	g.Expect(cp.ProviderCluster).To(Equal(dockerCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(wantCPMachineTemplate))
	g.Expect(cp.EtcdMachineTemplate).To(Equal(wantEtcdMachineTemplate))
}

func TestControlPlaneSpecNoKubeVersion(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := testClusterSpec()
	spec.Cluster.Spec.KubernetesVersion = ""

	_, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).To(MatchError(ContainSubstring("generating docker control plane yaml spec")))
}

func TestControlPlaneSpecUpdateMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()

	originalKubeadmControlPlane := kubeadmControlPlane()
	originalEtcdCluster := etcdCluster()
	originalEtcdCluster.Spec.InfrastructureTemplate.Name = "test-etcd-2"
	originalCPMachineTemplate := dockerMachineTemplate("test-control-plane-1")
	originalEtcdMachineTemplate := dockerMachineTemplate("test-etcd-2")
	wantKCP := originalKubeadmControlPlane.DeepCopy()
	wantEtcd := originalEtcdCluster.DeepCopy()
	wantCPtemplate := originalCPMachineTemplate.DeepCopy()
	wantEtcdTemplate := originalEtcdMachineTemplate.DeepCopy()

	originalCPMachineTemplate.Spec.Template.Spec.CustomImage = "old-custom-image"
	originalEtcdMachineTemplate.Spec.Template.Spec.CustomImage = "old-custom-image-etcd"

	client := test.NewFakeKubeClient(
		originalKubeadmControlPlane,
		originalEtcdCluster,
		originalCPMachineTemplate,
		originalEtcdMachineTemplate,
	)

	cpTaints := []corev1.Taint{
		{
			Key:    "foo",
			Value:  "bar",
			Effect: "PreferNoSchedule",
		},
	}

	spec.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints

	wantKCP.Spec.MachineTemplate.InfrastructureRef.Name = "test-control-plane-2"
	wantKCP.Spec.KubeadmConfigSpec.InitConfiguration.NodeRegistration.Taints = cpTaints
	wantKCP.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.Taints = cpTaints
	wantEtcd.Spec.InfrastructureTemplate.Name = "test-etcd-3"
	wantCPtemplate.Name = "test-control-plane-2"
	wantEtcdTemplate.Name = "test-etcd-3"

	cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(wantKCP))
	g.Expect(cp.EtcdCluster).To(Equal(wantEtcd))
	g.Expect(cp.ProviderCluster).To(Equal(dockerCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(wantCPtemplate))
	g.Expect(cp.EtcdMachineTemplate).To(Equal(wantEtcdTemplate))
}

func TestControlPlaneSpecNoChangesMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()
	originalKubeadmControlPlane := kubeadmControlPlane()
	originalEtcdCluster := etcdCluster()
	originalEtcdCluster.Spec.InfrastructureTemplate.Name = "test-etcd-1"
	originalCPMachineTemplate := dockerMachineTemplate("test-control-plane-1")
	originalEtcdMachineTemplate := dockerMachineTemplate("test-etcd-1")

	wantKCP := originalKubeadmControlPlane.DeepCopy()
	wantEtcd := originalEtcdCluster.DeepCopy()
	wantCPtemplate := originalCPMachineTemplate.DeepCopy()
	wantEtcdTemplate := originalEtcdMachineTemplate.DeepCopy()

	// This mimics what would happen if the objects were returned by a real api server
	// It helps make sure that the immutable object comparison is able to deal with these
	// kind of changes.
	originalCPMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())
	originalEtcdMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())

	// This is testing defaults. It's possible that some default logic will set items that are not set in our machine templates.
	// We need to take this into consideration when checking for equality.
	originalCPMachineTemplate.Spec.Template.Spec.ProviderID = ptr.String("default-id")
	originalEtcdMachineTemplate.Spec.Template.Spec.ProviderID = ptr.String("default-id")

	client := test.NewFakeKubeClient(
		originalKubeadmControlPlane,
		originalEtcdCluster,
		originalCPMachineTemplate,
		originalEtcdMachineTemplate,
	)

	cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(wantKCP))
	g.Expect(cp.EtcdCluster).To(Equal(wantEtcd))
	g.Expect(cp.ProviderCluster).To(Equal(dockerCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(wantCPtemplate))
	g.Expect(cp.EtcdMachineTemplate).To(Equal(wantEtcdTemplate))
}

func TestControPlaneSpecErrorFromClient(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()
	client := test.NewFakeKubeClientAlwaysError()

	_, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).To(MatchError(ContainSubstring("updating docker immutable object names")))
}

func TestControlPlaneSpecRegistryMirrorConfiguration(t *testing.T) {
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()

	tests := []struct {
		name         string
		mirrorConfig *anywherev1.RegistryMirrorConfiguration
		files        []bootstrapv1.File
	}{
		{
			name:         "insecure skip verify",
			mirrorConfig: test.RegistryMirrorInsecureSkipVerifyEnabled(),
			files:        test.RegistryMirrorConfigFilesInsecureSkipVerify(),
		},
		{
			name:         "insecure skip verify with ca cert",
			mirrorConfig: test.RegistryMirrorInsecureSkipVerifyEnabledAndCACert(),
			files:        test.RegistryMirrorConfigFilesInsecureSkipVerifyAndCACert(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := testClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.RegistryMirrorConfiguration = tt.mirrorConfig
			})
			wantCPMachineTemplate := dockerMachineTemplate("test-control-plane-1")
			wantEtcdMachineTemplate := dockerMachineTemplate("test-etcd-1")
			cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)

			g := NewWithT(t)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(cp).NotTo(BeNil())
			g.Expect(cp.Cluster).To(Equal(capiCluster()))
			g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, tt.files...)
				kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, test.RegistryMirrorPreKubeadmCommands()...)
			})))
			g.Expect(cp.EtcdCluster).To(Equal(etcdCluster(func(ec *etcdv1.EtcdadmCluster) {
				ec.Spec.EtcdadmConfigSpec.RegistryMirror = &etcdadmbootstrapv1.RegistryMirrorConfiguration{
					Endpoint: containerd.ToAPIEndpoint(registrymirror.FromClusterRegistryMirrorConfiguration(tt.mirrorConfig).CoreEKSAMirror()),
					CACert:   tt.mirrorConfig.CACertContent,
				}
			})))
			g.Expect(cp.ProviderCluster).To(Equal(dockerCluster()))
			g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(wantCPMachineTemplate))
			g.Expect(cp.EtcdMachineTemplate).To(Equal(wantEtcdMachineTemplate))
		})
	}
}

func TestControlPlaneUpgradeRolloutStrategy(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := testClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &anywherev1.ControlPlaneUpgradeRolloutStrategy{
			RollingUpdate: &anywherev1.ControlPlaneRollingUpdateParams{
				MaxSurge: 1,
			},
		}
	})

	cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane(func(k *controlplanev1.KubeadmControlPlane) {
		maxSurge := intstr.FromInt(1)
		k.Spec.RolloutStrategy = &controlplanev1.RolloutStrategy{
			RollingUpdate: &controlplanev1.RollingUpdate{
				MaxSurge: &maxSurge,
			},
		}
	})))
}

func testClusterSpec(opts ...test.ClusterSpecOpt) *cluster.Spec {
	name := "test"
	namespace := "test-namespace"
	devVersion := test.DevEksaVersion()

	clusterOpts := make([]test.ClusterSpecOpt, 0)
	clusterOpts = append(clusterOpts, func(s *cluster.Spec) {
		s.Cluster = &anywherev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: anywherev1.ClusterSpec{
				ClusterNetwork: anywherev1.ClusterNetwork{
					Services: anywherev1.Services{
						CidrBlocks: []string{"10.96.0.0/12"},
					},
					Pods: anywherev1.Pods{
						CidrBlocks: []string{"192.168.0.0/16"},
					},
				},
				ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
					Count: 3,
				},
				KubernetesVersion: "1.23",
				EksaVersion:       &devVersion,
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						Count:           ptr.Int(3),
						MachineGroupRef: &anywherev1.Ref{Name: name},
						Name:            "md-0",
					},
					{
						Count:           ptr.Int(3),
						MachineGroupRef: &anywherev1.Ref{Name: name},
						Name:            "md-1",
					},
				},
				ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
					Count: 3,
				},
				DatacenterRef: anywherev1.Ref{
					Kind: "DockerDatacenterConfig",
					Name: name,
				},
			},
		}

		s.VersionsBundles = map[anywherev1.KubernetesVersion]*cluster.VersionsBundle{
			"1.23": {
				KubeDistro: &cluster.KubeDistro{
					Kubernetes: cluster.VersionedRepository{
						Repository: "public.ecr.aws/eks-distro/kubernetes",
						Tag:        "v1.23.12-eks-1-23-6",
					},
					CoreDNS: cluster.VersionedRepository{
						Repository: "public.ecr.aws/eks-distro/coredns",
						Tag:        "v1.8.7-eks-1-23-6",
					},
					Etcd: cluster.VersionedRepository{
						Repository: "public.ecr.aws/eks-distro/etcd-io",
						Tag:        "v3.5.4-eks-1-23-6",
					},
					EtcdVersion: "3.5.4",
				},
				VersionsBundle: &releasev1alpha1.VersionsBundle{
					EksD: releasev1alpha1.EksDRelease{
						KindNode: releasev1alpha1.Image{
							Description: "kind/node container image",
							Name:        "kind/node",
							URI:         "public.ecr.aws/eks-anywhere/kubernetes-sigs/kind/node:v1.23.12-eks-d-1-23-6-eks-a-19",
						},
					},
				},
			},
		}
	})
	clusterOpts = append(clusterOpts, opts...)
	return test.NewClusterSpec(clusterOpts...)
}

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				APIServerPort: nil,
				ServiceDomain: "cluster.local",
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"10.96.0.0/12"},
				},
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"192.168.0.0/16"},
				},
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Name:       "test",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
			ManagedExternalEtcdRef: &corev1.ObjectReference{
				Kind:       "EtcdadmCluster",
				Name:       "test-etcd",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "DockerCluster",
				Name:       "test",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
		},
	}
}

func dockerCluster() *dockerv1.DockerCluster {
	return &dockerv1.DockerCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DockerCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func dockerMachineTemplate(name string) *dockerv1.DockerMachineTemplate {
	return &dockerv1.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DockerMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: dockerv1.DockerMachineTemplateSpec{
			Template: dockerv1.DockerMachineTemplateResource{
				Spec: dockerv1.DockerMachineSpec{
					CustomImage: "public.ecr.aws/eks-anywhere/kubernetes-sigs/kind/node:v1.23.12-eks-d-1-23-6-eks-a-19",
					ExtraMounts: []dockerv1.Mount{
						{
							ContainerPath: "/var/run/docker.sock",
							HostPath:      "/var/run/docker.sock",
							Readonly:      false,
						},
					},
					Bootstrapped: false,
				},
			},
		},
	}
}

func kubeadmControlPlane(opts ...func(*controlplanev1.KubeadmControlPlane)) *controlplanev1.KubeadmControlPlane {
	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: corev1.ObjectReference{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					Kind:       "DockerMachineTemplate",
					Name:       "test-control-plane-1",
					Namespace:  constants.EksaSystemNamespace,
				},
			},
			KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
					ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
					Etcd: bootstrapv1.Etcd{
						External: &bootstrapv1.ExternalEtcd{
							Endpoints: []string{},
							CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
							CertFile:  "/etc/kubernetes/pki/apiserver-etcd-client.crt",
							KeyFile:   "/etc/kubernetes/pki/apiserver-etcd-client.key",
						},
					},
					DNS: bootstrapv1.DNS{
						ImageMeta: bootstrapv1.ImageMeta{
							ImageRepository: "public.ecr.aws/eks-distro/coredns",
							ImageTag:        "v1.8.7-eks-1-23-6",
						},
					},
					APIServer: bootstrapv1.APIServer{
						CertSANs: []string{"localhost", "127.0.0.1"},
						ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{
								"audit-policy-file":   "/etc/kubernetes/audit-policy.yaml",
								"audit-log-path":      "/var/log/kubernetes/api-audit.log",
								"audit-log-maxage":    "30",
								"audit-log-maxbackup": "10",
								"audit-log-maxsize":   "512",
								"profiling":           "false",
								"tls-cipher-suites":   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							},
							ExtraVolumes: []bootstrapv1.HostPathMount{
								{
									HostPath:  "/etc/kubernetes/audit-policy.yaml",
									MountPath: "/etc/kubernetes/audit-policy.yaml",
									Name:      "audit-policy",
									PathType:  "File",
									ReadOnly:  true,
								},
								{
									HostPath:  "/var/log/kubernetes",
									MountPath: "/var/log/kubernetes",
									Name:      "audit-log-dir",
									PathType:  "DirectoryOrCreate",
									ReadOnly:  false,
								},
							},
						},
					},
					ControllerManager: bootstrapv1.ControlPlaneComponent{
						ExtraArgs: map[string]string{
							"enable-hostpath-provisioner": "true",
							"profiling":                   "false",
							"tls-cipher-suites":           "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						},
					},
					Scheduler: bootstrapv1.ControlPlaneComponent{
						ExtraArgs: map[string]string{
							"profiling":         "false",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						},
					},
				},
				Files: []bootstrapv1.File{
					{
						Path:  "/etc/kubernetes/audit-policy.yaml",
						Owner: "root:root",
						Content: `apiVersion: audit.k8s.io/v1beta1
kind: Policy
rules:
# Log aws-auth configmap changes
- level: RequestResponse
  namespaces: ["kube-system"]
  verbs: ["update", "patch", "delete"]
  resources:
  - group: "" # core
    resources: ["configmaps"]
    resourceNames: ["aws-auth"]
  omitStages:
  - "RequestReceived"
# The following requests were manually identified as high-volume and low-risk,
# so drop them.
- level: None
  users: ["system:kube-proxy"]
  verbs: ["watch"]
  resources:
  - group: "" # core
    resources: ["endpoints", "services", "services/status"]
- level: None
  users: ["kubelet"] # legacy kubelet identity
  verbs: ["get"]
  resources:
  - group: "" # core
    resources: ["nodes", "nodes/status"]
- level: None
  userGroups: ["system:nodes"]
  verbs: ["get"]
  resources:
  - group: "" # core
    resources: ["nodes", "nodes/status"]
- level: None
  users:
  - system:kube-controller-manager
  - system:kube-scheduler
  - system:serviceaccount:kube-system:endpoint-controller
  verbs: ["get", "update"]
  namespaces: ["kube-system"]
  resources:
  - group: "" # core
    resources: ["endpoints"]
- level: None
  users: ["system:apiserver"]
  verbs: ["get"]
  resources:
  - group: "" # core
    resources: ["namespaces", "namespaces/status", "namespaces/finalize"]
# Don't log HPA fetching metrics.
- level: None
  users:
  - system:kube-controller-manager
  verbs: ["get", "list"]
  resources:
  - group: "metrics.k8s.io"
# Don't log these read-only URLs.
- level: None
  nonResourceURLs:
  - /healthz*
  - /version
  - /swagger*
# Don't log events requests.
- level: None
  resources:
  - group: "" # core
    resources: ["events"]
# node and pod status calls from nodes are high-volume and can be large, don't log responses for expected updates from nodes
- level: Request
  users: ["kubelet", "system:node-problem-detector", "system:serviceaccount:kube-system:node-problem-detector"]
  verbs: ["update","patch"]
  resources:
  - group: "" # core
    resources: ["nodes/status", "pods/status"]
  omitStages:
  - "RequestReceived"
- level: Request
  userGroups: ["system:nodes"]
  verbs: ["update","patch"]
  resources:
  - group: "" # core
    resources: ["nodes/status", "pods/status"]
  omitStages:
  - "RequestReceived"
# deletecollection calls can be large, don't log responses for expected namespace deletions
- level: Request
  users: ["system:serviceaccount:kube-system:namespace-controller"]
  verbs: ["deletecollection"]
  omitStages:
  - "RequestReceived"
# Secrets, ConfigMaps, and TokenReviews can contain sensitive & binary data,
# so only log at the Metadata level.
- level: Metadata
  resources:
  - group: "" # core
    resources: ["secrets", "configmaps"]
  - group: authentication.k8s.io
    resources: ["tokenreviews"]
  omitStages:
    - "RequestReceived"
- level: Request
  resources:
  - group: ""
    resources: ["serviceaccounts/token"]
# Get repsonses can be large; skip them.
- level: Request
  verbs: ["get", "list", "watch"]
  resources:
  - group: "" # core
  - group: "admissionregistration.k8s.io"
  - group: "apiextensions.k8s.io"
  - group: "apiregistration.k8s.io"
  - group: "apps"
  - group: "authentication.k8s.io"
  - group: "authorization.k8s.io"
  - group: "autoscaling"
  - group: "batch"
  - group: "certificates.k8s.io"
  - group: "extensions"
  - group: "metrics.k8s.io"
  - group: "networking.k8s.io"
  - group: "policy"
  - group: "rbac.authorization.k8s.io"
  - group: "scheduling.k8s.io"
  - group: "settings.k8s.io"
  - group: "storage.k8s.io"
  omitStages:
  - "RequestReceived"
# Default level for known APIs
- level: RequestResponse
  resources:
  - group: "" # core
  - group: "admissionregistration.k8s.io"
  - group: "apiextensions.k8s.io"
  - group: "apiregistration.k8s.io"
  - group: "apps"
  - group: "authentication.k8s.io"
  - group: "authorization.k8s.io"
  - group: "autoscaling"
  - group: "batch"
  - group: "certificates.k8s.io"
  - group: "extensions"
  - group: "metrics.k8s.io"
  - group: "networking.k8s.io"
  - group: "policy"
  - group: "rbac.authorization.k8s.io"
  - group: "scheduling.k8s.io"
  - group: "settings.k8s.io"
  - group: "storage.k8s.io"
  omitStages:
  - "RequestReceived"
# Default level for all other requests.
- level: Metadata
  omitStages:
  - "RequestReceived"
`,
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						CRISocket: "/var/run/containerd/containerd.sock",
						KubeletExtraArgs: map[string]string{
							"cgroup-driver":     "cgroupfs",
							"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						},
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						CRISocket: "/var/run/containerd/containerd.sock",
						KubeletExtraArgs: map[string]string{
							"cgroup-driver":     "cgroupfs",
							"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						},
					},
				},
			},
			Replicas: ptr.Int32(3),
			Version:  "v1.23.12-eks-1-23-6",
		},
	}
	for _, opt := range opts {
		opt(kcp)
	}
	return kcp
}

func etcdCluster(opts ...func(*etcdv1.EtcdadmCluster)) *etcdv1.EtcdadmCluster {
	var etcdCluster *etcdv1.EtcdadmCluster = &etcdv1.EtcdadmCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EtcdadmCluster",
			APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-etcd",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: etcdv1.EtcdadmClusterSpec{
			EtcdadmConfigSpec: etcdadmbootstrapv1.EtcdadmConfigSpec{
				EtcdadmBuiltin: true,
				CloudInitConfig: &etcdadmbootstrapv1.CloudInitConfig{
					Version: "3.5.4",
				},
				CipherSuites: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			},
			InfrastructureTemplate: corev1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "DockerMachineTemplate",
				Name:       "test-etcd-1",
				Namespace:  constants.EksaSystemNamespace,
			},
			Replicas: ptr.Int32(3),
		},
	}
	for _, opt := range opts {
		opt(etcdCluster)
	}
	return etcdCluster
}
