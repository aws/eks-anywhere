package kubeconfig_test

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

var kindTypo = []byte(`apiVersion: v1\nkind: Conf`)

var goodKubeconfig = []byte(`
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:38471
  name: test
contexts:
- context:
    cluster: test
    user: test-admin
  name: test-admin@test
current-context: test-admin@test
kind: Config
preferences: {}
users:
- name: test-admin
  user:
    client-certificate-data: test
    client-key-data: test
`)

func TestFromEnvironment(t *testing.T) {
	t.Run("returns the filename from the env var", func(t *testing.T) {
		expected := "file one"
		t.Setenv("KUBECONFIG", expected)
		got := kubeconfig.FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("works with longer paths", func(t *testing.T) {
		expected := "/home/user/some/long/path/or file/directory structure/config"
		t.Setenv("KUBECONFIG", expected)
		got := kubeconfig.FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("returns the first file in a list", func(t *testing.T) {
		expected := "file one"
		t.Setenv("KUBECONFIG", expected+":filetwo")
		got := kubeconfig.FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("returns an empty string if no files are found", func(t *testing.T) {
		expected := ""
		t.Setenv("KUBECONFIG", "")
		got := kubeconfig.FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("trims whitespace, so as not to return 'empty' filenames", func(t *testing.T) {
		expected := ""
		t.Setenv("KUBECONFIG", " ")
		got := kubeconfig.FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}

		expected = ""
		t.Setenv("KUBECONFIG", "\t")
		got = kubeconfig.FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}

func TestValidateFilename(t *testing.T) {
	t.Run("reports errors from validator", func(t *testing.T) {
		badFile := test.WithFakeFileContents(t, bytes.NewReader(kindTypo))

		assert.Error(t, kubeconfig.ValidateFilename(badFile.Name()))
	})

	t.Run("reports errors for files that are empty", func(t *testing.T) {
		emptyFile := test.WithFakeFileContents(t, bytes.NewReader([]byte("")))

		assert.Error(t, kubeconfig.ValidateFilename(emptyFile.Name()))
	})

	t.Run("reports errors for files that don't exist", func(t *testing.T) {
		doesntExist := filepath.Join(t.TempDir(), "does-not-exist")
		test.RemoveFileIfExists(t, doesntExist)

		assert.ErrorIs(t, kubeconfig.ValidateFilename(doesntExist), fs.ErrNotExist)
	})

	t.Run("returns nil when valid", func(t *testing.T) {
		goodFile := test.WithFakeFileContents(t, bytes.NewReader(goodKubeconfig))

		assert.NoError(t, kubeconfig.ValidateFilename(goodFile.Name()))
	})

	t.Run("trims whitespace around a filename", func(t *testing.T) {
		goodFile := test.WithFakeFileContents(t, bytes.NewReader(goodKubeconfig))

		assert.NoError(t, kubeconfig.ValidateFilename("   "+goodFile.Name()+"\t\n"))
	})

	t.Run("reports errors for filenames that are the empty string", func(t *testing.T) {
		assert.Error(t, kubeconfig.ValidateFilename(""))
	})

	t.Run("reports errors for filenames that are only whitespace (as if it were the empty string)", func(t *testing.T) {
		assert.Error(t, kubeconfig.ValidateFilename("   \t \n"))
	})
}

func TestResolveFilename(t *testing.T) {
	t.Run("returns the flag value when provided", func(t *testing.T) {
		expected := "flag-provided-kubeconfig"
		filename := kubeconfig.ResolveFilename(expected, "cluster-name")

		assert.Equal(t, expected, filename)
	})

	t.Run("returns the cluster-name based filename when provided, and the flag value is empty", func(t *testing.T) {
		clusterName := "cluster-name"
		expected := kubeconfig.FromClusterName(clusterName)

		assert.Equal(t, expected, kubeconfig.ResolveFilename("", clusterName))
	})

	t.Run("returns the environment value if neither the flag or cluster-name values are provided", func(t *testing.T) {
		expected := "some-value"
		t.Setenv("KUBECONFIG", expected)

		assert.Equal(t, expected, kubeconfig.ResolveFilename("", ""))
	})
}

func TestResolveAndValidateFilename(t *testing.T) {
	t.Run("returns flag value when valid", func(t *testing.T) {
		goodFile := test.WithFakeFileContents(t, bytes.NewReader(goodKubeconfig))
		filename, err := kubeconfig.ResolveAndValidateFilename(goodFile.Name(), "")

		if assert.NoError(t, err) {
			assert.Equal(t, goodFile.Name(), filename)
		}
	})

	t.Run("returns error when invalid", func(t *testing.T) {
		goodFile := test.WithFakeFileContents(t, bytes.NewBufferString("lakjdf"))
		_, err := kubeconfig.ResolveAndValidateFilename(goodFile.Name(), "")

		assert.Error(t, err)
	})

	t.Run("returns an error if no kubeconfig is found", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "")
		_, err := kubeconfig.ResolveAndValidateFilename("", "")

		assert.Error(t, err)
	})

	t.Run("golden path", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "")
		_, err := kubeconfig.ResolveAndValidateFilename("", "")

		assert.Error(t, err)
	})
}
