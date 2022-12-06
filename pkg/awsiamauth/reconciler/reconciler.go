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
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/crypto"
)

// RemoteClientRegistry defines methods for remote cluster controller clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// Reconciler allows to reconcile AWSIamConfig.
type Reconciler struct {
	certgen              crypto.CertificateGenerator
	templateBuilder      *awsiamauth.TemplateBuilder
	generateUUID         UUIDGenerator
	client               client.Client
	remoteClientRegistry RemoteClientRegistry
}

// UUIDGenerator generates a new UUID.
type UUIDGenerator func() uuid.UUID

// New returns a new Reconciler.
func New(certgen crypto.CertificateGenerator, generateUUID UUIDGenerator, client client.Client, remoteClientRegistry RemoteClientRegistry) *Reconciler {
	return &Reconciler{
		certgen:              certgen,
		templateBuilder:      &awsiamauth.TemplateBuilder{},
		generateUUID:         generateUUID,
		client:               client,
		remoteClientRegistry: remoteClientRegistry,
	}
}

// EnsureCASecret ensures the AWS IAM Authenticator secret is present.
// It uses a controller.Result to indicate when requeues are needed.
func (r *Reconciler) EnsureCASecret(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	clusterName := cluster.Name
	s := &corev1.Secret{}
	err := r.client.Get(ctx, types.NamespacedName{Name: awsiamauth.CASecretName(clusterName), Namespace: constants.EksaSystemNamespace}, s)
	if apierrors.IsNotFound(err) {
		logger.Info("Creating aws-iam-authenticator CA secret")
		return r.createCASecret(ctx, clusterName)
	}
	if err != nil {
		return controller.Result{}, fmt.Errorf("fetching secret %s: %v", awsiamauth.CASecretName(clusterName), err)
	}

	logger.Info("aws-iam-authenticator CA secret found. Skipping secret create.", "Name", awsiamauth.CASecretName(clusterName))
	return controller.Result{}, nil
}

func (r *Reconciler) createCASecret(ctx context.Context, clusterName string) (controller.Result, error) {
	yaml, err := r.templateBuilder.GenerateCertKeyPairSecret(r.certgen, clusterName)
	if err != nil {
		return controller.Result{}, fmt.Errorf("generating aws-iam-authenticator ca secret: %v", err)
	}

	return controller.Result{}, serverside.ReconcileYaml(ctx, r.client, yaml)
}

// Reconcile takes the AWS IAM Authenticator installation to the desired state defined in AWSIAMConfig.
// It uses a controller.Result to indicate when requeues are needed.
// Intended to be used in a kubernetes controller.
func (r *Reconciler) Reconcile(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	clusterSpec, err := anywhereCluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	result, err := clusters.CheckControlPlaneReady(ctx, r.client, logger, cluster)
	if err != nil {
		return controller.Result{}, fmt.Errorf("checking controlplane ready: %v", err)
	}

	if result.Return() {
		return result, nil
	}

	rClient, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(cluster))
	if err != nil {
		return controller.Result{}, err
	}

	var clusterID uuid.UUID
	cm := &corev1.ConfigMap{}
	err = rClient.Get(ctx, types.NamespacedName{Name: awsiamauth.AwsIamAuthConfigMapName, Namespace: constants.KubeSystemNamespace}, cm)
	if apierrors.IsNotFound(err) {
		// If configmap is not found, this is a first time install of aws-iam-authenticator on the cluster.
		// We use a newly generated UUID.
		clusterID = r.generateUUID()
	} else if err != nil {
		return controller.Result{}, fmt.Errorf("fetching configmap %s: %v", awsiamauth.AwsIamAuthConfigMapName, err)
	}

	logger.Info("Applying aws-iam-authenticator manifest")
	return r.applyIAMAuthManifest(ctx, rClient, clusterSpec, clusterID)
}

func (r *Reconciler) applyIAMAuthManifest(ctx context.Context, client client.Client, clusterSpec *anywhereCluster.Spec, clusterID uuid.UUID) (controller.Result, error) {
	yaml, err := r.templateBuilder.GenerateManifest(clusterSpec, clusterID)
	if err != nil {
		return controller.Result{}, fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}

	return controller.Result{}, serverside.ReconcileYaml(ctx, client, yaml)
}
