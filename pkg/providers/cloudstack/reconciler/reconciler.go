package reconciler

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"time"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
}

type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

const defaultRequeueTime = time.Minute

type Reconciler struct {
	client  client.Client
	cmk     *executables.Cmk
	tracker *remote.ClusterCacheTracker
	log     logr.Logger
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	*serverside.ObjectApplier
}

func New(client client.Client, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry,
	cmk *executables.Cmk, tracker *remote.ClusterCacheTracker, log logr.Logger) *Reconciler {
	return &Reconciler{
		client:  client,
		cmk:     cmk,
		cniReconciler: cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ObjectApplier:        serverside.NewObjectApplier(client),
		tracker: tracker,
		log:     log,

	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "cloudstack")
	log.Info("Reconciling cloudstack provider reconciler")
	dataCenterConfig := &anywherev1.CloudStackDatacenterConfig{}
	dataCenterName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := r.client.Get(ctx, dataCenterName, dataCenterConfig); err != nil {
		return controller.Result{}, err
	}
	dataCenterConfig.SetDefaults()
	if !dataCenterConfig.Status.SpecValid {
		log.Info("Skipping cluster reconciliation because data center config is invalid", "data center", dataCenterConfig.Name)
		return controller.Result{
			Result: &ctrl.Result{
				Requeue:      true,
				RequeueAfter: defaultRequeueTime,
			},
		}, nil
	}

	machineConfigMap := map[string]*anywherev1.CloudStackMachineConfig{}

	for _, ref := range cluster.MachineConfigRefs() {
		machineConfig := &anywherev1.CloudStackMachineConfig{}
		machineConfigName := types.NamespacedName{Namespace: cluster.Namespace, Name: ref.Name}
		if err := r.client.Get(ctx, machineConfigName, machineConfig); err != nil {
			return controller.Result{}, err
		}
		machineConfigMap[ref.Name] = machineConfig
	}

	log.V(4).Info("Fetching cluster spec")
	specWithBundles, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner().Register(
		r.ValidateMachineConfigs,
		r.ReconcileControlPlane,
		r.CheckExternalEtcdReady,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
		r.ReconcileWorkers,
	).Run(ctx, log, specWithBundles)
}

func (r *Reconciler) ValidateMachineConfigs(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateMachineConfigs")
	secrets, err := r.fetchDatacenterSecrets(ctx, clusterSpec.CloudStackDatacenter)
	if err != nil {
		return controller.Result{}, fmt.Errorf("retreiving secrets from cloudstack datacenter config: %v", err)
	}
	execConfig, err := decoder.ParseCloudStackCredsFromSecrets(secrets)
	if err != nil {
		return controller.Result{}, err
	}
	r.cmk.SetExecConfig(execConfig)
	validator := cloudstack.NewValidator(r.cmk)
	cloudstackClusterSpec := cloudstack.NewSpec(clusterSpec, clusterSpec.CloudStackMachineConfigs, clusterSpec.CloudStackDatacenter)

	for _, machineConfig := range clusterSpec.CloudStackMachineConfigs {
		//if !machineConfig.Status.SpecValid {
		//	failureMessage := fmt.Sprintf("CloudStackMachineConfig %s is invalid", machineConfig.Name)
		//	if machineConfig.Status.FailureMessage != nil {
		//		failureMessage += ": " + *machineConfig.Status.FailureMessage
		//	}
		//
		//	log.Error(nil, failureMessage)
		//	clusterSpec.Cluster.Status.FailureMessage = &failureMessage
		//	return controller.Result{}, nil
		//}
		if err = validator.ValidateClusterMachineConfigs(ctx, cloudstackClusterSpec); err != nil {
			failureMessage := fmt.Sprintf("CloudStackMachineConfig %s is invalid", machineConfig.Name)
			log.Error(nil, failureMessage)
			clusterSpec.Cluster.Status.FailureMessage = &failureMessage
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log.Info("[DEBUG] ReconcileControlPlane", "name", clusterSpec.Cluster.Name)

	return r.Apply(ctx, func() ([]kubernetes.Object, error) {
		log.Info("[DEBUG] Inside generateObjects!")
		objects, err := cloudstack.ControlPlaneObjects(ctx, clusterSpec, log, clientutil.NewKubeClient(r.client))
		if err != nil {
			log.Error(err, "[DEBUG] We did not get objects, got err instead")
		}
		//log.Info("[DEBUG] We got objects!", "objects", objects)
		return objects, err
	})
}

func (r *Reconciler) CheckExternalEtcdReady(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckEtcdReady(ctx, r.client, log, clusterSpec.Cluster)
}

func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")

	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")

	return r.Apply(ctx, func() ([]kubernetes.Object, error) {
		return cloudstack.WorkersObjects(ctx, clusterSpec, clientutil.NewKubeClient(r.client))
	})
}

func (r *Reconciler) fetchDatacenterSecrets(ctx context.Context, cloudstackDatacenter *anywherev1.CloudStackDatacenterConfig) ([]apiv1.Secret, error) {
	var secrets []apiv1.Secret
	for _, az := range cloudstackDatacenter.Spec.AvailabilityZones {
		secret := &apiv1.Secret{}
		namespacedName := types.NamespacedName{
			Name:      az.CredentialsRef,
			Namespace: constants.EksaSystemNamespace,
		}
		if err := r.client.Get(ctx, namespacedName, secret); err != nil {
			return nil, err
		}
		secrets = append(secrets, *secret)
	}
	r.log.Info("Retrieved secrets (provider reconciler)", "secrets", secrets)

	return secrets, nil
}
