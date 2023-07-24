package clustermanager

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/utils/oci"
)

const (
	iptablesLegacyLabel           = "anywhere.eks.amazonaws.com/iptableslegacy"
	iptablesLegacyKubeProxyDSName = "kube-proxy-iptables-legacy"
	k8sAppLabel                   = "k8s-app"
	kubeProxyLabel                = "kube-proxy"
	kubeProxyDSName               = "kube-proxy"
	kubeProxyDSNamespace          = "kube-system"
)

var firstEKSDWithNewKubeProxy = map[anywherev1.KubernetesVersion]int{
	anywherev1.Kube122: 22,
	anywherev1.Kube123: 17,
	anywherev1.Kube124: 12,
	anywherev1.Kube125: 8,
}

// ClientFactory builds Kubernetes clients.
type ClientFactory interface {
	// BuildClientFromKubeconfig builds a Kubernetes client from a kubeconfig file.
	BuildClientFromKubeconfig(kubeconfigPath string) (kubernetes.Client, error)
}

// NewKubeProxyCLIUpgrader builds a new KubeProxyCLIUpgrader.
func NewKubeProxyCLIUpgrader(log logr.Logger, factory ClientFactory, opts ...KubeProxyCLIUpgraderOpt) KubeProxyCLIUpgrader {
	u := &KubeProxyCLIUpgrader{
		log:           log,
		clientFactory: factory,
		retrier:       *retrier.NewWithMaxRetries(12, time.Second),
	}

	for _, opt := range opts {
		opt(u)
	}

	return *u
}

// KubeProxyCLIUpgrader prepares a cluster for a kube-proxy upgrade.
// It's mostly a wrapper around [KubeProxyUpgrader] to be used from the CLI.
// It builds clients from kubeconfig files and facilitates mocking. It also uses a retrier
// around [KubeProxyCLIUpgrader] to deal with transient errors.
type KubeProxyCLIUpgrader struct {
	clientFactory ClientFactory
	log           logr.Logger
	retrier       retrier.Retrier
}

// KubeProxyCLIUpgraderOpt allows to customize a KubeProxyCLIUpgrader
// on construction.
type KubeProxyCLIUpgraderOpt func(*KubeProxyCLIUpgrader)

// KubeProxyCLIUpgraderRetrier allows to use a custom retrier.
func KubeProxyCLIUpgraderRetrier(retrier retrier.Retrier) KubeProxyCLIUpgraderOpt {
	return func(u *KubeProxyCLIUpgrader) {
		u.retrier = retrier
	}
}

// PrepareUpgrade perfoms the necessary steps prior to a kube-proxy upgrade.
func (u KubeProxyCLIUpgrader) PrepareUpgrade(ctx context.Context,
	spec *cluster.Spec,
	managementClusterKubeconfigPath, workloadClusterKubeconfigPath string,
) error {
	managementClusterClient, workloadClusterClient, err := u.buildClients(
		managementClusterKubeconfigPath, workloadClusterKubeconfigPath,
	)
	if err != nil {
		return err
	}

	up := NewKubeProxyUpgrader()

	return u.retrier.Retry(func() error {
		return up.PrepareForUpgrade(ctx, u.log, managementClusterClient, workloadClusterClient, spec)
	})
}

// CleanupAfterUpgrade perfoms the necessary steps after an upgrade.
func (u KubeProxyCLIUpgrader) CleanupAfterUpgrade(ctx context.Context,
	spec *cluster.Spec,
	managementClusterKubeconfigPath, workloadClusterKubeconfigPath string,
) error {
	managementClusterClient, workloadClusterClient, err := u.buildClients(
		managementClusterKubeconfigPath, workloadClusterKubeconfigPath,
	)
	if err != nil {
		return err
	}

	up := NewKubeProxyUpgrader()

	return u.retrier.Retry(func() error {
		return up.CleanupAfterUpgrade(ctx, u.log, managementClusterClient, workloadClusterClient, spec)
	})
}

func (u KubeProxyCLIUpgrader) buildClients(
	managementClusterKubeconfigPath, workloadClusterKubeconfigPath string,
) (managementClusterClient, workloadClusterClient kubernetes.Client, err error) {
	u.log.V(4).Info("Building client for management cluster", "kubeconfig", managementClusterKubeconfigPath)
	if err = u.retrier.Retry(func() error {
		managementClusterClient, err = u.clientFactory.BuildClientFromKubeconfig(managementClusterKubeconfigPath)
		return err
	}); err != nil {
		return nil, nil, err
	}

	u.log.V(4).Info("Building client for workload cluster", "kubeconfig", workloadClusterKubeconfigPath)
	if err = u.retrier.Retry(func() error {
		workloadClusterClient, err = u.clientFactory.BuildClientFromKubeconfig(workloadClusterKubeconfigPath)
		return err
	}); err != nil {
		return nil, nil, err
	}

	return managementClusterClient, workloadClusterClient, nil
}

// NewKubeProxyUpgrader builds a new KubeProxyUpgrader.
func NewKubeProxyUpgrader(opts ...KubeProxyUpgraderOpt) KubeProxyUpgrader {
	u := &KubeProxyUpgrader{
		updateKubeProxyRetries: 30,
		updateKubeProxyBackoff: 2 * time.Second,
	}

	for _, opt := range opts {
		opt(u)
	}

	return *u
}

// KubeProxyUpgrader prepares a cluster for a kube-proxy upgrade.
type KubeProxyUpgrader struct {
	updateKubeProxyRetries int
	updateKubeProxyBackoff time.Duration
}

// KubeProxyUpgraderOpt allows to customize a KubeProxyUpgraderOpt
// on construction.
type KubeProxyUpgraderOpt func(*KubeProxyUpgrader)

// WithUpdateKubeProxyTiming allows to customize the retry paramenter for the
// kube-proxy version update. This is for unit tests.
func WithUpdateKubeProxyTiming(retries int, backoff time.Duration) KubeProxyUpgraderOpt {
	return func(u *KubeProxyUpgrader) {
		u.updateKubeProxyRetries = retries
		u.updateKubeProxyBackoff = backoff
	}
}

// PrepareForUpgrade gets the workload cluster ready for a smooth transition between the
// old kube-proxy that always uses iptables legacy and the new one that detects the host preference
// and is able to work with nft as well. This is idempotent, so it can be called in a loop if transient
// errors are a risk.
func (u KubeProxyUpgrader) PrepareForUpgrade(ctx context.Context, log logr.Logger, managementClusterClient, workloadClusterClient kubernetes.Client, spec *cluster.Spec) error {
	kcp, err := getKubeadmControlPlane(ctx, managementClusterClient, spec.Cluster)
	if err != nil {
		return errors.Wrap(err, "reading the kubeadm control plane for an upgrade")
	}

	bundle := spec.ControlPlaneVersionsBundle()

	_, newVersion := oci.Split(bundle.KubeDistro.KubeProxy.URI)

	// If the new spec doesn't include the new kube-proxy or if the current cluster already has it, skip this
	if needsPrepare, err := needsKubeProxyPreUpgrade(spec, kcp); err != nil {
		return err
	} else if !needsPrepare {
		log.V(4).Info("Kube-proxy upgrade doesn't need special handling", "currentVersion", kcp.Spec.Version, "newVersion", newVersion)
		return nil
	}

	log.V(4).Info("Detected upgrade from kube-proxy with iptables legacy upgrade to new version", "currentVersion", kcp.Spec.Version, "newVersion", newVersion)

	// Add the annotation to the kcp so it doesn't undo our changes to the kube-proxy DS
	if err := annotateKCPWithSKipKubeProxy(ctx, log, managementClusterClient, kcp); err != nil {
		return err
	}

	// Add label to nodes so we can use nodeAffinity to control the kube-proxy scheduling
	if err := addIPTablesLegacyLabelToAllNodes(ctx, log, workloadClusterClient); err != nil {
		return err
	}

	originalKubeProxy, err := getKubeProxy(ctx, workloadClusterClient)
	if err != nil {
		return err
	}

	// Make sure original kube-proxy DS is only scheduled in new nodes and it stops running in current nodes.
	if err := restrictKubeProxyToNewNodes(ctx, workloadClusterClient, originalKubeProxy); err != nil {
		return err
	}

	// Once old kube-proxy pods are deleted, create the new DS that will only be scheduled in the old nodes.
	if err := createIPTablesLegacyKubeProxy(ctx, workloadClusterClient, kcp, originalKubeProxy); err != nil {
		return err
	}

	// Finally update the main kube-proxy DS to reflect the new version so all the new nodes
	// get that one scheduled from the beginning.
	log.V(4).Info("Updating kube-proxy DS version", "oldVersion", kcp.Spec.Version, "newVersion", newVersion)
	if err := u.ensureUpdateKubeProxyVersion(ctx, log, workloadClusterClient, spec); err != nil {
		return err
	}

	return nil
}

// CleanupAfterUpgrade cleanups all the leftover changes made by PrepareForUpgrade.
// It's idempotent so it can be call multiple timesm even if PrepareForUpgrade wasn't
// called before.
func (u KubeProxyUpgrader) CleanupAfterUpgrade(ctx context.Context, log logr.Logger, managementClusterClient, workloadClusterClient kubernetes.Client, spec *cluster.Spec) error {
	log.V(4).Info("Deleting iptables legacy kube-proxy", "name", iptablesLegacyKubeProxyDSName)
	if err := deleteIPTablesLegacyKubeProxy(ctx, workloadClusterClient); err != nil {
		return err
	}

	// Remove nodeAffinity from original kube-proxy. It's not strcitly necessary since there
	// won't be more nodes with that label, but it prevents future errors.
	kubeProxy, err := getKubeProxy(ctx, workloadClusterClient)
	if err != nil {
		return err
	}
	if kubeProxy.Spec.Template.Spec.Affinity != nil {
		kubeProxy.Spec.Template.Spec.Affinity = nil
		log.V(4).Info("Removing node-affinity from kube-proxy")
		if err := workloadClusterClient.Update(ctx, kubeProxy); err != nil {
			return errors.Wrap(err, "updating main kube-proxy version to remove nodeAffinity")
		}
	}

	// Remove the skip annotation from the kubeadm control plane so it starts reconciling the kube-proxy again
	kcp, err := getKubeadmControlPlane(ctx, managementClusterClient, spec.Cluster)
	if err != nil {
		return errors.Wrap(err, "reading the kubeadm control plane to cleanup the skip annotations")
	}

	if _, ok := kcp.Annotations[controlplanev1.SkipKubeProxyAnnotation]; !ok {
		return nil
	}

	delete(kcp.Annotations, controlplanev1.SkipKubeProxyAnnotation)
	log.V(4).Info("Removing skip kube-proxy annotation from KubeadmControlPlane")
	if err := managementClusterClient.Update(ctx, kcp); err != nil {
		return errors.Wrap(err, "preparing kcp for kube-proxy upgrade")
	}

	return nil
}

func specIncludesNewKubeProxy(spec *cluster.Spec) bool {
	bundle := spec.ControlPlaneVersionsBundle()
	return eksdIncludesNewKubeProxy(spec.Cluster.Spec.KubernetesVersion, bundle.KubeDistro.EKSD.Number)
}

func eksdIncludesNewKubeProxy(version anywherev1.KubernetesVersion, number int) bool {
	return number >= firstEKSDWithNewKubeProxy[version]
}

var eksDNumberRegex = regexp.MustCompile(`(?m)^.*-eks-(\d)-(\d+)-(\d+)$`)

func eksdVersionAndNumberFromTag(tag string) (anywherev1.KubernetesVersion, int, error) {
	matches := eksDNumberRegex.FindStringSubmatch(tag)
	if len(matches) != 4 {
		return "", 0, errors.Errorf("invalid eksd tag format %s", tag)
	}

	kubeMajor := matches[1]
	kubeMinor := matches[2]

	kubeVersion := anywherev1.KubernetesVersion(kubeMajor + "." + kubeMinor)

	numberStr := matches[3]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return "", 0, errors.Wrapf(err, "invalid number in eksd tag %s", tag)
	}

	return kubeVersion, number, nil
}

func needsKubeProxyPreUpgrade(spec *cluster.Spec, currentKCP *controlplanev1.KubeadmControlPlane) (bool, error) {
	currentKubeVersion, currentEKSDNumber, err := eksdVersionAndNumberFromTag(currentKCP.Spec.Version)
	if err != nil {
		return false, err
	}

	return specIncludesNewKubeProxy(spec) && !eksdIncludesNewKubeProxy(currentKubeVersion, currentEKSDNumber), nil
}

func annotateKCPWithSKipKubeProxy(ctx context.Context, log logr.Logger, c kubernetes.Client, kcp *controlplanev1.KubeadmControlPlane) error {
	log.V(4).Info("Adding skip annotation to kcp", "kcp", klog.KObj(kcp), "annotation", controlplanev1.SkipKubeProxyAnnotation)
	clientutil.AddAnnotation(kcp, controlplanev1.SkipKubeProxyAnnotation, "true")
	if err := c.Update(ctx, kcp); err != nil {
		return errors.Wrap(err, "preparing kcp for kube-proxy upgrade")
	}

	return nil
}

func addIPTablesLegacyLabelToAllNodes(ctx context.Context, log logr.Logger, client kubernetes.Client) error {
	nodeList := &corev1.NodeList{}
	if err := client.List(ctx, nodeList); err != nil {
		return errors.Wrap(err, "listing workload cluster nodes for kube-proxy upgrade")
	}

	nodes := make([]*corev1.Node, 0, len(nodeList.Items))
	for i := range nodeList.Items {
		nodes = append(nodes, &nodeList.Items[i])
	}

	log.V(4).Info("Adding iptables-legacy label to nodes", "nodes", klog.KObjSlice(nodes), "label", iptablesLegacyLabel)
	for i := range nodeList.Items {
		n := &nodeList.Items[i]
		clientutil.AddLabel(n, iptablesLegacyLabel, "true")
		if err := client.Update(ctx, n); err != nil {
			return errors.Wrap(err, "preparing workload cluster nodes for kube-proxy upgrade")
		}
	}

	return nil
}

func getKubeProxy(ctx context.Context, c kubernetes.Client) (*appsv1.DaemonSet, error) {
	kubeProxy := &appsv1.DaemonSet{}
	if err := c.Get(ctx, kubeProxyDSName, kubeProxyDSNamespace, kubeProxy); err != nil {
		return nil, errors.Wrap(err, "reading kube-proxy for upgrade")
	}

	return kubeProxy, nil
}

func getKubeadmControlPlane(ctx context.Context, c kubernetes.Client, cluster *anywherev1.Cluster) (*controlplanev1.KubeadmControlPlane, error) {
	key := controller.CAPIKubeadmControlPlaneKey(cluster)

	kubeadmControlPlane := &controlplanev1.KubeadmControlPlane{}
	if err := c.Get(ctx, key.Name, key.Namespace, kubeadmControlPlane); err != nil {
		return nil, err
	}
	return kubeadmControlPlane, nil
}

func addAntiNodeAffinityToKubeProxy(ctx context.Context, client kubernetes.Client, kubeProxy *appsv1.DaemonSet) error {
	kubeProxy.Spec.Template.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      iptablesLegacyLabel,
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
			},
		},
	}
	if err := client.Update(ctx, kubeProxy); err != nil {
		return errors.Wrap(err, "preparing main kube-proxty for upgrade")
	}

	return nil
}

func deleteAllOriginalKubeProxyPods(ctx context.Context, c kubernetes.Client) error {
	if err := c.DeleteAllOf(ctx, &corev1.Pod{},
		&kubernetes.DeleteAllOfOptions{
			Namespace: kubeProxyDSNamespace,
			HasLabels: map[string]string{
				k8sAppLabel: kubeProxyLabel,
			},
		},
	); err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "deleting kube-proxy pods before upgrade")
	}

	return nil
}

func restrictKubeProxyToNewNodes(ctx context.Context, client kubernetes.Client, kubeProxy *appsv1.DaemonSet) error {
	kubeProxy = kubeProxy.DeepCopy()
	// Add nodeAffinity to kube-proxy so it's not scheduled in new nodes without our label
	if err := addAntiNodeAffinityToKubeProxy(ctx, client, kubeProxy); err != nil {
		return err
	}

	// Delete original kube-proxy pods to ensure there is only one copy of kube-proxy running
	// on each node.
	if err := deleteAllOriginalKubeProxyPods(ctx, client); err != nil {
		return err
	}

	return nil
}

func iptablesLegacyKubeProxyFromCurrentDaemonSet(kcp *controlplanev1.KubeadmControlPlane, kubeProxy *appsv1.DaemonSet) *appsv1.DaemonSet {
	iptablesLegacyKubeProxy := kubeProxy.DeepCopy()

	// Generate a new DS with the old kube-proxy version with nodeAffinity so it only
	// gets scheduled in the old (current) nodes.
	iptablesLegacyKubeProxy.Name = iptablesLegacyKubeProxyDSName
	iptablesLegacyKubeProxy.ObjectMeta.ResourceVersion = ""
	iptablesLegacyKubeProxy.ObjectMeta.UID = ""
	image := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.ImageRepository + "/kube-proxy" +
		":" + kcp.Spec.Version
	iptablesLegacyKubeProxy.Spec.Template.Spec.Containers[0].Image = image
	iptablesLegacyKubeProxy.Spec.Template.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      iptablesLegacyLabel,
								Operator: corev1.NodeSelectorOpExists,
							},
						},
					},
				},
			},
		},
	}

	return iptablesLegacyKubeProxy
}

func createIPTablesLegacyKubeProxy(ctx context.Context, client kubernetes.Client, kcp *controlplanev1.KubeadmControlPlane, originalKubeProxy *appsv1.DaemonSet) error {
	iptablesLegacyKubeProxy := iptablesLegacyKubeProxyFromCurrentDaemonSet(kcp, originalKubeProxy)
	if err := client.Create(ctx, iptablesLegacyKubeProxy); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "creating secondary kube-proxy DS with iptables-legacy for old nodes")
	}

	return nil
}

func deleteIPTablesLegacyKubeProxy(ctx context.Context, client kubernetes.Client) error {
	iptablesLegacyKubeProxy := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      iptablesLegacyKubeProxyDSName,
			Namespace: kubeProxyDSNamespace,
		},
	}

	if err := client.Delete(ctx, iptablesLegacyKubeProxy); err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "deleting secondary kube-proxy DS with iptables-legacy")
	}

	return nil
}

func updateKubeProxyVersion(ctx context.Context, client kubernetes.Client, kubeProxy *appsv1.DaemonSet, image string) error {
	kubeProxy.Spec.Template.Spec.Containers[0].Image = image
	if err := client.Update(ctx, kubeProxy); err != nil {
		return errors.Wrap(err, "updating main kube-proxy version before upgrade")
	}

	return nil
}

func (u KubeProxyUpgrader) ensureUpdateKubeProxyVersion(ctx context.Context, log logr.Logger, client kubernetes.Client, spec *cluster.Spec) error {
	bundle := spec.ControlPlaneVersionsBundle()
	newKubeProxyImage := bundle.KubeDistro.KubeProxy.URI
	return retrier.Retry(u.updateKubeProxyRetries, u.updateKubeProxyBackoff, func() error {
		kubeProxy, err := getKubeProxy(ctx, client)
		if err != nil {
			return err
		}

		currentImage := kubeProxy.Spec.Template.Spec.Containers[0].Image
		if currentImage == newKubeProxyImage {
			log.V(4).Info("Kube-proxy image update seems stable", "wantImage", newKubeProxyImage, "currentImage", currentImage)
			return nil
		}

		log.V(4).Info("Kube-proxy image update has been reverted or was never updated", "wantImage", newKubeProxyImage, "currentImage", currentImage)
		log.V(4).Info("Updating Kube-proxy image", "newImage", newKubeProxyImage)
		if err := updateKubeProxyVersion(ctx, client, kubeProxy, newKubeProxyImage); err != nil {
			return err
		}

		return errors.Errorf("kube-proxy image update has been reverted from %s to %s", newKubeProxyImage, currentImage)
	})
}
