package cluster

var defaultManager *ConfigManager

func init() {
	var err error
	defaultManager, err = NewDefaultConfigManager()
	if err != nil {
		panic(err)
	}
}

func manager() *ConfigManager {
	return defaultManager
}

func NewDefaultConfigManager() (*ConfigManager, error) {
	m := NewConfigManager()
	err := m.Register(
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
		nutanixEntry(),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
