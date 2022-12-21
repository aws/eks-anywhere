package curatedpackages

import (
	"bytes"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/aws/eks-anywhere/internal/test"
)

func TestNewKubeClient(t *testing.T) {
	t.Parallel()
	cfg := test.UseEnvTest(t)
	rc := restConfigurator(func(_ []byte) (*rest.Config, error) { return cfg, nil })
	f := test.WithFakeFileContents(t, bytes.NewBufferString(""))

	t.Run("golden path", func(t *testing.T) {
		_, err := newKubeClient(f.Name(), rc)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
