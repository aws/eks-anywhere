module github.com/aws/eks-anywhere-test-tool

go 1.16

require (
	github.com/aws/aws-sdk-go v1.42.16
	github.com/aws/eks-anywhere v0.7.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

exclude sigs.k8s.io/cluster-api/test v1.0.0

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20220207154021-dcf66392d606
