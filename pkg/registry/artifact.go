package registry

// Artifact to head release dependency.
type Artifact struct {
	Registry   string
	Repository string
	Tag        string
	Digest     string
}

// NewArtifact creates a new artifact object.
func NewArtifact(registry, repository, tag, digest string) Artifact {
	return Artifact{
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		Digest:     digest,
	}
}

// Version returns tag or digest.
func (art *Artifact) Version() string {
	if art.Digest != "" {
		return "@" + art.Digest
	}
	return ":" + art.Tag
}

// VersionedImage returns full URI for image.
func (art *Artifact) VersionedImage() string {
	version := art.Version()
	return art.Registry + "/" + art.Repository + version
}
