package v1alpha1

// NOTE: yaml tags are required.  Any new fields you add must have yaml tags for the fields to be serialized.

// VSphereClusterConfigParams defines the parameters to generate vSphere cluster config.
type VSphereClusterConfigParams struct {
	PodsCidrBlocks        []string `yaml:"podsCidrBlocks"`
	ServicesCidrBlocks    []string `yaml:"servicesCidrBlocks"`
	CPCount               int      `yaml:"cpCount"`
	EtcdCount             int      `yaml:"etcdCount"`
	WorkerCount           int      `yaml:"workerCount"`
	CPEndpointHost        string   `yaml:"cpEndpointHost"`
	ManagementClusterName string   `yaml:"managementClusterName"`
	KubernetesVersion     string   `yaml:"kubernetesVersion"`
	Datacenter            string   `yaml:"datacenter"`
	Insecure              bool     `yaml:"insecure"`
	Network               string   `yaml:"network"`
	Server                string   `yaml:"server"`
	Thumbprint            string   `yaml:"thumbprint"`
	Datastore             string   `yaml:"datastore"`
	Folder                string   `yaml:"folder"`
	CPDiskGiB             int      `yaml:"cpDiskGiB"`
	CPMemoryMiB           int      `yaml:"cpMemoryMiB"`
	CPNumCPUs             int      `yaml:"cpNumCPUs"`
	EtcdDiskGiB           int      `yaml:"etcdDiskGiB"`
	EtcdMemoryMiB         int      `yaml:"etcdMemoryMiB"`
	EtcdNumCPUs           int      `yaml:"etcdNumCPUs"`
	WorkerDiskGiB         int      `yaml:"workerDiskGiB"`
	WorkerMemoryMiB       int      `yaml:"workerMemoryMiB"`
	WorkerNumCPUs         int      `yaml:"workerNumCPUs"`
	OSFamily              OSFamily `yaml:"osFamily"`
	ResourcePool          string   `yaml:"resourcePool"`
	Template              string   `yaml:"template"`
	SSHAuthorizedKeyFile  string   `yaml:"sshAuthorizedKeyFile"`
}
