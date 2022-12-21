package executables

import (
	"context"
	"time"
)

var (
	KubectlWaitRetryPolicy    = kubectlWaitRetryPolicy
	ClusterctlMoveRetryPolicy = clusterctlMoveRetryPolicy
)

func CallKubectlPrivateWait(k *Kubectl, ctx context.Context, kubeconfig string, timeoutTime time.Time, forCondition string, property string, namespace string) error {
	return k.wait(ctx, kubeconfig, timeoutTime, forCondition, property, namespace)
}
