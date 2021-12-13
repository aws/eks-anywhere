module github.com/aws/eks-anywhere

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.40
	github.com/aws/eks-anywhere/release v0.0.0-20211130194657-f6e9593c6551
	github.com/aws/eks-distro-build-tooling/release v0.0.0-20211103003257-a7e2379eae5e
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/mock v1.5.0
	github.com/google/go-github/v35 v35.2.0
	github.com/mrajashree/etcdadm-controller v0.1.3
	github.com/onsi/gomega v1.15.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2
	gopkg.in/square/go-jose.v2 v2.6.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/cluster-api v0.3.11-0.20210430210359-402a4524f006
	sigs.k8s.io/cluster-api-provider-vsphere v0.7.8
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.2.0
)

// exclude un-required transitive dependency from cluster-api-provider-vsphere v0.7.8
exclude sigs.k8s.io/cluster-api v0.3.14

// TODO: Once the repo is public, remove this so we use a versioned module
replace (
	github.com/aws/eks-anywhere/release => ./release
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc95
)
