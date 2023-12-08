package netest

import (
	"context"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"
)

const testImage = "public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53"

// VSphereOptions provides options for the ExecVSphereTests suite.
type VSphereOptions struct {
	AdditionalEndpoints []string
}

// ExecVSphereTests executes the test suite for a vSphere environment.
func ExecVSphereTests(ctx context.Context, i invoker.Invoker, b TestBroadcaster, opts VSphereOptions) Report {
	endpoints := append([]string{
		"public.ecr.aws",
		"anywhere-assets.eks.amazonaws.com",
		"distro.eks.amazonaws.com",
		"d2glxqk2uabbnd.cloudfront.net",
		"d5l0dvt14r5h8.cloudfront.net",
		"api.github.com",
	}, opts.AdditionalEndpoints...)

	var suite TestSuite

	for _, endpoint := range endpoints {
		suite.Add(ResolveTest(endpoint))
		suite.Add(PingTest(endpoint))
	}

	suite.Add(CrictlImagePullTest(testImage))

	var report Report
	for _, t := range suite {
		b.Broadcast(t.Summary)
		report.Result(t.Run(ctx, i))
	}

	return report
}
