package supportbundle

type Analyze struct {
	CustomResourceDefinition *customResourceDefinition `json:"customResourceDefinition,omitempty" yaml:"customResourceDefinition,omitempty"`
	Secret                   *analyzeSecret            `json:"secret,omitempty" yaml:"secret,omitempty"`
	ImagePullSecret          *imagePullSecret          `json:"imagePullSecret,omitempty" yaml:"imagePullSecret,omitempty"`
	DeploymentStatus         *deploymentStatus         `json:"deploymentStatus,omitempty" yaml:"deploymentStatus,omitempty"`
}

type customResourceDefinition struct {
	analyzeMeta                  `json:",inline" yaml:",inline"`
	Outcomes                     []*outcome `json:"outcomes" yaml:"outcomes"`
	CustomResourceDefinitionName string     `json:"customResourceDefinitionName" yaml:"customResourceDefinitionName"`
}

type analyzeSecret struct {
	analyzeMeta `json:",inline" yaml:",inline"`
	Outcomes    []*outcome `json:"outcomes" yaml:"outcomes"`
	SecretName  string     `json:"secretName" yaml:"secretName"`
	Namespace   string     `json:"namespace" yaml:"namespace"`
	Key         string     `json:"key,omitempty" yaml:"key,omitempty"`
}

type imagePullSecret struct {
	analyzeMeta  `json:",inline" yaml:",inline"`
	Outcomes     []*outcome `json:"outcomes" yaml:"outcomes"`
	RegistryName string     `json:"registryName" yaml:"registryName"`
}

type deploymentStatus struct {
	analyzeMeta `json:",inline" yaml:",inline"`
	Outcomes    []*outcome `json:"outcomes" yaml:"outcomes"`
	Namespace   string     `json:"namespace" yaml:"namespace"`
	Name        string     `json:"name" yaml:"name"`
}

type analyzeMeta struct {
	CheckName string `json:"checkName,omitempty" yaml:"checkName,omitempty"`
	Exclude   bool   `json:"exclude,omitempty" yaml:"exclude,omitempty"`
}
