package cluster

func clusterEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		Defaulters: []Defaulter{
			func(c *Config) {
				c.Cluster.SetDefaults()
			},
		},
	}
}
