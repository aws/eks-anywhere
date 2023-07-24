package cloudstack

import (
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// CloudStackMachineTemplateKind defines the K8s Kind corresponding with the MachineTemplate.
const CloudStackMachineTemplateKind = "CloudStackMachineTemplate"

func machineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration, kubeadmConfigTemplate *bootstrapv1.KubeadmConfigTemplate, cloudstackMachineTemplate *cloudstackv1.CloudStackMachineTemplate) *clusterv1.MachineDeployment {
	return clusterapi.MachineDeployment(clusterSpec, workerNodeGroupConfig, kubeadmConfigTemplate, cloudstackMachineTemplate)
}

// MachineDeployments returns generated CAPI MachineDeployment objects for a given cluster spec.
func MachineDeployments(clusterSpec *cluster.Spec, kubeadmConfigTemplates map[string]*bootstrapv1.KubeadmConfigTemplate, machineTemplates map[string]*cloudstackv1.CloudStackMachineTemplate) map[string]*clusterv1.MachineDeployment {
	m := make(map[string]*clusterv1.MachineDeployment, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		deployment := machineDeployment(clusterSpec, workerNodeGroupConfig,
			kubeadmConfigTemplates[workerNodeGroupConfig.Name],
			machineTemplates[workerNodeGroupConfig.Name],
		)

		m[workerNodeGroupConfig.Name] = deployment
	}
	return m
}

func generateMachineTemplateAnnotations(machineConfig *v1alpha1.CloudStackMachineConfigSpec) map[string]string {
	annotations := make(map[string]string, 0)
	if machineConfig.DiskOffering != nil {
		annotations[fmt.Sprintf("mountpath.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.DiskOffering.MountPath
		annotations[fmt.Sprintf("device.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.DiskOffering.Device
		annotations[fmt.Sprintf("filesystem.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.DiskOffering.Filesystem
		annotations[fmt.Sprintf("label.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.DiskOffering.Label
	}

	if machineConfig.Symlinks != nil {
		links := make([]string, 0)
		for key := range machineConfig.Symlinks {
			links = append(links, fmt.Sprintf("%s:%s", key, machineConfig.Symlinks[key]))
		}
		// sorting for unit test determinism
		sort.Strings(links)
		annotations[fmt.Sprintf("symlinks.%s", constants.CloudstackAnnotationSuffix)] = strings.Join(links, ",")
	}

	return annotations
}

func setDiskOffering(machineConfig *v1alpha1.CloudStackMachineConfigSpec, template *cloudstackv1.CloudStackMachineTemplate) {
	if machineConfig.DiskOffering == nil {
		return
	}

	template.Spec.Template.Spec.DiskOffering = cloudstackv1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: cloudstackv1.CloudStackResourceIdentifier{
			ID:   machineConfig.DiskOffering.Id,
			Name: machineConfig.DiskOffering.Name,
		},
		CustomSize: machineConfig.DiskOffering.CustomSize,
		MountPath:  machineConfig.DiskOffering.MountPath,
		Device:     machineConfig.DiskOffering.Device,
		Filesystem: machineConfig.DiskOffering.Filesystem,
		Label:      machineConfig.DiskOffering.Label,
	}
}

// MachineTemplate returns a generated CloudStackMachineTemplate object for a given EKS-A CloudStackMachineConfig.
func MachineTemplate(name string, machineConfig *v1alpha1.CloudStackMachineConfigSpec) *cloudstackv1.CloudStackMachineTemplate {
	template := &cloudstackv1.CloudStackMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: cloudstackv1.GroupVersion.String(),
			Kind:       CloudStackMachineTemplateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   constants.EksaSystemNamespace,
			Annotations: generateMachineTemplateAnnotations(machineConfig),
		},
		Spec: cloudstackv1.CloudStackMachineTemplateSpec{
			Template: cloudstackv1.CloudStackMachineTemplateResource{
				Spec: cloudstackv1.CloudStackMachineSpec{
					Details: machineConfig.UserCustomDetails,
					Offering: cloudstackv1.CloudStackResourceIdentifier{
						ID:   machineConfig.ComputeOffering.Id,
						Name: machineConfig.ComputeOffering.Name,
					},
					Template: cloudstackv1.CloudStackResourceIdentifier{
						ID:   machineConfig.Template.Id,
						Name: machineConfig.Template.Name,
					},
					AffinityGroupIDs: machineConfig.AffinityGroupIds,
					Affinity:         machineConfig.Affinity,
				},
			},
		},
	}
	setDiskOffering(machineConfig, template)

	return template
}
