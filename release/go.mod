module github.com/aws/eks-anywhere/release

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.40
	github.com/aws/aws-sdk-go-v2 v1.5.0
	github.com/aws/eks-distro-build-tooling/release v0.0.0-20211103003257-a7e2379eae5e
	github.com/fsouza/go-dockerclient v1.7.2
	github.com/ghodss/yaml v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.22.2
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.3
)
