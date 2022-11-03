package cluster

// NewDefaultConfigClientBuilder returns a ConfigClientBuilder with the
// default processors to build a Config.
func NewDefaultConfigClientBuilder() *ConfigClientBuilder {
	return NewConfigClientBuilder().Register(
		getVSphereDatacenter,
		getVSphereMachineConfigs,
		getSnowDatacenter,
		getSnowMachineConfigs,
		getSnowIdentitySecret,
		getOIDC,
		getAWSIam,
		getGitOps,
		getFluxConfig,
	)
}
