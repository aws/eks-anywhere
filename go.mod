module github.com/aws/eks-anywhere

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.40
	github.com/aws/eks-anywhere/release v0.0.0-20211130194657-f6e9593c6551
	github.com/aws/eks-distro-build-tooling/release v0.0.0-20211103003257-a7e2379eae5e
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/mock v1.6.0
	github.com/google/go-github/v35 v35.3.0
	github.com/google/uuid v1.3.0
	github.com/mrajashree/etcdadm-controller v1.0.0-rc3
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/tinkerbell/cluster-api-provider-tinkerbell v0.1.0
	github.com/tinkerbell/tink v0.6.0
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/cluster-api v1.0.2
	sigs.k8s.io/cluster-api-provider-vsphere v1.0.1
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.3.0
)

// exclude un-required transitive dependency from cluster-api-provider-vsphere v1.0.1
exclude sigs.k8s.io/cluster-api/test v1.0.0

// TODO: Once the repo is public, remove this so we use a versioned module
replace (
	github.com/aws/eks-anywhere/release => ./release
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20220207154021-dcf66392d606
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc95
	github.com/tinkerbell/cluster-api-provider-tinkerbell => github.com/pokearu/cluster-api-provider-tinkerbell v0.0.0-20220128001529-79d851d0861f
)
