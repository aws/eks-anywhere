package resource

import (
	"context"
	"strings"

	etcdv1alpha3 "github.com/mrajashree/etcdadm-controller/api/v1alpha3"
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
	ResourceUpdater
	now anywhereTypes.NowFunc
}

func (r *VsphereTemplate) TemplateResources(ctx context.Context, eksaCluster *anywherev1.Cluster, clusterSpec *cluster.Spec, vdc anywherev1.VSphereDatacenterConfig, cpVmc, workerVmc, etcdVmc anywherev1.VSphereMachineConfig) ([]*unstructured.Unstructured, error) {
	// control plane and etcd updates are prohibited in controller so those specs should not change
	templateBuilder := vsphere.NewVsphereTemplateBuilder(&vdc.Spec, &cpVmc.Spec, &workerVmc.Spec, &etcdVmc.Spec, r.now)
	clusterName := clusterSpec.ObjectMeta.Name

	oldVdc, err := r.ExistingVSphereDatacenterConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldCpVmc, err := r.ExistingVSphereControlPlaneMachineConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldEtcdVmc, err := r.ExistingVSphereEtcdMachineConfig(ctx, eksaCluster)
	if err != nil {
		return nil, err
	}
	oldWorkerVmc, err := r.ExistingVSphereWorkerMachineConfig(ctx, eksaCluster)
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

	var workloadTemplateName string
	updateWorkloadTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, oldWorkerVmc, &workerVmc)
	if updateWorkloadTemplate {
		workloadTemplateName = templateBuilder.WorkerMachineTemplateName(clusterName)
	} else {
		mcDeployment, err := r.MachineDeployment(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		workloadTemplateName = mcDeployment.Spec.Template.Spec.InfrastructureRef.Name
	}

	var etcdTemplateName string
	if eksaCluster.Spec.ExternalEtcdConfiguration != nil {
		updateEtcdTemplate := vsphere.AnyImmutableFieldChanged(oldVdc, &vdc, oldEtcdVmc, &etcdVmc)
		etcd, err := r.Etcd(ctx, eksaCluster)
		if err != nil {
			return nil, err
		}
		if updateEtcdTemplate {
			etcd.SetAnnotations(map[string]string{etcdv1alpha3.UpgradeInProgressAnnotation: "true"})
			if err := r.ApplyPatch(ctx, etcd, false); err != nil {
				return nil, err
			}
			etcdTemplateName = templateBuilder.EtcdMachineTemplateName(clusterName)
		} else {
			etcdTemplateName = etcd.Spec.InfrastructureTemplate.Name
		}
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = controlPlaneTemplateName
		values["vsphereControlPlaneSshAuthorizedKey"] = sshAuthorizedKey(cpVmc)
		values["vsphereEtcdSshAuthorizedKey"] = sshAuthorizedKey(etcdVmc)
		values["etcdTemplateName"] = etcdTemplateName
	}

	workersOpt := func(values map[string]interface{}) {
		values["workloadTemplateName"] = workloadTemplateName
		values["vsphereWorkerSshAuthorizedKey"] = sshAuthorizedKey(workerVmc)
	}

	return generateTemplateResources(templateBuilder, clusterSpec, cpOpt, workersOpt)
}

func generateTemplateResources(builder providers.TemplateBuilder, clusterSpec *cluster.Spec, cpOpt, workersOpt providers.BuildMapOption) ([]*unstructured.Unstructured, error) {
	cp, err := builder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	if err != nil {
		return nil, err
	}
	md, err := builder.GenerateCAPISpecWorkers(clusterSpec, workersOpt)
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
	workersOpt := func(values map[string]interface{}) {
		values["workloadTemplateName"] = mcDeployment.Spec.Template.Spec.InfrastructureRef.Name
	}
	return generateTemplateResources(templateBuilder, clusterSpec, cpOpt, workersOpt)
}

func sshAuthorizedKey(vmc anywherev1.VSphereMachineConfig) string {
	if len(vmc.Spec.Users) <= 0 || len(vmc.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		return ""
	}
	return vmc.Spec.Users[0].SshAuthorizedKeys[0]
}
