package cloudstack_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestWorkersSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	client := test.NewFakeKubeClient()
	format.MaxLength = 40000
	workers, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(
		clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
			KubeadmConfigTemplate:   kubeadmConfigTemplate(),
			MachineDeployment:       machineDeployment(),
			ProviderMachineTemplate: machineTemplate(),
		},
		clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
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
					md.Spec.Replicas = ptr.Int32(2)
				},
			),
			ProviderMachineTemplate: machineTemplate(
				func(vmt *cloudstackv1.CloudStackMachineTemplate) {
					vmt.Name = "test-md-1-1"
				},
			),
		},
	))
}

func TestWorkersSpecUpgradeCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	oldGroup1 := &clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: machineTemplate(),
	}
	oldGroup2 := &clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
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
				md.Spec.Replicas = ptr.Int32(2)
			},
		),
		ProviderMachineTemplate: machineTemplate(
			func(vmt *cloudstackv1.CloudStackMachineTemplate) {
				vmt.Name = "test-md-1-1"
			},
		),
	}

	// Always make copies before passing to client since it does modifies the api objects
	// Like for example, the ResourceVersion
	expectedGroup1 := oldGroup1.DeepCopy()
	expectedGroup2 := oldGroup2.DeepCopy()

	objs := make([]kubernetes.Object, 0, 6)
	objs = append(objs, oldGroup1.Objects()...)
	objs = append(objs, oldGroup2.Objects()...)
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)

	computeOffering := anywherev1.CloudStackResourceIdentifier{
		Name: "m4-medium",
	}

	// This will cause a change in the cloudstack machine templates, which is immutable
	spec.CloudStackMachineConfigs["test"].Spec.ComputeOffering = *computeOffering.DeepCopy()

	// This will cause a change in the kubeadmconfigtemplate which we also treat as immutable
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints = []corev1.Taint{}
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[1].Taints = []corev1.Taint{}

	expectedGroup1.MachineDeployment.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
	expectedGroup1.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"
	expectedGroup1.KubeadmConfigTemplate.Name = "test-md-0-2"
	expectedGroup1.KubeadmConfigTemplate.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints = []corev1.Taint{}
	expectedGroup1.ProviderMachineTemplate.Name = "test-md-0-2"
	expectedGroup1.ProviderMachineTemplate.Spec.Spec.Spec.Offering = cloudstackv1.CloudStackResourceIdentifier{
		ID:   computeOffering.Id,
		Name: computeOffering.Name,
	}

	expectedGroup2.MachineDeployment.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-2"
	expectedGroup2.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-2"
	expectedGroup2.KubeadmConfigTemplate.Name = "test-md-1-2"
	expectedGroup2.KubeadmConfigTemplate.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints = []corev1.Taint{}
	expectedGroup2.ProviderMachineTemplate.Name = "test-md-1-2"
	expectedGroup2.ProviderMachineTemplate.Spec.Spec.Spec.Offering = cloudstackv1.CloudStackResourceIdentifier{
		ID:   computeOffering.Id,
		Name: computeOffering.Name,
	}

	workers, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(*expectedGroup1, *expectedGroup2))
}

func TestWorkersSpecUpgradeClusterNoMachineTemplateChanges(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	oldGroup1 := &clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: machineTemplate(),
	}
	oldGroup2 := &clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
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
				md.Spec.Replicas = ptr.Int32(2)
			},
		),
		ProviderMachineTemplate: machineTemplate(
			func(vmt *cloudstackv1.CloudStackMachineTemplate) {
				vmt.Name = "test-md-1-1"
			},
		),
	}

	// Always make copies before passing to client since it does modifies the api objects
	// Like for example, the ResourceVersion
	expectedGroup1 := oldGroup1.DeepCopy()
	expectedGroup2 := oldGroup2.DeepCopy()

	// This mimics what would happen if the objects were returned by a real api server
	// It helps make sure that the immutable object comparison is able to deal with these
	// kind of changes.
	oldGroup1.ProviderMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())
	oldGroup2.ProviderMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())

	// This is testing defaults. We don't set ID in our machine templates,
	// but it's possible that some default logic does. We need to take this into
	// consideration when checking for equality.
	oldGroup1.ProviderMachineTemplate.Spec.Spec.Spec.ProviderID = ptr.String("default-id")
	oldGroup2.ProviderMachineTemplate.Spec.Spec.Spec.ProviderID = ptr.String("default-id")

	client := test.NewFakeKubeClient(
		oldGroup1.MachineDeployment,
		oldGroup1.KubeadmConfigTemplate,
		oldGroup1.ProviderMachineTemplate,
		oldGroup2.MachineDeployment,
		oldGroup2.KubeadmConfigTemplate,
		oldGroup2.ProviderMachineTemplate,
	)

	workers, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(*expectedGroup1, *expectedGroup2))
}

func TestWorkersSpecErrorFromClient(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	client := test.NewFakeKubeClientAlwaysError()
	_, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).To(MatchError(ContainSubstring("updating cloudstack worker immutable object names")))
}

func TestWorkersSpecMachineTemplateNotFound(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	client := test.NewFakeKubeClient(machineDeployment())
	_, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestWorkersSpecRegistryMirrorConfiguration(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	client := test.NewFakeKubeClient()
	format.MaxLength = 40000
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
			format.MaxLength = 40000

			spec.Cluster.Spec.RegistryMirrorConfiguration = tt.mirrorConfig
			workers, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(workers).NotTo(BeNil())
			g.Expect(workers.Groups).To(HaveLen(2))
			g.Expect(workers.Groups).To(ConsistOf(
				clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
						kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, tt.files...)
						preKubeadmCommands := append([]string{"swapoff -a"}, test.RegistryMirrorSudoPreKubeadmCommands()...)
						kct.Spec.Template.Spec.PreKubeadmCommands = append(preKubeadmCommands, kct.Spec.Template.Spec.PreKubeadmCommands[1:]...)
					}),
					MachineDeployment:       machineDeployment(),
					ProviderMachineTemplate: machineTemplate(),
				},
				clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					KubeadmConfigTemplate: kubeadmConfigTemplate(
						func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-1-1"
							kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, tt.files...)
							preKubeadmCommands := append([]string{"swapoff -a"}, test.RegistryMirrorSudoPreKubeadmCommands()...)
							kct.Spec.Template.Spec.PreKubeadmCommands = append(preKubeadmCommands, kct.Spec.Template.Spec.PreKubeadmCommands[1:]...)
						},
					),
					MachineDeployment: machineDeployment(
						func(md *clusterv1.MachineDeployment) {
							md.Name = "test-md-1"
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
							md.Spec.Replicas = ptr.Int32(2)
						},
					),
					ProviderMachineTemplate: machineTemplate(
						func(vmt *cloudstackv1.CloudStackMachineTemplate) {
							vmt.Name = "test-md-1-1"
						},
					),
				},
			))
		})
	}
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
			Labels:    map[string]string{"cluster.x-k8s.io/cluster-name": "test"},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: "test",
			Replicas:    ptr.Int32(3),
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels: map[string]string{"cluster.x-k8s.io/cluster-name": "test"},
				},
				Spec: clusterv1.MachineSpec{
					ClusterName: "test",
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &corev1.ObjectReference{
							Kind:       "KubeadmConfigTemplate",
							Name:       "test-md-0-1",
							APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
						},
					},
					InfrastructureRef: corev1.ObjectReference{
						Kind:       "CloudStackMachineTemplate",
						Name:       "test-md-0-1",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
					},
					Version: ptr.String("v1.21.2-eks-1-21-4"),
				},
			},
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
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
							Name:      "{{ ds.meta_data.hostname }}",
							CRISocket: "/var/run/containerd/containerd.sock",
							Taints: []corev1.Taint{
								{
									Key:       "key2",
									Value:     "val2",
									Effect:    "PreferNoSchedule",
									TimeAdded: nil,
								},
							},
							KubeletExtraArgs: map[string]string{
								"anonymous-auth":    "false",
								"provider-id":       "cloudstack:///'{{ ds.meta_data.instance_id }}'",
								"read-only-port":    "0",
								"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							},
						},
					},
					PreKubeadmCommands: []string{
						`swapoff -a`,
						`hostname "{{ ds.meta_data.hostname }}"`,
						`echo "::1         ipv6-localhost ipv6-loopback" >/etc/hosts`,
						`echo "127.0.0.1   localhost" >>/etc/hosts`,
						`echo "127.0.0.1   {{ ds.meta_data.hostname }}" >>/etc/hosts`,
						`echo "{{ ds.meta_data.hostname }}" >/etc/hostname`,
					},
					Users: []bootstrapv1.User{
						{
							Name:              "mySshUsername",
							Sudo:              ptr.String("ALL=(ALL) NOPASSWD:ALL"),
							SSHAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"},
						},
					},
					Format: bootstrapv1.Format("cloud-config"),
				},
			},
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

func machineTemplate(opts ...func(*cloudstackv1.CloudStackMachineTemplate)) *cloudstackv1.CloudStackMachineTemplate {
	o := &cloudstackv1.CloudStackMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudStackMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-md-0-1",
			Namespace: "eksa-system",
		},
		Spec: cloudstackv1.CloudStackMachineTemplateSpec{
			Spec: cloudstackv1.CloudStackMachineTemplateResource{
				Spec: cloudstackv1.CloudStackMachineSpec{
					Details: map[string]string{"foo": "bar"},
					Offering: cloudstackv1.CloudStackResourceIdentifier{
						Name: "m4-large",
					},
					Template: cloudstackv1.CloudStackResourceIdentifier{
						ID:   "",
						Name: "centos7-k8s-118",
					},
					AffinityGroupIDs: []string{"worker-affinity"},
					Affinity:         "",
				},
			},
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
