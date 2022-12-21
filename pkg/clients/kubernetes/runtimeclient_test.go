package kubernetes_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

func TestNewRuntimeClient(t *testing.T) {
	g := NewWithT(t)
	cfg := test.UseEnvTest(t)
	rc := kubernetes.RestConfigurator(func(_ []byte) (*rest.Config, error) { return cfg, nil })
	c, err := kubernetes.NewRuntimeClient([]byte{}, rc, runtime.NewScheme())
	g.Expect(err).To(BeNil())

	ctx := context.Background()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "name",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx",
					Name:  "nginx",
				},
			},
		},
	}
	err = c.Create(ctx, pod)
	g.Expect(err).To(BeNil())
}

func TestNewRuntimeClientInvalidRestConfig(t *testing.T) {
	g := NewWithT(t)
	rc := kubernetes.RestConfigurator(func(_ []byte) (*rest.Config, error) { return nil, errors.New("failed to build rest.Config") })
	_, err := kubernetes.NewRuntimeClient([]byte{}, rc, runtime.NewScheme())
	g.Expect(err).To(MatchError(ContainSubstring("failed to build rest.Config")))
}

func TestNewRuntimeClientInvalidScheme(t *testing.T) {
	g := NewWithT(t)
	cfg := test.UseEnvTest(t)
	rc := kubernetes.RestConfigurator(func(_ []byte) (*rest.Config, error) { return cfg, nil })
	_, err := kubernetes.NewRuntimeClient([]byte{}, rc, nil)
	g.Expect(err).To(MatchError(ContainSubstring("scheme was not provided")))
}

func TestNewRuntimeClientFromFilename(t *testing.T) {
	g := NewWithT(t)
	_, err := kubernetes.NewRuntimeClientFromFileName("file-does-not-exist.txt")
	g.Expect(err).To(MatchError(ContainSubstring("open file-does-not-exist.txt: no such file or directory")))
}
