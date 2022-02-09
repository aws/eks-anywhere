package framework

import (
	"os"
	"testing"
)

func CheckRequiredEnvVars(t *testing.T, requiredEnvVars []string) {
	for _, eVar := range requiredEnvVars {
		if _, ok := os.LookupEnv(eVar); !ok {
			t.Fatalf("Required env var [%s] not present", eVar)
		}
	}
}

func setKubeconfigEnvVar(t *testing.T, clusterName string) {
	err := os.Setenv("KUBECONFIG", clusterName+"/"+clusterName+"-eks-a-cluster.kubeconfig")
	if err != nil {
		t.Fatalf("Error setting KUBECONFIG env var: %v", err)
	}
}
