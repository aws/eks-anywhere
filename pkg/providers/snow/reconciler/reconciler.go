package reconciler

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
)

const defaultRequeueTime = time.Minute

const (
	controlSpecPlaneAppliedCondition clusterv1.ConditionType = "ControlPlaneSpecApplied"
	controlPlaneReadyCondition       clusterv1.ConditionType = "ControlPlaneReady"
	eksaLicense                      string                  = "EKSA_LICENSE"
)

type Reconciler struct {
	client     client.Client
	validator  *snow.AwsClientValidator
	defaulters *snow.Defaulters
}

func New(client client.Client, validator *snow.AwsClientValidator, defaulters *snow.Defaulters) *Reconciler {
	return &Reconciler{
		client:     client,
		validator:  validator,
		defaulters: defaulters,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	dataCenterConfig := &anywherev1.SnowDatacenterConfig{}
	dataCenterName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := r.client.Get(ctx, dataCenterName, dataCenterConfig); err != nil {
		return controller.Result{}, err
	}

	if err := setupEnvVars(ctx, r.client); err != nil {
		log.Error(err, "Failed to set up snow env vars")
		return controller.Result{}, err
	}

	if !dataCenterConfig.Status.SpecValid {
		log.Info("Skipping cluster reconciliation because datacenter config is invalid", "data center", dataCenterConfig.Name)
		return controller.Result{
			Result: &ctrl.Result{
				Requeue:      true,
				RequeueAfter: defaultRequeueTime,
			},
		}, nil
	}

	log.V(4).Info("Fetching cluster spec")
	kubeClient := clientutil.NewKubeClient(r.client)
	specWithBundles, err := c.BuildSpec(ctx, kubeClient, cluster)
	if err != nil {
		return controller.Result{}, err
	}

	configManager := snow.NewConfigManager(r.defaulters, r.validator)
	if err = configManager.SetDefaultsAndValidate(ctx, specWithBundles.Config); err != nil {
		return controller.Result{}, err
	}
	log.Info("cluster", "name", cluster.Name)

	if result, err := r.reconcileControlPlaneSpec(ctx, cluster, specWithBundles, kubeClient, log); err != nil {
		return result, err
	}

	capiCluster, result, errCAPICLuster := r.getCAPICluster(ctx, cluster, log)
	if errCAPICLuster != nil {
		return result, errCAPICLuster
	}

	if !conditions.IsTrue(capiCluster, controlPlaneReadyCondition) {
		log.Info("waiting for control plane to be ready", "cluster", capiCluster.Name, "kind", capiCluster.Kind)
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, err
	}

	return controller.Result{}, nil
}

func (r *Reconciler) reconcileControlPlaneSpec(ctx context.Context, cluster *anywherev1.Cluster, specWithBundles *c.Spec, kubeClient *clientutil.KubeClient, log logr.Logger) (controller.Result, error) {
	if !conditions.IsTrue(cluster, controlSpecPlaneAppliedCondition) {
		log.Info("Applying control plane spec", "name", cluster.Name)
		controlPlaneSpec, err := snow.ControlPlaneSpec(ctx, specWithBundles, kubeClient)
		if err != nil {
			return controller.Result{}, err
		}
		if err := serverside.ReconcileYaml(ctx, r.client, controlPlaneSpec); err != nil {
			return controller.Result{Result: &ctrl.Result{
				RequeueAfter: defaultRequeueTime,
			}}, err
		}
		conditions.MarkTrue(cluster, controlSpecPlaneAppliedCondition)
	}
	return controller.Result{}, nil
}

func (r *Reconciler) getCAPICluster(ctx context.Context, cluster *anywherev1.Cluster, log logr.Logger) (*clusterv1.Cluster, controller.Result, error) {
	capiCluster := &clusterv1.Cluster{}
	capiClusterName := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: cluster.Name}
	log.Info("Searching for CAPI cluster", "name", cluster.Name)
	if err := r.client.Get(ctx, capiClusterName, capiCluster); err != nil {
		return nil, controller.Result{Result: &ctrl.Result{
			Requeue:      true,
			RequeueAfter: defaultRequeueTime,
		}}, err
	}
	return capiCluster, controller.Result{}, nil
}

func setupEnvVars(ctx context.Context, cli client.Client) error {
	credentials, caBundle, err := getSnowCredentials(ctx, cli)
	if err != nil {
		return fmt.Errorf("failed getting snow credentials secret: %v", err)
	}

	if err := os.Setenv(snow.EksaSnowCredentialsFileKey, string(credentials)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", snow.EksaSnowCredentialsFileKey, err)
	}

	if err := os.Setenv(snow.EksaSnowCABundlesFileKey, string(caBundle)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", snow.EksaSnowCABundlesFileKey, err)
	}

	if _, ok := os.LookupEnv(eksaLicense); !ok {
		if err := os.Setenv(eksaLicense, ""); err != nil {
			return fmt.Errorf("unable to set %s: %v", eksaLicense, err)
		}
	}

	return nil
}
