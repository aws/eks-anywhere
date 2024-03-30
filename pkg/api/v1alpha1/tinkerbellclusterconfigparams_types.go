package v1alpha1

// NOTE: yaml tags are required.  Any new fields you add must have yaml tags for the fields to be serialized.

// TinkerbellClusterConfigParams defines the parameters to generate tinkerbell cluster config.
type TinkerbellClusterConfigParams struct {
	PodsCidrBlocks                       []string `yaml:"podsCidrBlocks"`
	ServicesCidrBlocks                   []string `yaml:"servicesCidrBlocks"`
	ManagementClusterName                string   `yaml:"managementClusterName"`
	KubernetesVersion                    string   `yaml:"kubernetesVersion"`
	OSFamily                             OSFamily `yaml:"osFamily"`
	OSImageURL                           string   `yaml:"osImageURL"`
	CPEndpointHost                       string   `yaml:"cpEndpointHost"`
	TinkerbellIP                         string   `yaml:"tinkerbellIP"`
	AdminIP                              string   `yaml:"adminIP"`
	CPCount                              int      `yaml:"cpCount"`
	WorkerCount                          int      `yaml:"workerCount"`
	HardwareCSV                          string   `yaml:"hardwareCSV"`
	SSHAuthorizedKeyFile                 string   `yaml:"sshAuthorizedKeyFile"`
	TinkerbellTemplateConfigTemplateFile string   `yaml:"tinkerbellTemplateConfigTemplateFile"`
}
