package registry

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
func (cache *Cache) Get(context StorageContext) (StorageClient, error) {
	aClient, found := cache.registries[context.host]
	if !found {
		aClient = NewOCIRegistry(context)
		err := aClient.Init()
		if err != nil {
			return nil, err
		}
		cache.registries[context.host] = aClient
	}
	return aClient, nil
}

// Set a client in the cache.
func (cache *Cache) Set(registryName string, client StorageClient) {
	cache.registries[registryName] = client
}
