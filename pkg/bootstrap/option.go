package bootstrap

func WithExtraDockerMounts() CreateOption {
	return func(cfg *kindExecConfig) {
		cfg.DockerExtraMounts = true
	}
}

func WithExtraPortMappings(ports []int) CreateOption {
	return func(cfg *kindExecConfig) {
		cfg.ExtraPortMappings = ports
	}
}

func WithEnv(env map[string]string) CreateOption {
	return func(cfg *kindExecConfig) {
		for name, value := range env {
			cfg.env[name] = value
		}
	}
}

func WithRegistryMirror(endpoint string, caCertFile string) CreateOption {
	return func(cfg *kindExecConfig) {
		cfg.RegistryMirrorEndpoint = endpoint
		cfg.RegistryCACertPath = caCertFile
	}
}
