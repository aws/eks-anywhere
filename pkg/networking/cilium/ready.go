package cilium

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
)

func checkDaemonSetReady(daemonSet *v1.DaemonSet) error {
	if daemonSet.Status.DesiredNumberScheduled != daemonSet.Status.NumberReady {
		return fmt.Errorf("daemonSet %s is not ready: %d/%d ready", daemonSet.Name, daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled)
	}
	return nil
}

func checkPreflightDaemonSetReady(ciliumDaemonSet, preflightDaemonSet *v1.DaemonSet) error {
	if ciliumDaemonSet.Status.NumberReady != preflightDaemonSet.Status.NumberReady {
		return fmt.Errorf("cilium preflight check DS is not ready: %d want and %d ready", ciliumDaemonSet.Status.NumberReady, preflightDaemonSet.Status.NumberReady)
	}
	return nil
}

func checkDeploymentReady(deployment *v1.Deployment) error {
	if deployment.Status.Replicas != deployment.Status.ReadyReplicas {
		return fmt.Errorf("deployment %s is not ready: %d/%d ready", deployment.Name, deployment.Status.ReadyReplicas, deployment.Status.Replicas)
	}
	return nil
}
