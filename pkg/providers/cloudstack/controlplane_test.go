package cloudstack

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	testClusterConfigFilename = "testdata/cluster_main.yaml"
)

func TestControlPlaneObjects(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane *ControlPlane
		expected     []kubernetes.Object
	}{
		{
			name: "stacked etcd",
			controlPlane: &ControlPlane{
				BaseControlPlane: BaseControlPlane{
					Cluster:                     capiCluster(),
					ProviderCluster:             cloudstackCluster(),
					KubeadmControlPlane:         kubeadmControlPlane(),
					ControlPlaneMachineTemplate: cloudstackMachineTemplate("controlplane-machinetemplate"),
				},
			},
			expected: []kubernetes.Object{
				capiCluster(),
				cloudstackCluster(),
				kubeadmControlPlane(),
				cloudstackMachineTemplate("controlplane-machinetemplate"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tc.controlPlane.Objects()).To(ConsistOf(tc.expected))
		})
	}
}

func TestControlPlaneSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane()))
	g.Expect(cp.ProviderCluster).To(Equal(cloudstackCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal("test-control-plane-1"))
}

func TestControlPlaneSpecNoChangesMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	originalKCP := kubeadmControlPlane()
	originalCPMachineTemplate := cloudstackMachineTemplate("test-control-plane-1")

	expectedKCP := originalKCP.DeepCopy()
	expectedCPtemplate := originalCPMachineTemplate.DeepCopy()

	client := test.NewFakeKubeClient(
		originalKCP,
		originalCPMachineTemplate,
	)

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(expectedKCP))
	g.Expect(cp.ProviderCluster).To(Equal(cloudstackCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(expectedCPtemplate))
}

func TestControlPlaneSpecUpdateMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	originalKubeadmControlPlane := kubeadmControlPlane()
	originalCPMachineTemplate := cloudstackMachineTemplate("test-control-plane")
	expectedKCP := originalKubeadmControlPlane.DeepCopy()
	expectedCPTemplate := originalCPMachineTemplate.DeepCopy()

	client := test.NewFakeKubeClient(
		originalKubeadmControlPlane,
		originalCPMachineTemplate,
	)
	cpTaints := []corev1.Taint{
		{
			Key:    "foo",
			Value:  "bar",
			Effect: "PreferNoSchedule",
		},
	}
	spec.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints

	expectedKCP.Spec.KubeadmConfigSpec.InitConfiguration.NodeRegistration.Taints = cpTaints
	expectedKCP.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.Taints = cpTaints

	expectedCPTemplate.Name = "test-control-plane-1"

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(capiCluster()))
	g.Expect(cp.KubeadmControlPlane).To(Equal(expectedKCP))
	g.Expect(cp.ProviderCluster).To(Equal(cloudstackCluster()))
	g.Expect(cp.ControlPlaneMachineTemplate).To(Equal(expectedCPTemplate))
}

func TestControlPlaneSpecRegistryMirrorConfiguration(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
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
			spec.Cluster.Spec.RegistryMirrorConfiguration = tt.mirrorConfig
			cp, err := ControlPlaneSpec(ctx, logger, client, spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(cp).NotTo(BeNil())
			g.Expect(cp.Cluster).To(Equal(capiCluster()))
			g.Expect(cp.KubeadmControlPlane).To(Equal(kubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, tt.files...)
				precmds := []string{"swapoff -a"}
				precmds = append(precmds, test.RegistryMirrorSudoPreKubeadmCommands()...)
				precmds2 := []string{
					"hostname \"{{ ds.meta_data.hostname }}\"",
					"echo \"::1         ipv6-localhost ipv6-loopback\" >/etc/hosts",
					"echo \"127.0.0.1   localhost\" >>/etc/hosts",
					"echo \"127.0.0.1   {{ ds.meta_data.hostname }}\" >>/etc/hosts",
					"echo \"{{ ds.meta_data.hostname }}\" >/etc/hostname",
					"if [ ! -L /var/log/kubernetes ] ;\n  then\n    mv /var/log/kubernetes /var/log/kubernetes-$(tr -dc A-Za-z0-9 < /dev/urandom | head -c 10) ;\n    mkdir -p /data-small/var/log/kubernetes && ln -s /data-small/var/log/kubernetes /var/log/kubernetes ;\n  else echo \"/var/log/kubernetes already symlnk\";\nfi",
				}
				kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(precmds, precmds2...)
			})))
			g.Expect(cp.ProviderCluster).To(Equal(cloudstackCluster()))
			g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal("test-control-plane-1"))
		})
	}
}

func TestControlPlaneSpecWithUpgradeRolloutStrategy(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	spec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &anywherev1.ControlPlaneUpgradeRolloutStrategy{
		RollingUpdate: &anywherev1.ControlPlaneRollingUpdateParams{
			MaxSurge: 1,
		},
	}

	cp, err := ControlPlaneSpec(ctx, logger, client, spec)
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

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
			Labels:    map[string]string{"cluster.x-k8s.io/cluster-name": "test"},
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				APIServerPort: nil,
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"10.96.0.0/12"},
				},
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"192.168.0.0/16"},
				},
			},
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "1.2.3.4",
				Port: 6443,
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Name:       "test",
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
			ManagedExternalEtcdRef: &corev1.ObjectReference{
				Kind:       "EtcdadmCluster",
				Name:       "test-etcd",
				APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "CloudStackCluster",
				Name:       "test",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
			},
		},
	}
}

func cloudstackCluster() *cloudstackv1.CloudStackCluster {
	return &cloudstackv1.CloudStackCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudStackCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: cloudstackv1.CloudStackClusterSpec{
			FailureDomains: []cloudstackv1.CloudStackFailureDomainSpec{
				{
					Name: "default-az-0",
					Zone: cloudstackv1.CloudStackZoneSpec{
						Name:    "zone1",
						ID:      "",
						Network: cloudstackv1.Network{ID: "", Type: "", Name: "net1"},
					},
					Account: "admin",
					Domain:  "domain1",
					ACSEndpoint: corev1.SecretReference{
						Name:      "global",
						Namespace: "eksa-system",
					},
				},
			},
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "1.2.3.4",
				Port: 6443,
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
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
					Kind:       "CloudStackMachineTemplate",
					Name:       "test-control-plane-1",
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
							ImageTag:        "v1.8.3-eks-1-21-4",
						},
					},
					APIServer: bootstrapv1.APIServer{
						ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{
								"audit-log-maxage":    "30",
								"audit-log-maxbackup": "10",
								"profiling":           "false",
								"tls-cipher-suites":   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"audit-log-maxsize":   "512",
								"audit-log-path":      "/var/log/kubernetes/api-audit.log",
								"audit-policy-file":   "/etc/kubernetes/audit-policy.yaml",
								"cloud-provider":      "external",
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
							"cloud-provider":    "external",
							"profiling":         "false",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
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
						Path:        "/etc/kubernetes/manifests/kube-vip.yaml",
						Owner:       "root:root",
						Permissions: "",
						Encoding:    "",
						Append:      false,
						Content:     "apiVersion: v1\nkind: Pod\nmetadata:\n  creationTimestamp: null\n  name: kube-vip\n  namespace: kube-system\nspec:\n  containers:\n  - args:\n    - manager\n    env:\n    - name: vip_arp\n      value: \"true\"\n    - name: port\n      value: \"6443\"\n    - name: vip_cidr\n      value: \"32\"\n    - name: cp_enable\n      value: \"true\"\n    - name: cp_namespace\n      value: kube-system\n    - name: vip_ddns\n      value: \"false\"\n    - name: vip_leaderelection\n      value: \"true\"\n    - name: vip_leaseduration\n      value: \"15\"\n    - name: vip_renewdeadline\n      value: \"10\"\n    - name: vip_retryperiod\n      value: \"2\"\n    - name: address\n      value: 1.2.3.4\n    image: public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.158\n    imagePullPolicy: IfNotPresent\n    name: kube-vip\n    resources: {}\n    securityContext:\n      capabilities:\n        add:\n        - NET_ADMIN\n        - NET_RAW\n    volumeMounts:\n    - mountPath: /etc/kubernetes/admin.conf\n      name: kubeconfig\n  hostNetwork: true\n  volumes:\n  - hostPath:\n      path: /etc/kubernetes/admin.conf\n    name: kubeconfig\nstatus: {}\n",
					},
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
						Name:      "{{ ds.meta_data.hostname }}",
						CRISocket: "/var/run/containerd/containerd.sock",
						KubeletExtraArgs: map[string]string{
							"read-only-port":    "0",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							"anonymous-auth":    "false",
							"provider-id":       "cloudstack:///'{{ ds.meta_data.instance_id }}'",
						},
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						Name:      "{{ ds.meta_data.hostname }}",
						CRISocket: "/var/run/containerd/containerd.sock",
						KubeletExtraArgs: map[string]string{
							"provider-id":       "cloudstack:///'{{ ds.meta_data.instance_id }}'",
							"read-only-port":    "0",
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							"anonymous-auth":    "false",
						},
					},
				},
				DiskSetup: &bootstrapv1.DiskSetup{
					Partitions: []bootstrapv1.Partition{
						{Device: "/dev/vdb", Layout: true, Overwrite: ptr.Bool(false), TableType: ptr.String("gpt")},
					},
					Filesystems: []bootstrapv1.Filesystem{
						{
							Device:     "/dev/vdb1",
							Filesystem: "ext4",
							Label:      "data_disk",
							Partition:  nil,
							Overwrite:  ptr.Bool(false),
							ReplaceFS:  nil,
							ExtraOpts: []string{
								"-E",
								"lazy_itable_init=1,lazy_journal_init=1",
							},
						},
					},
				},
				Mounts: []bootstrapv1.MountPoints{
					[]string{"LABEL=data_disk", "/data-small"},
				},
				PreKubeadmCommands: []string{
					"swapoff -a",
					"hostname \"{{ ds.meta_data.hostname }}\"",
					"echo \"::1         ipv6-localhost ipv6-loopback\" >/etc/hosts",
					"echo \"127.0.0.1   localhost\" >>/etc/hosts",
					"echo \"127.0.0.1   {{ ds.meta_data.hostname }}\" >>/etc/hosts",
					"echo \"{{ ds.meta_data.hostname }}\" >/etc/hostname",
					"if [ ! -L /var/log/kubernetes ] ;\n  then\n    mv /var/log/kubernetes /var/log/kubernetes-$(tr -dc A-Za-z0-9 < /dev/urandom | head -c 10) ;\n    mkdir -p /data-small/var/log/kubernetes && ln -s /data-small/var/log/kubernetes /var/log/kubernetes ;\n  else echo \"/var/log/kubernetes already symlnk\";\nfi",
				},
				Users: []bootstrapv1.User{
					{
						Name:              "mySshUsername",
						Sudo:              ptr.String("ALL=(ALL) NOPASSWD:ALL"),
						SSHAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
					},
				},
				Format:                   "cloud-config",
				UseExperimentalRetryJoin: true,
			},
			Replicas: ptr.Int32(3),
			Version:  "v1.21.2-eks-1-21-4",
		},
	}

	for _, opt := range opts {
		opt(kcp)
	}
	return kcp
}

func cloudstackMachineTemplate(name string) *cloudstackv1.CloudStackMachineTemplate {
	return &cloudstackv1.CloudStackMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudStackMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
			Annotations: map[string]string{
				"mountpath.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1":  "/data-small",
				"symlinks.cloudstack.anywhere.eks.amazonaws.com/v1alpha1":                "/var/log/kubernetes:/data-small/var/log/kubernetes",
				"device.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1":     "/dev/vdb",
				"filesystem.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1": "ext4",
				"label.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1":      "data_disk",
			},
		},
		Spec: cloudstackv1.CloudStackMachineTemplateSpec{
			Template: cloudstackv1.CloudStackMachineTemplateResource{
				Spec: cloudstackv1.CloudStackMachineSpec{
					Template: cloudstackv1.CloudStackResourceIdentifier{
						Name: "kubernetes_1_21",
					},
					Offering: cloudstackv1.CloudStackResourceIdentifier{
						Name: "m4-large",
					},
					DiskOffering: cloudstackv1.CloudStackResourceDiskOffering{
						CloudStackResourceIdentifier: cloudstackv1.CloudStackResourceIdentifier{ID: "", Name: "Small"},
						MountPath:                    "/data-small",
						Device:                       "/dev/vdb",
						Filesystem:                   "ext4",
						Label:                        "data_disk",
					},
					AffinityGroupIDs: []string{
						"control-plane-anti-affinity",
					},
				},
			},
		},
	}
}
