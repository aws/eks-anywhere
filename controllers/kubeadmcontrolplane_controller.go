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
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
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
	controlPlaneMachineLabel          = "cluster.x-k8s.io/control-plane-name"
)

// KubeadmControlPlaneReconciler reconciles a KubeadmControlPlaneReconciler object.
type KubeadmControlPlaneReconciler struct {
	client client.Client
	log    logr.Logger
}

// NewKubeadmControlPlaneReconciler returns a new instance of KubeadmControlPlaneReconciler.
func NewKubeadmControlPlaneReconciler(client client.Client) *KubeadmControlPlaneReconciler {
	return &KubeadmControlPlaneReconciler{
		client: client,
		log:    ctrl.Log.WithName("KubeadmControlPlaneController"),
	}
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplane,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplane/status,verbs=get

// Reconcile reconciles a KubeadmControlPlane object for in place upgrades.
func (r *KubeadmControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := r.log.WithValues("KubeadmControlPlane", req.NamespacedName)

	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.client.Get(ctx, req.NamespacedName, kcp); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if !r.inPlaceUpgradeNeeded(kcp) {
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
func (r *KubeadmControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1.KubeadmControlPlane{}).
		Complete(r)
}

func (r *KubeadmControlPlaneReconciler) reconcile(ctx context.Context, log logr.Logger, kcp *controlplanev1.KubeadmControlPlane) (ctrl.Result, error) {
	log.Info("Reconciling in place upgrade for control plane")
	if err := r.validateStackedEtcd(kcp); err != nil {
		log.Info("Stacked etcd validation failed, unable to reconcile for in place upgrade")
		return ctrl.Result{}, err
	}

	cpUpgrade := &anywherev1.ControlPlaneUpgrade{}
	cpuGetErr := r.client.Get(ctx, GetNamespacedNameType(cpUpgradeName(kcp.Name), constants.EksaSystemNamespace), cpUpgrade)

	mhc := &clusterv1.MachineHealthCheck{}
	if err := r.client.Get(ctx, GetNamespacedNameType(cpMachineHealthCheckName(kcp.Name), constants.EksaSystemNamespace), mhc); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, fmt.Errorf("getting MachineHealthCheck %s: %v", cpMachineHealthCheckName(kcp.Name), err)
	}
	mhcPatchHelper, err := patch.NewHelper(mhc, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	if kcp.Spec.Replicas != nil && (*kcp.Spec.Replicas == kcp.Status.UpdatedReplicas) {
		if cpuGetErr == nil && cpUpgrade.Status.Ready {
			log.Info("Control plane upgrade complete, deleting object", "ControlPlaneUpgrade", cpUpgrade.Name)
			if err := r.client.Delete(ctx, cpUpgrade); err != nil {
				return ctrl.Result{}, fmt.Errorf("deleting ControlPlaneUpgrade object: %v", err)
			}
			log.Info("Resuming control plane machine health check", "MachineHealthCheck", cpMachineHealthCheckName(kcp.Name))
			if err := resumeMachineHealthCheck(ctx, mhc, mhcPatchHelper); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating annotations for machine health check: %v", err)
			}
		} else if !apierrors.IsNotFound(cpuGetErr) {
			return ctrl.Result{}, fmt.Errorf("getting ControlPlaneUpgrade for KubeadmControlPlane %s: %v", kcp.Name, cpuGetErr)
		}

		log.Info("KubeadmControlPlane is ready, removing the \"in-place-upgrade-needed\" annotation")
		// Remove the in-place-upgrade-needed annotation only after the ControlPlaneUpgrade object is deleted
		delete(kcp.Annotations, kcpInPlaceUpgradeNeededAnnotation)
		return ctrl.Result{}, nil
	}

	if cpuGetErr != nil {
		if apierrors.IsNotFound(cpuGetErr) {
			log.Info("Creating ControlPlaneUpgrade object")
			machines, err := r.machinesToUpgrade(ctx, kcp)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("retrieving list of control plane machines: %v", err)
			}
			cpUpgrade, err := controlPlaneUpgrade(kcp, machines)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("generating ControlPlaneUpgrade: %v", err)
			}

			log.Info("Pausing control plane machine health check", "MachineHealthCheck", cpMachineHealthCheckName(kcp.Name))
			if err := pauseMachineHealthCheck(ctx, mhc, mhcPatchHelper); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating annotations for machine health check: %v", err)
			}

			if err := r.client.Create(ctx, cpUpgrade); client.IgnoreAlreadyExists(err) != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create ControlPlaneUpgrade for KubeadmControlPlane %s:  %v", kcp.Name, err)
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("getting ControlPlaneUpgrade for KubeadmControlPlane %s: %v", kcp.Name, cpuGetErr)
	}
	if !cpUpgrade.Status.Ready {
		return ctrl.Result{}, nil
	}

	log.Info("Control plane upgrade complete, deleting object", "ControlPlaneUpgrade", cpUpgrade.Name)
	if err := r.client.Delete(ctx, cpUpgrade); err != nil {
		return ctrl.Result{}, fmt.Errorf("deleting ControlPlaneUpgrade object: %v", err)
	}

	log.Info("Resuming control plane machine health check", "MachineHealthCheck", cpMachineHealthCheckName(kcp.Name))
	if err := resumeMachineHealthCheck(ctx, mhc, mhcPatchHelper); err != nil {
		return ctrl.Result{}, fmt.Errorf("updating annotations for machine health check: %v", err)
	}

	return ctrl.Result{}, nil
}

func (r *KubeadmControlPlaneReconciler) inPlaceUpgradeNeeded(kcp *controlplanev1.KubeadmControlPlane) bool {
	return strings.ToLower(kcp.Annotations[kcpInPlaceUpgradeNeededAnnotation]) == "true"
}

func (r *KubeadmControlPlaneReconciler) machinesToUpgrade(ctx context.Context, kcp *controlplanev1.KubeadmControlPlane) ([]corev1.ObjectReference, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: map[string]string{controlPlaneMachineLabel: kcp.Name}})
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

func (r *KubeadmControlPlaneReconciler) validateStackedEtcd(kcp *controlplanev1.KubeadmControlPlane) error {
	if kcp.Spec.KubeadmConfigSpec.ClusterConfiguration == nil {
		return fmt.Errorf("ClusterConfiguration not set for KubeadmControlPlane \"%s\", unable to retrieve etcd information", kcp.Name)
	}
	if kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local == nil {
		return fmt.Errorf("local etcd configuration is missing")
	}
	return nil
}

func pauseMachineHealthCheck(ctx context.Context, mhc *clusterv1.MachineHealthCheck, mhcPatchHelper *patch.Helper) error {
	annotations.AddAnnotations(mhc, map[string]string{clusterv1.PausedAnnotation: "true"})
	return mhcPatchHelper.Patch(ctx, mhc)
}

func resumeMachineHealthCheck(ctx context.Context, mhc *clusterv1.MachineHealthCheck, mhcPatchHelper *patch.Helper) error {
	delete(mhc.Annotations, clusterv1.PausedAnnotation)
	return mhcPatchHelper.Patch(ctx, mhc)
}

func controlPlaneUpgrade(kcp *controlplanev1.KubeadmControlPlane, machines []corev1.ObjectReference) (*anywherev1.ControlPlaneUpgrade, error) {
	kcpSpec, err := json.Marshal(kcp.Spec)
	if err != nil {
		return nil, fmt.Errorf("marshaling KCP spec: %v", err)
	}

	return &anywherev1.ControlPlaneUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpUpgradeName(kcp.Name),
			Namespace: constants.EksaSystemNamespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: kcp.APIVersion,
				Kind:       kcp.Kind,
				Name:       kcp.Name,
				UID:        kcp.UID,
			}},
		},
		Spec: anywherev1.ControlPlaneUpgradeSpec{
			ControlPlane: corev1.ObjectReference{
				Kind:      kcp.Kind,
				Namespace: kcp.Namespace,
				Name:      kcp.Name,
			},
			KubernetesVersion:      kcp.Spec.Version,
			EtcdVersion:            kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ImageTag,
			MachinesRequireUpgrade: machines,
			ControlPlaneSpecData:   base64.StdEncoding.EncodeToString(kcpSpec),
		},
	}, nil
}

func cpUpgradeName(kcpName string) string {
	return kcpName + "-cp-upgrade"
}

func cpMachineHealthCheckName(kcpName string) string {
	return fmt.Sprintf("%s-kcp-unhealthy", kcpName)
}
