package cmd

const (
	defaultVSphereParamsTemplate = `managementClusterName: <management cluster name>
podsCidrBlocks:
  - 192.168.192.0/18
servicesCidrBlocks:
  - 10.96.192.0/18
cpCount: 2
etcdCount: 3
workerCount: 3
cpEndpointHost: <control plane endpoint host ip>
kubernetesVersion: 1.28
datacenter: <vDatacenter>
insecure: true
network: <vCenterNetwork>
server: <serverIP>
thumbprint: <thumprint>
datastore: <vDatastore>
folder: <folder>
cpDiskGiB: 0
cpMemoryMiB: 0
cpNumCPUs: 0
etcdDiskGiB: 0
etcdMemoryMiB: 0
etcdNumCPUs: 0
workerDiskGiB: 256
workerMemoryMiB: 65536
workerNumCPUs: 16
osFamily: "ubuntu"
resourcePool: <resource pool>
template: <template name of OS>
sshAuthorizedKeyFile: <sshKey.pub>
`
)

// GetDefaultVSphereParamsTemplate returns the default VSphereParamsTemplate.
func GetDefaultVSphereParamsTemplate() string {
	return string(defaultVSphereParamsTemplate)
}
