package main

import (
	"os"

	"github.com/aws/eks-anywhere-test-tool/cmd"
)

func main() {
	if cmd.Execute() == nil {
		os.Exit(0)
	}
	os.Exit(-1)
}
