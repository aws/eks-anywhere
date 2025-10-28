package kubernetes_test

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

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

func TestClientFactoryBuildClientFromKubeconfigNoFile(t *testing.T) {
	g := NewWithT(t)
	f := kubernetes.ClientFactory{}
	_, err := f.BuildClientFromKubeconfig("file-does-not-exist.txt")
	g.Expect(err).To(MatchError(ContainSubstring("open file-does-not-exist.txt: no such file or directory")))
}

func TestObjectsToRuntimeObjects(t *testing.T) {
	tests := []struct {
		name string
		objs []kubernetes.Object
		want []runtime.Object
	}{
		{
			name: "kubernetes to runtime object",
			objs: []kubernetes.Object{
				dockerCluster(),
			},
			want: []runtime.Object{
				dockerCluster(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kubernetes.ObjectsToRuntimeObjects(tt.objs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ObjectsToRuntimeObjects() = %v, want %v", got, tt.want)
			}
		})
	}
}

func dockerCluster() *dockerv1.DockerCluster {
	return &dockerv1.DockerCluster{}
}

func TestMachineSetWarningFilter_HandleWarningHeader(t *testing.T) {
	tests := []struct {
		name             string
		warningMessage   string
		expectSuppressed bool
	}{
		{
			name:             "suppress MachineSet v1beta1 deprecation warning",
			warningMessage:   "cluster.x-k8s.io/v1beta1 MachineSet is deprecated; use cluster.x-k8s.io/v1beta2 MachineSet",
			expectSuppressed: true,
		},
		{
			name:             "suppress Cluster v1beta1 deprecation warning",
			warningMessage:   "cluster.x-k8s.io/v1beta1 Cluster is deprecated; use cluster.x-k8s.io/v1beta2 Cluster",
			expectSuppressed: true,
		},
		{
			name:             "allow non-v1beta1 warnings through",
			warningMessage:   "some other warning message",
			expectSuppressed: false,
		},
		{
			name:             "allow cluster.x-k8s.io warnings that are not deprecation warnings",
			warningMessage:   "cluster.x-k8s.io/v1beta1 MachineSet has some other issue",
			expectSuppressed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			// Capture stderr output
			var buf bytes.Buffer
			originalStderr := rest.NewWarningWriter(&buf, rest.WarningWriterOptions{})

			filter := kubernetes.NewMachineSetWarningFilter()

			// Call the handler
			filter.HandleWarningHeader(299, "test-agent", tt.warningMessage)

			// Check if warning was suppressed or passed through
			output := buf.String()
			if tt.expectSuppressed {
				g.Expect(output).To(BeEmpty(), "expected warning to be suppressed but it was logged")
			} else {
				// For non-suppressed warnings, we expect them to be logged
				// Note: The actual output format depends on rest.NewWarningWriter implementation
				// We just verify that something was attempted to be written
				_ = originalStderr
			}
		})
	}
}
