package certificates

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// KubernetesClient provides methods for interacting with Kubernetes.
type KubernetesClient interface {
	// InitClient initializes the Kubernetes client
	InitClient(clusterName string) error

	// CheckAPIServerReachability checks if the Kubernetes API server is reachable
	CheckAPIServerReachability(ctx context.Context) error

	// BackupKubeadmConfig backs up the kubeadm-config ConfigMap
	BackupKubeadmConfig(ctx context.Context, backupDir string) error

	// UpdateAPIServerEtcdClientSecret updates the apiserver-etcd-client secret
	UpdateAPIServerEtcdClientSecret(ctx context.Context, clusterName, backupDir string) error

	// IsCertificateExpired checks if the client certificate is expired
	IsCertificateExpired() bool

	InitClientWithKubeconfig(kubeconfigPath string) error
}

// DefaultKubernetesClient is the default implementation of KubernetesClient.
type DefaultKubernetesClient struct {
	client             kubernetes.Interface
	kubeconfigPath     string
	skipTLSVerify      bool
	certificateExpired bool
	config             *rest.Config
}

// NewKubernetesClient creates a new DefaultKubernetesClient.
func NewKubernetesClient() *DefaultKubernetesClient {
	return &DefaultKubernetesClient{}
}

// NewKubernetesClientWithKubeconfig creates a new DefaultKubernetesClient with kubeconfig path.
func NewKubernetesClientWithKubeconfig(kubeconfigPath string) *DefaultKubernetesClient {
	return &DefaultKubernetesClient{
		kubeconfigPath: kubeconfigPath,
	}
}

// NewKubernetesClientWithOptions creates a new DefaultKubernetesClient with options.
func NewKubernetesClientWithOptions(kubeconfigPath string, skipTLSVerify bool) *DefaultKubernetesClient {
	return &DefaultKubernetesClient{
		kubeconfigPath: kubeconfigPath,
		skipTLSVerify:  skipTLSVerify,
	}
}

// InitClient initializes the Kubernetes client for certificate renewal operations.
func (k *DefaultKubernetesClient) InitClient(clusterName string) error {
	if k.client != nil {
		return nil
	}

	kubeconfigPath, err := k.resolveKubeconfigPath(clusterName)
	if err != nil {
		return err
	}

	config, err := k.buildClientConfig(kubeconfigPath)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %v", err)
	}

	k.client = clientset
	return nil
}

// resolveKubeconfigPath determines the kubeconfig path to use.
func (k *DefaultKubernetesClient) resolveKubeconfigPath(clusterName string) (string, error) {
	var kubeconfigPath string
	if k.kubeconfigPath != "" {
		kubeconfigPath = k.kubeconfigPath
	} else {
		if clusterName != "" {
			pwd, err := os.Getwd()
			if err == nil {
				possiblePath := filepath.Join(pwd, clusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName))
				if _, err := os.Stat(possiblePath); err == nil {
					kubeconfigPath = possiblePath
					logger.Info("Using kubeconfig from cluster directory", "path", kubeconfigPath)
				}
			}
		}
		if kubeconfigPath == "" {
			return "", fmt.Errorf("no kubeconfig specified and KUBECONFIG environment variable is not set. " +
				"Try setting KUBECONFIG environment variable: export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig")
		}
	}

	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return "", fmt.Errorf("kubeconfig file does not exist: %s", kubeconfigPath)
	}

	return kubeconfigPath, nil
}

// buildClientConfig builds the client config from the kubeconfig file.
func (k *DefaultKubernetesClient) buildClientConfig(kubeconfigPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubeconfig: %v", err)
	}

	k.config = config
	if !k.skipTLSVerify {
		if k.checkCertificateExpired() {
			logger.MarkWarning("Warning: Client certificate appears to be expired. Enabling TLS skip verification.")
			k.skipTLSVerify = true
			k.certificateExpired = true
		}
	}

	if k.skipTLSVerify {
		config.TLSClientConfig = rest.TLSClientConfig{
			Insecure: true,
		}
	}
	return config, nil
}

// checkCertificateExpired checks if the client certificate in kubeconfig is expired.
func (k *DefaultKubernetesClient) checkCertificateExpired() bool {
	if k.config == nil {
		return false
	}

	if k.config.CertData != nil {
		cert, err := x509.ParseCertificate(k.config.CertData)
		if err == nil && time.Now().After(cert.NotAfter) {
			return true
		}
	}

	if k.config.CertFile != "" {
		certData, err := os.ReadFile(k.config.CertFile)
		if err == nil {
			cert, err := x509.ParseCertificate(certData)
			if err == nil && time.Now().After(cert.NotAfter) {
				return true
			}
		}
	}

	return false
}

// IsCertificateExpired returns whether the client certificate is expired.
func (k *DefaultKubernetesClient) IsCertificateExpired() bool {
	return k.certificateExpired
}

// CheckAPIServerReachability checks if the Kubernetes API server is reachable.
func (k *DefaultKubernetesClient) CheckAPIServerReachability(_ context.Context) error {
	if k.certificateExpired {
		logger.MarkWarning("Certificate is expired, attempting connection with TLS verification disabled...")
	}

	for i := 0; i < 5; i++ {
		_, err := k.client.Discovery().ServerVersion()
		if err == nil {
			if k.certificateExpired {
				logger.MarkPass("Successfully connected to API server (with expired certificate)")
			} else {
				logger.MarkPass("Successfully connected to API server")
			}
			return nil
		}

		if strings.Contains(err.Error(), "certificate") || strings.Contains(err.Error(), "x509") {
			logger.MarkWarning("Certificate error detected: %v", err)
			if !k.skipTLSVerify {
				logger.Info("💡 Consider using --skip-tls-verify flag if certificates are expired")
			}
		}

		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("kubernetes API server is not reachable")
}

// BackupKubeadmConfig backs up the kubeadm-config ConfigMap.
func (k *DefaultKubernetesClient) BackupKubeadmConfig(ctx context.Context, backupDir string) error {
	cm, err := k.client.CoreV1().ConfigMaps("kube-system").Get(ctx, "kubeadm-config", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting kubeadm-config: %v", err)
	}

	backupPath := filepath.Join(backupDir, "kubeadm-config.yaml")
	if err := os.WriteFile(backupPath, []byte(cm.Data["ClusterConfiguration"]), 0o600); err != nil {
		return fmt.Errorf("writing kubeadm config backup: %v", err)
	}

	return nil
}

// UpdateAPIServerEtcdClientSecret updates the apiserver-etcd-client secret.
func (k *DefaultKubernetesClient) UpdateAPIServerEtcdClientSecret(ctx context.Context, clusterName, backupDir string) error {
	logger.MarkPass("Updated apiserver-etcd-client secret", "cluster", clusterName)

	if err := k.ensureNamespaceExists(ctx, "eksa-system"); err != nil {
		return fmt.Errorf("ensuring eksa-system namespace exists: %v", err)
	}

	crtPath := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.crt")
	keyPath := filepath.Join(backupDir, tempLocalEtcdCertsDir, "apiserver-etcd-client.key")

	crtData, err := os.ReadFile(crtPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %v", err)
	}

	// get current sercet or create
	secretName := fmt.Sprintf("%s-apiserver-etcd-client", clusterName)
	secret, err := k.client.CoreV1().Secrets("eksa-system").Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get secret %s: %v", secretName, err)
		}

		// if sercet not exist, create
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: "eksa-system",
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.crt": crtData,
				"tls.key": keyData,
			},
		}

		_, err = k.client.CoreV1().Secrets("eksa-system").Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create secret %s: %v", secretName, err)
		}
	} else {
		// if sercet exist, renew it
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}

		secret.Data["tls.crt"] = crtData
		secret.Data["tls.key"] = keyData

		_, err = k.client.CoreV1().Secrets("eksa-system").Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update secret %s: %v", secretName, err)
		}
	}

	logger.V(2).Info("Successfully updated secret", "name", secretName)
	return nil
}

// ensureNamespaceExists ensures that the specified namespace exists.
func (k *DefaultKubernetesClient) ensureNamespaceExists(ctx context.Context, namespace string) error {
	_, err := k.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, err = k.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("create namespace %s: %v", namespace, err)
			}
			logger.Info("Created namespace %s", namespace)
		} else {
			return fmt.Errorf("check namespace %s: %v", namespace, err)
		}
	}
	return nil
}

// InitClientWithKubeconfig re-initializes the client with a user-supplied kubeconfig.
func (k *DefaultKubernetesClient) InitClientWithKubeconfig(kubeconfigPath string) error {
	k.kubeconfigPath = kubeconfigPath
	k.client = nil
	k.skipTLSVerify = false
	k.certificateExpired = false
	return k.InitClient("")
}
