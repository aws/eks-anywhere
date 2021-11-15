package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	dockerp "github.com/aws/eks-anywhere/pkg/providers/docker"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type docker struct {
	client eksaClient
}

type dockerClusterSpec struct {
	cluster          *anywhere.Cluster
	spec             *cluster.Spec
	datacenterConfig *anywhere.DockerDatacenterConfig
}

func (d *docker) ReconcileControlPlane(ctx context.Context, cluster *anywhere.Cluster) error {
	spec, err := d.client.BuildClusterSpec(ctx, cluster)
	if err != nil {
		return err
	}

	// Generate CAPI CP yaml
	controlPlaneSpec, err := generateControlPlaneCAPISpecForCreate(spec)
	if err != nil {
		return err
	}

	// Convert yaml CP spec to objects
	objs, err := yamlToUnstructured(controlPlaneSpec)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		obj.SetNamespace(cluster.Namespace)
		// TODO: this is super hacky. We should probably do this in the template
		//  Once we move to structs and generate them individually this should be way easier
		if needsClusterLabelNameControlPlane(obj) {
			labels := obj.GetLabels()
			if labels == nil {
				labels = make(map[string]string, 1)
			}
			labels[ClusterLabelName] = cluster.Name
			obj.SetLabels(labels)
		}
		if err := d.client.Create(ctx, &obj); err != nil {
			if apierrors.IsAlreadyExists(err) {
				// TODO: log this so we know the object already exits bc with the current logic it shouldn't
				//  I just don't have the logger here yet
			}
			return err
		}
	}

	return nil
}

func needsClusterLabelNameControlPlane(obj unstructured.Unstructured) bool {
	return isCAPICluster(obj) || isCAPIControlPlane(obj)
}

func isCAPICluster(obj unstructured.Unstructured) bool {
	return obj.GetKind() == "Cluster" && strings.HasPrefix(obj.GetAPIVersion(), "cluster.x-k8s.io")
}

func isCAPIControlPlane(obj unstructured.Unstructured) bool {
	return obj.GetKind() == "KubeadmControlPlane" && strings.HasPrefix(obj.GetAPIVersion(), "controlplane.cluster.x-k8s.io")
}

// copy paste from the provider
func generateControlPlaneCAPISpecForCreate(spec *cluster.Spec) (controlPlaneSpec []byte, err error) {
	templateBuilder := dockerp.NewDockerTemplateBuilder(time.Now)
	clusterName := spec.Cluster.Name
	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = templateBuilder.CPMachineTemplateName(clusterName)
		values["etcdTemplateName"] = templateBuilder.EtcdMachineTemplateName(clusterName)
	}

	controlPlaneSpec, err = templateBuilder.GenerateCAPISpecControlPlane(spec, cpOpt)
	if err != nil {
		return nil, err
	}

	return controlPlaneSpec, nil
}

func (d *docker) ReconcileWorkers(ctx context.Context, cluster *anywhere.Cluster) error {
	spec, err := d.client.BuildClusterSpec(ctx, cluster)
	if err != nil {
		return err
	}

	// Generate CAPI CP yaml
	controlPlaneSpec, err := generateWorkersCAPISpecForCreate(spec)
	if err != nil {
		return err
	}

	// Convert yaml CP spec to objects
	objs, err := yamlToUnstructured(controlPlaneSpec)
	if err != nil {
		return err
	}

	machineDeployments := 0
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
			// TODO: Super hacky, docker doesn't have machine group configs so we can't use the name here
			// we fake the name by using the order, but this is going to break at some point
			// We can just add the label on the template directly, but we should have a way to map these to the machine groups
			labels[MachineGroupLabelName] = fmt.Sprintf("%s-machine-group-%d", cluster.Name, machineDeployments)
			machineDeployments++
			labels[MachineDeploymentLabelType] = MachineDeploymentWorkersType
			obj.SetLabels(labels)
		}
		if err := d.client.Create(ctx, &obj); err != nil {
			if apierrors.IsAlreadyExists(err) {
				// TODO: log this so we know the object already exits bc with the current logic it shouldn't
				//  I just don't have the logger here yet
			}
			return err
		}
	}

	return nil
}

func generateWorkersCAPISpecForCreate(spec *cluster.Spec) (controlPlaneSpec []byte, err error) {
	templateBuilder := dockerp.NewDockerTemplateBuilder(time.Now)
	clusterName := spec.Cluster.Name

	workersOpts := func(values map[string]interface{}) {
		values["workloadTemplateName"] = templateBuilder.WorkerMachineTemplateName(clusterName)
	}
	workersSpec, err := templateBuilder.GenerateCAPISpecWorkers(spec, workersOpts)
	if err != nil {
		return nil, err
	}

	return workersSpec, nil
}

func needsClusterLabelsWorkers(obj unstructured.Unstructured) bool {
	return isMachineDeployment(obj)
}

func isMachineDeployment(obj unstructured.Unstructured) bool {
	return obj.GetKind() == "MachineDeployment" && strings.HasPrefix(obj.GetAPIVersion(), "cluster.x-k8s.io")
}
