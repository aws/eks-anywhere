package providers

func ConfigsMapToSlice(c map[string]MachineConfig) []MachineConfig {
	configs := make([]MachineConfig, 0, len(c))
	for _, config := range c {
		configs = append(configs, config)
	}

	return configs
}
