package reconciler

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

const defaultRequeueTime = time.Minute

// TODO move these constants
const (
	managedEtcdReadyCondition             clusterv1.ConditionType = "ManagedEtcdReady"
	controlSpecPlaneAppliedCondition      clusterv1.ConditionType = "ControlPlaneSpecApplied"
	workerNodeSpecPlaneAppliedCondition   clusterv1.ConditionType = "WorkerNodeSpecApplied"
	extraObjectsSpecPlaneAppliedCondition clusterv1.ConditionType = "ExtraObjectsSpecApplied"
	cniSpecAppliedCondition               clusterv1.ConditionType = "CNISpecApplied"
	controlPlaneReadyCondition            clusterv1.ConditionType = "ControlPlaneReady"
)

type Reconciler struct {
	client    client.Client
	validator *vsphere.Validator
	defaulter *vsphere.Defaulter
	tracker   *remote.ClusterCacheTracker
}

func New(client client.Client, validator *vsphere.Validator, defaulter *vsphere.Defaulter, tracker *remote.ClusterCacheTracker) *Reconciler {
	return &Reconciler{
		client:    client,
		validator: validator,
		defaulter: defaulter,
		tracker:   tracker,
	}
}

func VsphereCredentials(ctx context.Context, cli client.Client) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: "eksa-system",
		Name:      vsphere.CredentialsObjectName,
	}
	if err := cli.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func SetupEnvVars(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig, cli client.Client) error {
	secret, err := VsphereCredentials(ctx, cli)
	if err != nil {
		return fmt.Errorf("failed getting vsphere credentials secret: %v", err)
	}

	vsphereUsername := secret.Data["username"]
	vspherePassword := secret.Data["password"]

	if err := os.Setenv(config.EksavSphereUsernameKey, string(vsphereUsername)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", config.EksavSphereUsernameKey, err)
	}

	if err := os.Setenv(config.EksavSpherePasswordKey, string(vspherePassword)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", config.EksavSpherePasswordKey, err)
	}

	if err := vsphere.SetupEnvVars(vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting env vars: %v", err)
	}

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	dataCenterConfig := &anywherev1.VSphereDatacenterConfig{}
	dataCenterName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := r.client.Get(ctx, dataCenterName, dataCenterConfig); err != nil {
		return controller.Result{}, err
	}
	// Set up envs for executing Govc cmd and default values for datacenter config
	if err := SetupEnvVars(ctx, dataCenterConfig, r.client); err != nil {
		log.Error(err, "Failed to set up env vars and default values for VsphereDatacenterConfig")
		return controller.Result{}, err
	}
	if !dataCenterConfig.Status.SpecValid {
		log.Info("Skipping cluster reconciliation because data center config is invalid", "data center", dataCenterConfig.Name)
		return controller.Result{
			Result: &ctrl.Result{
				Requeue:      true,
				RequeueAfter: defaultRequeueTime,
			},
		}, nil
	}

	machineConfigMap := map[string]*anywherev1.VSphereMachineConfig{}

	for _, ref := range cluster.MachineConfigRefs() {
		machineConfig := &anywherev1.VSphereMachineConfig{}
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

	vsphereClusterSpec := vsphere.NewSpec(specWithBundles, machineConfigMap, dataCenterConfig)

	if err := r.validator.ValidateClusterMachineConfigs(ctx, vsphereClusterSpec); err != nil {
		return controller.Result{}, err
	}

	workerNodeGroupMachineSpecs := make(map[string]anywherev1.VSphereMachineConfigSpec, len(cluster.Spec.WorkerNodeGroupConfigurations))
	for _, wnConfig := range cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = machineConfigMap[wnConfig.MachineGroupRef.Name].Spec
	}

	cp := machineConfigMap[specWithBundles.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	var etcdSpec *anywherev1.VSphereMachineConfigSpec
	if specWithBundles.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcd := machineConfigMap[specWithBundles.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		etcdSpec = &etcd.Spec
	}

	templateBuilder := vsphere.NewVsphereTemplateBuilder(&dataCenterConfig.Spec, &cp.Spec, etcdSpec, workerNodeGroupMachineSpecs, time.Now, true)
	clusterName := cluster.ObjectMeta.Name

	kubeadmconfigTemplateNames := make(map[string]string, len(cluster.Spec.WorkerNodeGroupConfigurations))
	workloadTemplateNames := make(map[string]string, len(cluster.Spec.WorkerNodeGroupConfigurations))

	for _, wnConfig := range cluster.Spec.WorkerNodeGroupConfigurations {
		kubeadmconfigTemplateNames[wnConfig.Name] = common.KubeadmConfigTemplateName(cluster.Name, wnConfig.MachineGroupRef.Name, time.Now)
		workloadTemplateNames[wnConfig.Name] = common.WorkerMachineTemplateName(cluster.Name, wnConfig.Name, time.Now)
		templateBuilder.WorkerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name]
	}

	cpOpt := func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, time.Now)
		controlPlaneUser := machineConfigMap[cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
		values["vsphereControlPlaneSshAuthorizedKey"] = controlPlaneUser.SshAuthorizedKeys[0]

		if cluster.Spec.ExternalEtcdConfiguration != nil {
			etcdUser := machineConfigMap[cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
			values["vsphereEtcdSshAuthorizedKey"] = etcdUser.SshAuthorizedKeys[0]
		}

		values["etcdTemplateName"] = common.EtcdMachineTemplateName(clusterName, time.Now)
	}
	log.Info("cluster", "name", cluster.Name)

	if result, err := r.reconcileControlPlaneSpec(ctx, cluster, templateBuilder, specWithBundles, cpOpt, log); err != nil {
		return result, err
	}

	if result, err := r.reconcileWorkerNodeSpec(ctx, cluster, templateBuilder, specWithBundles, workloadTemplateNames, kubeadmconfigTemplateNames); err != nil {
		return result, err
	}

	capiCluster, result, errCAPICLuster := r.getCAPICluster(ctx, cluster, log)
	if errCAPICLuster != nil {
		return result, errCAPICLuster
	}

	// wait for etcd if necessary
	if cluster.Spec.ExternalEtcdConfiguration != nil {
		if !conditions.Has(capiCluster, managedEtcdReadyCondition) || conditions.IsFalse(capiCluster, managedEtcdReadyCondition) {
			log.Info("Waiting for etcd to be ready", "cluster", cluster.Name)
			return controller.Result{Result: &ctrl.Result{
				RequeueAfter: defaultRequeueTime,
			}}, nil
		}
	}

	if !conditions.IsTrue(capiCluster, controlPlaneReadyCondition) {
		log.Info("waiting for control plane to be ready", "cluster", capiCluster.Name, "kind", capiCluster.Kind)
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, err
	}

	if result, err := r.reconcileExtraObjects(ctx, cluster, capiCluster, specWithBundles); err != nil {
		return result, err
	}

	if result, err := r.reconcileCNI(ctx, cluster, capiCluster, specWithBundles, log); err != nil {
		return result, err
	}

	return controller.Result{}, nil
}

func (r *Reconciler) reconcileCNI(ctx context.Context, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *c.Spec, log logr.Logger) (controller.Result, error) {
	if !conditions.Has(cluster, cniSpecAppliedCondition) || conditions.IsFalse(capiCluster, cniSpecAppliedCondition) {
		log.Info("Getting remote client", "client for cluster", capiCluster.Name)
		key := client.ObjectKey{
			Namespace: capiCluster.Namespace,
			Name:      capiCluster.Name,
		}
		remoteClient, err := r.tracker.GetClient(ctx, key)
		if err != nil {
			return controller.Result{}, err
		}

		log.Info("About to apply CNI")

		helm := executables.NewHelm(executables.NewExecutable("helm"), executables.WithInsecure())
		cilium := cilium.NewCilium(nil, helm)

		if err != nil {
			return controller.Result{}, err
		}
		ciliumSpec, err := cilium.GenerateManifest(ctx, specWithBundles, []string{constants.CapvSystemNamespace})
		if err != nil {
			return controller.Result{}, err
		}
		if err := serverside.ReconcileYaml(ctx, remoteClient, ciliumSpec); err != nil {
			return controller.Result{}, err
		}
		conditions.MarkTrue(cluster, cniSpecAppliedCondition)
	}
	return controller.Result{}, nil
}

func (r *Reconciler) reconcileExtraObjects(ctx context.Context, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *c.Spec) (controller.Result, error) {
	if !conditions.IsTrue(capiCluster, extraObjectsSpecPlaneAppliedCondition) {
		extraObjects := c.BuildExtraObjects(specWithBundles)

		for _, spec := range extraObjects.Values() {
			if err := serverside.ReconcileYaml(ctx, r.client, spec); err != nil {
				return controller.Result{}, err
			}
		}
		conditions.MarkTrue(cluster, extraObjectsSpecPlaneAppliedCondition)
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

func (r *Reconciler) reconcileWorkerNodeSpec(
	ctx context.Context, cluster *anywherev1.Cluster, templateBuilder providers.TemplateBuilder,
	specWithBundles *c.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string,
) (controller.Result, error) {
	if !conditions.IsTrue(cluster, workerNodeSpecPlaneAppliedCondition) {
		workersSpec, err := templateBuilder.GenerateCAPISpecWorkers(specWithBundles, workloadTemplateNames, kubeadmconfigTemplateNames)
		if err != nil {
			return controller.Result{}, err
		}

		if err := serverside.ReconcileYaml(ctx, r.client, workersSpec); err != nil {
			return controller.Result{}, err
		}

		conditions.MarkTrue(cluster, workerNodeSpecPlaneAppliedCondition)
	}
	return controller.Result{}, nil
}

func (r *Reconciler) reconcileControlPlaneSpec(ctx context.Context, cluster *anywherev1.Cluster, templateBuilder providers.TemplateBuilder, specWithBundles *c.Spec, cpOpt func(values map[string]interface{}), log logr.Logger) (controller.Result, error) {
	if !conditions.IsTrue(cluster, controlSpecPlaneAppliedCondition) {
		log.Info("Applying control plane spec", "name", cluster.Name)
		controlPlaneSpec, err := templateBuilder.GenerateCAPISpecControlPlane(specWithBundles, cpOpt)
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
