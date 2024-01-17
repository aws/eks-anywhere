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
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	kcpInPlaceUpgradeNeededAnnotation = "controlplane.clusters.x-k8s.io/in-place-upgrade-needed"
)

// KubeadmControlPlaneInPlaceUpgradeReconciler reconciles a KubeadmControlPlaneInPlaceUpgradeReconciler object.
type KubeadmControlPlaneInPlaceUpgradeReconciler struct {
	client client.Client
	log    logr.Logger
}

// NewKubeadmControlPlaneInPlaceUpgradeReconciler returns a new instance of KubeadmControlPlaneInPlaceUpgradeReconciler.
func NewKubeadmControlPlaneInPlaceUpgradeReconciler(client client.Client) *KubeadmControlPlaneInPlaceUpgradeReconciler {
	return &KubeadmControlPlaneInPlaceUpgradeReconciler{
		client: client,
		log:    ctrl.Log.WithName("KubeadmControlPlaneInPlaceUpgradeController"),
	}
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplane,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplane/status,verbs=get
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines,verbs=get,list

// Reconcile reconciles a KubeadmControlPlane object for in place upgrades.
func (r *KubeadmControlPlaneInPlaceUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := r.log.WithValues("KubeadmControlPlane", req.NamespacedName)

	log.Info("Reconciling KubeadmControlPlane object")
	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.client.Get(ctx, req.NamespacedName, kcp); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if !r.inPlaceUpgradeNeeded(kcp) {
		log.Info("In place upgraded needed annotation not detected, nothing to do")
		return ctrl.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(kcp, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch after each reconciliation in case annotation is removed.
		if err := patchHelper.Patch(ctx, kcp); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the "in-place-upgrade-needed" annotation is not set.
		// We do this to be able to update the status continuously until it becomes ready,
		// since there might be changes in state of the world that don't trigger reconciliation requests
		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && r.inPlaceUpgradeNeeded(kcp) {
			result = ctrl.Result{RequeueAfter: 10 * time.Second}
		}
	}()

	return r.reconcile(ctx, log, kcp)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeadmControlPlaneInPlaceUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1.KubeadmControlPlane{}).
		Complete(r)
}

func (r *KubeadmControlPlaneInPlaceUpgradeReconciler) reconcile(ctx context.Context, log logr.Logger, kcp *controlplanev1.KubeadmControlPlane) (ctrl.Result, error) {
	log.Info("Reconciling in place upgrade for control plane")
	if kcp.Spec.Replicas != nil && (*kcp.Spec.Replicas == kcp.Status.UpdatedReplicas) {
		log.Info("KubeadmControlPlane is ready, nothing else to reconcile for in place upgrade")
		// Remove in-place-upgrade-needed annotation
		delete(kcp.Annotations, kcpInPlaceUpgradeNeededAnnotation)
		return ctrl.Result{}, nil
	}
	cpUpgrade := &anywherev1.ControlPlaneUpgrade{}
	if err := r.client.Get(ctx, GetNamespacedNameType(cpUpgradeName(kcp.Name), constants.EksaSystemNamespace), cpUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Creating control plane upgrade object")
			machines, err := r.machinesToUpgrade(ctx, kcp)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("retrieving list of control plane machines: %v", err)
			}
			if err := r.client.Create(ctx, controlPlaneUpgrade(kcp, machines)); client.IgnoreAlreadyExists(err) != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create control plane upgrade for KubeadmControlPlane %s:  %v", kcp.Name, err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting control plane upgrade for KubeadmControlPlane %s: %v", kcp.Name, err)
	}
	if !cpUpgrade.Status.Ready {
		return ctrl.Result{}, nil
	}
	// TODO: update status for templates and other resources
	log.Info("Control plane upgrade complete, deleting object")
	if err := r.client.Delete(ctx, cpUpgrade); err != nil {
		return ctrl.Result{}, fmt.Errorf("deleting control plane upgrade object: %v", err)
	}

	return ctrl.Result{}, nil
}

func (r *KubeadmControlPlaneInPlaceUpgradeReconciler) inPlaceUpgradeNeeded(kcp *controlplanev1.KubeadmControlPlane) bool {
	_, ok := kcp.Annotations[kcpInPlaceUpgradeNeededAnnotation]
	return ok
}

func (r *KubeadmControlPlaneInPlaceUpgradeReconciler) machinesToUpgrade(ctx context.Context, kcp *controlplanev1.KubeadmControlPlane) ([]corev1.ObjectReference, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: map[string]string{controlPlaneLabel: kcp.Name}})
	if err != nil {
		return nil, err
	}
	machineList := &clusterv1.MachineList{}
	if err := r.client.List(ctx, machineList, &client.ListOptions{LabelSelector: selector, Namespace: kcp.Namespace}); err != nil {
		return nil, err
	}
	machines := collections.FromMachineList(machineList).SortedByCreationTimestamp()
	machineObjects := make([]corev1.ObjectReference, 0, len(machines))
	for _, machine := range machines {
		machineObjects = append(machineObjects,
			corev1.ObjectReference{
				Kind:      machine.Kind,
				Namespace: machine.Namespace,
				Name:      machine.Name,
			},
		)
	}
	return machineObjects, nil
}

func controlPlaneUpgrade(kcp *controlplanev1.KubeadmControlPlane, machines []corev1.ObjectReference) *anywherev1.ControlPlaneUpgrade {
	return &anywherev1.ControlPlaneUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpUpgradeName(kcp.Name),
			Namespace: constants.EksaSystemNamespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         kcp.APIVersion,
				Kind:               kcp.Kind,
				Name:               kcp.Name,
				UID:                kcp.UID,
				Controller:         pointer.Bool(true),
				BlockOwnerDeletion: pointer.Bool(true),
			}},
		},
		Spec: anywherev1.ControlPlaneUpgradeSpec{
			ControlPlane: corev1.ObjectReference{
				Kind:      kcp.Kind,
				Namespace: kcp.Namespace,
				Name:      kcp.Name,
			},
			KubernetesVersion:      "v1.28.3-eks-1-28-9",
			EtcdVersion:            "v3.5.9-eks-1-28-9",
			MachinesRequireUpgrade: machines,
		},
	}
}

func cpUpgradeName(kcpName string) string {
	return kcpName + "-cp-upgrade"
}
