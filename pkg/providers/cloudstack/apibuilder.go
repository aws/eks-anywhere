package cloudstack

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
)

const (
	CloudStackMachineTemplateKind = "CloudStackMachineTemplate"
	CloudStackInfrastructureAPIVersion = "infrastructure.cluster.x-k8s.io/v1beta2"
)

func machineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration, kubeadmConfigTemplate *bootstrapv1.KubeadmConfigTemplate, cloudstackMachineTemplate *cloudstackv1.CloudStackMachineTemplate) clusterv1.MachineDeployment {
	return clusterapi.MachineDeployment(clusterSpec, workerNodeGroupConfig, kubeadmConfigTemplate, cloudstackMachineTemplate)
}

func MachineDeployments(clusterSpec *cluster.Spec, kubeadmConfigTemplates map[string]*bootstrapv1.KubeadmConfigTemplate, machineTemplates map[string]*cloudstackv1.CloudStackMachineTemplate) map[string]*clusterv1.MachineDeployment {
	m := make(map[string]*clusterv1.MachineDeployment, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		deployment := machineDeployment(clusterSpec, workerNodeGroupConfig,
			kubeadmConfigTemplates[workerNodeGroupConfig.Name],
			machineTemplates[workerNodeGroupConfig.Name],
		)
		m[workerNodeGroupConfig.Name] = &deployment
	}
	return m
}

func fillMachineTemplateAnnotations(machineConfig *v1alpha1.CloudStackMachineConfig) map[string]string {
	annotations := make(map[string]string, 0)
	if machineConfig.Spec.DiskOffering != nil {
		annotations[fmt.Sprintf("mountpath.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.Spec.DiskOffering.MountPath
		annotations[fmt.Sprintf("device.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.Spec.DiskOffering.Device
		annotations[fmt.Sprintf("filesystem.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.Spec.DiskOffering.Filesystem
		annotations[fmt.Sprintf("label.diskoffering.%s", constants.CloudstackAnnotationSuffix)] = machineConfig.Spec.DiskOffering.Label
	}

	if machineConfig.Spec.Symlinks != nil {
		links := make([]string, 0)
		for key := range machineConfig.Spec.Symlinks {
			links = append(links, fmt.Sprintf("%s:%s",key, machineConfig.Spec.Symlinks[key]))
		}
		annotations[fmt.Sprintf("symlinks.%s", constants.CloudstackAnnotationSuffix)] = strings.Join(links, ",")
	}

	return annotations
}

func setDiskOffering(machineConfig *v1alpha1.CloudStackMachineConfig, template *cloudstackv1.CloudStackMachineTemplate) {
	if machineConfig.Spec.DiskOffering == nil {
		return
	}

	template.Spec.Spec.Spec.DiskOffering = cloudstackv1.CloudStackResourceDiskOffering{
		CloudStackResourceIdentifier: cloudstackv1.CloudStackResourceIdentifier{
			ID: machineConfig.Spec.DiskOffering.Id,
			Name: machineConfig.Spec.DiskOffering.Name,
		},
		CustomSize:                   machineConfig.Spec.DiskOffering.CustomSize,
		MountPath:                    machineConfig.Spec.DiskOffering.MountPath,
		Device:                       machineConfig.Spec.DiskOffering.Device,
		Filesystem:                   machineConfig.Spec.DiskOffering.Filesystem,
		Label:                        machineConfig.Spec.DiskOffering.Label,
	}
}

func CloudStackMachineTemplate(name string, machineConfig *v1alpha1.CloudStackMachineConfig) *cloudstackv1.CloudStackMachineTemplate {
	template := &cloudstackv1.CloudStackMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: CloudStackInfrastructureAPIVersion,
			Kind:       CloudStackMachineTemplateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
			Annotations: fillMachineTemplateAnnotations(machineConfig),
		},
		Spec: cloudstackv1.CloudStackMachineTemplateSpec{
			Spec: cloudstackv1.CloudStackMachineTemplateResource{
				Spec: cloudstackv1.CloudStackMachineSpec{
					Details: machineConfig.Spec.UserCustomDetails,
					Offering: cloudstackv1.CloudStackResourceIdentifier{
						ID:   machineConfig.Spec.ComputeOffering.Id,
						Name: machineConfig.Spec.ComputeOffering.Name,
					},
					Template: cloudstackv1.CloudStackResourceIdentifier{
						ID:   machineConfig.Spec.Template.Id,
						Name: machineConfig.Spec.Template.Name,
					},
					AffinityGroupIDs: machineConfig.Spec.AffinityGroupIds,
					Affinity: machineConfig.Spec.Affinity,
				},
			},
		},
	}
	setDiskOffering(machineConfig, template)

	return template
}
