package docker

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type savedImage struct {
	Config   string
	RepoTags []string
	Layers   []string
}

type savedConfig struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant"`
	RootFS       struct {
		Type    string
		DiffIDs []string `json:"diff_ids"`
	} `json:"rootfs"`
}

type archiveFile struct {
	digest string
	diffID string
	config []byte
}

type savedDescriptor struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Annotations map[string]string `json:"annotations"`
	Platform    *savedPlatform    `json:"platform,omitempty"`
}

type savedPlatform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

type savedIndex struct {
	SchemaVersion int               `json:"schemaVersion"`
	Manifests     []savedDescriptor `json:"manifests"`
}

type savedOCIManifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	Config        savedDescriptor   `json:"config"`
	Layers        []savedDescriptor `json:"layers"`
}

// Docker save contains configuration IDs and uncompressed layer digests. Verify
// both before reusing cached tags; a tag match alone can conceal a damaged archive.
func archiveImageIDs(filename string) (map[string]string, error) {
	manifest, index, err := readSavedManifest(filename)
	if err != nil {
		return nil, err
	}
	files, err := readSavedFiles(filename, manifest, index)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]string)
	for _, image := range manifest {
		id, err := verifySavedImage(image, files)
		if err != nil {
			return nil, err
		}
		for _, tag := range image.RepoTags {
			if previous, ok := ids[tag]; ok && previous != id {
				return nil, fmt.Errorf("conflicting saved image tag %q", tag)
			}
			ids[tag] = id
		}
	}
	if err := verifySavedIndex(index, manifest, files, ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func verifySavedImage(image savedImage, files map[string]archiveFile) (string, error) {
	config, ok := files[image.Config]
	if !ok || config.digest != configPathDigest(image.Config) {
		return "", fmt.Errorf("invalid saved image configuration %q", image.Config)
	}
	var parsed savedConfig
	if err := json.Unmarshal(config.config, &parsed); err != nil {
		return "", err
	}
	if err := verifySavedLayers(image, parsed, files); err != nil {
		return "", err
	}
	return config.digest, nil
}

func verifySavedLayers(image savedImage, parsed savedConfig, files map[string]archiveFile) error {
	if parsed.RootFS.Type != "layers" || len(parsed.RootFS.DiffIDs) != len(image.Layers) {
		return fmt.Errorf("invalid saved image layers for %q", image.Config)
	}
	for i, name := range image.Layers {
		layer, ok := files[name]
		if !ok || layer.diffID != parsed.RootFS.DiffIDs[i] {
			return fmt.Errorf("invalid saved image layer %q", name)
		}
		if strings.HasPrefix(name, "blobs/sha256/") && layer.digest != configPathDigest(name) {
			return fmt.Errorf("invalid saved image blob %q", name)
		}
	}
	return nil
}

func readSavedManifest(filename string) ([]savedImage, *savedIndex, error) {
	var manifest []savedImage
	var index *savedIndex
	seen := make(map[string]bool)
	err := walkSavedArchive(filename, func(header *tar.Header, reader io.Reader) error {
		if seen[header.Name] {
			return fmt.Errorf("duplicate saved image file %q", header.Name)
		}
		seen[header.Name] = true
		if header.Name != "manifest.json" && header.Name != "index.json" {
			return nil
		}
		data, err := readArchiveMetadata(header, reader)
		if err != nil {
			return err
		}
		if header.Name == "index.json" {
			return json.Unmarshal(data, &index)
		}
		return json.Unmarshal(data, &manifest)
	})
	if err != nil {
		return nil, nil, err
	}
	if len(manifest) == 0 {
		return nil, nil, fmt.Errorf("Docker save manifest is missing or empty")
	}
	return manifest, index, nil
}

func readSavedFiles(filename string, manifest []savedImage, index *savedIndex) (map[string]archiveFile, error) {
	configs := make(map[string]bool)
	layers := make(map[string]bool)
	for _, image := range manifest {
		configs[image.Config] = true
		for _, layer := range image.Layers {
			layers[layer] = true
		}
	}
	if index != nil {
		for _, descriptor := range index.Manifests {
			configs[descriptorPath(descriptor.Digest)] = true
		}
	}
	files := make(map[string]archiveFile)
	err := walkSavedArchive(filename, func(header *tar.Header, reader io.Reader) error {
		switch {
		case configs[header.Name]:
			data, err := readArchiveMetadata(header, reader)
			if err != nil {
				return err
			}
			digest := sha256.Sum256(data)
			files[header.Name] = archiveFile{digest: "sha256:" + hex.EncodeToString(digest[:]), config: data}
		case layers[header.Name]:
			layer, err := readSavedLayer(reader)
			if err != nil {
				return err
			}
			files[header.Name] = layer
		}
		return nil
	})
	return files, err
}

func readSavedLayer(reader io.Reader) (archiveFile, error) {
	rawHash := sha256.New()
	buffered := bufio.NewReader(io.TeeReader(reader, rawHash))
	magic, err := buffered.Peek(2)
	if err != nil && err != io.EOF {
		return archiveFile{}, err
	}
	var uncompressed io.Reader = buffered
	if len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		decompressed, err := gzip.NewReader(buffered)
		if err != nil {
			return archiveFile{}, err
		}
		defer decompressed.Close()
		uncompressed = decompressed
	}
	diffHash := sha256.New()
	if _, err := io.Copy(diffHash, uncompressed); err != nil {
		return archiveFile{}, err
	}
	return archiveFile{
		digest: "sha256:" + hex.EncodeToString(rawHash.Sum(nil)),
		diffID: "sha256:" + hex.EncodeToString(diffHash.Sum(nil)),
	}, nil
}

func verifySavedIndex(index *savedIndex, saved []savedImage, files map[string]archiveFile, ids map[string]string) error {
	if index == nil {
		return nil
	}
	if index.SchemaVersion != 2 || len(index.Manifests) == 0 {
		return fmt.Errorf("saved OCI index is invalid")
	}
	seen := make(map[string]bool)
	for _, descriptor := range index.Manifests {
		manifest, err := readSavedOCIManifest(descriptor, files)
		if err != nil {
			return err
		}
		tag := descriptor.Annotations["io.containerd.image.name"]
		if ids[tag] != manifest.Config.Digest || seen[tag] {
			return fmt.Errorf("saved OCI tag %q does not match Docker save manifest", tag)
		}
		seen[tag] = true
		if err := verifySavedOCILayers(manifest, saved, files); err != nil {
			return err
		}
	}
	if len(seen) != len(ids) {
		return fmt.Errorf("saved OCI tags do not match Docker save manifest")
	}
	return nil
}

func readSavedOCIManifest(descriptor savedDescriptor, files map[string]archiveFile) (*savedOCIManifest, error) {
	// A multi-platform index needs platform-specific cache inspection. Retain
	// Docker's loader for that layout instead of accepting only one platform.
	if descriptor.MediaType != "application/vnd.oci.image.manifest.v1+json" &&
		descriptor.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		return nil, fmt.Errorf("unsupported saved OCI manifest type %q", descriptor.MediaType)
	}
	file, ok := files[descriptorPath(descriptor.Digest)]
	if !ok || file.digest != descriptor.Digest {
		return nil, fmt.Errorf("invalid saved OCI manifest %q", descriptor.Digest)
	}
	var manifest savedOCIManifest
	if err := json.Unmarshal(file.config, &manifest); err != nil {
		return nil, err
	}
	if manifest.SchemaVersion != 2 {
		return nil, fmt.Errorf("saved OCI manifest schema is unsupported")
	}
	config, ok := files[descriptorPath(manifest.Config.Digest)]
	if !ok || config.digest != manifest.Config.Digest {
		return nil, fmt.Errorf("invalid saved OCI configuration %q", manifest.Config.Digest)
	}
	if err := verifySavedPlatform(descriptor.Platform, config); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func verifySavedPlatform(platform *savedPlatform, config archiveFile) error {
	if platform == nil {
		return nil
	}
	var parsed savedConfig
	if err := json.Unmarshal(config.config, &parsed); err != nil {
		return err
	}
	if platform.Architecture != parsed.Architecture || platform.OS != parsed.OS || platform.Variant != parsed.Variant {
		return fmt.Errorf("saved OCI platform does not match image configuration")
	}
	return nil
}

func verifySavedOCILayers(manifest *savedOCIManifest, saved []savedImage, files map[string]archiveFile) error {
	var layers []string
	for _, image := range saved {
		if files[image.Config].digest == manifest.Config.Digest {
			layers = image.Layers
			break
		}
	}
	if len(layers) != len(manifest.Layers) {
		return fmt.Errorf("saved OCI layers do not match Docker save manifest")
	}
	for i, layer := range manifest.Layers {
		file, ok := files[descriptorPath(layer.Digest)]
		if !ok || file.digest != layer.Digest || file.digest != files[layers[i]].digest {
			return fmt.Errorf("invalid saved OCI layer %q", layer.Digest)
		}
	}
	return nil
}

func descriptorPath(digest string) string {
	return "blobs/sha256/" + strings.TrimPrefix(digest, "sha256:")
}

func walkSavedArchive(filename string, visit func(*tar.Header, io.Reader) error) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := tar.NewReader(file)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if path.IsAbs(header.Name) {
			return fmt.Errorf("absolute saved image path %q", header.Name)
		}
		for _, component := range strings.Split(header.Name, "/") {
			if component == ".." {
				return fmt.Errorf("parent traversal in saved image path %q", header.Name)
			}
		}
		header.Name = path.Clean(header.Name)
		switch header.Typeflag {
		case tar.TypeReg, tar.TypeRegA, tar.TypeDir:
		default:
			return fmt.Errorf("unsupported saved image entry type %d for %q", header.Typeflag, header.Name)
		}
		if err := visit(header, reader); err != nil {
			return err
		}
	}
}

func readArchiveMetadata(header *tar.Header, reader io.Reader) ([]byte, error) {
	const maxMetadataSize = 4 << 20
	if header.Size > maxMetadataSize {
		return nil, fmt.Errorf("saved image metadata %q exceeds size limit", header.Name)
	}
	return io.ReadAll(reader)
}

func configPathDigest(name string) string {
	var value string
	switch {
	case strings.HasPrefix(name, "blobs/sha256/"):
		value = strings.TrimPrefix(name, "blobs/sha256/")
	case strings.HasSuffix(name, ".json") && !strings.Contains(name, "/"):
		value = strings.TrimSuffix(name, ".json")
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return ""
	}
	return "sha256:" + value
}
