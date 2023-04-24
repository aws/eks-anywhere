package main

import (
	"fmt"
	"os"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd"
	"github.com/aws/eks-anywhere/pkg/eksctl"
)

func main() {
	if eksctl.Enabled() {
		err := eksctl.ValidateVersion()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}
	if cmd.Execute() == nil {
		os.Exit(0)
	}
	os.Exit(-1)
}
