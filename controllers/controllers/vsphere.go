package controllers

import (
	"context"
	"time"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	vspherep "github.com/aws/eks-anywhere/pkg/providers/vsphere"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type vsphere struct {
	client eksaClient
}

type vsphereClusterSpec struct {
	cluster                   *anywhere.Cluster
	spec                      *cluster.Spec
	datacenterConfig          *anywhere.VSphereDatacenterConfig
	controlPlaneMachineConfig *anywhere.VSphereMachineConfig
	etcdMachineConfig         *anywhere.VSphereMachineConfig
	workerMachineConfigs      []anywhere.VSphereMachineConfig
}

func (v *vsphereClusterSpec) buildTemplateBuilder() providers.TemplateBuilder {
	var etcdMachineConfigSpec *anywhere.VSphereMachineConfigSpec
	if v.etcdMachineConfig != nil {
		etcdMachineConfigSpec = &v.etcdMachineConfig.Spec
	}

	return vspherep.NewVsphereTemplateBuilder(
		&v.datacenterConfig.Spec,
		&v.controlPlaneMachineConfig.Spec,
		&v.workerMachineConfigs[0].Spec, // TODO: this sucks but that's how the template builder works right now
		etcdMachineConfigSpec,
		time.Now,
	)
}

func (v *vsphere) ReconcileControlPlane(ctx context.Context, cluster *anywhere.Cluster) error {
	spec, err := v.buildVSphereSpec(ctx, cluster)
	if err != nil {
		return err
	}

	// Generate CAPI CP yaml
	controlPlaneSpec, err := generateVSphereControlPlaneCAPISpecForCreate(spec)
	if err != nil {
		return err
	}

	return createControlPlaneObjectsFromYamlManifest(ctx, v.client, cluster, controlPlaneSpec)
}

func (v *vsphere) ReconcileWorkers(ctx context.Context, cluster *anywhere.Cluster) error {
	spec, err := v.buildVSphereSpec(ctx, cluster)
	if err != nil {
		return err
	}

	// Generate CAPI workers yaml
	workersSpec, err := generateVSphereWorkersCAPISpecForCreate(spec)
	if err != nil {
		return err
	}

	// Convert yaml workers spec to objects
	objs, err := yamlToUnstructured(workersSpec)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		obj.SetNamespace(cluster.Namespace)
		// TODO: this is super hacky. We should probably do this in the template
		//  Once we move to structs and generate them individually this should be way easier
		if needsClusterLabelsWorkers(obj) {
			labels := obj.GetLabels()
			if labels == nil {
				labels = make(map[string]string, 3)
			}
			labels[ClusterLabelName] = cluster.Name
			// TODO: hacky, assigns the same label to all machine deployments
			labels[MachineGroupLabelName] = cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
			labels[MachineDeploymentLabelType] = MachineDeploymentWorkersType
			obj.SetLabels(labels)
		}
		if err := v.client.Create(ctx, &obj); err != nil {
			if apierrors.IsAlreadyExists(err) {
				// TODO: log this so we know the object already exits bc with the current logic it shouldn't
				//  I just don't have the logger here yet
			}
			return err
		}
	}

	return nil
}

func (v *vsphere) buildVSphereSpec(ctx context.Context, cluster *anywhere.Cluster) (*vsphereClusterSpec, error) {
	vsphereSpec := &vsphereClusterSpec{cluster: cluster}
	var err error

	vsphereSpec.spec, err = v.client.BuildClusterSpec(ctx, cluster)
	if err != nil {
		return nil, err
	}

	vsphereSpec.datacenterConfig, err = v.client.GetVSphereDatacenter(ctx, cluster)
	if err != nil {
		return nil, err
	}

	vsphereSpec.controlPlaneMachineConfig, err = v.client.GetVSphereControlPlaneMachineConfig(ctx, cluster)
	if err != nil {
		return nil, err
	}

	vsphereSpec.workerMachineConfigs, err = v.client.GetVSphereWorkersMachineConfig(ctx, cluster)
	if err != nil {
		return nil, err
	}

	if cluster.Spec.ExternalEtcdConfiguration != nil {
		vsphereSpec.etcdMachineConfig, err = v.client.GetVSphereEtcdMachineConfig(ctx, cluster)
		if err != nil {
			return nil, err
		}
	}

	return vsphereSpec, nil
}

// almost copy paste from the provider, just get the ssh keys directly from the spec
// TODO: if ssh keys need some kind of transformation, we should probably store them in the status during validation instead of updating whatever the user set
func generateVSphereControlPlaneCAPISpecForCreate(vsphereSpec *vsphereClusterSpec) (controlPlaneSpec []byte, err error) {
	templateBuilder := vsphereSpec.buildTemplateBuilder()
	clusterName := vsphereSpec.spec.Cluster.Name

	etcdSSHKey := ""
	if vsphereSpec.cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdSSHKey = vsphereSpec.etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys[0]
	}

	controlPlaneSSHKey := ""
	if len(vsphereSpec.controlPlaneMachineConfig.Spec.Users) > 0 && len(vsphereSpec.controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys) > 0 {
		controlPlaneSSHKey = vsphereSpec.controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys[0]
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = templateBuilder.CPMachineTemplateName(clusterName)
		values["vsphereControlPlaneSshAuthorizedKey"] = controlPlaneSSHKey
		values["vsphereEtcdSshAuthorizedKey"] = etcdSSHKey
		values["etcdTemplateName"] = templateBuilder.EtcdMachineTemplateName(clusterName)
	}

	// TODO: we need to change a bunch of things in the vsphere template so objects can either be shared between clusters or they have unique names. For example vsphere-csi-controller
	return templateBuilder.GenerateCAPISpecControlPlane(vsphereSpec.spec, cpOpt)
}

// almost copy paste from the provider, just get the ssh key directly from the spec
func generateVSphereWorkersCAPISpecForCreate(vsphereSpec *vsphereClusterSpec) (controlPlaneSpec []byte, err error) {
	templateBuilder := vsphereSpec.buildTemplateBuilder()
	clusterName := vsphereSpec.spec.Cluster.Name

	workerSSHKey := ""
	if len(vsphereSpec.workerMachineConfigs) > 0 {
		firstWorkerMachineConfigSpec := vsphereSpec.workerMachineConfigs[0].Spec
		if len(firstWorkerMachineConfigSpec.Users) > 0 && len(firstWorkerMachineConfigSpec.Users[0].SshAuthorizedKeys) > 0 {
			workerSSHKey = firstWorkerMachineConfigSpec.Users[0].SshAuthorizedKeys[0]
		}
	}

	workersOpt := func(values map[string]interface{}) {
		values["workloadTemplateName"] = templateBuilder.WorkerMachineTemplateName(clusterName)
		values["vsphereWorkerSshAuthorizedKey"] = workerSSHKey
	}
	return templateBuilder.GenerateCAPISpecWorkers(vsphereSpec.spec, workersOpt)
}
