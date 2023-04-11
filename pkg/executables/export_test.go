package executables

import (
	"context"
	"time"
)

func KubectlWaitRetryPolicy(k *Kubectl, totalRetries int, err error) (retry bool, wait time.Duration) {
	return k.kubectlWaitRetryPolicy(totalRetries, err)
}

func CallKubectlPrivateWait(k *Kubectl, ctx context.Context, kubeconfig string, timeoutTime time.Time, forCondition string, property string, namespace string) error {
	return k.wait(ctx, kubeconfig, timeoutTime, forCondition, property, namespace)
}
