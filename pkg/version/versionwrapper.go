package version

// VersionClient exists to make mocking the version package easier.
type VersionClient struct{}

// Get returns the version info of eksa.
func (vc *VersionClient) Get() Info {
	return Get()
}
