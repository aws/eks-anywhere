package version

type VersionClient struct{}

func (vc *VersionClient) Get() Info {
	return Get()
}

func newVersionClient() *VersionClient {
	return &VersionClient{}
}
