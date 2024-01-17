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
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// MachineDeploymentInPlaceUpgradeReconciler reconciles a MachineDeploymentInPlaceUpgradeReconciler object.
type MachineDeploymentInPlaceUpgradeReconciler struct {
	client client.Client
	log    logr.Logger
}

// NewMachineDeploymentInPlaceUpgradeReconciler returns a new instance of MachineDeploymentInPlaceUpgradeReconciler.
func NewMachineDeploymentInPlaceUpgradeReconciler(client client.Client) *MachineDeploymentInPlaceUpgradeReconciler {
	return &MachineDeploymentInPlaceUpgradeReconciler{
		client: client,
		log:    ctrl.Log.WithName("MachineDeploymentInPlaceUpgradeController"),
	}
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployment,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployment/status,verbs=get
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines,verbs=get,list

// Reconcile reconciles a MachineDeployment object for in place upgrades.
func (r *MachineDeploymentInPlaceUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := r.log.WithValues("MachineDeployment", req.NamespacedName)

	log.Info("Reconciling MachineDeployment object")
	md := &clusterv1.MachineDeployment{}
	if err := r.client.Get(ctx, req.NamespacedName, md); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if !r.inPlaceUpgradeNeeded(md) {
		log.Info("In place upgraded needed annotation not detected, nothing to do")
		return ctrl.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(md, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch after each reconciliation in case annotation is removed.
		if err := patchHelper.Patch(ctx, md); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the "in-place-upgrade-needed" annotation is not set.
		// We do this to be able to update the status continuously until it becomes ready,
		// since there might be changes in state of the world that don't trigger reconciliation requests
		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && r.inPlaceUpgradeNeeded(md) {
			result = ctrl.Result{RequeueAfter: 10 * time.Second}
		}
	}()

	// Reconcile the MachineDeployment deletion if the DeletionTimestamp is set.
	if !md.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, md)
	}

	// AddFinalizer	is idempotent
	controllerutil.AddFinalizer(md, mdUpgradeFinalizerName)

	return r.reconcile(ctx, log, md)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineDeploymentInPlaceUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.MachineDeployment{}).
		Complete(r)
}

func (r *MachineDeploymentInPlaceUpgradeReconciler) reconcile(ctx context.Context, log logr.Logger, md *clusterv1.MachineDeployment) (ctrl.Result, error) {
	log.Info("Reconciling in place upgrade for workers")
	/*if md.Status.UpdatedReplicas == md.Spec.Replicas {
		log.Info("MachineDeployment is ready, nothing else to reconcile for in place upgrade")
		// Remove annotation
	}*/
	mdUpgrade := mdUpgrade(md)
	if err := r.client.Get(ctx, GetNamespacedNameType(md.Name, constants.EksaSystemNamespace), mdUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Creating machine deployment upgrade object")
			if err := r.client.Create(ctx, mdUpgrade); client.IgnoreAlreadyExists(err) != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create machine deployment upgrade for MachineDeployment %s:  %v", md.Name, err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting machine deployment upgrade for MachineDeployment %s: %v", md.Name, err)
	}
	if !mdUpgrade.Status.Ready {
		return ctrl.Result{}, nil
	}
	// TODO: update status for templates and other resources
	log.Info("Machine deployment upgrade complete, deleting object")
	if err := r.client.Delete(ctx, mdUpgrade); err != nil {
		return ctrl.Result{}, fmt.Errorf("deleting machine deployment upgrade object: %v", err)
	}

	return ctrl.Result{}, nil
}

func (r *MachineDeploymentInPlaceUpgradeReconciler) reconcileDelete(ctx context.Context, log logr.Logger, md *clusterv1.MachineDeployment) (ctrl.Result, error) {
	log.Info("Reconciling MachineDeployment deletion of in place upgrade resources")
	mdUpgrade := mdUpgrade(md)
	if err := r.client.Get(ctx, GetNamespacedNameType(md.Name, constants.EksaSystemNamespace), mdUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Machine deployment in place upgrade object not found, skipping MachineDeployment deletion")
		} else {
			return ctrl.Result{}, fmt.Errorf("getting machine deployment upgrade for cluster %s: %v", md.Name, err)
		}
	} else {
		log.Info("Deleting machine deployment upgrade", "MachineDeploymentUpgrade", mdUpgradeName(md.Name))
		if err := r.client.Delete(ctx, mdUpgrade); err != nil {
			return ctrl.Result{}, fmt.Errorf("deleting machine deployment upgrade object: %v", err)
		}
	}

	// Remove the finalizer on MachineDeployment object
	controllerutil.RemoveFinalizer(md, mdUpgradeFinalizerName)
	return ctrl.Result{}, nil
}

func (r *MachineDeploymentInPlaceUpgradeReconciler) inPlaceUpgradeNeeded(md *clusterv1.MachineDeployment) bool {
	return strings.ToLower(md.Annotations[constants.InPlaceUpgradeNeededAnnotation]) == "true"
}

func mdUpgrade(md *clusterv1.MachineDeployment) *anywherev1.MachineDeploymentUpgrade {
	return &anywherev1.MachineDeploymentUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mdUpgradeName(md.Name),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.MachineDeploymentUpgradeSpec{
			MachineDeployment: corev1.ObjectReference{
				Kind:      md.Kind,
				Namespace: constants.EksaSystemNamespace,
				Name:      md.Name,
			},
		},
	}
}

func mdUpgradeName(mdName string) string {
	return mdName + "-md-upgrade"
}
