module github.com/aws/eks-anywhere-test-tool

go 1.16

require (
	github.com/aws/aws-sdk-go v1.42.16
	github.com/aws/eks-anywhere v0.6.2-0.20211130214857-f40ef7755a29
	github.com/go-logr/logr v1.2.1 // indirect
	github.com/go-logr/zapr v1.2.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

replace github.com/aws/eks-anywhere => ../../../../
