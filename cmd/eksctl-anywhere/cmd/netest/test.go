package netest

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"
)

// ResolveTest creates a test that validates hostname can be resolved.
func ResolveTest(hostname string) Test {
	return Test{
		Summary: fmt.Sprintf("Ensure hostname can be resolved %v", hostname),
		Run: func(ctx context.Context, i invoker.Invoker) TestResult {
			return toResult(i.Invoke(ctx, fmt.Sprintf("nslookup %s", hostname)))
		},
	}
}

// PingTest creates a test that validates endpoint is pingable.
func PingTest(endpoint string) Test {
	return Test{
		Summary: fmt.Sprintf("Enure endpoint is pingable %v", endpoint),
		Run: func(ctx context.Context, i invoker.Invoker) TestResult {
			return toResult(i.Invoke(ctx, fmt.Sprintf("ping -c 1 %s", endpoint)))
		},
	}
}

// CrictlImagePullTest creates a test that validated image can be pulled using crictl.
func CrictlImagePullTest(image string) Test {
	return Test{
		Summary: fmt.Sprintf("Ensure Docker image can be pulled (%v)", image),
		Run: func(ctx context.Context, i invoker.Invoker) TestResult {
			i.Invoke(ctx, fmt.Sprintf("sudo crictl rmi %s", image))
			defer i.Invoke(ctx, fmt.Sprintf("sudo crictl rmi %s", image))
			return toResult(i.Invoke(ctx, fmt.Sprintf("sudo crictl pull %s", image)))
		},
	}
}
