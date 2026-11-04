package docker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type cacheLoader struct {
	ids      map[string]string
	loads    int
	loadErr  error
	cacheErr error
}

func (c *cacheLoader) LoadFromFile(context.Context, string) error {
	c.loads++
	return c.loadErr
}

func (c *cacheLoader) ImageID(_ context.Context, image string) (string, error) {
	return c.ids[image], c.cacheErr
}

func TestDiskSourceVerifiedCache(t *testing.T) {
	for _, tt := range []struct {
		name     string
		change   func(map[string][]byte, []savedImage, *cacheLoader)
		images   []string
		wantLoad bool
	}{
		{name: "all images cached"},
		{name: "duplicate requested tag", images: []string{"example.test/image:one", "example.test/image:one"}},
		{name: "missing cache image", change: func(_ map[string][]byte, _ []savedImage, cache *cacheLoader) {
			delete(cache.ids, "example.test/image:two")
		}, wantLoad: true},
		{name: "different cached platform", change: func(_ map[string][]byte, _ []savedImage, cache *cacheLoader) {
			cache.ids["example.test/image:one"] = "sha256:" + digest([]byte("arm64 config"))
		}, wantLoad: true},
		{name: "cache lookup error", change: func(_ map[string][]byte, _ []savedImage, cache *cacheLoader) {
			cache.cacheErr = errors.New("daemon unavailable")
		}, wantLoad: true},
		{name: "missing archive image", images: []string{"example.test/image:absent"}, wantLoad: true},
		{name: "corrupt manifest", change: func(files map[string][]byte, _ []savedImage, _ *cacheLoader) {
			files["manifest.json"] = []byte("{")
		}, wantLoad: true},
		{name: "corrupt config", change: func(files map[string][]byte, manifest []savedImage, _ *cacheLoader) {
			files[manifest[0].Config] = []byte("{}")
		}, wantLoad: true},
		{name: "corrupt layer", change: func(files map[string][]byte, manifest []savedImage, _ *cacheLoader) {
			files[manifest[0].Layers[0]] = []byte("damaged")
		}, wantLoad: true},
		{name: "missing layer", change: func(files map[string][]byte, manifest []savedImage, _ *cacheLoader) {
			delete(files, manifest[0].Layers[0])
		}, wantLoad: true},
		{name: "missing config", change: func(files map[string][]byte, manifest []savedImage, _ *cacheLoader) {
			delete(files, manifest[0].Config)
		}, wantLoad: true},
		{name: "OCI-only archive", change: func(files map[string][]byte, _ []savedImage, _ *cacheLoader) {
			delete(files, "manifest.json")
		}, wantLoad: true},
		{name: "empty request", images: []string{}, wantLoad: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			files, manifest, ids := savedFixture(t, false, false)
			cache := &cacheLoader{ids: ids}
			if tt.change != nil {
				tt.change(files, manifest, cache)
			}
			images := tt.images
			if images == nil {
				images = []string{"example.test/image:one", "example.test/image:two"}
			}
			filename := writeSavedArchive(t, files)
			if err := NewDiskSource(cache, filename).Load(context.Background(), images...); err != nil {
				t.Fatal(err)
			}
			if (cache.loads == 1) != tt.wantLoad {
				t.Fatalf("Docker loads = %d; want load %v", cache.loads, tt.wantLoad)
			}
		})
	}
}

func TestDiskSourceOCICache(t *testing.T) {
	for _, tt := range []struct {
		name     string
		change   func(*testing.T, map[string][]byte)
		wantLoad bool
	}{
		{name: "OCI Docker save"},
		{name: "corrupt OCI manifest", change: func(t *testing.T, files map[string][]byte) {
			index := fixtureIndex(t, files)
			files[descriptorPath(index.Manifests[0].Digest)] = []byte("{}")
		}, wantLoad: true},
		{name: "different tag", change: func(t *testing.T, files map[string][]byte) {
			index := fixtureIndex(t, files)
			index.Manifests[0].Annotations["io.containerd.image.name"] = "example.test/other:one"
			files["index.json"] = fixtureJSON(t, index)
		}, wantLoad: true},
		{name: "multi-platform index", change: func(t *testing.T, files map[string][]byte) {
			index := fixtureIndex(t, files)
			index.Manifests[0].MediaType = "application/vnd.oci.image.index.v1+json"
			files["index.json"] = fixtureJSON(t, index)
		}, wantLoad: true},
		{name: "different platform annotation", change: func(t *testing.T, files map[string][]byte) {
			index := fixtureIndex(t, files)
			index.Manifests[0].Platform = &savedPlatform{Architecture: "arm64", OS: "linux"}
			files["index.json"] = fixtureJSON(t, index)
		}, wantLoad: true},
		{name: "different OCI layer list", change: func(t *testing.T, files map[string][]byte) {
			index := fixtureIndex(t, files)
			descriptor := &index.Manifests[0]
			var manifest savedOCIManifest
			if err := json.Unmarshal(files[descriptorPath(descriptor.Digest)], &manifest); err != nil {
				t.Fatal(err)
			}
			manifest.Layers = nil
			data := fixtureJSON(t, manifest)
			descriptor.Digest = "sha256:" + digest(data)
			files[descriptorPath(descriptor.Digest)] = data
			files["index.json"] = fixtureJSON(t, index)
		}, wantLoad: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			files, _, ids := savedFixture(t, true, true)
			if tt.change != nil {
				tt.change(t, files)
			}
			cache := &cacheLoader{ids: ids}
			if err := NewDiskSource(cache, writeSavedArchive(t, files)).Load(context.Background(),
				"example.test/image:one", "example.test/image:two"); err != nil {
				t.Fatal(err)
			}
			if (cache.loads == 1) != tt.wantLoad {
				t.Fatalf("Docker loads = %d; want load %v", cache.loads, tt.wantLoad)
			}
		})
	}
}

func TestDiskSourceLoadErrorPreserved(t *testing.T) {
	expected := errors.New("invalid archive")
	cache := &cacheLoader{loadErr: expected}
	err := NewDiskSource(cache, filepath.Join(t.TempDir(), "missing.tar")).Load(context.Background(), "image:one")
	if !errors.Is(err, expected) || cache.loads != 1 {
		t.Fatalf("load error = %v, loads = %d", err, cache.loads)
	}
}

func TestSavedArchiveDuplicateAndTruncated(t *testing.T) {
	files, _, _ := savedFixture(t, false, false)
	for _, name := range []string{"duplicate", "truncated"} {
		t.Run(name, func(t *testing.T) {
			filename := writeSavedArchive(t, files)
			data, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "duplicate":
				// Insert a second copy of the archive before the end marker.
				data = append(data[:len(data)-1024], data...)
			case "truncated":
				data = data[:len(data)/2]
			}
			if err := os.WriteFile(filename, data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := archiveImageIDs(filename); err == nil {
				t.Fatal("damaged archive was accepted")
			}
		})
	}
}

func TestDiskSourceUnsupportedEntriesUseDockerLoad(t *testing.T) {
	files, manifest, ids := savedFixture(t, true, true)
	protected := []string{"manifest.json", manifest[0].Config, manifest[0].Layers[0]}
	for _, entryType := range []byte{tar.TypeSymlink, tar.TypeLink, tar.TypeFifo, tar.TypeChar, tar.TypeBlock} {
		for _, name := range protected {
			t.Run(fmt.Sprintf("type=%d/path=%s", entryType, name), func(t *testing.T) {
				filename := writeSavedArchive(t, files)
				header := &tar.Header{Name: name, Typeflag: entryType, Mode: 0o600}
				if entryType == tar.TypeSymlink || entryType == tar.TypeLink {
					header.Linkname = "replacement"
				}
				appendSavedHeader(t, filename, header)
				expected := errors.New("Docker rejected unsupported archive entry")
				cache := &cacheLoader{ids: ids, loadErr: expected}
				err := NewDiskSource(cache, filename).Load(context.Background(),
					"example.test/image:one", "example.test/image:two")
				if !errors.Is(err, expected) || cache.loads != 1 {
					t.Fatalf("error = %v, Docker loads = %d", err, cache.loads)
				}
			})
		}
	}
}

func TestDiskSourceDirectoryEntries(t *testing.T) {
	files, _, ids := savedFixture(t, true, true)
	for _, tt := range []struct {
		name     string
		wantLoad bool
	}{
		{name: "blobs/"},
		{name: "manifest.json", wantLoad: true},
		{name: "/absolute/", wantLoad: true},
		{name: "blobs/../outside/", wantLoad: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			filename := writeSavedArchive(t, files)
			appendSavedHeader(t, filename, &tar.Header{Name: tt.name, Typeflag: tar.TypeDir, Mode: 0o755})
			cache := &cacheLoader{ids: ids}
			if err := NewDiskSource(cache, filename).Load(context.Background(),
				"example.test/image:one", "example.test/image:two"); err != nil {
				t.Fatal(err)
			}
			if (cache.loads == 1) != tt.wantLoad {
				t.Fatalf("Docker loads = %d; want load %v", cache.loads, tt.wantLoad)
			}
		})
	}
}

func appendSavedHeader(t *testing.T, filename string, header *tar.Header) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	buffer.Write(data[:len(data)-1024])
	writer := tar.NewWriter(&buffer)
	if err := writer.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, buffer.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}

func savedFixture(t *testing.T, oci, compressed bool) (map[string][]byte, []savedImage, map[string]string) {
	t.Helper()
	files := make(map[string][]byte)
	ids := make(map[string]string)
	var manifest []savedImage
	index := savedIndex{SchemaVersion: 2}
	layer := []byte("shared uncompressed layer")
	rawLayer := layer
	layerPath := "shared/layer.tar"
	if compressed {
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		if _, err := writer.Write(layer); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		rawLayer = buffer.Bytes()
	}
	if oci {
		layerPath = "blobs/sha256/" + digest(rawLayer)
	}
	files[layerPath] = rawLayer
	for _, tag := range []string{"one", "two"} {
		config := fixtureJSON(t, map[string]interface{}{
			"architecture": "amd64", "os": "linux", "variant": tag,
			"rootfs": map[string]interface{}{"type": "layers", "diff_ids": []string{"sha256:" + digest(layer)}},
		})
		configPath := digest(config) + ".json"
		if oci {
			configPath = "blobs/sha256/" + digest(config)
		}
		files[configPath] = config
		fullTag := "example.test/image:" + tag
		ids[fullTag] = "sha256:" + digest(config)
		manifest = append(manifest, savedImage{Config: configPath, RepoTags: []string{fullTag}, Layers: []string{layerPath}})
		if oci {
			ociManifest := savedOCIManifest{
				SchemaVersion: 2,
				Config:        savedDescriptor{Digest: ids[fullTag]},
				Layers:        []savedDescriptor{{Digest: "sha256:" + digest(rawLayer)}},
			}
			data := fixtureJSON(t, ociManifest)
			manifestDigest := "sha256:" + digest(data)
			files[descriptorPath(manifestDigest)] = data
			index.Manifests = append(index.Manifests, savedDescriptor{
				MediaType: "application/vnd.oci.image.manifest.v1+json", Digest: manifestDigest,
				Annotations: map[string]string{"io.containerd.image.name": fullTag},
			})
		}
	}
	files["manifest.json"] = fixtureJSON(t, manifest)
	if oci {
		files["index.json"] = fixtureJSON(t, index)
	}
	return files, manifest, ids
}

func writeSavedArchive(t *testing.T, files map[string][]byte) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "images.tar")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer := tar.NewWriter(file)
	for name, data := range files {
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
	return filename
}

func fixtureJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func fixtureIndex(t *testing.T, files map[string][]byte) savedIndex {
	t.Helper()
	var index savedIndex
	if err := json.Unmarshal(files["index.json"], &index); err != nil {
		t.Fatal(err)
	}
	return index
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
