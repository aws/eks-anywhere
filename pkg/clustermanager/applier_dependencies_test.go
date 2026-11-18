package clustermanager_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

type dependencyCheckingClient struct {
	kubernetes.Client
	children []kubernetes.Object
}

func (c dependencyCheckingClient) ApplyServerSide(ctx context.Context, manager string, obj kubernetes.Object, options ...kubernetes.ApplyServerSideOption) error {
	if _, ok := obj.(*anywherev1.Cluster); ok {
		for _, child := range c.children {
			stored := child.DeepCopyObject().(kubernetes.Object)
			if err := c.Get(ctx, child.GetName(), child.GetNamespace(), stored); err != nil {
				return fmt.Errorf("cluster is visible before its referenced config: %w", err)
			}
		}
	}
	if err := c.Client.ApplyServerSide(ctx, manager, obj, options...); err != nil {
		return err
	}
	if _, ok := obj.(*anywherev1.Cluster); ok {
		current := &anywherev1.Cluster{}
		if err := c.Get(ctx, obj.GetName(), obj.GetNamespace(), current); err != nil {
			return err
		}
		current.Status.ObservedGeneration = current.Generation
		for _, conditionType := range []anywherev1.ConditionType{
			anywherev1.ControlPlaneReadyCondition, anywherev1.DefaultCNIConfiguredCondition,
			anywherev1.WorkersReadyCondition, anywherev1.ReadyCondition,
		} {
			current.Status.Conditions = append(current.Status.Conditions, anywherev1.Condition{
				Type: conditionType, Status: corev1.ConditionTrue,
			})
		}
		return c.Update(ctx, current)
	}
	return nil
}

func TestApplierReferencedConfigsExistWhenClusterIsApplied(t *testing.T) {
	for _, existingCluster := range []bool{false, true} {
		t.Run(fmt.Sprintf("existingCluster=%t", existingCluster), func(t *testing.T) {
			tt := newApplierTest(t)
			var existing []kubernetes.Object
			if existingCluster {
				tt.spec.Cluster.ResourceVersion = "999"
				existing = append(existing, tt.spec.Cluster)
			}
			tt.client = test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(existing)...)
			checked := dependencyCheckingClient{
				Client: tt.client, children: tt.spec.ChildObjects(),
			}
			tt.client = checked
			tt.clientFactory.EXPECT().BuildClientFromKubeconfig(tt.mgmtCluster.KubeconfigFile).Return(checked, nil)
			a := clustermanager.NewApplier(tt.log, tt.clientFactory,
				clustermanager.WithApplierApplyClusterTimeout(0),
				clustermanager.WithApplierRetryBackOff(time.Millisecond),
				clustermanager.WithApplierWaitForClusterReconcile(time.Second),
				clustermanager.WithApplierWaitForFailureMessage(0),
			)
			tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(Succeed())
		})
	}
}
