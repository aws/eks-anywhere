package docker_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestWorkersSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()
	client := test.NewFakeKubeClient()

	workers, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(
		clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
			KubeadmConfigTemplate:   kubeadmConfigTemplate(),
			MachineDeployment:       machineDeployment(),
			ProviderMachineTemplate: dockerMachineTemplate("test-md-0-1"),
		},
		clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
			KubeadmConfigTemplate: kubeadmConfigTemplate(
				func(kct *bootstrapv1.KubeadmConfigTemplate) {
					kct.Name = "test-md-1-1"
				},
			),
			MachineDeployment: machineDeployment(
				func(md *clusterv1.MachineDeployment) {
					md.Name = "test-md-1"
					md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
					md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
				},
			),
			ProviderMachineTemplate: dockerMachineTemplate("test-md-1-1"),
		},
	))
}

func TestWorkersSpecUpgradeCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()

	currentGroup1 := clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: dockerMachineTemplate("test-md-0-1"),
	}

	currentGroup2 := clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
		KubeadmConfigTemplate: kubeadmConfigTemplate(
			func(kct *bootstrapv1.KubeadmConfigTemplate) {
				kct.Name = "test-md-1-1"
			},
		),
		MachineDeployment: machineDeployment(
			func(md *clusterv1.MachineDeployment) {
				md.Name = "test-md-1"
				md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
				md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
			},
		),
		ProviderMachineTemplate: dockerMachineTemplate("test-md-1-1"),
	}

	// Always make copies before passing to client since it does modifies the api objects
	// Like for example, the ResourceVersion
	expectedGroup1 := currentGroup1.DeepCopy()
	expectedGroup2 := currentGroup2.DeepCopy()

	objs := make([]kubernetes.Object, 0, 6)
	objs = append(objs, currentGroup1.Objects()...)
	objs = append(objs, currentGroup2.Objects()...)
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)

	// This will cause a change in the kubeadmconfigtemplate which we also treat as immutable
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints = []corev1.Taint{
		{
			Key:    "a",
			Value:  "accept",
			Effect: corev1.TaintEffectNoSchedule,
		},
	}
	expectedGroup1.KubeadmConfigTemplate.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints = []corev1.Taint{
		{
			Key:    "a",
			Value:  "accept",
			Effect: corev1.TaintEffectNoSchedule,
		},
	}
	expectedGroup1.KubeadmConfigTemplate.Name = "test-md-0-2"
	expectedGroup1.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"

	// This will cause a change in the docker machine templates, which are immutable
	spec.VersionsBundles["1.23"].EksD.KindNode = releasev1.Image{
		URI: "my-new-kind-image:tag",
	}
	expectedGroup1.ProviderMachineTemplate.Spec.Template.Spec.CustomImage = "my-new-kind-image:tag"
	expectedGroup1.ProviderMachineTemplate.Name = "test-md-0-2"
	expectedGroup1.MachineDeployment.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"

	expectedGroup2.ProviderMachineTemplate.Spec.Template.Spec.CustomImage = "my-new-kind-image:tag"
	expectedGroup2.ProviderMachineTemplate.Name = "test-md-1-2"
	expectedGroup2.MachineDeployment.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-2"

	workers, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(*expectedGroup1, *expectedGroup2))
}

func TestWorkersSpecUpgradeClusterRemoveLabels(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()

	kct := kubeadmConfigTemplate()
	kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{
		"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"cgroup-driver":     "cgroupfs",
		"node-labels":       "foo=bar",
		"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
	}

	currentGroup1 := clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
		KubeadmConfigTemplate:   kct,
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: dockerMachineTemplate("test-md-0-1"),
	}

	currentGroup2 := clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
		KubeadmConfigTemplate: kubeadmConfigTemplate(
			func(kct *bootstrapv1.KubeadmConfigTemplate) {
				kct.Name = "test-md-1-1"
			},
		),
		MachineDeployment: machineDeployment(
			func(md *clusterv1.MachineDeployment) {
				md.Name = "test-md-1"
				md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
				md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
			},
		),
		ProviderMachineTemplate: dockerMachineTemplate("test-md-1-1"),
	}

	// Always make copies before passing to client since it does modifies the api objects
	// Like for example, the ResourceVersion
	expectedGroup1 := currentGroup1.DeepCopy()
	expectedGroup2 := currentGroup2.DeepCopy()

	objs := make([]kubernetes.Object, 0, 6)
	objs = append(objs, currentGroup1.Objects()...)
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)

	// This will cause a change in the kubeadmconfigtemplate which we also treat as immutable
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Labels = map[string]string{}

	expectedGroup1.KubeadmConfigTemplate.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{
		"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"cgroup-driver":     "cgroupfs",
		"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
	}
	expectedGroup1.KubeadmConfigTemplate.Name = "test-md-0-2"
	expectedGroup1.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"

	workers, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(*expectedGroup1, *expectedGroup2))
}

func TestWorkersSpecNoMachineTemplateChanges(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()

	currentGroup1 := clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: dockerMachineTemplate("test-md-0-1"),
	}

	currentGroup2 := clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
		KubeadmConfigTemplate: kubeadmConfigTemplate(
			func(kct *bootstrapv1.KubeadmConfigTemplate) {
				kct.Name = "test-md-1-1"
			},
		),
		MachineDeployment: machineDeployment(
			func(md *clusterv1.MachineDeployment) {
				md.Name = "test-md-1"
				md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
				md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
			},
		),
		ProviderMachineTemplate: dockerMachineTemplate("test-md-1-1"),
	}

	// Always make copies before passing to client since it does modifies the api objects
	// Like for example, the ResourceVersion
	expectedGroup1 := currentGroup1.DeepCopy()
	expectedGroup2 := currentGroup2.DeepCopy()

	// This mimics what would happen if the objects were returned by a real api server
	// It helps make sure that the immutable object comparison is able to deal with these
	// kind of changes.
	currentGroup1.ProviderMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())
	currentGroup1.ProviderMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())

	// This is testing defaults. It's possible that some default logic will set items that are not set in our machine templates.
	// We need to take this into consideration when checking for equality.
	currentGroup1.ProviderMachineTemplate.Spec.Template.Spec.ProviderID = ptr.String("default-id")
	currentGroup2.ProviderMachineTemplate.Spec.Template.Spec.ProviderID = ptr.String("default-id")

	objs := make([]kubernetes.Object, 0, 6)
	objs = append(objs, currentGroup1.Objects()...)
	objs = append(objs, currentGroup2.Objects()...)
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)

	workers, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(*expectedGroup1, *expectedGroup2))
}

func TestWorkersSpecErrorFromClient(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()
	client := test.NewFakeKubeClientAlwaysError()
	_, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).To(MatchError(ContainSubstring("updating docker worker immutable object names")))
}

func TestWorkersSpecMachineTemplateNotFound(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec()
	client := test.NewFakeKubeClient(machineDeployment())
	_, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestWorkersSpecRegistryMirrorConfiguration(t *testing.T) {
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
			spec := testClusterSpec()
			spec.Cluster.Spec.RegistryMirrorConfiguration = tt.mirrorConfig
			workers, err := docker.WorkersSpec(ctx, logger, client, spec)

			g := NewWithT(t)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(workers).NotTo(BeNil())
			g.Expect(workers.Groups).To(HaveLen(2))
			g.Expect(workers.Groups).To(ConsistOf(
				clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
					KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
						kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, tt.files...)
						kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, test.RegistryMirrorPreKubeadmCommands()...)
					}),
					MachineDeployment:       machineDeployment(),
					ProviderMachineTemplate: dockerMachineTemplate("test-md-0-1"),
				},
				clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
					KubeadmConfigTemplate: kubeadmConfigTemplate(
						func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-1-1"
							kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, tt.files...)
							kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, test.RegistryMirrorPreKubeadmCommands()...)
						},
					),
					MachineDeployment: machineDeployment(
						func(md *clusterv1.MachineDeployment) {
							md.Name = "test-md-1"
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
						},
					),
					ProviderMachineTemplate: dockerMachineTemplate("test-md-1-1"),
				},
			))
		})
	}
}

func TestWorkersSpecUpgradeRolloutStrategy(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := testClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
			{
				Count:           ptr.Int(3),
				MachineGroupRef: &anywherev1.Ref{Name: "test"},
				Name:            "md-0",
				UpgradeRolloutStrategy: &anywherev1.WorkerNodesUpgradeRolloutStrategy{
					RollingUpdate: &anywherev1.WorkerNodesRollingUpdateParams{
						MaxSurge:       1,
						MaxUnavailable: 0,
					},
				},
			},
		}
	})
	client := test.NewFakeKubeClient()

	workers, err := docker.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(1))
	g.Expect(workers.Groups).To(ConsistOf(
		clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
			KubeadmConfigTemplate: kubeadmConfigTemplate(),
			MachineDeployment: machineDeployment(func(m *clusterv1.MachineDeployment) {
				maxSurge := intstr.FromInt(1)
				maxUnavailable := intstr.FromInt(0)
				m.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
					RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
						MaxSurge:       &maxSurge,
						MaxUnavailable: &maxUnavailable,
					},
				}
			}),
			ProviderMachineTemplate: dockerMachineTemplate("test-md-0-1"),
		},
	))
}

func kubeadmConfigTemplate(opts ...func(*bootstrapv1.KubeadmConfigTemplate)) *bootstrapv1.KubeadmConfigTemplate {
	o := &bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmConfigTemplate",
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-md-0-1",
			Namespace: "eksa-system",
		},
		Spec: bootstrapv1.KubeadmConfigTemplateSpec{
			Template: bootstrapv1.KubeadmConfigTemplateResource{
				Spec: bootstrapv1.KubeadmConfigSpec{
					JoinConfiguration: &bootstrapv1.JoinConfiguration{
						NodeRegistration: bootstrapv1.NodeRegistrationOptions{
							CRISocket: "/var/run/containerd/containerd.sock",
							KubeletExtraArgs: map[string]string{
								"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"cgroup-driver":     "cgroupfs",
								"eviction-hard":     "nodefs.available<0%,nodefs.inodesFree<0%,imagefs.available<0%",
							},
							Taints: []corev1.Taint{},
						},
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func machineDeployment(opts ...func(*clusterv1.MachineDeployment)) *clusterv1.MachineDeployment {
	o := &clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-md-0",
			Namespace: "eksa-system",
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: "test",
			Replicas:    ptr.Int32(3),
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{},
				Spec: clusterv1.MachineSpec{
					ClusterName: "test",
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &corev1.ObjectReference{
							Kind:       "KubeadmConfigTemplate",
							Name:       "test-md-0-1",
							APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
							Namespace:  "eksa-system",
						},
					},
					InfrastructureRef: corev1.ObjectReference{
						Kind:       "DockerMachineTemplate",
						Name:       "test-md-0-1",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
						Namespace:  "eksa-system",
					},
					Version: ptr.String("v1.23.12-eks-1-23-6"),
				},
			},
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
