package cmd

const (
	defaultTinkerbellParamsTemplate = `managementClusterName: <management cluster name>
podsCidrBlocks:
  - 192.168.64.0/18
servicesCidrBlocks:
  - 10.96.0.0/12
kubernetesVersion: 1.26
cpCount: 1
workerCount: 2
cpEndpointHost: <control plane endpoint host ip>
tinkerbellIP: <tinkerbellIP>
adminIP: <admin machine ip>
osFamily: ubuntu
osImageURL: <osImageURL of K8s 1.26>
hardwareCSV: <hardware CSV file>
sshAuthorizedKeyFile: <sshKey.pub file>
tinkerbellTemplateConfigTemplateFile: tinkerbellTemplateConfigTemplate.yaml
`
)

// GetDefaultTinkerbellParamsTemplate returns the default TinkerbellParamsTemplate.
func GetDefaultTinkerbellParamsTemplate() string {
	return string(defaultTinkerbellParamsTemplate)
}
