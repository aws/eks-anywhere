package netest

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"
)

const testImage = "public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53"

type VSphereOptions struct {
	AdditionalEndpoints []string
}

/*

1. some tests may not be appropriate for a provider
2. some tests may be appropriate but run using different commands

Building a test report

* Create a report datastructure
* Leverage the error interface to return multierrors


*/

type Outcome int

const (
	Fail Outcome = iota
	Pass
)

type Result struct {
	Cmd     string
	Outcome Outcome
	Error   string
}

func ExecVSphereTests(ctx context.Context, i invoker.Invoker, opts VSphereOptions) []Result {
	endpoints := append([]string{
		"public.ecr.aws",
		"anywhere-assets.eks.amazonaws.com",
		"distro.eks.amazonaws.com",
		"d2glxqk2uabbnd.cloudfront.net",
		"d5l0dvt14r5h8.cloudfront.net",
		"api.github.com",
	}, opts.AdditionalEndpoints...)

	var results []Result

	for _, endpoint := range endpoints {
		outcome := i.Invoke(ctx, fmt.Sprintf("nslookup %s", endpoint))
		results = append(results, toResult(outcome))
	}

	for _, endpoint := range endpoints {
		outcome := i.Invoke(ctx, fmt.Sprintf("ping -c 1 %s", endpoint))
		results = append(results, toResult(outcome))
	}

	// Remove image before and after to circumvent old pulls and be a good citizen.
	i.Invoke(ctx, fmt.Sprintf("sudo crictl rmi %s", testImage))
	outcome := i.Invoke(ctx, fmt.Sprintf("sudo crictl pull %s", testImage))
	results = append(results, toResult(outcome))
	i.Invoke(ctx, fmt.Sprintf("sudo crictl rmi %s", testImage))

	return results
}

// func ExecCloudstackTests(ctx context.Context, i Invoker, opts CloudStackOptions) error {

// }

func toResult(r invoker.Outcome) Result {
	res := Result{
		Cmd:     r.Cmd,
		Outcome: Pass,
	}

	if r.Error != nil {
		res.Outcome = Fail
		res.Error = r.Stderr.String()
	}

	return res
}
