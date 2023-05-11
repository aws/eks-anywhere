package version

// Client exists to make mocking the version package easier.
type Client struct{}

// Get returns the version info of eksa.
func (vc *Client) Get() Info {
	return Get()
}
