package cluster

func SetConfigDefaults(c *Config) {
	manager().SetDefaults(c)
}
