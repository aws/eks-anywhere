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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	// mdUpgradeFinalizerName is the finalizer added to MachineDeploymentUpgrade objects to handle deletion.
	mdUpgradeFinalizerName = "machinedeploymentupgrades.anywhere.eks.amazonaws.com/finalizer"
)

// MachineDeploymentUpgradeReconciler reconciles a MachineDeploymentUpgrade object.
type MachineDeploymentUpgradeReconciler struct {
	client client.Client
	log    logr.Logger
}

// NewMachineDeploymentUpgradeReconciler returns a new instance of MachineDeploymentUpgradeReconciler.
func NewMachineDeploymentUpgradeReconciler(client client.Client) *MachineDeploymentUpgradeReconciler {
	return &MachineDeploymentUpgradeReconciler{
		client: client,
		log:    ctrl.Log.WithName("MachineDeploymentUpgradeController"),
	}
}

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=machinedeploymentupgrades,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=machinedeploymentupgrades/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=machinedeploymentupgrades/finalizers,verbs=update

// Reconcile reconciles a MachineDeploymentUpgrade object.
// nolint:gocyclo
// TODO: Reduce high cyclomatic complexity: https://github.com/aws/eks-anywhere-internal/issues/2119
func (r *MachineDeploymentUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := r.log.WithValues("MachineDeploymentUpgrade", req.NamespacedName)

	log.Info("Reconciling machine deployment upgrade object")
	mdUpgrade := &anywherev1.MachineDeploymentUpgrade{}
	if err := r.client.Get(ctx, req.NamespacedName, mdUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, err
	}

	patchHelper, err := patch.NewHelper(mdUpgrade, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	md := &clusterv1.MachineDeployment{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: mdUpgrade.Spec.MachineDeployment.Name, Namespace: mdUpgrade.Spec.MachineDeployment.Namespace}, md); err != nil {
		return ctrl.Result{}, fmt.Errorf("getting MachineDeployment %s: %v", mdUpgrade.Spec.MachineDeployment.Name, err)
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: map[string]string{"cluster.x-k8s.io/deployment-name": mdUpgrade.Spec.MachineDeployment.Name}})
	if err != nil {
		return ctrl.Result{}, err
	}

	machineList := &clusterv1.MachineList{}
	if err := r.client.List(ctx, machineList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	machines := collections.FromMachineList(machineList).SortedByCreationTimestamp()

	defer func() {
		err := r.updateStatus(ctx, log, mdUpgrade, machines)
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
		if err := patchMachineDeploymentUpgrade(ctx, patchHelper, mdUpgrade, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the Ready condition is false.
		// We do this to be able to update the status continuously until it becomes ready,
		// since there might be changes in state of the world that don't trigger reconciliation requests

		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && !mdUpgrade.Status.Ready {
			result = ctrl.Result{RequeueAfter: 10 * time.Second}
		}
	}()

	// Reconcile the MachineDeploymentUpgrade deletion if the DeletionTimestamp is set.
	if !mdUpgrade.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, mdUpgrade, machines)
	}

	// AddFinalizer	is idempotent
	controllerutil.AddFinalizer(mdUpgrade, mdUpgradeFinalizerName)

	return r.reconcile(ctx, log, mdUpgrade, md, machines)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineDeploymentUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.MachineDeploymentUpgrade{}).
		Complete(r)
}

func (r *MachineDeploymentUpgradeReconciler) reconcile(ctx context.Context, log logr.Logger, mdUpgrade *anywherev1.MachineDeploymentUpgrade, md *clusterv1.MachineDeployment, machines []*clusterv1.Machine) (ctrl.Result, error) {
	log.Info("Upgrading all worker nodes")
	for _, machine := range machines {
		nodeUpgrade, err := getNodeUpgrade(ctx, r.client, nodeUpgraderName(machine.Name))
		if err != nil {
			if apierrors.IsNotFound(err) {
				if md.Spec.Template.Spec.Version == nil {
					return ctrl.Result{}, fmt.Errorf("failed to get kubernetes version for machine deployment %s", md.Name)
				}
				nodeUpgrade = mdNodeUpgrader(machine, *md.Spec.Template.Spec.Version)
				if err := r.client.Create(ctx, nodeUpgrade); err != nil {
					return ctrl.Result{}, fmt.Errorf("failed to create node upgrader for machine %s:  %v", machine.Name, err)
				}
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, fmt.Errorf("getting node upgrader for machine %s: %v", machine.Name, err)
		}
		if !nodeUpgrade.Status.Completed {
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *MachineDeploymentUpgradeReconciler) reconcileDelete(ctx context.Context, log logr.Logger, mdUpgrade *anywherev1.MachineDeploymentUpgrade, machines []*clusterv1.Machine) (ctrl.Result, error) {
	log.Info("Reconcile MachineDeploymentUpgrade deletion")

	for _, machine := range machines {
		nodeUpgrade, err := getNodeUpgrade(ctx, r.client, nodeUpgraderName(machine.Name))
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("Node Upgrader not found, skipping node upgrade deletion")
			} else {
				return ctrl.Result{}, fmt.Errorf("getting node upgrader for machine %s: %v", machine.Name, err)
			}
		} else {
			log.Info("Deleting node upgrader", "Machine", machine.Name)
			if err := r.client.Delete(ctx, nodeUpgrade); err != nil {
				return ctrl.Result{}, fmt.Errorf("deleting node upgrader: %v", err)
			}
		}
	}

	// Remove the finalizer on MachineDeploymentUpgrade object
	controllerutil.RemoveFinalizer(mdUpgrade, mdUpgradeFinalizerName)
	return ctrl.Result{}, nil
}

func patchMachineDeploymentUpgrade(ctx context.Context, patchHelper *patch.Helper, mdUpgrade *anywherev1.MachineDeploymentUpgrade, patchOpts ...patch.Option) error {
	// Always attempt to patch the object and status after each reconciliation.
	return patchHelper.Patch(ctx, mdUpgrade, patchOpts...)
}

func (r *MachineDeploymentUpgradeReconciler) updateStatus(ctx context.Context, log logr.Logger, mdUpgrade *anywherev1.MachineDeploymentUpgrade, machines []*clusterv1.Machine) error {
	// When MachineDeploymentUpgrade is fully deleted, we do not need to update the status. Without this check
	// the subsequent patch operations would fail if the status is updated after it is fully deleted.
	if !mdUpgrade.DeletionTimestamp.IsZero() && len(mdUpgrade.GetFinalizers()) == 0 {
		log.Info("MachineDeploymentUpgrade is deleted, skipping status update")
		return nil
	}

	log.Info("Updating MachineDeploymentUpgrade status")

	nodesUpgradeCompleted := 0
	nodesUpgradeRequired := len(machines)
	machineStates := []v1alpha1.MachineState{}

	for _, machine := range machines {
		nodeUpgrade, err := getNodeUpgrade(ctx, r.client, nodeUpgraderName(machine.Name))
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("Node upgrader not found for the machine yet", "Machine", machine.Name)
				continue
			} else {
				return err
			}
		}
		machineStates = append(machineStates, anywherev1.MachineState{Name: machine.Name, Upgraded: nodeUpgrade.Status.Completed})
		if nodeUpgrade.Status.Completed {
			nodesUpgradeCompleted++
			nodesUpgradeRequired--
		}
	}
	log.Info("Worker nodes ready", "total", mdUpgrade.Status.Upgraded, "need-upgrade", mdUpgrade.Status.RequireUpgrade)
	mdUpgrade.Status.Upgraded = int64(nodesUpgradeCompleted)
	mdUpgrade.Status.RequireUpgrade = int64(nodesUpgradeRequired)
	mdUpgrade.Status.Ready = nodesUpgradeRequired == 0
	mdUpgrade.Status.MachineState = machineStates
	return nil
}

func mdNodeUpgrader(machine *clusterv1.Machine, kubernetesVersion string) *anywherev1.NodeUpgrade {
	return &anywherev1.NodeUpgrade{
		ObjectMeta: v1.ObjectMeta{
			Name:      nodeUpgraderName(machine.Name),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.NodeUpgradeSpec{
			Machine: corev1.ObjectReference{
				Kind:      machine.Kind,
				Namespace: constants.EksaSystemNamespace,
				Name:      machine.Name,
			},
			KubernetesVersion: kubernetesVersion,
		},
	}
}
