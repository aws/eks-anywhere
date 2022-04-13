package cluster

func clusterEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		Defaulters: []Defaulter{
			func(c *Config) error {
				c.Cluster.SetDefaults()
				return nil
			},
		},
		Validations: []Validation{
			func(c *Config) error {
				return c.Cluster.Validate()
			},
		},
	}
}
