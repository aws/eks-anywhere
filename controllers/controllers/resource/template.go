package resource

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/templater"
	anywhereTypes "github.com/aws/eks-anywhere/pkg/types"
)

type DockerTemplate struct {
	ResourceFetcher
	now anywhereTypes.NowFunc
}

type VsphereTemplate struct {
	ResourceFetcher
	now anywhereTypes.NowFunc
}

func (r *VsphereTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, vdcSpec anywherev1.VSphereDatacenterConfig, vmcSpec anywherev1.VSphereMachineConfig) ([]*unstructured.Unstructured, error) {
	// passing the same vspheremachineconfig for worker and cp as control plane updates are prohibited in controller
	templateBuilder := vsphere.NewVsphereTemplateBuilder(&vdcSpec.Spec, &vmcSpec.Spec, &vmcSpec.Spec, &vmcSpec.Spec, r.now)
	clusterName := clusterSpec.ObjectMeta.Name
	kubeadmControlPlane, err := r.ControlPlane(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldVdcSpec, err := r.ExistingVSphereDatacenterConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldVmcSpec, err := r.ExistingVSphereWorkerMachineConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	var workloadTemplateName string
	updateWorkloadTemplate := vsphere.AnyImmutableFieldChanged(oldVdcSpec, &vdcSpec, oldVmcSpec, &vmcSpec)
	if updateWorkloadTemplate {
		workloadTemplateName = templateBuilder.WorkerMachineTemplateName(clusterName)
	} else {
		mcDeployment, err := r.MachineDeployment(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		workloadTemplateName = mcDeployment.Spec.Template.Spec.InfrastructureRef.Name
	}
	if len(vmcSpec.Spec.Users) <= 0 {
		vmcSpec.Spec.Users = []anywherev1.UserConfiguration{{}}
	}
	if len(vmcSpec.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		vmcSpec.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}
	valuesOpt := func(values map[string]interface{}) {
		values["needsNewControlPlaneTemplate"] = false // not supported in flux
		values["controlPlaneTemplateName"] = kubeadmControlPlane.Spec.InfrastructureTemplate.Name
		values["needsNewWorkloadTemplate"] = updateWorkloadTemplate
		values["workloadTemplateName"] = workloadTemplateName
		values["clusterName"] = clusterName
		values["vsphereWorkerSshAuthorizedKey"] = vmcSpec.Spec.Users[0].SshAuthorizedKeys[0]
		values["needsNewEtcdTemplate"] = false
	}
	return generateTemplateResources(templateBuilder, clusterSpec, valuesOpt)
}

func generateTemplateResources(builder providers.TemplateBuilder, clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) ([]*unstructured.Unstructured, error) {
	cp, err := builder.GenerateClusterApiSpecCP(clusterSpec, buildOptions...)
	if err != nil {
		return nil, err
	}
	md, err := builder.GenerateClusterApiSpecMD(clusterSpec, buildOptions...)
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
	templateBuilder := docker.NewDockerTemplateBuilder(r.now)
	mcDeployment, err := r.MachineDeployment(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	kubeadmControlPlane, err := r.ControlPlane(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	valuesOpt := func(values map[string]interface{}) {
		values["clusterName"] = clusterSpec.ObjectMeta.Name
		values["needsNewControlPlaneTemplate"] = false
		values["controlPlaneTemplateName"] = kubeadmControlPlane.Spec.InfrastructureTemplate.Name
		values["needsNewWorkloadTemplate"] = false
		values["workloadTemplateName"] = mcDeployment.Spec.Template.Spec.InfrastructureRef.Name
		values["needsNewEtcdTemplate"] = false
	}
	return generateTemplateResources(templateBuilder, clusterSpec, valuesOpt)
}
