package cni

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	eksacluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
)

const defaultRequeueTime = time.Second * 10

type CiliumReconciler struct {
	Client  client.Client
	Log     logr.Logger
	tracker *remote.ClusterCacheTracker
}

func (cr *CiliumReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *eksacluster.Spec) (controller.Result, error) {
	cr.Log.Info("Getting remote client", "client for cluster", capiCluster.Name)
	key := client.ObjectKey{
		Namespace: capiCluster.Namespace,
		Name:      capiCluster.Name,
	}
	remoteClient, err := cr.tracker.GetClient(ctx, key)
	if err != nil {
		return controller.Result{}, err
	}

	cr.Log.Info("Applying CNI")
	ciliumDS := &v1.DaemonSet{}
	ciliumDSName := types.NamespacedName{Namespace: "kube-system", Name: cilium.DaemonSetName}
	err = remoteClient.Get(ctx, ciliumDSName, ciliumDS)
	if err != nil {
		if apierrors.IsNotFound(err) {
			cr.Log.Info("Deploying Cilium DS")
			helm := executables.NewHelm(executables.NewExecutable("helm"), executables.WithInsecure())

			ci := cilium.NewCilium(nil, helm)

			ciliumSpec, err := ci.GenerateManifest(ctx, specWithBundles, []string{constants.CapvSystemNamespace})
			if err != nil {
				return controller.Result{}, err
			}
			if err := serverside.ReconcileYaml(ctx, remoteClient, ciliumSpec); err != nil {
				return controller.Result{}, err
			}
			return controller.Result{}, err
		}

		return controller.Result{}, err
	}

	// upgrade cilium
	cr.Log.Info("Upgrading Cilium")
	needsUpgrade, err := ciliumNeedsUpgrade(ctx, cr.Log, remoteClient, specWithBundles)
	if err != nil {
		return controller.Result{}, err
	}

	if !needsUpgrade {
		cr.Log.Info("Skipping Cilium")
		return controller.Result{}, nil
	}

	helm := executables.NewHelm(executables.NewExecutable("helm"))
	templater := cilium.NewTemplater(helm)
	preflight, err := templater.GenerateUpgradePreflightManifest(ctx, specWithBundles)
	if err != nil {
		return controller.Result{}, err
	}

	cr.Log.Info("Installing Cilium upgrade preflight manifest")
	if err := serverside.ReconcileYaml(ctx, remoteClient, preflight); err != nil {
		return controller.Result{}, err
	}

	ciliumDS = &v1.DaemonSet{}
	ciliumDSName = types.NamespacedName{Namespace: "kube-system", Name: "cilium"}
	if err := remoteClient.Get(ctx, ciliumDSName, ciliumDS); err != nil {
		cr.Log.Info("Cilium DS not found")
		return controller.Result{}, err
	}

	if err := cilium.CheckDaemonSetReady(ciliumDS); err != nil {
		cr.Log.Info("Cilium DS not ready")
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, nil
	}

	preFlightCiliumDS := &v1.DaemonSet{}
	preFlightCiliumDSName := types.NamespacedName{Namespace: "kube-system", Name: cilium.PreflightDaemonSetName}
	if err := remoteClient.Get(ctx, preFlightCiliumDSName, preFlightCiliumDS); err != nil {
		cr.Log.Info("Preflight Cilium DS not found.")
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, err
	}

	if err := cilium.CheckPreflightDaemonSetReady(ciliumDS, preFlightCiliumDS); err != nil {
		cr.Log.Info("Preflight DS not ready ")
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, err
	}

	cr.Log.Info("Deleting Preflight Cilium objects")
	if err := serverside.DeleteYaml(ctx, remoteClient, preflight); err != nil {
		cr.Log.Info("Error deleting Preflight Cilium objects")
		return controller.Result{}, err
	}

	cr.Log.Info("Generating Cilium upgrade manifest")
	upgradeManifest, err := templater.GenerateUpgradeManifest(ctx, specWithBundles, specWithBundles)
	if err != nil {
		return controller.Result{}, err
	}

	if err := serverside.ReconcileYaml(ctx, remoteClient, upgradeManifest); err != nil {
		return controller.Result{}, err
	}
	return controller.Result{}, nil
}

func ciliumNeedsUpgrade(ctx context.Context, log logr.Logger, client client.Client, clusterSpec *eksacluster.Spec) (bool, error) {
	log.Info("Checking if Cilium DS needs upgrade")
	needsUpgrade, err := ciliumDSNeedsUpgrade(ctx, log, client, clusterSpec)
	if err != nil {
		return false, err
	}

	if needsUpgrade {
		log.Info("Cilium DS needs upgrade")
		return true, nil
	}

	log.Info("Checking if Cilium operator deployment needs upgrade")
	needsUpgrade, err = ciliumOperatorNeedsUpgrade(ctx, log, client, clusterSpec)
	if err != nil {
		return false, err
	}

	if needsUpgrade {
		log.Info("Cilium operator deployment needs upgrade")
		return true, nil
	}

	return false, nil
}

func ciliumDSNeedsUpgrade(ctx context.Context, log logr.Logger, client client.Client, clusterSpec *eksacluster.Spec) (bool, error) {
	ds, err := getCiliumDS(ctx, client)
	if err != nil {
		return false, err
	}

	if ds == nil {
		log.Info("Cilium DS doesn't exist")
		return true, nil
	}

	dsImage := clusterSpec.VersionsBundle.Cilium.Cilium.VersionedImage()
	containers := make([]corev1.Container, 0, len(ds.Spec.Template.Spec.Containers)+len(ds.Spec.Template.Spec.InitContainers))
	for _, c := range containers {
		if c.Image != dsImage {
			log.Info("Cilium DS container needs upgrade", "container", c.Name)
			return true, nil
		}
	}

	return false, nil
}

func ciliumOperatorNeedsUpgrade(ctx context.Context, log logr.Logger, client client.Client, clusterSpec *eksacluster.Spec) (bool, error) {
	operator, err := getCiliumDeployment(ctx, client)
	if err != nil {
		return false, err
	}

	if operator == nil {
		log.Info("Cilium operator deployment doesn't exist")
		return true, nil
	}

	operatorImage := clusterSpec.VersionsBundle.Cilium.Operator.VersionedImage()
	if len(operator.Spec.Template.Spec.Containers) == 0 {
		return false, errors.New("cilium-operator deployment doesn't have any containers")
	}

	if operator.Spec.Template.Spec.Containers[0].Image != operatorImage {
		return true, nil
	}

	return false, nil
}

func getCiliumDS(ctx context.Context, client client.Client) (*v1.DaemonSet, error) {
	ds := &v1.DaemonSet{}
	err := client.Get(ctx, types.NamespacedName{Name: "cilium", Namespace: "kube-system"}, ds)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func getCiliumDeployment(ctx context.Context, client client.Client) (*v1.Deployment, error) {
	deployment := &v1.Deployment{}
	err := client.Get(ctx, types.NamespacedName{Name: cilium.DeploymentName, Namespace: "kube-system"}, deployment)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func NewCiliumReconciler(
	client client.Client, log logr.Logger, tracker *remote.ClusterCacheTracker,
) *CiliumReconciler {
	return &CiliumReconciler{
		Client:  client,
		Log:     log,
		tracker: tracker,
	}
}
