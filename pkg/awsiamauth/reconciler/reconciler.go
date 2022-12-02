package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/crypto"
)

// AWSIamConfigReconciler is an interface for reconciling AWSIAMConfig in the Cluster reconciler.
type AWSIamConfigReconciler interface {
	ReconcileAWSIAMAuthCASecret(ctx context.Context, logger logr.Logger, client client.Client, clusterName string) (controller.Result, error)
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, cluster *anywherev1.Cluster) (controller.Result, error)
}

// RemoteClientRegistry is an interface that defines methods for remote clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// Reconciler allows to reconcile AWSIamConfig.
type Reconciler struct {
	certgen              crypto.CertificateGenerator
	templateBuilder      *awsiamauth.TemplateBuilder
	clusterID            uuid.UUID
	remoteClientRegistry RemoteClientRegistry
}

// New returns a new Reconciler.
func New(certgen crypto.CertificateGenerator, clusterID uuid.UUID, remoteClientRegistry RemoteClientRegistry) *Reconciler {
	return &Reconciler{
		certgen:              certgen,
		templateBuilder:      &awsiamauth.TemplateBuilder{},
		clusterID:            clusterID,
		remoteClientRegistry: remoteClientRegistry,
	}
}

// ReconcileAWSIAMAuthCASecret ensures the AWS IAM Authenticator secret is present.
// It uses a controller.Result to indicate when requeues are needed.
func (r *Reconciler) ReconcileAWSIAMAuthCASecret(ctx context.Context, logger logr.Logger, client client.Client, clusterName string) (controller.Result, error) {
	// Fetch the CA Secret
	s := &corev1.Secret{}
	err := client.Get(ctx, types.NamespacedName{Name: awsiamauth.GetAwsIamAuthCaSecretName(clusterName), Namespace: constants.EksaSystemNamespace}, s)
	if apierrors.IsNotFound(err) {
		logger.Info("Applying aws-iam-authenticator CA secret")
		return r.applyCASecret(ctx, client, clusterName)
	}
	if err != nil {
		return controller.Result{}, fmt.Errorf("fetching secret %s: %v", awsiamauth.GetAwsIamAuthCaSecretName(clusterName), err)
	}

	// If Secret is present, no-op
	logger.Info("aws-iam-authenticator CA secret found", "Name", awsiamauth.GetAwsIamAuthCaSecretName(clusterName))
	return controller.Result{}, nil
}

func (r *Reconciler) applyCASecret(ctx context.Context, client client.Client, clusterName string) (controller.Result, error) {
	// Generate k8s secret yaml for aws-iam-authenticator
	yaml, err := r.templateBuilder.GenerateCertKeyPairSecret(r.certgen, clusterName)
	if err != nil {
		return controller.Result{}, fmt.Errorf("generating aws-iam-authenticator ca secret: %v", err)
	}

	return controller.Result{}, serverside.ReconcileYaml(ctx, client, yaml)
}

// Reconcile takes the AWS IAM Authenticator indentity provider to the desired state defined in AWSIAMConfig.
// It uses a controller.Result to indicate when requeues are needed.
// Intended to be used in a kubernetes controller.
func (r *Reconciler) Reconcile(ctx context.Context, logger logr.Logger, client client.Client, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Build the cluster spec for the Cluster
	clusterSpec, err := anywhereCluster.BuildSpec(ctx, clientutil.NewKubeClient(client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	// Get the client object for the workload cluster
	rClient, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(cluster))
	if err != nil {
		return controller.Result{}, err
	}

	var clusterID uuid.UUID
	// Fetch the aws-iam-authenticator config map.
	cm := &corev1.ConfigMap{}
	err = rClient.Get(ctx, types.NamespacedName{Name: awsiamauth.AwsIamAuthConfigMapName, Namespace: constants.KubeSystemNamespace}, cm)
	if apierrors.IsNotFound(err) {
		// Use a new clusterID
		clusterID = r.clusterID
		return r.applyIAMAuthManifest(ctx, rClient, clusterSpec, clusterID)
	}
	if err != nil {
		return controller.Result{}, fmt.Errorf("fetching configmap %s: %v", awsiamauth.AwsIamAuthConfigMapName, err)
	}

	logger.Info("Applying aws-iam-authenticator manifest")
	return r.applyIAMAuthManifest(ctx, rClient, clusterSpec, clusterID)
}

func (r *Reconciler) applyIAMAuthManifest(ctx context.Context, client client.Client, clusterSpec *anywhereCluster.Spec, clusterID uuid.UUID) (controller.Result, error) {
	// Generate AWS IAM Authenticator manifest yaml
	yaml, err := r.templateBuilder.GenerateManifest(clusterSpec, clusterID)
	if err != nil {
		return controller.Result{}, fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}

	return controller.Result{}, serverside.ReconcileYaml(ctx, client, yaml)
}
