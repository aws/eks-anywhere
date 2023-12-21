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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

	defer func() {
		err := r.updateStatus(ctx, log, mdUpgrade)
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
		return r.reconcileDelete(ctx, log, mdUpgrade)
	}

	// AddFinalizer	is idempotent
	controllerutil.AddFinalizer(mdUpgrade, mdUpgradeFinalizerName)

	return r.reconcile(ctx, log, mdUpgrade)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineDeploymentUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.MachineDeploymentUpgrade{}).
		Complete(r)
}

func (r *MachineDeploymentUpgradeReconciler) reconcile(ctx context.Context, log logr.Logger, mdUpgrade *anywherev1.MachineDeploymentUpgrade) (ctrl.Result, error) {
	log.Info("Upgrading all worker nodes")
	for _, machineRef := range mdUpgrade.Spec.MachinesRequireUpgrade {
		nodeUpgrade, err := getNodeUpgrade(ctx, r.client, nodeUpgraderName(machineRef.Name))
		if err != nil {
			if apierrors.IsNotFound(err) {
				nodeUpgrade = mdNodeUpgrader(machineRef, kubernetesVersion)
				if err := r.client.Create(ctx, nodeUpgrade); err != nil {
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

func (r *MachineDeploymentUpgradeReconciler) reconcileDelete(ctx context.Context, log logr.Logger, mdUpgrade *anywherev1.MachineDeploymentUpgrade) (ctrl.Result, error) {
	log.Info("Reconcile MachineDeploymentUpgrade deletion")

	for _, machineRef := range mdUpgrade.Spec.MachinesRequireUpgrade {
		nodeUpgrade, err := getNodeUpgrade(ctx, r.client, nodeUpgraderName(machineRef.Name))
		if err != nil {
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

	// Remove the finalizer on MachineDeploymentUpgrade object
	controllerutil.RemoveFinalizer(mdUpgrade, mdUpgradeFinalizerName)
	return ctrl.Result{}, nil
}

func patchMachineDeploymentUpgrade(ctx context.Context, patchHelper *patch.Helper, mdUpgrade *anywherev1.MachineDeploymentUpgrade, patchOpts ...patch.Option) error {
	// Always attempt to patch the object and status after each reconciliation.
	return patchHelper.Patch(ctx, mdUpgrade, patchOpts...)
}

func (r *MachineDeploymentUpgradeReconciler) updateStatus(ctx context.Context, log logr.Logger, mdUpgrade *anywherev1.MachineDeploymentUpgrade) error {
	// When MachineDeploymentUpgrade is fully deleted, we do not need to update the status. Without this check
	// the subsequent patch operations would fail if the status is updated after it is fully deleted.
	if !mdUpgrade.DeletionTimestamp.IsZero() && len(mdUpgrade.GetFinalizers()) == 0 {
		log.Info("MachineDeploymentUpgrade is deleted, skipping status update")
		return nil
	}

	log.Info("Updating MachineDeploymentUpgrade status")

	nodesUpgradeCompleted := 0
	nodesUpgradeRequired := len(mdUpgrade.Spec.MachinesRequireUpgrade)

	for _, machine := range mdUpgrade.Spec.MachinesRequireUpgrade {
		nodeUpgrade, err := getNodeUpgrade(ctx, r.client, nodeUpgraderName(machine.Name))
		if err != nil {
			return err
		}
		if nodeUpgrade.Status.Completed {
			nodesUpgradeCompleted++
			nodesUpgradeRequired--

		}
	}
	log.Info("Worker nodes ready", "total", mdUpgrade.Status.Upgraded, "need-upgrade", mdUpgrade.Status.RequireUpgrade)
	mdUpgrade.Status.Upgraded = int64(nodesUpgradeCompleted)
	mdUpgrade.Status.RequireUpgrade = int64(nodesUpgradeRequired)
	mdUpgrade.Status.Ready = nodesUpgradeRequired == 0
	return nil
}

func mdNodeUpgrader(machineRef anywherev1.Ref, kubernetesVersion string) *anywherev1.NodeUpgrade {
	return &anywherev1.NodeUpgrade{
		ObjectMeta: v1.ObjectMeta{
			Name:      nodeUpgraderName(machineRef.Name),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.NodeUpgradeSpec{
			Machine: corev1.ObjectReference{
				Kind:      machineRef.Kind,
				Namespace: constants.EksaSystemNamespace,
				Name:      machineRef.Name,
			},
			KubernetesVersion: kubernetesVersion,
		},
	}
}
