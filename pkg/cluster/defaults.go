package cluster

func SetConfigDefaults(c *Config) error {
	return manager().SetDefaults(c)
}
