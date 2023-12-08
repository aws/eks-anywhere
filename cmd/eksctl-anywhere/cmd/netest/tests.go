package netest

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"
)

const testImage = "public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53"

// Outcome is the outcome of executing a netest test suite.
type Outcome int

const (
	// Fail indicates a test suite failed.
	Fail Outcome = iota

	// Pass indicates a test suite passed.
	Pass
)

// Result provides detail on the result of executing an atomic test.
type Result struct {
	Cmd     string
	Outcome Outcome
	Error   string
}

// Report is a collect of test atomic results.
type Report []Result

// VSphereOptions provides options for the ExecVSphereTests suite.
type VSphereOptions struct {
	AdditionalEndpoints []string
}

// ExecVSphereTests executes the test suite for a vSphere environment.
func ExecVSphereTests(ctx context.Context, i invoker.Invoker, opts VSphereOptions) Report {
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
