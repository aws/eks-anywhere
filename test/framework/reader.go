package framework

import "github.com/aws/eks-anywhere/pkg/files"

func newFileReader() *files.Reader {
	return files.NewReader(files.WithEKSAUserAgent("e2e-test", testBranch()))
}
