/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/external"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// controlPlaneUpgradeFinalizerName is the finalizer added to NodeUpgrade objects to handle deletion.
const (
	controlPlaneUpgradeFinalizerName      = "controlplaneupgrades.anywhere.eks.amazonaws.com/finalizer"
	kubeadmClusterConfigurationAnnotation = "controlplane.cluster.x-k8s.io/kubeadm-cluster-configuration"
	cloneFromNameAnnotationInfraMachine   = "cluster.x-k8s.io/cloned-from-name"
)

// ControlPlaneUpgradeReconciler reconciles a ControlPlaneUpgradeReconciler object.
type ControlPlaneUpgradeReconciler struct {
	client client.Client
	log    logr.Logger
}

// NewControlPlaneUpgradeReconciler returns a new instance of ControlPlaneUpgradeReconciler.
func NewControlPlaneUpgradeReconciler(client client.Client) *ControlPlaneUpgradeReconciler {
	return &ControlPlaneUpgradeReconciler{
		client: client,
		log:    ctrl.Log.WithName("ControlPlaneUpgradeController"),
	}
}

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=controlplaneupgrades,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=controlplaneupgrades/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=controlplaneupgrades/finalizers,verbs=update
//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=kubeadmconfigs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tinkerbellmachines;vspheremachines,verbs=get;list;update;patch

// Reconcile reconciles a ControlPlaneUpgrade object.
// nolint:gocyclo
func (r *ControlPlaneUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := r.log.WithValues("ControlPlaneUpgrade", req.NamespacedName)

	log.Info("Reconciling ControlPlaneUpgrade object")
	cpUpgrade := &anywherev1.ControlPlaneUpgrade{}
	if err := r.client.Get(ctx, req.NamespacedName, cpUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, err
	}

	patchHelper, err := patch.NewHelper(cpUpgrade, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		err := r.updateStatus(ctx, log, cpUpgrade)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}

		// We want the observedGeneration to indicate, that the status shown is up-to-date given the desired spec of the same generation.
		// However, if there is an error while updating the status, we may get a partial status update, In this case,
		// a partially updated status is not considered up to date, so we should not update the observedGeneration

		// Patch ObservedGeneration only if the reconciliation completed without error
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchHelper.Patch(ctx, cpUpgrade, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the Cluster ready condition is false.
		// We do this to be able to update the status continuously until the cluster becomes ready,
		// since there might be changes in state of the world that don't trigger reconciliation requests

		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && !cpUpgrade.Status.Ready {
			result = ctrl.Result{RequeueAfter: 10 * time.Second}
		}
	}()

	// Reconcile the NodeUpgrade deletion if the DeletionTimestamp is set.
	if !cpUpgrade.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, cpUpgrade)
	}
	controllerutil.AddFinalizer(cpUpgrade, controlPlaneUpgradeFinalizerName)
	return r.reconcile(ctx, log, cpUpgrade)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.ControlPlaneUpgrade{}).
		Complete(r)
}

func (r *ControlPlaneUpgradeReconciler) reconcile(ctx context.Context, log logr.Logger, cpUpgrade *anywherev1.ControlPlaneUpgrade) (ctrl.Result, error) {
	var firstControlPlane bool
	// return early if controlplane upgrade is already complete
	if cpUpgrade.Status.Ready {
		log.Info("All Control Plane nodes are upgraded")
		return ctrl.Result{}, nil
	}

	log.Info("Upgrading all Control Plane nodes")

	for idx, machineRef := range cpUpgrade.Spec.MachinesRequireUpgrade {
		firstControlPlane = idx == 0
		nodeUpgrade := nodeUpgrader(machineRef, cpUpgrade.Spec.KubernetesVersion, cpUpgrade.Spec.EtcdVersion, firstControlPlane)
		if err := r.client.Get(ctx, GetNamespacedNameType(nodeUpgraderName(machineRef.Name), constants.EksaSystemNamespace), nodeUpgrade); err != nil {
			if apierrors.IsNotFound(err) {
				if err := r.client.Create(ctx, nodeUpgrade); client.IgnoreAlreadyExists(err) != nil {
					return ctrl.Result{}, fmt.Errorf("failed to create node upgrader for machine %s:  %v", machineRef.Name, err)
				}
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, fmt.Errorf("getting node upgrader for machine %s: %v", machineRef.Name, err)
		}
		if !nodeUpgrade.Status.Completed {
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func nodeUpgrader(machineRef corev1.ObjectReference, kubernetesVersion, etcdVersion string, firstControlPlane bool) *anywherev1.NodeUpgrade {
	return &anywherev1.NodeUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeUpgraderName(machineRef.Name),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.NodeUpgradeSpec{
			Machine: corev1.ObjectReference{
				Kind:      machineRef.Kind,
				Namespace: constants.EksaSystemNamespace,
				Name:      machineRef.Name,
			},
			KubernetesVersion:     kubernetesVersion,
			EtcdVersion:           &etcdVersion,
			FirstNodeToBeUpgraded: firstControlPlane,
		},
	}
}

func (r *ControlPlaneUpgradeReconciler) reconcileDelete(ctx context.Context, log logr.Logger, cpUpgrade *anywherev1.ControlPlaneUpgrade) (ctrl.Result, error) {
	log.Info("Reconcile ControlPlaneUpgrade deletion")

	for _, machineRef := range cpUpgrade.Spec.MachinesRequireUpgrade {
		nodeUpgrade := &anywherev1.NodeUpgrade{}
		if err := r.client.Get(ctx, GetNamespacedNameType(nodeUpgraderName(machineRef.Name), constants.EksaSystemNamespace), nodeUpgrade); err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("Node Upgrader not found, skipping node upgrade deletion")
			} else {
				return ctrl.Result{}, fmt.Errorf("getting node upgrader for machine %s: %v", machineRef.Name, err)
			}
		} else {
			log.Info("Deleting node upgrader", "Machine", machineRef.Name)
			if err := r.client.Delete(ctx, nodeUpgrade); err != nil {
				return ctrl.Result{}, fmt.Errorf("deleting node upgrader: %v", err)
			}
		}
	}

	// Remove the finalizer on ControlPlaneUpgrade objext
	controllerutil.RemoveFinalizer(cpUpgrade, controlPlaneUpgradeFinalizerName)
	return ctrl.Result{}, nil
}

func (r *ControlPlaneUpgradeReconciler) updateStatus(ctx context.Context, log logr.Logger, cpUpgrade *anywherev1.ControlPlaneUpgrade) error {
	// When ControlPlaneUpgrade is fully deleted, we do not need to update the status. Without this check
	// the subsequent patch operations would fail if the status is updated after it is fully deleted.

	if !cpUpgrade.DeletionTimestamp.IsZero() && len(cpUpgrade.GetFinalizers()) == 0 {
		log.Info("ControlPlaneUpgrade is deleted, skipping status update")
		return nil
	}

	log.Info("Updating ControlPlaneUpgrade status")
	nodeUpgrade := &anywherev1.NodeUpgrade{}
	nodesUpgradeCompleted := 0
	nodesUpgradeRequired := len(cpUpgrade.Spec.MachinesRequireUpgrade)
	for _, machineRef := range cpUpgrade.Spec.MachinesRequireUpgrade {
		if err := r.client.Get(ctx, GetNamespacedNameType(nodeUpgraderName(machineRef.Name), constants.EksaSystemNamespace), nodeUpgrade); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("getting node upgrader for machine %s: %v", machineRef.Name, err)
		}
		if nodeUpgrade.Status.Completed {
			if err := r.updateResources(ctx, log, cpUpgrade, nodeUpgrade); err != nil {
				return err
			}
			nodesUpgradeCompleted++
			nodesUpgradeRequired--
		}
	}
	log.Info("Control Plane Nodes ready", "upgraded", cpUpgrade.Status.Upgraded, "need-upgrade", cpUpgrade.Status.RequireUpgrade)
	cpUpgrade.Status.Upgraded = int64(nodesUpgradeCompleted)
	cpUpgrade.Status.RequireUpgrade = int64(nodesUpgradeRequired)
	cpUpgrade.Status.Ready = nodesUpgradeRequired == 0
	return nil
}

func (r *ControlPlaneUpgradeReconciler) updateResources(ctx context.Context, log logr.Logger, cpUpgrade *anywherev1.ControlPlaneUpgrade, nodeUpgrade *anywherev1.NodeUpgrade) error {
	machine, err := getCapiMachine(ctx, r.client, nodeUpgrade)
	if err != nil {
		return err
	}

	log = log.WithValues("Machine", machine.Name)
	machinePatchHelper, err := patch.NewHelper(machine, r.client)
	if err != nil {
		return err
	}

	log.Info("Updating K8s version and kubeadmClusterConfiguration annotation in machine")
	kcpSpec, err := decodeAndUnmarshalKcpSpecData(cpUpgrade.Spec.ControlPlaneSpecData)
	if err != nil {
		return err
	}

	// Update the machine kubeadmClusterConfiguration annotation
	kcc, err := json.Marshal(kcpSpec.KubeadmConfigSpec.ClusterConfiguration)
	if err != nil {
		return fmt.Errorf("marshaling KCP cluster configuration: %v", err)
	}

	// Update the machine k8s version and update the KubeadmClusterConfiguration annotation
	machine.Spec.Version = &nodeUpgrade.Spec.KubernetesVersion
	annotations.AddAnnotations(machine, map[string]string{kubeadmClusterConfigurationAnnotation: string(kcc)})

	if err := machinePatchHelper.Patch(ctx, machine); err != nil {
		return fmt.Errorf("updating spec for machine %s: %v", machine.Name, err)
	}

	if err := r.updateKubeadmConfig(ctx, log, kcpSpec, machine); err != nil {
		return fmt.Errorf("updating kubeadm config: %v", err)
	}

	if err := r.updateInfraMachine(ctx, log, kcpSpec, machine); err != nil {
		return fmt.Errorf("updating infra machine: %v", err)
	}

	return nil
}

func (r *ControlPlaneUpgradeReconciler) updateKubeadmConfig(ctx context.Context, log logr.Logger, kcpSpec *controlplanev1.KubeadmControlPlaneSpec, machine *clusterv1.Machine) error {
	bootstrapRef := machine.Spec.Bootstrap.ConfigRef
	if bootstrapRef == nil {
		return fmt.Errorf("bootstrap config for machine %s is nil", machine.Name)
	}

	kc := &bootstrapv1.KubeadmConfig{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: bootstrapRef.Name, Namespace: machine.Namespace}, kc); err != nil {
		return fmt.Errorf("retrieving bootstrap config for machine %s: %v", machine.Name, err)
	}

	patchHelper, err := patch.NewHelper(kc, r.client)
	if err != nil {
		return fmt.Errorf("initializing patch helper for KubeadmConfig %s: %v", kc.Name, err)
	}

	kcSpec := kcpSpec.KubeadmConfigSpec.DeepCopy()
	if kc.Spec.InitConfiguration == nil {
		kcSpec.InitConfiguration = nil
	}

	if kc.Spec.JoinConfiguration == nil {
		kcSpec.JoinConfiguration = nil
	}

	kc.Spec = *kcSpec
	log.Info("Patching KubeadmConfig", "KubeadmConfig", kc.Name)
	if err := patchHelper.Patch(ctx, kc); err != nil {
		return fmt.Errorf("patching KubeadmConfig %s for Machine %s: %v", kc.Name, machine.Name, err)
	}

	return nil
}

func (r *ControlPlaneUpgradeReconciler) updateInfraMachine(ctx context.Context, log logr.Logger, kcpSpec *controlplanev1.KubeadmControlPlaneSpec, machine *clusterv1.Machine) error {
	infraMachineObj, err := external.Get(ctx, r.client, &machine.Spec.InfrastructureRef, machine.Namespace)
	if err != nil {
		return fmt.Errorf("retrieving infra machine %s for machine %s: %v", machine.Spec.InfrastructureRef.Name, machine.Name, err)
	}
	patchHelper, err := patch.NewHelper(infraMachineObj, r.client)
	if err != nil {
		return err
	}
	log.Info("Updating infra machine template annotation in infra machine", "InfrastructureRef.Name", kcpSpec.MachineTemplate.InfrastructureRef.Name)
	// Update the cloned-from-name annotation to match the updated infra machine template name in KubeadmControlPlane
	annotations := infraMachineObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[cloneFromNameAnnotationInfraMachine] = kcpSpec.MachineTemplate.InfrastructureRef.Name
	infraMachineObj.SetAnnotations(annotations)

	if err := patchHelper.Patch(ctx, infraMachineObj); err != nil {
		return fmt.Errorf("updating spec for infra machine %s: %v", infraMachineObj.GetName(), err)
	}
	return nil
}

func decodeAndUnmarshalKcpSpecData(kcpSpecData string) (*controlplanev1.KubeadmControlPlaneSpec, error) {
	kcpSpec := &controlplanev1.KubeadmControlPlaneSpec{}
	decodedKcpSpec, err := base64.StdEncoding.DecodeString(kcpSpecData)
	if err != nil {
		return nil, fmt.Errorf("decoding cpUpgrade.Spec.ControlPlaneSpec: %v", err)
	}
	if err := json.Unmarshal(decodedKcpSpec, kcpSpec); err != nil {
		return nil, fmt.Errorf("unmarshaling cpUpgrade.Spec.ControlPlaneSpec: %v", err)
	}
	return kcpSpec, nil
}

func getCapiMachine(ctx context.Context, client client.Client, nodeUpgrade *anywherev1.NodeUpgrade) (*clusterv1.Machine, error) {
	machine := &clusterv1.Machine{}
	if err := client.Get(ctx, GetNamespacedNameType(nodeUpgrade.Spec.Machine.Name, nodeUpgrade.Spec.Machine.Namespace), machine); err != nil {
		return nil, fmt.Errorf("getting machine %s: %v", nodeUpgrade.Spec.Machine.Name, err)
	}
	return machine, nil
}
