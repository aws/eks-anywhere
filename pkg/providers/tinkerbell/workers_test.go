package tinkerbell_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestWorkersSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_multiple_node_groups.yaml")
	client := test.NewFakeKubeClient()

	workers, err := tinkerbell.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(
		clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
			KubeadmConfigTemplate:   kubeadmConfigTemplate(),
			MachineDeployment:       machineDeployment(),
			ProviderMachineTemplate: machineTemplate(),
		},
		clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
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
					md.Spec.Replicas = ptr.Int32(1)
					md.Labels["pool"] = "md-1"
					md.Spec.Template.ObjectMeta.Labels["pool"] = "md-1"
					md.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
						Type: "",
						RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
							MaxUnavailable: &intstr.IntOrString{Type: 0, IntVal: 3, StrVal: ""},
							MaxSurge:       &intstr.IntOrString{Type: 0, IntVal: 5, StrVal: ""},
							DeletePolicy:   nil,
						},
					}
				},
			),
			ProviderMachineTemplate: machineTemplate(
				func(tmt *tinkerbellv1.TinkerbellMachineTemplate) {
					tmt.Name = "test-md-1-1"
				},
			),
		},
	))
}

func TestWorkersSpecNewClusterInPlaceRolloutStrategy(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_multiple_node_groups.yaml")
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].UpgradeRolloutStrategy = &anywherev1.WorkerNodesUpgradeRolloutStrategy{
		Type: "InPlace",
	}
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[1].UpgradeRolloutStrategy = &anywherev1.WorkerNodesUpgradeRolloutStrategy{
		Type: "InPlace",
	}
	client := test.NewFakeKubeClient()

	workers, err := tinkerbell.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(
		clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
			KubeadmConfigTemplate: kubeadmConfigTemplate(),
			MachineDeployment: machineDeployment(
				func(md *clusterv1.MachineDeployment) {
					md.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
						Type: "InPlace",
					}
				},
			),
			ProviderMachineTemplate: machineTemplate(),
		},
		clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
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
					md.Spec.Replicas = ptr.Int32(1)
					md.Labels["pool"] = "md-1"
					md.Spec.Template.ObjectMeta.Labels["pool"] = "md-1"
					md.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
						Type: "InPlace",
					}
				},
			),
			ProviderMachineTemplate: machineTemplate(
				func(tmt *tinkerbellv1.TinkerbellMachineTemplate) {
					tmt.Name = "test-md-1-1"
				},
			),
		},
	))
}

func TestWorkersSpecUpgradeClusterNoMachineTemplateChanges(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_multiple_node_groups.yaml")
	oldGroup1 := &clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
		KubeadmConfigTemplate:   kubeadmConfigTemplate(),
		MachineDeployment:       machineDeployment(),
		ProviderMachineTemplate: machineTemplate(),
	}
	oldGroup2 := &clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
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
				md.Spec.Replicas = ptr.Int32(1)
				md.Labels["pool"] = "md-1"
				md.Spec.Template.ObjectMeta.Labels["pool"] = "md-1"
				md.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
					Type: "",
					RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
						MaxUnavailable: &intstr.IntOrString{Type: 0, IntVal: 3, StrVal: ""},
						MaxSurge:       &intstr.IntOrString{Type: 0, IntVal: 5, StrVal: ""},
						DeletePolicy:   nil,
					},
				}
			},
		),
		ProviderMachineTemplate: machineTemplate(
			func(vmt *tinkerbellv1.TinkerbellMachineTemplate) {
				vmt.Name = "test-md-1-1"
			},
		),
	}

	expectedGroup1 := oldGroup1.DeepCopy()
	expectedGroup2 := oldGroup2.DeepCopy()

	oldGroup1.ProviderMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())
	oldGroup2.ProviderMachineTemplate.CreationTimestamp = metav1.NewTime(time.Now())

	client := test.NewFakeKubeClient(
		oldGroup1.MachineDeployment,
		oldGroup1.KubeadmConfigTemplate,
		oldGroup1.ProviderMachineTemplate,
		oldGroup2.MachineDeployment,
		oldGroup2.KubeadmConfigTemplate,
		oldGroup2.ProviderMachineTemplate,
	)

	workers, err := tinkerbell.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(2))
	g.Expect(workers.Groups).To(ConsistOf(*expectedGroup1, *expectedGroup2))
}

func TestWorkersSpecMachineTemplateNotFound(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_multiple_node_groups.yaml")
	client := test.NewFakeKubeClient(machineDeployment())
	_, err := tinkerbell.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestWorkersSpecErrorFromClient(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_multiple_node_groups.yaml")
	client := test.NewFakeKubeClientAlwaysError()

	_, err := tinkerbell.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).To(HaveOccurred())
}

func TestWorkersSpecRegistryMirrorInsecureSkipVerify(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_multiple_node_groups.yaml")
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
			name:         "insecure skip verify with cacert",
			mirrorConfig: test.RegistryMirrorInsecureSkipVerifyEnabledAndCACert(),
			files:        test.RegistryMirrorConfigFilesInsecureSkipVerifyAndCACert(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec.Cluster.Spec.RegistryMirrorConfiguration = tt.mirrorConfig
			workers, err := tinkerbell.WorkersSpec(ctx, logger, client, spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(workers).NotTo(BeNil())
			g.Expect(workers.Groups).To(HaveLen(2))
			g.Expect(workers.Groups).To(ConsistOf(
				clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
					KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
						kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, tt.files...)
						kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, test.RegistryMirrorSudoPreKubeadmCommands()...)
					}),
					MachineDeployment:       machineDeployment(),
					ProviderMachineTemplate: machineTemplate(),
				},
				clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
					KubeadmConfigTemplate: kubeadmConfigTemplate(
						func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-1-1"
							kct.Spec.Template.Spec.Files = append(kct.Spec.Template.Spec.Files, tt.files...)
							kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, test.RegistryMirrorSudoPreKubeadmCommands()...)
						},
					),
					MachineDeployment: machineDeployment(
						func(md *clusterv1.MachineDeployment) {
							md.Name = "test-md-1"
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
							md.Spec.Replicas = ptr.Int32(1)
							md.Labels["pool"] = "md-1"
							md.Spec.Template.ObjectMeta.Labels["pool"] = "md-1"
							md.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
								Type: "",
								RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
									MaxUnavailable: &intstr.IntOrString{Type: 0, IntVal: 3, StrVal: ""},
									MaxSurge:       &intstr.IntOrString{Type: 0, IntVal: 5, StrVal: ""},
									DeletePolicy:   nil,
								},
							}
						},
					),
					ProviderMachineTemplate: machineTemplate(
						func(tmt *tinkerbellv1.TinkerbellMachineTemplate) {
							tmt.Name = "test-md-1-1"
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
			Labels:    map[string]string{"cluster.x-k8s.io/cluster-name": "test", "pool": "md-0"},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: "test",
			Replicas:    ptr.Int32(1),
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels: map[string]string{"cluster.x-k8s.io/cluster-name": "test", "pool": "md-0"},
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
						Kind:       "TinkerbellMachineTemplate",
						Name:       "test-md-0-1",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					},
					Version: ptr.String("v1.21.2-eks-1-21-4"),
				},
			},
			Strategy: &clusterv1.MachineDeploymentStrategy{
				Type: "",
				RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: 0, IntVal: 0, StrVal: ""},
					MaxSurge:       &intstr.IntOrString{Type: 0, IntVal: 1, StrVal: ""},
					DeletePolicy:   nil,
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
							Name:      "",
							CRISocket: "",
							Taints:    nil,
							KubeletExtraArgs: map[string]string{
								"anonymous-auth":    "false",
								"provider-id":       "PROVIDER_ID",
								"read-only-port":    "0",
								"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							},
						},
					},
					Users: []bootstrapv1.User{
						{
							Name:              "tink-user",
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

func machineTemplate(opts ...func(*tinkerbellv1.TinkerbellMachineTemplate)) *tinkerbellv1.TinkerbellMachineTemplate {
	o := &tinkerbellv1.TinkerbellMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-md-0-1",
			Namespace: "eksa-system",
		},
		Spec: tinkerbellv1.TinkerbellMachineTemplateSpec{
			Template: tinkerbellv1.TinkerbellMachineTemplateResource{
				Spec: tinkerbellv1.TinkerbellMachineSpec{
					TemplateOverride: "global_timeout: 6000\nid: \"\"\nname: tink-test\ntasks:\n- actions:\n  - environment:\n      COMPRESSED: \"true\"\n      DEST_DISK: /dev/sda\n      IMG_URL: \"\"\n    image: image2disk:v1.0.0\n    name: stream-image\n    timeout: 360\n  - environment:\n      BLOCK_DEVICE: /dev/sda2\n      CHROOT: \"y\"\n      CMD_LINE: apt -y update && apt -y install openssl\n      DEFAULT_INTERPRETER: /bin/sh -c\n      FS_TYPE: ext4\n    image: cexec:v1.0.0\n    name: install-openssl\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        network:\n          version: 2\n          renderer: networkd\n          ethernets:\n              eno1:\n                  dhcp4: true\n              eno2:\n                  dhcp4: true\n              eno3:\n                  dhcp4: true\n              eno4:\n                  dhcp4: true\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/netplan/config.yaml\n      DIRMODE: \"0755\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0644\"\n      UID: \"0\"\n    image: writefile:v1.0.0\n    name: write-netplan\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        datasource:\n          Ec2:\n            metadata_urls: []\n            strict_id: false\n        system_info:\n          default_user:\n            name: tink\n            groups: [wheel, adm]\n            sudo: [\"ALL=(ALL) NOPASSWD:ALL\"]\n            shell: /bin/bash\n        manage_etc_hosts: localhost\n        warnings:\n          dsid_missing_source: off\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0600\"\n    image: writefile:v1.0.0\n    name: add-tink-cloud-init-config\n    timeout: 90\n  - environment:\n      CONTENTS: |\n        datasource: Ec2\n      DEST_DISK: /dev/sda2\n      DEST_PATH: /etc/cloud/ds-identify.cfg\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0600\"\n      UID: \"0\"\n    image: writefile:v1.0.0\n    name: add-tink-cloud-init-ds-config\n    timeout: 90\n  - environment:\n      BLOCK_DEVICE: /dev/sda2\n      FS_TYPE: ext4\n    image: kexec:v1.0.0\n    name: kexec-image\n    pid: host\n    timeout: 90\n  name: tink-test\n  volumes:\n  - /dev:/dev\n  - /dev/console:/dev/console\n  - /lib/firmware:/lib/firmware:ro\n  worker: '{{.device_1}}'\nversion: \"0.1\"\n",
					HardwareAffinity: &tinkerbellv1.HardwareAffinity{
						Required: []tinkerbellv1.HardwareAffinityTerm{
							{
								LabelSelector: metav1.LabelSelector{MatchLabels: map[string]string{"type": "worker"}},
							},
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
