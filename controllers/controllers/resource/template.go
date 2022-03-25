package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/templater"
	anywhereTypes "github.com/aws/eks-anywhere/pkg/types"
)

const (
	ConfigMapKind       = "ConfigMap"
	EKSIamConfigMapName = "aws-auth"
)

type DockerTemplate struct {
	ResourceFetcher
	now anywhereTypes.NowFunc
}

type VsphereTemplate struct {
	ResourceFetcher
	ResourceUpdater
	now anywhereTypes.NowFunc
}

type CloudStackTemplate struct {
	ResourceFetcher
	ResourceUpdater
	now anywhereTypes.NowFunc
}

type AWSIamConfigTemplate struct {
	ResourceFetcher
}

type TinkerbellTemplate struct {
	ResourceFetcher
	now anywhereTypes.NowFunc
}

func (r *VsphereTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, vdc anywherev1.VSphereDatacenterConfig, cpVmc, etcdVmc anywherev1.VSphereMachineConfig, workerVmcs map[string]anywherev1.VSphereMachineConfig) ([]*unstructured.Unstructured, error) {
	workerNodeGroupMachineSpecs := make(map[string]anywherev1.VSphereMachineConfigSpec, len(workerVmcs))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = workerVmcs[wnConfig.MachineGroupRef.Name].Spec
	}
	// control plane and etcd updates are prohibited in controller so those specs should not change
	templateBuilder := vsphere.NewVsphereTemplateBuilder(&vdc.Spec, &cpVmc.Spec, &etcdVmc.Spec, workerNodeGroupMachineSpecs, r.now, true)
	clusterName := clusterSpec.Cluster.Name

	oldVdc, err := r.ExistingVSphereDatacenterConfig(ctx, eksaCluster, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0])
	if err != nil {
		return nil, err
	}
	oldCpVmc, err := r.ExistingVSphereControlPlaneMachineConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}

	var controlPlaneTemplateName string
	updateControlPlaneTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, oldCpVmc, &cpVmc)
	if updateControlPlaneTemplate {
		controlPlaneTemplateName = common.CPMachineTemplateName(clusterName, r.now)
	} else {
		cp, err := r.ControlPlane(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		controlPlaneTemplateName = cp.Spec.MachineTemplate.InfrastructureRef.Name
	}

	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		oldWn, err := r.ExistingWorkerNodeGroupConfig(ctx, eksaCluster, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		vmc := workerVmcs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		oldVmc, err := r.ExistingVSphereWorkerMachineConfig(ctx, eksaCluster, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		if vsphere.NeedsNewKubeadmConfigTemplate(&workerNodeGroupConfiguration, oldWn, oldVmc, &vmc) {
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = common.KubeadmConfigTemplateName(clusterName, workerNodeGroupConfiguration.Name, r.now)
		} else {
			md, err := r.MachineDeployment(ctx, eksaCluster, workerNodeGroupConfiguration)
			if err != nil {
				return nil, err
			}
			workloadKubeadmConfigTemplateName := md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = workloadKubeadmConfigTemplateName
		}

		updateWorkloadTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, oldVmc, &vmc)
		if updateWorkloadTemplate {
			workloadTemplateName := common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, r.now)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			md, err := r.MachineDeployment(ctx, eksaCluster, workerNodeGroupConfiguration)
			if err != nil {
				return nil, err
			}
			workloadTemplateName := md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
	}

	var etcdTemplateName string
	if eksaCluster.Spec.ExternalEtcdConfiguration != nil {
		oldEtcdVmc, err := r.ExistingVSphereEtcdMachineConfig(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		updateEtcdTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, oldEtcdVmc, &etcdVmc)
		etcd, err := r.Etcd(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		if updateEtcdTemplate {
			etcd.SetAnnotations(map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"})
			if err := r.ApplyPatch(ctx, etcd, false); err != nil {
				return nil, err
			}
			etcdTemplateName = common.EtcdMachineTemplateName(clusterName, r.now)
		} else {
			etcdTemplateName = etcd.Spec.InfrastructureTemplate.Name
		}
	}

	// Get vsphere credentials so that the template can apply correctly instead of with empty values
	credSecret, err := r.VSphereCredentials(ctx)
	if err != nil {
		return nil, err
	}
	usernameBytes, ok := credSecret.Data["username"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve username from secret")
	}
	passwordBytes, ok := credSecret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve password from secret")
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["vsphereControlPlaneSshAuthorizedKey"] = sshAuthorizedKey(cpVmc.Spec.Users)
		values["vsphereEtcdSshAuthorizedKey"] = sshAuthorizedKey(etcdVmc.Spec.Users)
		values["etcdTemplateName"] = etcdTemplateName
		values["eksaVsphereUsername"] = string(usernameBytes)
		values["eksaVspherePassword"] = string(passwordBytes)
	}

	return generateTemplateResources(templateBuilder, clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames, cpOpt, nil)
}

func (r *CloudStackTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, csdc anywherev1.CloudStackDatacenterConfig, cpCsmc, etcdCsmc anywherev1.CloudStackMachineConfig, workerCsmcs map[string]anywherev1.CloudStackMachineConfig) ([]*unstructured.Unstructured, error) {
	if len(workerCsmcs) > 1 {
		return nil, fmt.Errorf("multiple worker node group configuration is not supported in CloudStack provider")
	}
	workerNodeGroupMachineSpecs := make(map[string]anywherev1.CloudStackMachineConfigSpec, len(workerCsmcs))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = workerCsmcs[wnConfig.MachineGroupRef.Name].Spec
	}
	// control plane and etcd updates are prohibited in controller so those specs should not change
	templateBuilder := cloudstack.NewCloudStackTemplateBuilder(&csdc.Spec, &cpCsmc.Spec, &etcdCsmc.Spec, workerNodeGroupMachineSpecs, r.now)
	clusterName := clusterSpec.Cluster.Name

	oldCsdc, err := r.ExistingCloudStackDatacenterConfig(ctx, eksaCluster, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0])
	if err != nil {
		return nil, err
	}
	oldCpCsmc, err := r.ExistingCloudStackControlPlaneMachineConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}

	controlPlaneTemplateName, err := r.getControlPlaneTemplateName(ctx, eksaCluster, oldCsdc, csdc, oldCpCsmc, cpCsmc, clusterName)
	if err != nil {
		return nil, err
	}

	kubeadmconfigTemplateNames, err := r.getKubeadmconfigTemplateNames(ctx, eksaCluster, clusterSpec, clusterName)
	if err != nil {
		return nil, err
	}
	workloadTemplateNames, err := r.getWorkloadTemplateNames(ctx, eksaCluster, clusterSpec, oldCsdc, csdc, workerCsmcs, clusterName)
	if err != nil {
		return nil, err
	}
	etcdTemplateName, err := r.getEtcdTemplateName(ctx, eksaCluster, oldCsdc, csdc, etcdCsmc, clusterName)
	if err != nil {
		return nil, err
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["cloudstackControlPlaneSshAuthorizedKey"] = sshAuthorizedKey(cpCsmc.Spec.Users)
		values["cloudstackEtcdSshAuthorizedKey"] = sshAuthorizedKey(etcdCsmc.Spec.Users)
		values["etcdTemplateName"] = etcdTemplateName
	}

	workersOpt := func(values map[string]interface{}) {
		values["workloadTemplateName"] = workloadTemplateNames[clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name]
		values["cloudStackWorkerSshAuthorizedKey"] = sshAuthorizedKey(workerNodeGroupMachineSpecs[clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name].Users)
	}

	return generateTemplateResources(templateBuilder, clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames, cpOpt, workersOpt)
}

func (r *CloudStackTemplate) getControlPlaneTemplateName(ctx context.Context, eksaCluster *anywherev1.Cluster, oldCsdc *anywherev1.CloudStackDatacenterConfig, csdc anywherev1.CloudStackDatacenterConfig, oldCpCsmc *anywherev1.CloudStackMachineConfig, cpCsmc anywherev1.CloudStackMachineConfig, clusterName string) (string, error) {
	var controlPlaneTemplateName string
	updateControlPlaneTemplate := cloudstack.AnyImmutableFieldChanged(oldCsdc, &csdc, oldCpCsmc, &cpCsmc)
	if updateControlPlaneTemplate {
		controlPlaneTemplateName = common.CPMachineTemplateName(clusterName, r.now)
	} else {
		cp, err := r.ControlPlane(ctx, eksaCluster)
		if err != nil {
			return "", err
		}
		controlPlaneTemplateName = cp.Spec.MachineTemplate.InfrastructureRef.Name
	}
	return controlPlaneTemplateName, nil
}

func (r *CloudStackTemplate) getEtcdTemplateName(ctx context.Context, eksaCluster *anywherev1.Cluster, oldCsdc *anywherev1.CloudStackDatacenterConfig, csdc anywherev1.CloudStackDatacenterConfig, etcdCsmc anywherev1.CloudStackMachineConfig, clusterName string) (string, error) {
	var etcdTemplateName string
	if eksaCluster.Spec.ExternalEtcdConfiguration != nil {
		oldEtcdCsmc, err := r.ExistingCloudStackEtcdMachineConfig(ctx, eksaCluster)
		if err != nil {
			return "", err
		}
		updateEtcdTemplate := cloudstack.AnyImmutableFieldChanged(oldCsdc, &csdc, oldEtcdCsmc, &etcdCsmc)
		etcd, err := r.Etcd(ctx, eksaCluster)
		if err != nil {
			return "", err
		}
		if updateEtcdTemplate {
			etcd.SetAnnotations(map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"})
			if err := r.ApplyPatch(ctx, etcd, false); err != nil {
				return "", err
			}
			etcdTemplateName = common.EtcdMachineTemplateName(clusterName, r.now)
		} else {
			etcdTemplateName = etcd.Spec.InfrastructureTemplate.Name
		}
	}
	return etcdTemplateName, nil
}

func (r *CloudStackTemplate) getKubeadmconfigTemplateNames(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, clusterName string) (map[string]string, error) {
	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		oldWn, err := r.ExistingWorkerNodeGroupConfig(ctx, eksaCluster, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		if cloudstack.NeedsNewKubeadmConfigTemplate(&workerNodeGroupConfiguration, oldWn) {
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = common.KubeadmConfigTemplateName(clusterName, workerNodeGroupConfiguration.Name, r.now)
		} else {
			md, err := r.MachineDeployment(ctx, eksaCluster, workerNodeGroupConfiguration)
			if err != nil {
				return nil, err
			}
			workloadKubeadmConfigTemplateName := md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
			kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name] = workloadKubeadmConfigTemplateName
		}
	}
	return kubeadmconfigTemplateNames, nil
}

func (r *CloudStackTemplate) getWorkloadTemplateNames(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, oldCsdc *anywherev1.CloudStackDatacenterConfig, csdc anywherev1.CloudStackDatacenterConfig, workerCsmcs map[string]anywherev1.CloudStackMachineConfig, clusterName string) (map[string]string, error) {
	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		csmc := workerCsmcs[workerNodeGroupConfiguration.MachineGroupRef.Name]
		oldCsmc, err := r.ExistingCloudStackWorkerMachineConfig(ctx, eksaCluster, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		updateWorkloadTemplate := cloudstack.AnyImmutableFieldChanged(oldCsdc, &csdc, oldCsmc, &csmc)
		if updateWorkloadTemplate {
			workloadTemplateName := common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, r.now)
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		} else {
			md, err := r.MachineDeployment(ctx, eksaCluster, workerNodeGroupConfiguration)
			if err != nil {
				return nil, err
			}
			workloadTemplateName := md.Spec.Template.Spec.InfrastructureRef.Name
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = workloadTemplateName
		}
	}
	return workloadTemplateNames, nil
}

// TODO(pokearu): This method is currently not used. Need to add logic in reconciler for TinkerbellDatacenterKind
func (r *TinkerbellTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, tdc anywherev1.TinkerbellDatacenterConfig, cpTmc, etcdTmc anywherev1.TinkerbellMachineConfig, workerTmc map[string]anywherev1.TinkerbellMachineConfig) ([]*unstructured.Unstructured, error) {
	workerNodeGroupMachineSpecs := make(map[string]anywherev1.TinkerbellMachineConfigSpec, len(workerTmc))
	for _, wnConfig := range workerTmc {
		workerNodeGroupMachineSpecs[wnConfig.Name] = wnConfig.Spec
	}
	templateBuilder := tinkerbell.NewTinkerbellTemplateBuilder(&tdc.Spec, &cpTmc.Spec, &etcdTmc.Spec, workerNodeGroupMachineSpecs, r.now)
	cp, err := r.ControlPlane(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	var etcdTemplateName string
	if eksaCluster.Spec.ExternalEtcdConfiguration != nil {
		etcd, err := r.Etcd(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		etcdTemplateName = etcd.Spec.InfrastructureTemplate.Name
	}

	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		mcDeployment, err := r.MachineDeployment(ctx, eksaCluster, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		workloadTemplateNames[workerNodeGroupConfiguration.Name] = mcDeployment.Spec.Template.Spec.InfrastructureRef.Name
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = cp.Spec.MachineTemplate.InfrastructureRef.Name
		values["tinkerbellControlPlaneSshAuthorizedKey"] = sshAuthorizedKey(cpTmc.Spec.Users)
		values["tinkerbellEtcdSshAuthorizedKey"] = sshAuthorizedKey(etcdTmc.Spec.Users)
		values["etcdTemplateName"] = etcdTemplateName
	}

	return generateTemplateResources(templateBuilder, clusterSpec, nil, nil, cpOpt, nil)
}

func generateTemplateResources(builder providers.TemplateBuilder, clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string, cpOpt providers.BuildMapOption, workerOpt providers.BuildMapOption) ([]*unstructured.Unstructured, error) {
	cp, err := builder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, err
	}
	md, err := builder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	if err != nil {
		return nil, err
	}
	content := templater.AppendYamlResources(cp, md)
	var resources []*unstructured.Unstructured
	templates := strings.Split(string(content), "---")
	for _, template := range templates {
		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(template), u); err != nil {
			continue
		}
		if u.GetKind() != "" {
			resources = append(resources, u)
		}
	}
	return resources, nil
}

func (r *DockerTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec) ([]*unstructured.Unstructured, error) {
	clusterName := clusterSpec.Cluster.Name
	bundle := clusterSpec.VersionsBundle
	templateBuilder := docker.NewDockerTemplateBuilder(r.now)

	existingVersion, err := r.ExistingKubeVersion(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	existingControlPlaneNodeImage, err := r.ExistingControlPlaneKindNodeImage(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	// Check to see if there is any change the Kubernetes tag that requires a new template in order to specify the new
	// node image
	kubeVersionChanged := existingVersion != bundle.KubeDistro.Kubernetes.Tag
	var controlPlaneTemplateName string
	if kubeVersionChanged {
		controlPlaneTemplateName = common.CPMachineTemplateName(clusterName, r.now)
	} else {
		kubeadmControlPlane, err := r.ControlPlane(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		controlPlaneTemplateName = kubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name
	}

	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		existingWorkerNodeImage, err := r.ExistingWorkerKindNodeImage(ctx, eksaCluster, workerNodeGroupConfiguration)
		if err != nil {
			return nil, err
		}
		// If Kubernetes version did not change but we have a newer kind node image for the same version, we will roll
		// out new worker nodes to consume the latest image instead of only change on Kubernetes versions for control
		// plane
		if kubeVersionChanged || existingWorkerNodeImage != bundle.EksD.KindNode.VersionedImage() {
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = common.WorkerMachineTemplateName(clusterName, workerNodeGroupConfiguration.Name, r.now)
		} else {
			mcDeployment, err := r.MachineDeployment(ctx, eksaCluster, workerNodeGroupConfiguration)
			if err != nil {
				return nil, err
			}
			workloadTemplateNames[workerNodeGroupConfiguration.Name] = mcDeployment.Spec.Template.Spec.InfrastructureRef.Name
		}
	}

	var etcdTemplateName string
	if eksaCluster.Spec.ExternalEtcdConfiguration != nil {
		if kubeVersionChanged {
			etcdTemplateName = common.EtcdMachineTemplateName(clusterName, r.now)
		} else {
			etcd, err := r.Etcd(ctx, eksaCluster)
			if err != nil {
				return nil, err
			}
			etcdTemplateName = etcd.Spec.InfrastructureTemplate.Name
		}
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["etcdTemplateName"] = etcdTemplateName
		if !kubeVersionChanged {
			values["kindNodeImage"] = existingControlPlaneNodeImage
		}
	}

	return generateTemplateResources(templateBuilder, clusterSpec, workloadTemplateNames, nil, cpOpt, nil)
}

func sshAuthorizedKey(users []anywherev1.UserConfiguration) string {
	if len(users) <= 0 || len(users[0].SshAuthorizedKeys) <= 0 {
		return ""
	}
	return users[0].SshAuthorizedKeys[0]
}

func (r *AWSIamConfigTemplate) TemplateResources(ctx context.Context, clusterSpec *cluster.Spec) ([]*unstructured.Unstructured, error) {
	var resources []*unstructured.Unstructured
	templateBuilder := awsiamauth.NewAwsIamAuthTemplateBuilder()
	content, err := templateBuilder.GenerateManifest(clusterSpec, uuid.Nil)
	if err != nil {
		return nil, err
	}
	templates := strings.Split(string(content), "---")
	for _, template := range templates {
		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(template), u); err != nil {
			continue
		}

		// Only reconcile IAM role mappings ConfigMap
		if u.GetKind() == ConfigMapKind && u.GetName() == EKSIamConfigMapName {
			resources = append(resources, u)
		}
	}
	return resources, nil
}
