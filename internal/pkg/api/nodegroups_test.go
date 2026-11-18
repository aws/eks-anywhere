package api_test

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestWithTaintRemainsValidAfterGitConfigRetries(t *testing.T) {
	g := NewWithT(t)
	cluster := &anywherev1.Cluster{}
	unrelated := corev1.Taint{Key: "unrelated", Value: "value2", Effect: corev1.TaintEffectNoSchedule}
	otherEffect := corev1.Taint{Key: "key1", Value: "value1", Effect: corev1.TaintEffectPreferNoSchedule}
	api.WithWorkerNodeGroup("worker-0", api.WithCount(2), api.WithTaint(unrelated))(cluster)
	api.WithWorkerNodeGroup("worker-1", api.WithCount(1), api.WithTaint(otherEffect))(cluster)
	taint := corev1.Taint{Key: "key1", Value: "value1", Effect: corev1.TaintEffectNoExecute}
	update := api.WithWorkerNodeGroup("worker-1", api.WithTaint(taint))
	for range 9 {
		// Git retries reload the modified manifest before applying the fillers again.
		data, err := yaml.Marshal(cluster)
		g.Expect(err).NotTo(HaveOccurred())
		cluster = &anywherev1.Cluster{}
		g.Expect(yaml.Unmarshal(data, cluster)).To(Succeed())
		update(cluster)
	}
	worker := cluster.Spec.WorkerNodeGroupConfigurations[1]
	node := corev1.Node{Spec: corev1.NodeSpec{Taints: []corev1.Taint{otherEffect, taint}}}
	g.Expect(api.ValidateWorkerNodeTaints(worker, node)).To(Succeed())
	g.Expect(worker.Taints).To(Equal([]corev1.Taint{otherEffect, taint}))
	g.Expect(cluster.Spec.WorkerNodeGroupConfigurations[0].Taints).To(Equal([]corev1.Taint{unrelated}))
	g.Expect(*cluster.Spec.WorkerNodeGroupConfigurations[0].Count).To(Equal(2))
}

func TestWithTaintPreservesOtherValuesAndEffects(t *testing.T) {
	g := NewWithT(t)
	worker := &anywherev1.WorkerNodeGroupConfiguration{}
	api.WithTaint(corev1.Taint{Key: "key1", Value: "old", Effect: corev1.TaintEffectNoExecute})(worker)
	api.WithTaint(corev1.Taint{Key: "key1", Value: "another", Effect: corev1.TaintEffectNoSchedule})(worker)
	replacement := corev1.Taint{Key: "key1", Value: "new", Effect: corev1.TaintEffectNoExecute}
	api.WithTaint(replacement)(worker)
	g.Expect(worker.Taints).To(Equal([]corev1.Taint{
		{Key: "key1", Value: "old", Effect: corev1.TaintEffectNoExecute},
		{Key: "key1", Value: "another", Effect: corev1.TaintEffectNoSchedule},
		replacement,
	}))
}
