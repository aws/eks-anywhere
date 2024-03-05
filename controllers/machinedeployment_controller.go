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
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	mdInPlaceUpgradeNeededAnnotation = "machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed"
	workerMachineLabel               = "cluster.x-k8s.io/deployment-name"
	machineDeploymentKind            = "MachineDeployment"
)

// MachineDeploymentReconciler reconciles a MachineDeploymentReconciler object.
type MachineDeploymentReconciler struct {
	// client reads from a cache and is not a fully direct client.
	client client.Client
	// uncachedClient reads directly from the API server and is slightly slower.
	uncachedClient client.Reader
	log            logr.Logger
}

// NewMachineDeploymentReconciler returns a new instance of MachineDeploymentReconciler.
func NewMachineDeploymentReconciler(client client.Client, uncachedClient client.Reader) *MachineDeploymentReconciler {
	return &MachineDeploymentReconciler{
		client:         client,
		uncachedClient: uncachedClient,
		log:            ctrl.Log.WithName("MachineDeploymentController"),
	}
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployment,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployment/status,verbs=get

// Reconcile reconciles a MachineDeployment object for in place upgrades.
func (r *MachineDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := r.log.WithValues("MachineDeployment", req.NamespacedName)

	md := &clusterv1.MachineDeployment{}
	if err := r.uncachedClient.Get(ctx, req.NamespacedName, md); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if !r.inPlaceUpgradeNeeded(md) {
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

	return r.reconcile(ctx, log, md)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.MachineDeployment{}).
		Complete(r)
}

func (r *MachineDeploymentReconciler) reconcile(ctx context.Context, log logr.Logger, md *clusterv1.MachineDeployment) (ctrl.Result, error) {
	log.Info("Reconciling in place upgrade for workers")
	if md.Spec.Template.Spec.Version == nil {
		log.Info("Kubernetes version not present, unable to reconcile for in place upgrade")
		return ctrl.Result{}, fmt.Errorf("unable to retrieve kubernetes version from MachineDeployment \"%s\"", md.ObjectMeta.Name)
	}

	mhc := &clusterv1.MachineHealthCheck{}
	if err := r.client.Get(ctx, GetNamespacedNameType(mdMachineHealthCheckName(md.ObjectMeta.Name), constants.EksaSystemNamespace), mhc); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, fmt.Errorf("getting MachineHealthCheck %s: %v", mdMachineHealthCheckName(md.ObjectMeta.Name), err)
	}
	mhcPatchHelper, err := patch.NewHelper(mhc, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	machineList, machineRefList, err := r.machinesToUpgrade(ctx, md)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("retrieving list of control plane machines: %v", err)
	}

	mdUpgrade := &anywherev1.MachineDeploymentUpgrade{}
	mduGetErr := r.client.Get(ctx, GetNamespacedNameType(mdUpgradeName(md.ObjectMeta.Name), constants.EksaSystemNamespace), mdUpgrade)
	if mduGetErr == nil {
		machinesUpgraded := true
		for i := range machineList {
			m := machineList[i]
			if m.Spec.Version == nil || *m.Spec.Version != mdUpgrade.Spec.KubernetesVersion {
				machinesUpgraded = false
				break
			}
		}
		if machinesUpgraded && mdUpgrade.Status.Ready {
			log.Info("Machine deployment upgrade complete, deleting object", "MachineDeploymentUpgrade", mdUpgrade.Name)
			if err := r.client.Delete(ctx, mdUpgrade); err != nil {
				return ctrl.Result{}, fmt.Errorf("deleting MachineDeploymentUpgrade object: %v", err)
			}
			log.Info("Resuming machine deployment machine health check", "MachineHealthCheck", mdMachineHealthCheckName(md.ObjectMeta.Name))
			if err := resumeMachineHealthCheck(ctx, mhc, mhcPatchHelper); err != nil {
				return ctrl.Result{}, fmt.Errorf("updating annotations for machine health check: %v", err)
			}
			log.Info("MachineDeployment is ready, removing the \"in-place-upgrade-needed\" annotation")
			// Remove the in-place-upgrade-needed annotation only after the MachineDeploymentUpgrade object is deleted
			delete(md.Annotations, mdInPlaceUpgradeNeededAnnotation)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	if apierrors.IsNotFound(mduGetErr) {
		log.Info("Creating MachineDeploymentUpgrade object")
		mdUpgrade, err := machineDeploymentUpgrade(md, machineRefList)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("generating MachineDeploymentUpgrade: %v", err)
		}

		log.Info("Pausing machine deployment machine health check", "MachineHealthCheck", mdMachineHealthCheckName(md.ObjectMeta.Name))
		if err := pauseMachineHealthCheck(ctx, mhc, mhcPatchHelper); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating annotations for machine health check: %v", err)
		}

		if err := r.client.Create(ctx, mdUpgrade); client.IgnoreAlreadyExists(err) != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create MachineDeploymentUpgrade for MachineDeployment %s:  %v", md.ObjectMeta.Name, err)
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, fmt.Errorf("getting MachineDeploymentUpgrade for MachineDeployment %s: %v", md.ObjectMeta.Name, err)
}

func (r *MachineDeploymentReconciler) inPlaceUpgradeNeeded(md *clusterv1.MachineDeployment) bool {
	return strings.ToLower(md.Annotations[mdInPlaceUpgradeNeededAnnotation]) == "true"
}

func (r *MachineDeploymentReconciler) machinesToUpgrade(ctx context.Context, md *clusterv1.MachineDeployment) ([]*clusterv1.Machine, []corev1.ObjectReference, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: map[string]string{workerMachineLabel: md.ObjectMeta.Name}})
	if err != nil {
		return nil, nil, err
	}
	machineList := &clusterv1.MachineList{}
	if err := r.client.List(ctx, machineList, &client.ListOptions{LabelSelector: selector, Namespace: md.ObjectMeta.Namespace}); err != nil {
		return nil, nil, err
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
	return machines, machineObjects, nil
}

func machineDeploymentUpgrade(md *clusterv1.MachineDeployment, machines []corev1.ObjectReference) (*anywherev1.MachineDeploymentUpgrade, error) {
	msSpec, err := json.Marshal(md.Spec.Template.Spec)
	if err != nil {
		return nil, fmt.Errorf("marshaling Machine spec: %v", err)
	}
	return &anywherev1.MachineDeploymentUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mdUpgradeName(md.ObjectMeta.Name),
			Namespace: constants.EksaSystemNamespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterv1.GroupVersion.String(),
				Kind:       machineDeploymentKind,
				Name:       md.ObjectMeta.Name,
				UID:        md.ObjectMeta.UID,
			}},
		},
		Spec: anywherev1.MachineDeploymentUpgradeSpec{
			MachineDeployment: corev1.ObjectReference{
				Kind:      machineDeploymentKind,
				Namespace: md.ObjectMeta.Namespace,
				Name:      md.ObjectMeta.Name,
			},
			KubernetesVersion:      *md.Spec.Template.Spec.Version,
			MachinesRequireUpgrade: machines,
			MachineSpecData:        base64.StdEncoding.EncodeToString(msSpec),
		},
	}, nil
}

func mdUpgradeName(mdName string) string {
	return mdName + "-md-upgrade"
}

func mdMachineHealthCheckName(mdName string) string {
	return fmt.Sprintf("%s-worker-unhealthy", mdName)
}
