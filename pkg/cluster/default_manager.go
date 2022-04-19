package cluster

var defaultManager *ConfigManager

func init() {
	defaultManager = NewConfigManager()
	err := defaultManager.Register(
		clusterEntry(),
		oidcEntry(),
		awsIamEntry(),
		gitOpsEntry(),
		fluxEntry(),
		vsphereEntry(),
		cloudstackEntry(),
		dockerEntry(),
		snowEntry(),
		tinkerbellEntry(),
	)
	if err != nil {
		panic(err)
	}
}

func manager() *ConfigManager {
	return defaultManager
}
