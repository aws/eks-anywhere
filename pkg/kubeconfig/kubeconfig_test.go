package kubeconfig

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var kindTypo = []byte(`apiVersion: v1\nkind: Conf`)

var goodKubeconfigFile = []byte(`
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

func TestValidateFile(t *testing.T) {
	t.Run("reports errors from validator", func(t *testing.T) {
		badFile := withFakeFileContents(t, bytes.NewReader(kindTypo))
		if err := ValidateFile(badFile.Name()); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("reports errors for files that are empty", func(t *testing.T) {
		emptyFile := withFakeFileContents(t, bytes.NewReader([]byte("")))
		if err := ValidateFile(emptyFile.Name()); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("reports errors for files that don't exist", func(t *testing.T) {
		doesntExist := filepath.Join(t.TempDir(), "does-not-exist")
		err := ValidateFile(doesntExist)
		if err == nil || !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("expected fs.IsNotExist, got %s", err)
		}
	})

	t.Run("returns nil when valid", func(t *testing.T) {
		goodFile := withFakeFileContents(t, bytes.NewReader(goodKubeconfigFile))
		if err := ValidateFile(goodFile.Name()); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
	})
}

// withFakeFile returns a throwaway file in a test-specific directory.
//
// The file is automatically closed and removed when the test ends.
func withFakeFile(t *testing.T) (f *os.File) {
	f, err := ioutil.TempFile(t.TempDir(), "fake-file")
	if err != nil {
		t.Fatalf("opening temp file: %s", err)
	}

	t.Cleanup(func() {
		f.Close()
	})

	return f
}

// withFakeFileContents returns a file containing some data.
//
// The file is automatically closed and removed when the test ends.
func withFakeFileContents(t *testing.T, r io.Reader) (f *os.File) {
	f = withFakeFile(t)
	_, err := io.Copy(f, r)
	if err != nil {
		t.Fatalf("copying contents into fake file %q: %s", f.Name(), err)
	}

	return f
}

func TestFromEnvironment(t *testing.T) {
	t.Run("returns the filename from the env var", func(t *testing.T) {
		expected := "file one"
		t.Setenv("KUBECONFIG", expected)
		got := FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("works with longer paths", func(t *testing.T) {
		expected := "/home/user/some/long/path/or file/directory structure/config"
		t.Setenv("KUBECONFIG", expected)
		got := FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("returns the first file in a list", func(t *testing.T) {
		expected := "file one"
		t.Setenv("KUBECONFIG", expected+":filetwo")
		got := FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("returns an empty string if no files are found", func(t *testing.T) {
		expected := ""
		t.Setenv("KUBECONFIG", "")
		got := FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("trims whitespace, so as not to return 'empty' filenames", func(t *testing.T) {
		expected := ""
		t.Setenv("KUBECONFIG", " ")
		got := FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}

		expected = ""
		t.Setenv("KUBECONFIG", "\t")
		got = FromEnvironment()
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}
