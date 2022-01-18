package resource

import (
	"context"
	"fmt"
	"strings"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/templater"
	anywhereTypes "github.com/aws/eks-anywhere/pkg/types"
)

const ConfigMapKind = "ConfigMap"

type DockerTemplate struct {
	ResourceFetcher
	now anywhereTypes.NowFunc
}

type VsphereTemplate struct {
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

func (r *VsphereTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, vdc anywherev1.VSphereDatacenterConfig, cpVmc, etcdVmc anywherev1.VSphereMachineConfig, workerVmc []*anywherev1.VSphereMachineConfig) ([]*unstructured.Unstructured, error) {
	workerNodeGroupMachineSpecs := make(map[string]anywherev1.VSphereMachineConfigSpec, len(workerVmc))
	for _, wnConfig := range workerVmc {
		workerNodeGroupMachineSpecs[wnConfig.Name] = wnConfig.Spec
	}
	// control plane and etcd updates are prohibited in controller so those specs should not change
	templateBuilder := vsphere.NewVsphereTemplateBuilder(&vdc.Spec, &cpVmc.Spec, &etcdVmc.Spec, workerNodeGroupMachineSpecs, r.now)
	clusterName := clusterSpec.ObjectMeta.Name

	oldVdc, err := r.ExistingVSphereDatacenterConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldCpVmc, err := r.ExistingVSphereControlPlaneMachineConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldWorkerVmcs, err := r.ExistingVSphereWorkerMachineConfigs(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}

	var controlPlaneTemplateName string
	updateControlPlaneTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, oldCpVmc, &cpVmc)
	if updateControlPlaneTemplate {
		controlPlaneTemplateName = templateBuilder.CPMachineTemplateName(clusterName)
	} else {
		cp, err := r.ControlPlane(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		controlPlaneTemplateName = cp.Spec.InfrastructureTemplate.Name
	}

	var workloadTemplateNames []string
	for _, vmc := range workerVmc {
		oldVmc := oldWorkerVmcs[vmc.Name]
		updateWorkloadTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, &oldVmc, vmc)
		if updateWorkloadTemplate {
			workloadTemplateName := templateBuilder.WorkerMachineTemplateName(clusterName, clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Name)
			workloadTemplateNames = append(workloadTemplateNames, workloadTemplateName)
		} else {
			mcDeployments, err := r.MachineDeployments(ctx, eksaCluster)
			if err != nil {
				return nil, err
			}
			for _, md := range mcDeployments {
				workloadTemplateName := md.Spec.Template.Spec.InfrastructureRef.Name
				workloadTemplateNames = append(workloadTemplateNames, workloadTemplateName)
			}
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
			etcdTemplateName = templateBuilder.EtcdMachineTemplateName(clusterName)
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

	return generateTemplateResources(templateBuilder, clusterSpec, workloadTemplateNames, cpOpt)
}

func (r *TinkerbellTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, tdc anywherev1.TinkerbellDatacenterConfig, cpTmc, etcdTmc anywherev1.TinkerbellMachineConfig, workerTmc map[string]anywherev1.TinkerbellMachineConfig) ([]*unstructured.Unstructured, error) {
	workerNodeGroupMachineSpecs := make(map[string]anywherev1.TinkerbellMachineConfigSpec, len(workerTmc))
	for _, wnConfig := range workerTmc {
		workerNodeGroupMachineSpecs[wnConfig.Name] = wnConfig.Spec
	}
	var workerTemplateNames []string
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

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = cp.Spec.InfrastructureTemplate.Name
		values["tinkerbellControlPlaneSshAuthorizedKey"] = sshAuthorizedKey(cpTmc.Spec.Users)
		values["tinkerbellEtcdSshAuthorizedKey"] = sshAuthorizedKey(etcdTmc.Spec.Users)
		values["etcdTemplateName"] = etcdTemplateName
	}

	return generateTemplateResources(templateBuilder, clusterSpec, workerTemplateNames, cpOpt)
}

func generateTemplateResources(builder providers.TemplateBuilder, clusterSpec *cluster.Spec, templateNames []string, cpOpt providers.BuildMapOption) ([]*unstructured.Unstructured, error) {
	cp, err := builder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, err
	}
	md, err := builder.GenerateCAPISpecWorkers(clusterSpec)
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
	var workerTemplateNames []string
	templateBuilder := docker.NewDockerTemplateBuilder(r.now)
	kubeadmControlPlane, err := r.ControlPlane(ctx, eksaCluster)
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

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = kubeadmControlPlane.Spec.InfrastructureTemplate.Name
		values["etcdTemplateName"] = etcdTemplateName
	}
	return generateTemplateResources(templateBuilder, clusterSpec, workerTemplateNames, cpOpt)
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
	content, err := templateBuilder.GenerateManifest(clusterSpec)
	if err != nil {
		return nil, err
	}
	templates := strings.Split(string(content), "---")
	for _, template := range templates {
		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(template), u); err != nil {
			continue
		}
		if u.GetKind() == ConfigMapKind {
			resources = append(resources, u)
		}
	}
	return resources, nil
}
