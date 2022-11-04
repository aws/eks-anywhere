package test

// Reader is a commonly used interface in multiple packages
// We are replicating here just to create a single mock we can
// use in multiple test packages.
type Reader interface {
	ReadFile(url string) ([]byte, error)
}
