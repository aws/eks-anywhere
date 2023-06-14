package clusters

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/semver"
)

// CleanupStatusAfterValidate removes errors from the cluster status. Intended to be used as a reconciler phase
// after all validation phases have been executed.
func CleanupStatusAfterValidate(_ context.Context, _ logr.Logger, spec *cluster.Spec) (controller.Result, error) {
	spec.Cluster.Status.FailureMessage = nil
	return controller.Result{}, nil
}

// ClusterValidator runs cluster level validations.
type ClusterValidator struct {
	client client.Client
}

// NewClusterValidator returns a validator that will run cluster level validations.
func NewClusterValidator(client client.Client) *ClusterValidator {
	return &ClusterValidator{
		client: client,
	}
}

// ValidateManagementClusterName checks if the management cluster specified in the workload cluster spec is valid.
func (v *ClusterValidator) ValidateManagementClusterName(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	mgmtCluster := &anywherev1.Cluster{}
	mgmtClusterKey := client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cluster.Spec.ManagementCluster.Name,
	}
	if err := v.client.Get(ctx, mgmtClusterKey, mgmtCluster); err != nil {
		if apierrors.IsNotFound(err) {
			err := fmt.Errorf("unable to retrieve management cluster %v: %v", cluster.Spec.ManagementCluster.Name, err)
			log.Error(err, "Invalid cluster configuration")
			return err
		}
	}
	if mgmtCluster.IsManaged() {
		err := fmt.Errorf("%s is not a valid management cluster", mgmtCluster.Name)
		log.Error(err, "Invalid cluster configuration")
		return err
	}

	return nil
}

// ValidateEksaVersionExists checks if the cluster's eksa version exists.
func (v *ClusterValidator) ValidateEksaVersionExists(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	reader := files.NewReader()
	mReader := manifests.NewReader(reader)

	if _, err := mReader.ReadReleaseForVersion(string(*cluster.Spec.EksaVersion)); err != nil {
		err := fmt.Errorf("cluster eksaVersion is not valid: %v", cluster.Spec.EksaVersion)
		log.Error(err, "Invalid cluster configuration")
		return err
	}

	return nil
}

// ValidateManagementClusterEksaVersion checks if the workload cluster's eksa version is valid compared to the management cluster.
func (v *ClusterValidator) ValidateManagementClusterEksaVersion(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	mgmtCluster := &anywherev1.Cluster{}
	mgmtClusterKey := client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cluster.Spec.ManagementCluster.Name,
	}
	if err := v.client.Get(ctx, mgmtClusterKey, mgmtCluster); err != nil {
		if apierrors.IsNotFound(err) {
			err := fmt.Errorf("unable to retrieve management cluster %v: %v", cluster.Spec.ManagementCluster.Name, err)
			log.Error(err, "Invalid cluster configuration")
			return err
		}
	}

	mVersion, err := semver.New(string(*mgmtCluster.Spec.EksaVersion))
	if err != nil {
		err := fmt.Errorf("management cluster eksaVersion is invalid: %w", err)
		log.Error(err, "Invalid management cluster configuration")
		return err
	}

	wVersion, err := semver.New(string(*cluster.Spec.EksaVersion))
	if err != nil {
		err := fmt.Errorf("workload cluster eksaVersion is invalid: %w", err)
		log.Error(err, "Invalid cluster configuration")
		return err
	}

	if wVersion.GreaterThan(mVersion) {
		err := fmt.Errorf("Cannot upgrade workload cluster with version %v while management cluster is an older version %v", wVersion, mVersion)
		log.Error(err, "Invalid cluster configuration")
		return err
	}

	return nil
}

func (v *ClusterValidator) ValidateManagementWorkloadSkew(ctx context.Context, log logr.Logger, mgmtCluster *anywherev1.Cluster) error {
	workloads := &anywherev1.ClusterList{}
	if err := v.client.List(ctx, workloads); err != nil {
		if apierrors.IsNotFound(err) {
			err := fmt.Errorf("unable to retrieve clusters: %v", err)
			log.Error(err, "Invalid cluster configuration")
			return err
		}
	}

	mVersion, err := semver.New(string(*mgmtCluster.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("management cluster eksaVersion is invalid: %w", err)
	}

	for _, w := range workloads.Items {
		if w.Spec.ManagementCluster.Name != mgmtCluster.Name || w.Name == mgmtCluster.Name {
			continue
		}

		if w.Spec.EksaVersion == nil {
			return fmt.Errorf("workload cluster eksaVersion cannot be nil: %v", w)
		}

		wVersion, err := semver.New(string(*w.Spec.EksaVersion))
		if err != nil {
			return fmt.Errorf("workload cluster eksaVersion is invalid: %v", w)
		}

		minorSkew := mVersion.Minor - wVersion.Minor

		if !mVersion.SameMajor(wVersion) || minorSkew > 1 {
			return fmt.Errorf("Cannot upgrade management cluster to %v. There can only be a skew of one eksa minor version against workload cluster %s: %v", mVersion, w.Name, wVersion)
		}
	}

	return nil
}
