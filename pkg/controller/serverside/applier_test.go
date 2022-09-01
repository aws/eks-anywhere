package serverside_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
)

func TestObjectApplierApplySuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	namespace := env.CreateNamespaceForTest(ctx, t)
	generator := generator(namespace)

	a := serverside.NewObjectApplier(env.Client())
	result, err := a.Apply(ctx, generator)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestObjectApplierApplyErrorFromGenerator(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	a := serverside.NewObjectApplier(env.Client())
	_, err := a.Apply(ctx, func() ([]kubernetes.Object, error) {
		return nil, errors.New("failed generating")
	})
	g.Expect(err).To(MatchError(ContainSubstring("failed generating")))
}

func TestObjectApplierApplyErrorApplying(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	a := serverside.NewObjectApplier(env.Client())
	_, err := a.Apply(ctx, func() ([]kubernetes.Object, error) {
		// this is an invalid object
		return []kubernetes.Object{&corev1.ConfigMap{}}, nil
	})
	g.Expect(err).To(MatchError(ContainSubstring("resource name may not be empty")))
}

func generator(namespace string) serverside.ObjectGenerator {
	return func() ([]kubernetes.Object, error) {
		return []kubernetes.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cm-1",
					Namespace: namespace,
				},
			},
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cm-1",
					Namespace: namespace,
				},
			},
		}, nil
	}
}
