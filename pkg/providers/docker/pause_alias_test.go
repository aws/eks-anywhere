package docker

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestPauseImageAliasUsesKubeadmVersion(t *testing.T) {
	for _, tt := range []struct {
		name    string
		version string
		source  string
		alias   string
	}{
		{"137", "v1.37.1-example", "registry.example.test/kubernetes/pause:v1.37.1-example", "registry.example.test/kubernetes/pause:3.10.2"},
		{"older", "v1.31.4-example", "registry.example.test/kubernetes/pause:v1.31.4-example", "registry.example.test/kubernetes/pause:3.10"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir, env := pauseAliasTestEnvironment(t, tt.alias, true)
			command := pauseImageAliasCommand(tt.version, "registry.example.test/kubernetes", tt.source)
			run := exec.Command("sh", "-c", command)
			run.Env = env
			if output, err := run.CombinedOutput(); err != nil {
				t.Fatalf("alias command failed: %v, %s", err, output)
			}
			want := []string{"config", "images", "list", "--kubernetes-version", tt.version, "--image-repository", "registry.example.test/kubernetes"}
			if got := readArguments(t, filepath.Join(dir, "kubeadm-args")); !reflect.DeepEqual(got, want) {
				t.Fatalf("kubeadm args %q, want %q", got, want)
			}
			want = []string{"--namespace", "k8s.io", "images", "tag", "--force", tt.source, tt.alias}
			if got := readArguments(t, filepath.Join(dir, "ctr-args")); !reflect.DeepEqual(got, want) {
				t.Fatalf("ctr args %q, want %q", got, want)
			}
			if data, err := os.ReadFile(filepath.Join(dir, "alias")); err != nil || string(data) != "cached-image-digest" {
				t.Fatalf("alias does not reference the cached image: %v", err)
			}
		})
	}
}

func TestPauseImageAliasMissingCacheFails(t *testing.T) {
	dir, env := pauseAliasTestEnvironment(t, "registry.example.test/kubernetes/pause:3.10.2", false)
	run := exec.Command("sh", "-c", pauseImageAliasCommand("v1.37.1-example", "registry.example.test/kubernetes", "registry.example.test/pause:v1.37.1-example"))
	run.Env = env
	if output, err := run.CombinedOutput(); err == nil {
		t.Fatalf("missing cached image must fail: %s", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "alias")); !os.IsNotExist(err) {
		t.Fatal("missing cached image produced a pause alias")
	}
}

func TestPauseImageAliasMissingKubeadmImageFails(t *testing.T) {
	_, env := pauseAliasTestEnvironment(t, "", true)
	run := exec.Command("sh", "-c", pauseImageAliasCommand("v1.37.1-example", "registry.example.test/kubernetes", "registry.example.test/pause:v1.37.1-example"))
	run.Env = env
	if output, err := run.CombinedOutput(); err == nil {
		t.Fatalf("kubeadm output without pause image must fail: %s", output)
	}
}

func TestPauseImageAliasMissingBundleImage(t *testing.T) {
	values := map[string]interface{}{}
	err := populatePauseImageAlias(&cluster.VersionsBundle{KubeDistro: &cluster.KubeDistro{}}, values)
	if err == nil {
		t.Fatal("missing bundled pause image must fail")
	}
	if _, ok := values["pauseImageAliasCommand"]; ok {
		t.Fatal("missing bundled image produced an alias command")
	}
}

func pauseAliasTestEnvironment(t *testing.T, expectedImage string, cached bool) (string, []string) {
	t.Helper()
	dir := t.TempDir()
	if cached {
		if err := os.WriteFile(filepath.Join(dir, "cache"), []byte("cached-image-digest"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	scripts := map[string]string{
		"kubeadm": "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$KUBEADM_ARGS\"\nprintf '%s\\n' \"$EXPECTED_PAUSE_IMAGE\"\n",
		"ctr":     "#!/bin/sh\n[ -f \"$CACHE_FILE\" ] || exit 1\nprintf '%s\\n' \"$@\" > \"$CTR_ARGS\"\ncp \"$CACHE_FILE\" \"$ALIAS_FILE\"\n",
	}
	for name, script := range scripts {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	env := append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"),
		"KUBEADM_ARGS="+filepath.Join(dir, "kubeadm-args"), "CTR_ARGS="+filepath.Join(dir, "ctr-args"),
		"EXPECTED_PAUSE_IMAGE="+expectedImage, "CACHE_FILE="+filepath.Join(dir, "cache"),
		"ALIAS_FILE="+filepath.Join(dir, "alias"))
	return dir, env
}

func readArguments(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}
