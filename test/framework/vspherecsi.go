package framework

import (
	"context"
	"strings"
	"testing"
)

const (
	csiDeployment              = "vsphere-csi-controller"
	csiDaemonSet               = "vsphere-csi-node"
	csiStorageClassName        = "standard"
	csiStorageClassProvisioner = "csi.vsphere.vmware.com"
	kubeSystemNameSpace        = "kube-system"
)

// ValidateVSphereCSI checks whether vsphere csi exists as expected or not.
func (e *ClusterE2ETest) ValidateVSphereCSI(installed bool) {
	ctx := context.Background()
	_, err := e.KubectlClient.GetDeployment(ctx, csiDeployment, kubeSystemNameSpace, e.cluster().KubeconfigFile)
	if err != nil {
		handleError(e.T, installed, err)
	}
	_, err = e.KubectlClient.GetDaemonSet(ctx, csiDaemonSet, kubeSystemNameSpace, e.cluster().KubeconfigFile)
	if err != nil {
		handleError(e.T, installed, err)
	}
	storageclass, err := e.KubectlClient.GetStorageClass(ctx, csiStorageClassName, e.cluster().KubeconfigFile)
	if err != nil {
		handleError(e.T, installed, err)
	}
	if installed && storageclass.Provisioner != csiStorageClassProvisioner {
		e.T.Fatalf("provisioners don't match. got: %v, want: %v", storageclass.Provisioner, csiStorageClassProvisioner)
	}
}

func handleError(t *testing.T, installed bool, err error) {
	if installed || !strings.Contains(err.Error(), "not found") {
		t.Fatal(err)
	}
}

// DeleteVSphereCSI removes the vsphere csi from the cluster.
func (e *ClusterE2ETest) DeleteVSphereCSI() {
	ctx := context.Background()
	err := e.KubectlClient.Delete(ctx, "deployment", csiDeployment, kubeSystemNameSpace, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}
	err = e.KubectlClient.Delete(ctx, "daemonset", csiDaemonSet, kubeSystemNameSpace, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}
	err = e.KubectlClient.DeleteClusterObject(ctx, "storageclass", csiStorageClassName, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}
}
