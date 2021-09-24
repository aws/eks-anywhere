package supportbundle

type Analyze struct {
	CustomResourceDefinition *customResourceDefinition `json:"customResourceDefinition,omitempty"`
	Secret                   *analyzeSecret            `json:"secret,omitempty"`
	ImagePullSecret          *imagePullSecret          `json:"imagePullSecret,omitempty"`
	DeploymentStatus         *deploymentStatus         `json:"deploymentStatus,omitempty"`
}

type customResourceDefinition struct {
	analyzeMeta                  `json:",inline"`
	Outcomes                     []*outcome `json:"outcomes"`
	CustomResourceDefinitionName string     `json:"customResourceDefinitionName"`
}

type analyzeSecret struct {
	analyzeMeta `json:",inline"`
	Outcomes    []*outcome `json:"outcomes"`
	SecretName  string     `json:"secretName"`
	Namespace   string     `json:"namespace"`
	Key         string     `json:"key,omitempty"`
}

type imagePullSecret struct {
	analyzeMeta  `json:",inline"`
	Outcomes     []*outcome `json:"outcomes"`
	RegistryName string     `json:"registryName"`
}

type deploymentStatus struct {
	analyzeMeta `json:",inline"`
	Outcomes    []*outcome `json:"outcomes"`
	Namespace   string     `json:"namespace"`
	Name        string     `json:"name"`
}

type analyzeMeta struct {
	CheckName string `json:"checkName,omitempty"`
	Exclude   bool   `json:"exclude,omitempty"`
}
