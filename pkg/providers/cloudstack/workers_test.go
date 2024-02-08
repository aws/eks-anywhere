package cloudstack_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestWorkersSpec(t *testing.T) {
	logger := test.NewNullLogger()
	ctx := context.Background()

	spec := test.NewFullClusterSpec(t, "testdata/test_worker_spec.yaml")

	for _, tc := range []struct {
		Name      string
		Configure func(*cluster.Spec)
		Exists    func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]
		Expect    func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]
	}{
		// Create
		{
			Name: "Create",
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate:   kubeadmConfigTemplate(),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
		},
		{
			Name: "CreateMultipleWorkerNodeGroups",
			Configure: func(s *cluster.Spec) {
				// Re-use the existing worker node group.
				s.Cluster.Spec.WorkerNodeGroupConfigurations = append(
					s.Cluster.Spec.WorkerNodeGroupConfigurations,
					s.Cluster.Spec.WorkerNodeGroupConfigurations[0],
				)
				s.Cluster.Spec.WorkerNodeGroupConfigurations[1].Name = "md-1"
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate:   kubeadmConfigTemplate(),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-1-1"
						}),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Name = "test-md-1"
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-1-1"
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-1-1"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-1-1"
						}),
					},
				}
			},
		},
		{
			Name: "CreateTaints",
			Configure: func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints = []corev1.Taint{
					{
						Key:    "test-taint",
						Value:  "value",
						Effect: "Effect",
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints = []corev1.Taint{
								{
									Key:    "test-taint",
									Value:  "value",
									Effect: "Effect",
								},
							}
						}),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {}),
					},
				}
			},
		},
		{
			Name: "CreateDiskOffering",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.DiskOffering = &anywherev1.CloudStackResourceDiskOffering{
					CustomSize: 10,
					MountPath:  "/mnt/sda",
					Device:     "/dev/sda",
					Filesystem: "ext3",
					Label:      "label",
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment:     machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							spec := &csmt.Spec.Template.Spec

							clientutil.AddAnnotation(csmt, "device.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "/dev/sda")
							clientutil.AddAnnotation(csmt, "filesystem.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "ext3")
							clientutil.AddAnnotation(csmt, "label.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "label")
							clientutil.AddAnnotation(csmt, "mountpath.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "/mnt/sda")

							spec.DiskOffering = cloudstackv1.CloudStackResourceDiskOffering{
								CustomSize: 10,
								MountPath:  "/mnt/sda",
								Device:     "/dev/sda",
								Filesystem: "ext3",
								Label:      "label",
							}
						}),
					},
				}
			},
		},
		{
			Name: "CreateSymlinks",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.Symlinks = map[string]string{"foo": "bar"}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Spec.Template.Spec.PreKubeadmCommands = append(
								kct.Spec.Template.Spec.PreKubeadmCommands,
								"if [ ! -L foo ] ;\n  then\n    mv foo foo-$(tr -dc A-Za-z0-9 \u003c /dev/urandom | head -c 10) ;\n    mkdir -p bar \u0026\u0026 ln -s bar foo ;\n  else echo \"foo already symlnk\" ;\nfi",
							)
						}),
						MachineDeployment: machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							clientutil.AddAnnotation(csmt, "symlinks.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "foo:bar")
						}),
					},
				}
			},
		},
		{
			Name: "CreateAffinityGroupIDs",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.AffinityGroupIds = []string{"affinity_group_id"}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment:     machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Spec.Template.Spec.AffinityGroupIDs = []string{"affinity_group_id"}
						}),
					},
				}
			},
		},
		{
			Name: "CreateUserCustomDetails",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.UserCustomDetails = map[string]string{"qux": "baz"}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment:     machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Spec.Template.Spec.Details = map[string]string{"qux": "baz"}
						}),
					},
				}
			},
		},

		// Upgrade
		{
			Name: "UpgradeTaints",
			Configure: func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints = []corev1.Taint{
					{
						Key:    "change-taint",
						Value:  "value",
						Effect: "Effect",
					},
				}
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints = []corev1.Taint{
								{
									Key:    "test-taint",
									Value:  "value",
									Effect: "Effect",
								},
							}
						}),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-0-2"
							kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints = []corev1.Taint{
								{
									Key:    "change-taint",
									Value:  "value",
									Effect: "Effect",
								},
							}
						}),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
		},
		{
			Name: "UpgradeComputeOffering",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.ComputeOffering = anywherev1.CloudStackResourceIdentifier{
					Name: "m4-medium",
				}
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate:   kubeadmConfigTemplate(),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-0-1"
						}),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-1"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.Offering = cloudstackv1.CloudStackResourceIdentifier{
								Name: "m4-medium",
							}
						}),
					},
				}
			},
		},
		{
			Name: "UpgradeDiskOffering",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.DiskOffering = &anywherev1.CloudStackResourceDiskOffering{
					CustomSize: 10,
					MountPath:  "/mnt/sda",
					Device:     "/dev/sda",
					Filesystem: "ext3",
					Label:      "label",
				}
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate:   kubeadmConfigTemplate(),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.DiskOffering = cloudstackv1.CloudStackResourceDiskOffering{
								CustomSize: 10,
								MountPath:  "/mnt/sda",
								Device:     "/dev/sda",
								Filesystem: "ext3",
								Label:      "label",
							}
							clientutil.AddAnnotation(csmt, "device.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "/dev/sda")
							clientutil.AddAnnotation(csmt, "filesystem.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "ext3")
							clientutil.AddAnnotation(csmt, "label.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "label")
							clientutil.AddAnnotation(csmt, "mountpath.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "/mnt/sda")
						}),
					},
				}
			},
		},
		{
			Name: "UpgradeSymlinks",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.Symlinks = map[string]string{"foo": "bar"}
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate:   kubeadmConfigTemplate(),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-0-2"
							kct.Spec.Template.Spec.PreKubeadmCommands = append(
								kct.Spec.Template.Spec.PreKubeadmCommands,
								"if [ ! -L foo ] ;\n  then\n    mv foo foo-$(tr -dc A-Za-z0-9 \u003c /dev/urandom | head -c 10) ;\n    mkdir -p bar \u0026\u0026 ln -s bar foo ;\n  else echo \"foo already symlnk\" ;\nfi",
							)
						}),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							clientutil.AddAnnotation(csmt, "symlinks.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "foo:bar")
						}),
					},
				}
			},
		},
		{
			Name: "UpgradeAffinityGroups",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.AffinityGroupIds = []string{"changed"}
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate:   kubeadmConfigTemplate(),
						MachineDeployment:       machineDeployment(),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.AffinityGroupIDs = []string{"changed"}
						}),
					},
				}
			},
		},
		{
			Name: "UpgradeUserCustomDetails",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.UserCustomDetails = map[string]string{"qux": "baz"}
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment:     machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Spec.Template.Spec.Details = map[string]string{"foo": "bar"}
						}),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.Details = map[string]string{"qux": "baz"}
						}),
					},
				}
			},
		},

		// Remove
		{
			Name: "RemoveDiskOffering",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.DiskOffering = nil
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.DiskOffering = cloudstackv1.CloudStackResourceDiskOffering{
								CustomSize: 10,
								MountPath:  "/mnt/sda",
								Device:     "/dev/sda",
								Filesystem: "ext3",
								Label:      "label",
							}
							clientutil.AddAnnotation(csmt, "device.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "/dev/sda")
							clientutil.AddAnnotation(csmt, "filesystem.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "ext3")
							clientutil.AddAnnotation(csmt, "label.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "label")
							clientutil.AddAnnotation(csmt, "mountpath.diskoffering.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "/mnt/sda")
						}),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-3"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-3"
						}),
					},
				}
			},
		},
		{
			Name: "RemoveSymlinks",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.Symlinks = nil
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-0-2"
							kct.Spec.Template.Spec.PreKubeadmCommands = append(
								kct.Spec.Template.Spec.PreKubeadmCommands,
								"if [ ! -L foo ] ;\n  then\n    mv foo foo-$(tr -dc A-Za-z0-9 \u003c /dev/urandom | head -c 10) ;\n    mkdir -p bar \u0026\u0026 ln -s bar foo ;\n  else echo \"foo already symlnk\" ;\nfi",
							)
						}),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							clientutil.AddAnnotation(csmt, "symlinks.cloudstack.anywhere.eks.amazonaws.com/v1alpha1", "foo:bar")
						}),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(func(kct *bootstrapv1.KubeadmConfigTemplate) {
							kct.Name = "test-md-0-2"
						}),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.Bootstrap.ConfigRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(),
					},
				}
			},
		},
		{
			Name: "RemoveAffinityGroups",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.AffinityGroupIds = nil
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment:     machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Spec.Template.Spec.AffinityGroupIDs = []string{"affinity_group_id"}
						}),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.AffinityGroupIDs = nil
						}),
					},
				}
			},
		},
		{
			Name: "RemoveUserCustomDetails",
			Configure: func(s *cluster.Spec) {
				s.CloudStackMachineConfigs["test"].Spec.UserCustomDetails = nil
			},
			Exists: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment:     machineDeployment(),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Spec.Template.Spec.Details = map[string]string{"foo": "bar"}
						}),
					},
				}
			},
			Expect: func() []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate] {
				return []clusterapi.WorkerGroup[*cloudstackv1.CloudStackMachineTemplate]{
					{
						KubeadmConfigTemplate: kubeadmConfigTemplate(),
						MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
							md.Spec.Template.Spec.InfrastructureRef.Name = "test-md-0-2"
						}),
						ProviderMachineTemplate: machineTemplate(func(csmt *cloudstackv1.CloudStackMachineTemplate) {
							csmt.Name = "test-md-0-2"
							csmt.Spec.Template.Spec.Details = nil
						}),
					},
				}
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			g := NewWithT(t)

			// Copy the foundational spec already read from disk so we don't pollute tests.
			spec := spec.DeepCopy()

			if tc.Configure != nil {
				tc.Configure(spec)
			}

			// Build a client with all the objects that should already exist in the cluster.
			var objects []kubernetes.Object
			if tc.Exists != nil {
				for _, group := range tc.Exists() {
					objects = append(objects, group.Objects()...)
				}
			}
			client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objects)...)

			expect := tc.Expect()

			workers, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
			g.Expect(err).NotTo(HaveOccurred())

			// Optionally dump expect and got. This proved useful in debugging as the Ginkgo output
			// gets truncated. Compare the files in your IDE.
			if os.Getenv("T_DUMP") == "true" {
				expectGroups, _ := json.MarshalIndent(expect, "", "\t")
				receivedGroups, _ := json.MarshalIndent(workers.Groups, "", "\t")
				_ = os.WriteFile("groups_expected.json", expectGroups, 0o666)
				_ = os.WriteFile("groups_received.json", receivedGroups, 0o666)
				_ = os.WriteFile("groups_expected_received.diff", []byte(cmp.Diff(expectGroups, receivedGroups)), 0o666)
			}

			g.Expect(workers).NotTo(BeNil())
			g.Expect(workers.Groups).To(HaveLen(len(expect)))
			for _, e := range expect {
				g.Expect(workers.Groups).To(ContainElement(e))
			}
		})
	}
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

func TestWorkersSpecUpgradeRolloutStrategy(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := test.NewFullClusterSpec(t, "testdata/test_worker_spec.yaml")
	spec.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
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
	client := test.NewFakeKubeClient()

	workers, err := cloudstack.WorkersSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(workers).NotTo(BeNil())
	g.Expect(workers.Groups).To(HaveLen(1))
	g.Expect(workers.Groups[0].MachineDeployment).To(Equal(machineDeployment(func(m *clusterv1.MachineDeployment) {
		maxSurge := intstr.FromInt(1)
		maxUnavailable := intstr.FromInt(0)
		m.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
			RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
				MaxSurge:       &maxSurge,
				MaxUnavailable: &maxUnavailable,
			},
		}
	})))
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
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
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
							SSHAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
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
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-md-0-1",
			Namespace: "eksa-system",
		},
		Spec: cloudstackv1.CloudStackMachineTemplateSpec{
			Template: cloudstackv1.CloudStackMachineTemplateResource{
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
