package netest

import (
	"context"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"
)

type Test struct {
	Summary string
	Run     func(context.Context, invoker.Invoker) TestResult
}

type TestSuite []Test

func (suite *TestSuite) Add(t Test) {
	*suite = append(*suite, t)
}
