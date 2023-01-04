package registry

import "fmt"

// Cache storage client for an OCI registry.
type Cache struct {
	registries map[string]StorageClient
}

// NewCache creates an OCI registry client.
func NewCache() *Cache {
	return &Cache{
		registries: make(map[string]StorageClient),
	}
}

// Get cached registry client or make it.
func (cache *Cache) Get(host string, certFile string, insecure bool) (StorageClient, error) {
	aClient, found := cache.registries[host]
	if !found {
		aClient = NewOCIRegistry(host, certFile, insecure)
		err := aClient.Init()
		if err != nil {
			return nil, fmt.Errorf("error with repository %s: %v", host, err)
		}
		cache.registries[host] = aClient
	}
	return aClient, nil
}
