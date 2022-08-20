package cilium

import (
	"fmt"

	v1 "k8s.io/api/apps/v1"
)

func CheckDaemonSetReady(daemonSet *v1.DaemonSet) error {
	if err := checkDaemonSetObservedGeneration(daemonSet); err != nil {
		return err
	}

	if daemonSet.Status.DesiredNumberScheduled != daemonSet.Status.NumberReady {
		return fmt.Errorf("daemonSet %s is not ready: %d/%d ready", daemonSet.Name, daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled)
	}
	return nil
}

func CheckPreflightDaemonSetReady(ciliumDaemonSet, preflightDaemonSet *v1.DaemonSet) error {
	if err := checkDaemonSetObservedGeneration(ciliumDaemonSet); err != nil {
		return err
	}
	if err := checkDaemonSetObservedGeneration(preflightDaemonSet); err != nil {
		return err
	}

	if ciliumDaemonSet.Status.NumberReady != preflightDaemonSet.Status.NumberReady {
		return fmt.Errorf("cilium preflight check DS is not ready: %d want and %d ready", ciliumDaemonSet.Status.NumberReady, preflightDaemonSet.Status.NumberReady)
	}
	return nil
}

func CheckDeploymentReady(deployment *v1.Deployment) error {
	if err := checkDeploymentObservedGeneration(deployment); err != nil {
		return err
	}

	if deployment.Status.Replicas != deployment.Status.ReadyReplicas {
		return fmt.Errorf("deployment %s is not ready: %d/%d ready", deployment.Name, deployment.Status.ReadyReplicas, deployment.Status.Replicas)
	}
	return nil
}

func checkDaemonSetObservedGeneration(daemonSet *v1.DaemonSet) error {
	observedGeneration := daemonSet.Status.ObservedGeneration
	generation := daemonSet.Generation
	if observedGeneration != generation {
		return fmt.Errorf("daemonSet %s status needs to be refreshed: observed generation is %d, want %d", daemonSet.Name, observedGeneration, generation)
	}

	return nil
}

func checkDeploymentObservedGeneration(deployment *v1.Deployment) error {
	observedGeneration := deployment.Status.ObservedGeneration
	generation := deployment.Generation
	if observedGeneration != generation {
		return fmt.Errorf("deployment %s status needs to be refreshed: observed generation is %d, want %d", deployment.Name, observedGeneration, generation)
	}

	return nil
}
