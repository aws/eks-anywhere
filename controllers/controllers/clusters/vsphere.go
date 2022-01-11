package clusters

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
	"github.com/aws/eks-anywhere/controllers/controllers/resource"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// Struct that holds common methods and properties
type VSphereReconciler struct {
	Client    client.Client
	Log       logr.Logger
	Validator *vsphere.Validator
	Defaulter *vsphere.Defaulter
}

type VSphereClusterReconciler struct {
	VSphereReconciler
	*providerClusterReconciler

	capiResourceFetcher *resource.CapiResourceFetcher
}

func NewVSphereReconciler(client client.Client, log logr.Logger, validator *vsphere.Validator, defaulter *vsphere.Defaulter) *VSphereClusterReconciler {
	capiResourceFetcher := resource.NewCAPIResourceFetcher(client, log)
	return &VSphereClusterReconciler{
		VSphereReconciler: VSphereReconciler{
			Client:    client,
			Log:       log,
			Validator: validator,
			Defaulter: defaulter,
		},
		providerClusterReconciler: &providerClusterReconciler{},
		capiResourceFetcher:       capiResourceFetcher,
	}
}

func (v *VSphereReconciler) VsphereCredentials(ctx context.Context) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: "eksa-system",
		Name:      vsphere.CredentialsObjectName,
	}
	if err := v.Client.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func (v *VSphereReconciler) SetupEnvsAndDefaults(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig) error {
	secret, err := v.VsphereCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed getting vsphere credentials secret: %v", err)
	}

	vsphereUsername := secret.Data["username"]
	vspherePassword := secret.Data["password"]

	if err := os.Setenv(vsphere.EksavSphereUsernameKey, string(vsphereUsername)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", vsphere.EksavSphereUsernameKey, err)
	}

	if err := os.Setenv(vsphere.EksavSpherePasswordKey, string(vspherePassword)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", vsphere.EksavSpherePasswordKey, err)
	}

	if err := vsphere.SetupEnvVars(vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting env vars: %v", err)
	}

	if err := v.Defaulter.SetDefaultsForDatacenterConfig(ctx, vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	return nil
}

func (v *VSphereClusterReconciler) bundles(ctx context.Context, name, namespace string) (*releasev1alpha1.Bundles, error) {
	clusterBundle := &releasev1alpha1.Bundles{}
	err := v.capiResourceFetcher.FetchObjectByName(ctx, name, namespace, clusterBundle)
	if err != nil {
		return nil, err
	}
	return clusterBundle, nil
}

func (v *VSphereClusterReconciler) FetchAppliedSpec(ctx context.Context, cs *anywherev1.Cluster) (*c.Spec, error) {
	return c.BuildSpecForCluster(ctx, cs, v.bundles, nil)
}

func (v *VSphereClusterReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error) {
	dataCenterConfig := &anywherev1.VSphereDatacenterConfig{}
	dataCenterName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := v.Client.Get(ctx, dataCenterName, dataCenterConfig); err != nil {
		return reconciler.Result{}, err
	}
	// Set up envs for executing Govc cmd and default values for datacenter config
	if err := v.SetupEnvsAndDefaults(ctx, dataCenterConfig); err != nil {
		v.Log.Error(err, "Failed to set up env vars and default values for VsphereDatacenterConfig")
		return reconciler.Result{}, err
	}
	if !dataCenterConfig.Status.SpecValid {
		v.Log.Info("Skipping cluster reconciliation because data center config is invalid %v", dataCenterConfig)
		return reconciler.Result{}, nil
	}

	machineConfigMap := map[string]*anywherev1.VSphereMachineConfig{}

	for _, ref := range cluster.MachineConfigRefs() {
		machineConfig := &anywherev1.VSphereMachineConfig{}
		machineConfigName := types.NamespacedName{Namespace: cluster.Namespace, Name: ref.Name}
		if err := v.Client.Get(ctx, machineConfigName, machineConfig); err != nil {
			return reconciler.Result{}, err
		}
		machineConfigMap[ref.Name] = machineConfig
	}

	clusterSpec, err := v.FetchAppliedSpec(ctx, cluster)
	if err != nil {
		return reconciler.Result{}, err
	}
	vshepreClusterSpec := vsphere.NewSpec(clusterSpec, machineConfigMap, dataCenterConfig)

	if err := v.Validator.ValidateClusterMachineConfigs(ctx, vshepreClusterSpec); err != nil {
		return reconciler.Result{}, err
	}

	return reconciler.Result{}, nil
}
