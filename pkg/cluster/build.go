package cluster

// NewDefaultConfigClientBuilder returns a ConfigClientBuilder with the
// default processors to build a Config
func NewDefaultConfigClientBuilder() *ConfigClientBuilder {
	return NewConfigClientBuilder().Register(
		getSnowDatacenter,
		getSnowMachineConfigs,
		getCloudStackDatacenter,
		getCloudStackMachineConfigs,
		getOIDC,
		getAWSIam,
		getGitOps,
		getFluxConfig,
	)
}
