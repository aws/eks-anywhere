package main

import (
	"os"

	"github.com/aws/eks-anywhere/cmd/eks-a-tool/cmd"
)

func main() {
	if cmd.Execute() == nil {
		os.Exit(0)
	}
	os.Exit(-1)
}
