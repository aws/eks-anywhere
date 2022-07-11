package reconciler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	eksacluster "github.com/aws/eks-anywhere/pkg/cluster"
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

const defaultRequeueTime = time.Second * 10

// TODO move these constants
const (
	managedEtcdReadyCondition             clusterv1.ConditionType = "ManagedEtcdReady"
	controlSpecPlaneAppliedCondition      clusterv1.ConditionType = "ControlPlaneSpecApplied"
	workerNodeSpecPlaneAppliedCondition   clusterv1.ConditionType = "WorkerNodeSpecApplied"
	extraObjectsSpecPlaneAppliedCondition clusterv1.ConditionType = "ExtraObjectsSpecApplied"
	controlPlaneReadyCondition            clusterv1.ConditionType = "ControlPlaneReady"
)

// Struct that holds common methods and properties
type VSphereReconciler struct {
	Client    client.Client
	Log       logr.Logger
	Validator *vsphere.Validator
	Defaulter *vsphere.Defaulter
	tracker   *remote.ClusterCacheTracker
}

type VSphereClusterReconciler struct {
	VSphereReconciler
}

func NewVSphereReconciler(client client.Client, log logr.Logger, validator *vsphere.Validator, defaulter *vsphere.Defaulter, tracker *remote.ClusterCacheTracker) *VSphereClusterReconciler {
	return &VSphereClusterReconciler{
		VSphereReconciler: VSphereReconciler{
			Client:    client,
			Log:       log,
			Validator: validator,
			Defaulter: defaulter,
			tracker:   tracker,
		},
	}
}

func VsphereCredentials(ctx context.Context, cli client.Client) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
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

	if err := os.Setenv(vsphere.EksavSphereUsernameKey, string(vsphereUsername)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", vsphere.EksavSphereUsernameKey, err)
	}

	if err := os.Setenv(vsphere.EksavSpherePasswordKey, string(vspherePassword)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", vsphere.EksavSpherePasswordKey, err)
	}

	if err := vsphere.SetupEnvVars(vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting env vars: %v", err)
	}

	return nil
}

func (v *VSphereClusterReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (controller.Result, error) {
	dataCenterConfig := &anywherev1.VSphereDatacenterConfig{}
	dataCenterName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := v.Client.Get(ctx, dataCenterName, dataCenterConfig); err != nil {
		return controller.Result{}, err
	}
	// Set up envs for executing Govc cmd and default values for datacenter config
	if err := SetupEnvVars(ctx, dataCenterConfig, v.Client); err != nil {
		v.Log.Error(err, "Failed to set up env vars and default values for VsphereDatacenterConfig")
		return controller.Result{}, err
	}
	if !dataCenterConfig.Status.SpecValid {
		v.Log.Info("Skipping cluster reconciliation because data center config is invalid", "data center", dataCenterConfig.Name)
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
		if err := v.Client.Get(ctx, machineConfigName, machineConfig); err != nil {
			return controller.Result{}, err
		}
		machineConfigMap[ref.Name] = machineConfig
	}

	v.Log.V(4).Info("Fetching bundle", "cluster name", cluster.Spec.ManagementCluster.Name)

	bundlesCluster := cluster
	if cluster.Spec.BundlesRef == nil {
		managementCluster := &anywherev1.Cluster{}
		managementClusterName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.ManagementCluster.Name}
		if err := v.Client.Get(ctx, managementClusterName, managementCluster); err != nil {
			return controller.Result{}, err
		}

		bundlesCluster = managementCluster
	}

	specWithBundles, err := eksacluster.BuildSpec(ctx, clientutil.NewKubeClient(v.Client), bundlesCluster)
	if err != nil {
		return controller.Result{}, err
	}

	vsphereClusterSpec := vsphere.NewSpec(specWithBundles, machineConfigMap, dataCenterConfig)

	if err := v.Validator.ValidateClusterMachineConfigs(ctx, vsphereClusterSpec); err != nil {
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
	v.Log.Info("cluster", "name", cluster.Name)

	if result, err := v.reconcileControlPlaneSpec(ctx, cluster, templateBuilder, specWithBundles, cpOpt); err != nil {
		return result, err
	}

	if result, err := v.reconcileWorkerNodeSpec(ctx, cluster, templateBuilder, specWithBundles, workloadTemplateNames, kubeadmconfigTemplateNames); err != nil {
		return result, err
	}

	capiCluster, result, errCAPICLuster := v.getCAPICluster(ctx, cluster)
	if errCAPICLuster != nil {
		return result, errCAPICLuster
	}

	// wait for etcd if necessary
	if cluster.Spec.ExternalEtcdConfiguration != nil {
		if !conditions.Has(capiCluster, managedEtcdReadyCondition) || conditions.IsFalse(capiCluster, managedEtcdReadyCondition) {
			v.Log.Info("Waiting for etcd to be ready", "cluster", cluster.Name)
			return controller.Result{Result: &ctrl.Result{
				RequeueAfter: defaultRequeueTime,
			}}, nil
		}
	}

	if !conditions.IsTrue(capiCluster, controlPlaneReadyCondition) {
		v.Log.Info("waiting for control plane to be ready", "cluster", capiCluster.Name, "kind", capiCluster.Kind)
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, nil
	}

	if result, err := v.reconcileExtraObjects(ctx, cluster, capiCluster, specWithBundles); err != nil {
		return result, err
	}

	if result, err := v.reconcileCNI(ctx, cluster, capiCluster, specWithBundles); err != nil {
		return result, err
	}

	return controller.Result{}, nil
}

func (v *VSphereClusterReconciler) reconcileCNI(ctx context.Context, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *eksacluster.Spec) (controller.Result, error) {
	v.Log.Info("Getting remote client", "client for cluster", capiCluster.Name)
	key := client.ObjectKey{
		Namespace: capiCluster.Namespace,
		Name:      capiCluster.Name,
	}
	remoteClient, err := v.tracker.GetClient(ctx, key)
	if err != nil {
		return controller.Result{}, err
	}

	v.Log.Info("Applying CNI")
	ciliumDS := &v1.DaemonSet{}
	ciliumDSName := types.NamespacedName{Namespace: "kube-system", Name: cilium.DaemonSetName}
	err = remoteClient.Get(ctx, ciliumDSName, ciliumDS)
	if err != nil {
		if apierrors.IsNotFound(err) {
			v.Log.Info("Deploying Cilium DS")
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
	v.Log.Info("Upgrading Cilium")
	needsUpgrade, err := ciliumNeedsUpgrade(ctx, v.Log, remoteClient, specWithBundles)
	if err != nil {
		return controller.Result{}, err
	}

	if !needsUpgrade {
		v.Log.Info("Skipping Cilium")
		return controller.Result{}, nil
	}

	helm := executables.NewHelm(executables.NewExecutable("helm"))
	templater := cilium.NewTemplater(helm)
	preflight, err := templater.GenerateUpgradePreflightManifest(ctx, specWithBundles)
	if err != nil {
		return controller.Result{}, err
	}

	v.Log.Info("Installing Cilium upgrade preflight manifest")
	if err := serverside.ReconcileYaml(ctx, remoteClient, preflight); err != nil {
		return controller.Result{}, err
	}

	ciliumDS = &v1.DaemonSet{}
	ciliumDSName = types.NamespacedName{Namespace: "kube-system", Name: "cilium"}
	if err := remoteClient.Get(ctx, ciliumDSName, ciliumDS); err != nil {
		v.Log.Info("Cilium DS not found")
		return controller.Result{}, err
	}

	if err := cilium.CheckDaemonSetReady(ciliumDS); err != nil {
		v.Log.Info("Cilium DS not ready")
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, nil
	}

	preFlightCiliumDS := &v1.DaemonSet{}
	preFlightCiliumDSName := types.NamespacedName{Namespace: "kube-system", Name: cilium.PreflightDaemonSetName}
	if err := remoteClient.Get(ctx, preFlightCiliumDSName, preFlightCiliumDS); err != nil {
		v.Log.Info("Preflight Cilium DS not found.")
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, err
	}

	if err := cilium.CheckPreflightDaemonSetReady(ciliumDS, preFlightCiliumDS); err != nil {
		v.Log.Info("Preflight DS not ready ")
		return controller.Result{Result: &ctrl.Result{
			RequeueAfter: defaultRequeueTime,
		}}, err
	}

	v.Log.Info("Deleting Preflight Cilium objects")
	if err := serverside.DeleteYaml(ctx, remoteClient, preflight); err != nil {
		v.Log.Info("Error deleting Preflight Cilium objects")
		return controller.Result{}, err
	}

	v.Log.Info("Generating Cilium upgrade manifest")
	upgradeManifest, err := templater.GenerateUpgradeManifest(ctx, specWithBundles, specWithBundles)
	if err != nil {
		return controller.Result{}, err
	}

	if err := serverside.ReconcileYaml(ctx, remoteClient, upgradeManifest); err != nil {
		return controller.Result{}, err
	}
	return controller.Result{}, nil
}

func (v *VSphereClusterReconciler) reconcileExtraObjects(ctx context.Context, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *eksacluster.Spec) (controller.Result, error) {
	if !conditions.IsTrue(capiCluster, extraObjectsSpecPlaneAppliedCondition) {
		extraObjects := eksacluster.BuildExtraObjects(specWithBundles)

		for _, spec := range extraObjects.Values() {
			if err := serverside.ReconcileYaml(ctx, v.Client, spec); err != nil {
				return controller.Result{}, err
			}
		}
		conditions.MarkTrue(cluster, extraObjectsSpecPlaneAppliedCondition)
	}
	return controller.Result{}, nil
}

func (v *VSphereClusterReconciler) getCAPICluster(ctx context.Context, cluster *anywherev1.Cluster) (*clusterv1.Cluster, controller.Result, error) {
	capiCluster := &clusterv1.Cluster{}
	capiClusterName := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: cluster.Name}
	v.Log.Info("Searching for CAPI cluster", "name", cluster.Name)
	if err := v.Client.Get(ctx, capiClusterName, capiCluster); err != nil {
		return nil, controller.Result{Result: &ctrl.Result{
			Requeue:      true,
			RequeueAfter: defaultRequeueTime,
		}}, err
	}
	return capiCluster, controller.Result{}, nil
}

func (v *VSphereClusterReconciler) reconcileWorkerNodeSpec(
	ctx context.Context, cluster *anywherev1.Cluster, templateBuilder providers.TemplateBuilder,
	specWithBundles *eksacluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string,
) (controller.Result, error) {
	if !conditions.IsTrue(cluster, workerNodeSpecPlaneAppliedCondition) {
		workersSpec, err := templateBuilder.GenerateCAPISpecWorkers(specWithBundles, workloadTemplateNames, kubeadmconfigTemplateNames)
		if err != nil {
			return controller.Result{}, err
		}

		if err := serverside.ReconcileYaml(ctx, v.Client, workersSpec); err != nil {
			return controller.Result{}, err
		}

		conditions.MarkTrue(cluster, workerNodeSpecPlaneAppliedCondition)
	}
	return controller.Result{}, nil
}

func (v *VSphereClusterReconciler) reconcileControlPlaneSpec(ctx context.Context, cluster *anywherev1.Cluster, templateBuilder providers.TemplateBuilder, specWithBundles *eksacluster.Spec, cpOpt func(values map[string]interface{})) (controller.Result, error) {
	if !conditions.IsTrue(cluster, controlSpecPlaneAppliedCondition) {
		v.Log.Info("Applying control plane spec", "name", cluster.Name)
		controlPlaneSpec, err := templateBuilder.GenerateCAPISpecControlPlane(specWithBundles, cpOpt)
		if err != nil {
			return controller.Result{}, err
		}
		if err := serverside.ReconcileYaml(ctx, v.Client, controlPlaneSpec); err != nil {
			return controller.Result{Result: &ctrl.Result{
				RequeueAfter: defaultRequeueTime,
			}}, err
		}
		conditions.MarkTrue(cluster, controlSpecPlaneAppliedCondition)
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
