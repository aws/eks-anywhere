package artifacts_test

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts/mocks"
	"github.com/aws/eks-anywhere/pkg/docker"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type archiveCache struct {
	ids    map[string]string
	loads  int
	pushed map[string]int
	lock   sync.Mutex
}

func (c *archiveCache) ImageID(_ context.Context, image string) (string, error) {
	return c.ids[image], nil
}

func (c *archiveCache) LoadFromFile(context.Context, string) error {
	c.loads++
	return errors.New("unexpected duplicate Docker load: no space left on device")
}

func (c *archiveCache) TagImage(context.Context, string, string) error { return nil }

func (c *archiveCache) PushImage(_ context.Context, image, _ string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.pushed[image]++
	return nil
}

func TestImportSeparateToolsAndBundleArchivesReuseVerifiedCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	charts := mocks.NewMockChartImporter(ctrl)
	files := mocks.NewMockFileImporter(ctrl)
	unpack := mocks.NewMockUnPackager(ctrl)
	ctx := context.Background()
	folder := t.TempDir()
	tools := releasev1.Image{URI: "example.test/tools:v1"}
	component := releasev1.Image{URI: "example.test/component:v1"}
	bundle := &releasev1.Bundles{Spec: releasev1.BundlesSpec{
		VersionsBundles: []releasev1.VersionsBundle{{Eksa: releasev1.EksaBundle{CliTools: tools}}},
	}}
	cache := &archiveCache{
		ids: map[string]string{
			tools.VersionedImage():     writeDockerArchive(t, filepath.Join(folder, "tools.tar"), tools.VersionedImage()),
			component.VersionedImage(): writeDockerArchive(t, filepath.Join(folder, "images.tar"), component.VersionedImage()),
		},
		pushed: make(map[string]int),
	}
	destination := docker.NewRegistryDestination(cache, "mirror.example.test")
	unpack.EXPECT().UnPackage("download.tar", folder).Return(nil)
	first := artifacts.ImportToolsImage{
		Bundles: bundle, InputFile: "download.tar", TmpArtifactsFolder: folder, UnPackager: unpack,
		ImageMover: docker.NewImageMover(docker.NewDiskSource(cache, filepath.Join(folder, "tools.tar")), destination),
	}
	if err := first.Run(ctx); err != nil {
		t.Fatal(err)
	}
	reader.EXPECT().ReadImagesFromBundles(ctx, bundle).Return([]releasev1.Image{tools, component}, nil)
	reader.EXPECT().ReadChartsFromBundles(ctx, bundle).Return(nil)
	charts.EXPECT().Import(ctx).Return(nil)
	files.EXPECT().Push(ctx, bundle)
	remaining := artifacts.Import{
		Reader: reader, Bundles: bundle, ChartImporter: charts, FileImporter: files, TmpArtifactsFolder: folder,
		ImageMover: docker.NewImageMover(docker.NewDiskSource(cache, filepath.Join(folder, "images.tar")), destination),
	}
	if err := remaining.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if cache.loads != 0 || cache.pushed[tools.VersionedImage()] != 1 || cache.pushed[component.VersionedImage()] != 1 {
		t.Fatalf("Docker loads=%d, pushed=%v", cache.loads, cache.pushed)
	}
}

func writeDockerArchive(t *testing.T, filename, tag string) string {
	t.Helper()
	layer := []byte("verified shared layer")
	layerHash := sha256.Sum256(layer)
	config, err := json.Marshal(map[string]interface{}{
		"architecture": "amd64", "os": "linux",
		"rootfs": map[string]interface{}{"type": "layers", "diff_ids": []string{"sha256:" + hex.EncodeToString(layerHash[:])}},
	})
	if err != nil {
		t.Fatal(err)
	}
	configHash := sha256.Sum256(config)
	configName := hex.EncodeToString(configHash[:]) + ".json"
	manifest, err := json.Marshal([]map[string]interface{}{{
		"Config": configName, "RepoTags": []string{tag}, "Layers": []string{"layer/layer.tar"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	writer := tar.NewWriter(file)
	for name, data := range map[string][]byte{configName: config, "manifest.json": manifest, "layer/layer.tar": layer} {
		if err := writer.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return "sha256:" + hex.EncodeToString(configHash[:])
}
