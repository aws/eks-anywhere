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

type VSphereReconciler struct {
	*providerClusterReconciler

	client    client.Client
	log       logr.Logger
	validator *vsphere.Validator
	defaulter *vsphere.Defaulter

	capiResourceFetcher *resource.CapiResourceFetcher
}

func NewVSphereReconciler(client client.Client, log logr.Logger, validator *vsphere.Validator, defaulter *vsphere.Defaulter) *VSphereReconciler {
	capiResourceFetcher := resource.NewCAPIResourceFetcher(client, log)
	return &VSphereReconciler{
		providerClusterReconciler: &providerClusterReconciler{},
		client:                    client,
		log:                       log,
		validator:                 validator,
		defaulter:                 defaulter,
		capiResourceFetcher:       capiResourceFetcher,
	}
}

// TODO remove code
func (v *VSphereReconciler) vsphereCredentials(ctx context.Context) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: "eksa-system",
		Name:      vsphere.CredentialsObjectName,
	}
	if err := v.client.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// TODO remove code
func (v *VSphereReconciler) setupEnvsAndDefaults(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig) error {
	secret, err := v.vsphereCredentials(ctx)
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

	if err := v.defaulter.SetDefaultsForDatacenterConfig(ctx, vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting default values for vsphere datacenter config: %v", err)
	}

	return nil
}

func (v *VSphereReconciler) bundles(ctx context.Context, name, namespace string) (*releasev1alpha1.Bundles, error) {
	clusterBundle := &releasev1alpha1.Bundles{}
	err := v.capiResourceFetcher.FetchObjectByName(ctx, name, namespace, clusterBundle)
	if err != nil {
		return nil, err
	}
	return clusterBundle, nil
}

func (v *VSphereReconciler) FetchAppliedSpec(ctx context.Context, cs *anywherev1.Cluster) (*c.Spec, error) {
	return c.BuildSpecForCluster(ctx, cs, v.bundles, nil)
}

func (v *VSphereReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error) {
	dc := &anywherev1.VSphereDatacenterConfig{}
	dcName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := v.client.Get(ctx, dcName, dc); err != nil {
		return reconciler.Result{}, err
	}
	// Set up envs for executing Govc cmd and default values for datacenter config
	if err := v.setupEnvsAndDefaults(ctx, dc); err != nil {
		v.log.Error(err, "Failed to set up env vars and default values for VsphereDatacenterConfig")
		return reconciler.Result{}, err
	}

	mcMap := map[string]*anywherev1.VSphereMachineConfig{}

	if cluster.Spec.ExternalEtcdConfiguration != nil {
		mc := &anywherev1.VSphereMachineConfig{}
		mcName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}
		if err := v.client.Get(ctx, mcName, mc); err != nil {
			return reconciler.Result{}, err
		}
		v.log.V(4).Info("Using etcd machine config %v", mc)
		mcMap[cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] = mc
	}

	cpMachine := &anywherev1.VSphereMachineConfig{}
	cpMachineNameName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name}
	if err := v.client.Get(ctx, cpMachineNameName, cpMachine); err != nil {
		return reconciler.Result{}, err
	}
	v.log.V(4).Info("Using cp machine config %v", cpMachine)
	mcMap[cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] = cpMachine

	wnMachine := &anywherev1.VSphereMachineConfig{}
	wnMachineName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name}
	if err := v.client.Get(ctx, wnMachineName, wnMachine); err != nil {
		return reconciler.Result{}, err
	}
	v.log.V(4).Info("Using wn machine config %v", wnMachine)
	mcMap[cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name] = wnMachine

	cs, err := v.FetchAppliedSpec(ctx, cluster)
	if err != nil {
		return reconciler.Result{}, err
	}
	vs := vsphere.NewSpec(cs, mcMap, dc)

	if err := v.validator.ValidateCluster(ctx, vs); err != nil {
		return reconciler.Result{}, err
	}

	return reconciler.Result{}, nil
}
