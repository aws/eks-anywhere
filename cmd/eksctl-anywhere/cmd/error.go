package cmd

import "fmt"

type missingKubeconfigError struct {
	path        string
	clusterName string
}

func (m missingKubeconfigError) Error() string {
	return fmt.Sprintf("kubeconfig missing for cluster %v: kubeconfig path=%v", m.clusterName, m.path)
}
