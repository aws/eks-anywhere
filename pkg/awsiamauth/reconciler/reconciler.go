package reconciler

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
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
	secretName := awsiamauth.CASecretName(clusterName)

	s := &corev1.Secret{}
	err := r.client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: constants.EksaSystemNamespace}, s)
	if apierrors.IsNotFound(err) {
		logger.Info("Creating aws-iam-authenticator CA secret")
		return controller.Result{}, r.createCASecret(ctx, cluster)
	}
	if err != nil {
		return controller.Result{}, fmt.Errorf("fetching secret %s: %v", secretName, err)
	}

	logger.Info("aws-iam-authenticator CA secret found. Skipping secret create.", "name", secretName)
	return controller.Result{}, nil
}

func (r *Reconciler) createCASecret(ctx context.Context, cluster *anywherev1.Cluster) error {
	yaml, err := r.templateBuilder.GenerateCertKeyPairSecret(r.certgen, cluster.Name)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator ca secret: %v", err)
	}

	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		return fmt.Errorf("converting aws-iam-authenticator ca secret yaml to objects: %v", err)
	}

	for _, o := range objs {
		if err := r.client.Create(ctx, o); err != nil {
			return fmt.Errorf("creating aws-iam-authenticator ca secret: %v", err)
		}
	}

	return nil
}

// Reconcile takes the AWS IAM Authenticator installation to the desired state defined in AWSIAMConfig.
// It uses a controller.Result to indicate when requeues are needed.
// Intended to be used in a kubernetes controller.
func (r *Reconciler) Reconcile(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	clusterSpec, err := anywhereCluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	// CheckControlPlaneReady was not meant to be used here.
	// It was intended as a phase in cluster reconciliation.
	// TODO (pokearu): Break down the function to better reuse it.
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
		// The configmap clusterID and kubeconfig token need to match. Hence the kubeconfig secret is created for first install.
		clusterID = r.generateUUID()
		logger.Info("Creating aws-iam-authenticator kubeconfig secret")
		if err := r.createKubeconfigSecret(ctx, clusterSpec, cluster, clusterID); err != nil {
			return controller.Result{}, err
		}
	} else if err != nil {
		return controller.Result{}, fmt.Errorf("fetching configmap %s: %v", awsiamauth.AwsIamAuthConfigMapName, err)
	}

	logger.Info("Applying aws-iam-authenticator manifest")
	if err := r.applyIAMAuthManifest(ctx, rClient, clusterSpec, clusterID); err != nil {
		return controller.Result{}, fmt.Errorf("applying aws-iam-authenticator manifest: %v", err)
	}

	return controller.Result{}, nil
}

func (r *Reconciler) applyIAMAuthManifest(ctx context.Context, client client.Client, clusterSpec *anywhereCluster.Spec, clusterID uuid.UUID) error {
	yaml, err := r.templateBuilder.GenerateManifest(clusterSpec, clusterID)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}

	return serverside.ReconcileYaml(ctx, client, yaml)
}

func (r *Reconciler) createKubeconfigSecret(ctx context.Context, clusterSpec *anywhereCluster.Spec, cluster *anywherev1.Cluster, clusterID uuid.UUID) error {
	apiServerEndpoint := fmt.Sprintf("https://%s:6443", cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)

	clusterCaSecretName := clusterapi.ClusterCASecretName(cluster.Name)
	clusterCaSecret := &corev1.Secret{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: clusterCaSecretName, Namespace: constants.EksaSystemNamespace}, clusterCaSecret); err != nil {
		return fmt.Errorf("fetching cluster ca secret: %v", err)
	}

	clusterTLSCrt, ok := clusterCaSecret.Data["tls.crt"]
	if !ok {
		return fmt.Errorf("tls.crt key not found in cluster CA secret %s", clusterCaSecret.Name)
	}

	yaml, err := r.templateBuilder.GenerateKubeconfig(clusterSpec, clusterID, apiServerEndpoint, base64.StdEncoding.EncodeToString(clusterTLSCrt))
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	kubeconfigSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      awsiamauth.KubeconfigSecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
		StringData: map[string]string{
			"value": string(yaml),
		},
	}

	if err := r.client.Create(ctx, kubeconfigSecret); err != nil {
		return fmt.Errorf("creating aws-iam-authenticator kubeconfig secret: %v", err)
	}

	return nil
}

// ReconcileDelete deletes any AWS Iam authenticator specific resources leftover on the eks-a cluster.
func (r *Reconciler) ReconcileDelete(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) error {
	secretName := awsiamauth.CASecretName(cluster.Name)
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      secretName,
			Namespace: constants.EksaSystemNamespace,
		},
	}

	logger.Info("Deleting aws-iam-authenticator ca secret", "name", secretName, "namespace", constants.EksaSystemNamespace)
	if err := r.deleteObject(ctx, secret); err != nil {
		return err
	}

	kubeconfigSecName := awsiamauth.KubeconfigSecretName(cluster.Name)
	kubeConfigSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      kubeconfigSecName,
			Namespace: constants.EksaSystemNamespace,
		},
	}

	logger.Info("Deleting aws-iam-authenticator kubeconfig secret", "name", kubeconfigSecName, "namespace", constants.EksaSystemNamespace)
	if err := r.deleteObject(ctx, kubeConfigSecret); err != nil {
		return err
	}

	return nil
}

// deleteObject deletes a kubernetes object. It's idempotent, if the object doesn't exist,
// it doesn't return an error.
func (r *Reconciler) deleteObject(ctx context.Context, obj client.Object) error {
	err := r.client.Delete(ctx, obj)
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting aws-iam-authenticator %s %s: %v", obj.GetObjectKind(), obj.GetName(), err)
	}

	return nil
}
