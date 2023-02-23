package cluster

// NewDefaultConfigClientBuilder returns a ConfigClientBuilder with the
// default processors to build a Config.
func NewDefaultConfigClientBuilder() *ConfigClientBuilder {
	return NewConfigClientBuilder().Register(
		getCloudStackDatacenter,
		getTinkerbellMachineConfigs,
		getTinkerbellDatacenter,
		getDockerDatacenter,
		getVSphereDatacenter,
		getVSphereMachineConfigs,
		getSnowDatacenter,
		getSnowMachineConfigsAndIPPools,
		getSnowIdentitySecret,
		getOIDC,
		getAWSIam,
		getGitOps,
		getFluxConfig,
	)
}
