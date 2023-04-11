package framework

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/storage/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

const (
	csiDeployment              = "vsphere-csi-controller"
	csiDaemonSet               = "vsphere-csi-node"
	csiStorageClassName        = "standard"
	csiStorageClassProvisioner = "csi.vsphere.vmware.com"
	kubeSystemNameSpace        = "kube-system"
)

const defaultMaxRetries = 5

// ValidateVSphereCSI checks whether vsphere csi exists as expected or not.
func (e *ClusterE2ETest) ValidateVSphereCSI(installed bool) {
	ctx := context.Background()
	maxRetries := defaultMaxRetries
	if !installed {
		maxRetries = 1
	}
	err := e.getDeployment(ctx, maxRetries)
	if err != nil {
		handleError(e.T, installed, err)
	}
	err = e.getDaemonSet(ctx, maxRetries)
	if err != nil {
		handleError(e.T, installed, err)
	}
	storageclass, err := e.getStorageClass(ctx, maxRetries)
	if err != nil {
		handleError(e.T, installed, err)
	}
	if installed && storageclass.Provisioner != csiStorageClassProvisioner {
		e.T.Fatalf("provisioners don't match. got: %v, want: %v", storageclass.Provisioner, csiStorageClassProvisioner)
	}
	err = e.getClusterResourceSet(ctx, maxRetries)
	if err != nil {
		handleError(e.T, installed, err)
	}
}

func handleError(t T, installed bool, err error) {
	if installed || !strings.Contains(err.Error(), "not found") {
		t.Fatal(err)
	}
}

// DeleteVSphereCSI removes the vsphere csi from the cluster.
func (e *ClusterE2ETest) DeleteVSphereCSI() {
	ctx := context.Background()
	e.deleteVsphereCSIResources(ctx)
	csiClusterResourceSetName := fmt.Sprintf("%s-csi", e.ClusterName)
	opts := &kubernetes.KubectlDeleteOptions{
		Name:      csiClusterResourceSetName,
		Namespace: constants.EksaSystemNamespace,
	}
	err := e.KubectlClient.Delete(ctx, "clusterresourceset", e.Cluster().KubeconfigFile, opts)
	if err != nil {
		e.T.Fatal(err)
	}
}

// DeleteWorkloadVsphereCSI removes the vsphere CSI from a workload cluster.
func (w *WorkloadCluster) DeleteWorkloadVsphereCSI() {
	ctx := context.Background()
	w.deleteVsphereCSIResources(ctx)
}

func (e *ClusterE2ETest) deleteVsphereCSIResources(ctx context.Context) {
	opts := &kubernetes.KubectlDeleteOptions{
		Name:      csiDeployment,
		Namespace: kubeSystemNameSpace,
	}
	err := e.KubectlClient.Delete(ctx, "deployment", e.Cluster().KubeconfigFile, opts)
	if err != nil {
		e.T.Fatal(err)
	}

	opts = &kubernetes.KubectlDeleteOptions{
		Name:      csiDaemonSet,
		Namespace: kubeSystemNameSpace,
	}
	err = e.KubectlClient.Delete(ctx, "daemonset", e.Cluster().KubeconfigFile, opts)
	if err != nil {
		e.T.Fatal(err)
	}

	err = e.KubectlClient.DeleteClusterObject(ctx, "storageclass", csiStorageClassName, e.Cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}
}

func (e *ClusterE2ETest) getDeployment(ctx context.Context, retries int) error {
	return retrier.Retry(retries, time.Second*5, func() error {
		_, err := e.KubectlClient.GetDeployment(ctx, csiDeployment, kubeSystemNameSpace, e.Cluster().KubeconfigFile)
		if err != nil {
			return err
		}
		return nil
	})
}

func (e *ClusterE2ETest) getDaemonSet(ctx context.Context, retries int) error {
	return retrier.Retry(retries, time.Second*5, func() error {
		_, err := e.KubectlClient.GetDaemonSet(ctx, csiDaemonSet, kubeSystemNameSpace, e.Cluster().KubeconfigFile)
		if err != nil {
			return err
		}
		return nil
	})
}

func (e *ClusterE2ETest) getStorageClass(ctx context.Context, retries int) (*v1.StorageClass, error) {
	var storageclass *v1.StorageClass
	err := retrier.Retry(retries, time.Second*5, func() error {
		s, err := e.KubectlClient.GetStorageClass(ctx, csiStorageClassName, e.Cluster().KubeconfigFile)
		if err != nil {
			return err
		}
		storageclass = s
		return nil
	})
	return storageclass, err
}

func (e *ClusterE2ETest) getClusterResourceSet(ctx context.Context, retries int) error {
	return retrier.Retry(retries, time.Second*5, func() error {
		csiClusterResourceSetName := fmt.Sprintf("%s-csi", e.ClusterName)
		_, err := e.KubectlClient.GetClusterResourceSet(ctx, e.Cluster().KubeconfigFile, csiClusterResourceSetName, constants.EksaSystemNamespace)
		if err != nil {
			return err
		}
		return nil
	})
}
