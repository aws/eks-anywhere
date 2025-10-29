package reconciler

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
)

// CNIReconciler is an interface for reconciling CNI in the Tinkerbell cluster reconciler.
type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
}

// RemoteClientRegistry is an interface that defines methods for remote clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// IPValidator is an interface that defines methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *cluster.Spec) (controller.Result, error)
}

// Reconciler reconciles a Nutanix cluster.
type Reconciler struct {
	client               client.Client
	validator            *nutanix.Validator
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	ipValidator          IPValidator
	*serverside.ObjectApplier
}

// New defines a new Nutanix reconciler.
func New(client client.Client, validator *nutanix.Validator, cniReconciler CNIReconciler, registry RemoteClientRegistry, ipValidator IPValidator) *Reconciler {
	return &Reconciler{
		client:               client,
		validator:            validator,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: registry,
		ipValidator:          ipValidator,
		ObjectApplier:        serverside.NewObjectApplier(client),
	}
}

func getSecret(ctx context.Context, kubectl client.Client, secretName, secretNS string) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: secretNS,
		Name:      secretName,
	}
	if err := kubectl.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// GetNutanixCredsFromSecret returns the Nutanix credentials from a secret.
func GetNutanixCredsFromSecret(ctx context.Context, kubectl client.Client, secretName, secretNS string) (credentials.BasicAuthCredential, error) {
	secret, err := getSecret(ctx, kubectl, secretName, secretNS)
	if err != nil {
		return credentials.BasicAuthCredential{}, fmt.Errorf("failed getting nutanix credentials secret: %v", err)
	}

	creds, err := credentials.ParseCredentials(secret.Data["credentials"])
	if err != nil {
		return credentials.BasicAuthCredential{}, fmt.Errorf("failed parsing nutanix credentials: %v", err)
	}

	return credentials.BasicAuthCredential{PrismCentral: credentials.PrismCentralBasicAuth{
		BasicAuth: credentials.BasicAuth{
			Username: creds.Username,
			Password: creds.Password,
		},
	}}, nil
}

func (r *Reconciler) reconcileClusterSecret(ctx context.Context, log logr.Logger, c *cluster.Spec) (controller.Result, error) {
	eksaSecret := &apiv1.Secret{}
	eksaSecretKey := client.ObjectKey{
		Namespace: constants.EksaSystemNamespace,
		Name:      nutanix.EKSASecretName(c),
	}

	if err := r.client.Get(ctx, eksaSecretKey, eksaSecret); err != nil {
		log.Error(err, "Failed to get EKS-A secret %s/%s", constants.EksaSystemNamespace, c.NutanixDatacenter.Spec.CredentialRef.Name)
		return controller.Result{}, err
	}

	capxSecret := &apiv1.Secret{}
	capxSecretKey := client.ObjectKey{
		Namespace: constants.EksaSystemNamespace,
		Name:      nutanix.CAPXSecretName(c),
	}
	if err := r.client.Get(ctx, capxSecretKey, capxSecret); err == nil {
		if reflect.DeepEqual(eksaSecret.Data, capxSecret.Data) {
			return controller.Result{}, nil
		}
		capxSecret.Data = eksaSecret.Data
		if err := r.client.Update(ctx, capxSecret); err != nil {
			log.Error(err, "Failed to update CAPX secret %s/%s", constants.EksaSystemNamespace, c.Cluster.Name)
			return controller.Result{}, err
		}
		return controller.Result{}, nil
	}

	capxSecret = &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
			Name:      nutanix.CAPXSecretName(c),
		},
		Data: eksaSecret.Data,
	}
	if err := r.client.Create(ctx, capxSecret); err != nil {
		log.Error(err, "Failed to create CAPX secret %s/%s", constants.EksaSystemNamespace, c.Cluster.Name)
		return controller.Result{}, err
	}

	return controller.Result{}, nil
}

// Reconcile reconciles the cluster to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, c *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "nutanix")

	clusterSpec, err := cluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), c)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner[*cluster.Spec]().Register(
		r.reconcileClusterSecret,
		r.ipValidator.ValidateControlPlaneIP,
		r.ValidateClusterSpec,
		clusters.CleanupStatusAfterValidate,
		r.ReconcileControlPlane,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

// ReconcileCNI reconciles the CNI to the desired state.
func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")

	c, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, c, clusterSpec)
}

// ValidateClusterSpec performs additional, context-aware validations on the cluster spec.
func (r *Reconciler) ValidateClusterSpec(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateClusterSpec")

	creds, err := GetNutanixCredsFromSecret(ctx, r.client, clusterSpec.NutanixDatacenter.Spec.CredentialRef.Name, "eksa-system")
	if err != nil {
		return controller.Result{}, err
	}

	if err := r.validator.ValidateClusterSpec(ctx, clusterSpec, creds); err != nil {
		log.Error(err, "Invalid cluster spec", "cluster", clusterSpec.Cluster.Name)
		failureMessage := err.Error()
		clusterSpec.Cluster.SetFailure(anywherev1.ClusterInvalidReason, failureMessage)
		return controller.ResultWithReturn(), nil
	}

	os.Setenv(constants.EksaNutanixUsernameKey, creds.PrismCentral.Username)
	os.Setenv(constants.EksaNutanixPasswordKey, creds.PrismCentral.Password)

	return controller.Result{}, nil
}

// ReconcileControlPlane reconciles the control plane to the desired state.
func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")

	cp, err := nutanix.ControlPlaneSpec(ctx, log, clientutil.NewKubeClient(r.client), clusterSpec)
	if err != nil {
		return controller.Result{}, err
	}

	// Set owner references on ClusterResourceSets and related objects
	if err := r.ensureOwnerReferences(ctx, log, clusterSpec, cp); err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileControlPlane(ctx, log, r.client, toClientControlPlane(cp))
}

func toClientControlPlane(cp *nutanix.ControlPlane) *clusters.ControlPlane {
	other := make([]client.Object, 0, len(cp.ConfigMaps)+len(cp.ClusterResourceSets)+len(cp.Secrets)+1)
	for _, o := range cp.ClusterResourceSets {
		other = append(other, o)
	}
	for _, o := range cp.ConfigMaps {
		other = append(other, o)
	}
	for _, o := range cp.Secrets {
		other = append(other, o)
	}

	return &clusters.ControlPlane{
		Cluster:                     cp.Cluster,
		ProviderCluster:             cp.ProviderCluster,
		KubeadmControlPlane:         cp.KubeadmControlPlane,
		ControlPlaneMachineTemplate: cp.ControlPlaneMachineTemplate,
		EtcdCluster:                 cp.EtcdCluster,
		EtcdMachineTemplate:         cp.EtcdMachineTemplate,
		Other:                       other,
	}
}

// ReconcileWorkers reconciles the workers to the desired state.
func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, spec *cluster.Spec) (controller.Result, error) {
	if spec.NutanixDatacenter == nil {
		return controller.Result{}, nil
	}
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")
	w, err := nutanix.WorkersSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileWorkersForEKSA(ctx, log, r.client, spec.Cluster, clusters.ToWorkers(w))
}

// CheckControlPlaneReady checks whether the control plane for an eks-a cluster is ready or not.
// Requeues with the appropriate wait times whenever the cluster is not ready yet.
func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

// ensureOwnerReferences ensures that ClusterResourceSets, ConfigMaps, and Secrets have proper owner references
// to the CAPI Cluster for garbage collection.
func (r *Reconciler) ensureOwnerReferences(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec, cp *nutanix.ControlPlane) error {
	// Get the CAPI cluster to use as owner
	capiCluster := &clusterv1.Cluster{}
	clusterKey := client.ObjectKey{
		Name:      clusterSpec.Cluster.Name,
		Namespace: constants.EksaSystemNamespace,
	}

	if err := r.client.Get(ctx, clusterKey, capiCluster); err != nil {
		// If the cluster doesn't exist yet, skip setting owner references - it will be set on next reconciliation
		log.V(5).Info("CAPI cluster not found yet, skipping owner reference setting", "cluster", clusterSpec.Cluster.Name)
		return nil
	}

	log.Info("Ensuring owner references for Nutanix cluster resources",
		"clusterResourceSets", len(cp.ClusterResourceSets),
		"configMaps", len(cp.ConfigMaps),
		"secrets", len(cp.Secrets))

	// Set owner references on ClusterResourceSets
	if err := r.setOwnerReferencesOnObjects(ctx, log, capiCluster, toClientObjects(cp.ClusterResourceSets)); err != nil {
		return fmt.Errorf("setting owner references on ClusterResourceSets: %w", err)
	}

	// Set owner references on ConfigMaps
	if err := r.setOwnerReferencesOnObjects(ctx, log, capiCluster, toClientObjects(cp.ConfigMaps)); err != nil {
		return fmt.Errorf("setting owner references on ConfigMaps: %w", err)
	}

	// Set owner references on Secrets
	if err := r.setOwnerReferencesOnObjects(ctx, log, capiCluster, toClientObjects(cp.Secrets)); err != nil {
		return fmt.Errorf("setting owner references on Secrets: %w", err)
	}

	return nil
}

// setOwnerReferencesOnObjects sets the owner reference on a list of objects.
func (r *Reconciler) setOwnerReferencesOnObjects(ctx context.Context, log logr.Logger, owner *clusterv1.Cluster, objects []client.Object) error {
	for _, obj := range objects {
		// Get the current object from the cluster to check if owner reference already exists
		currentObj := obj.DeepCopyObject().(client.Object)
		objKey := client.ObjectKey{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		}

		err := r.client.Get(ctx, objKey, currentObj)
		if err != nil {
			// If the object doesn't exist yet, skip it - owner reference will be set when it's created
			log.V(5).Info("Object not found yet, skipping owner reference", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
			continue
		}

		// Check if owner reference already exists
		ownerRefs := currentObj.GetOwnerReferences()
		hasOwnerRef := false
		for _, ref := range ownerRefs {
			if ref.UID == owner.UID {
				hasOwnerRef = true
				break
			}
		}

		if hasOwnerRef {
			// Owner reference already exists, skip
			continue
		}

		// Set the owner reference
		if err := controllerutil.SetOwnerReference(owner, currentObj, r.client.Scheme()); err != nil {
			return fmt.Errorf("setting owner reference on %s %s: %w", currentObj.GetObjectKind().GroupVersionKind().Kind, currentObj.GetName(), err)
		}

		// Update the object with the new owner reference
		if err := r.client.Update(ctx, currentObj); err != nil {
			return fmt.Errorf("updating %s %s with owner reference: %w", currentObj.GetObjectKind().GroupVersionKind().Kind, currentObj.GetName(), err)
		}

		log.Info("Set owner reference on object",
			"kind", currentObj.GetObjectKind().GroupVersionKind().Kind,
			"name", currentObj.GetName(),
			"owner", owner.Name)
	}

	return nil
}

// toClientObjects converts a slice of any kubernetes object type to []client.Object.
func toClientObjects[T client.Object](objs []T) []client.Object {
	result := make([]client.Object, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj)
	}
	return result
}
