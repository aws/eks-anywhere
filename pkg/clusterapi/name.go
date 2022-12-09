package clusterapi

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
)

var nameRegex = regexp.MustCompile(`(.*?)(-)(\d+)$`)

// Object represents a kubernetes API object.
type Object[O kubernetes.Object] interface {
	kubernetes.Object
	DeepCopy() O
}

// ObjectComparator returns true only if only both kubernetes Object's are identical
// Most of the time, this only requires comparing the Spec field, but that can variate
// from object to object.
type ObjectComparator[O Object[O]] func(current, new O) bool

// ObjectRetriever gets a kubernetes API object using the provided client
// If the object doesn't exist, it returns a NotFound error.
type ObjectRetriever[O Object[O]] func(ctx context.Context, client kubernetes.Client, name, namespace string) (O, error)

// IncrementName takes an object name and increments the suffix number by one.
// This method is used for updating objects (e.g. machinetemplate, kubeadmconfigtemplate) that are either immutable
// or require recreation to trigger machine rollout. The original object name should follow the name convention of
// alphanumeric followed by dash digits, e.g. abc-1, md-0, kct-2. An error will be raised if the original name does not follow
// this pattern.
func IncrementName(name string) (string, error) {
	match := nameRegex.FindStringSubmatch(name)
	if match == nil {
		return "", fmt.Errorf(`invalid format of name [name=%s]. Name has to follow regex pattern "(-)(\d+)$", e.g. machinetemplate-cp-1`, name)
	}

	n, err := strconv.Atoi(match[3])
	if err != nil {
		return "", fmt.Errorf("converting object suffix to int: %v", err)
	}

	return ObjectName(match[1], n+1), nil
}

// IncrementNameWithFallbackDefault calls the IncrementName and fallbacks to use the default name if IncrementName
// returns an error. This method is used to accommodate for any objects with name breaking changes from a previous version.
// For example, in beta capi snowmachinetemplate is named after the eks-a snowmachineconfig name, without the '-1' suffix.
// We set the object name to the default new machinetemplate name after detecting the invalid old name.
func IncrementNameWithFallbackDefault(name, defaultName string) string {
	n, err := IncrementName(name)
	if err != nil {
		logger.V(4).Info("Unable to increment object name (might due to changes of name format), fallback to the default name", "error", err.Error())
		return defaultName
	}
	return n
}

func ObjectName(baseName string, version int) string {
	return fmt.Sprintf("%s-%d", baseName, version)
}

func DefaultObjectName(baseName string) string {
	return ObjectName(baseName, 1)
}

func KubeadmControlPlaneName(clusterSpec *cluster.Spec) string {
	return clusterSpec.Cluster.GetName()
}

func MachineDeploymentName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	// Adding cluster name prefix guarantees the machine deployment name uniqueness
	// among clusters under the same management cluster setting.
	return clusterWorkerNodeGroupName(clusterSpec, workerNodeGroupConfig)
}

func DefaultKubeadmConfigTemplateName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	return DefaultObjectName(clusterWorkerNodeGroupName(clusterSpec, workerNodeGroupConfig))
}

func clusterWorkerNodeGroupName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	return fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfig.Name)
}

// ControlPlaneMachineTemplateName sets the default object name on the control plane machine template.
func ControlPlaneMachineTemplateName(cluster *v1alpha1.Cluster) string {
	return DefaultObjectName(fmt.Sprintf("%s-control-plane", cluster.Name))
}

// EtcdMachineTemplateName sets the default object name on the etcd machine template.
func EtcdMachineTemplateName(cluster *v1alpha1.Cluster) string {
	return DefaultObjectName(fmt.Sprintf("%s-etcd", cluster.Name))
}

func WorkerMachineTemplateName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	return DefaultObjectName(fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfig.Name))
}

func ControlPlaneMachineHealthCheckName(clusterSpec *cluster.Spec) string {
	return fmt.Sprintf("%s-kcp-unhealthy", KubeadmControlPlaneName(clusterSpec))
}

func WorkerMachineHealthCheckName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	return fmt.Sprintf("%s-worker-unhealthy", MachineDeploymentName(clusterSpec, workerNodeGroupConfig))
}

// EnsureNewNameIfChanged updates an object's name if such object is different from its current state in the cluster.
func EnsureNewNameIfChanged[M Object[M]](ctx context.Context,
	client kubernetes.Client,
	retrieve ObjectRetriever[M],
	equal ObjectComparator[M],
	new M,
) error {
	current, err := retrieve(ctx, client, new.GetName(), new.GetNamespace())
	if apierrors.IsNotFound(err) {
		// if object doesn't exist with same name in same namespace, no need to compare, there won't be a conflict
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "reading %s %s/%s from API",
			new.GetObjectKind().GroupVersionKind().Kind,
			new.GetNamespace(),
			new.GetName(),
		)
	}

	if !equal(new, current) {
		newName, err := IncrementName(new.GetName())
		if err != nil {
			return errors.Wrapf(err, "incrementing name for %s %s/%s",
				new.GetObjectKind().GroupVersionKind().Kind,
				new.GetNamespace(),
				new.GetName(),
			)
		}

		new.SetName(newName)
	}

	return nil
}

// ClusterCASecretName returns the name of the cluster CA secret for the cluster.
func ClusterCASecretName(clusterName string) string {
	return fmt.Sprintf("%s-ca", clusterName)
}
