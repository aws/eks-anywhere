package framework

import (
	"os"
)

func checkRequiredEnvVars(t T, requiredEnvVars []string) {
	for _, eVar := range requiredEnvVars {
		if _, ok := os.LookupEnv(eVar); !ok {
			t.Fatalf("Required env var [%s] not present", eVar)
		}
	}
}

func setKubeconfigEnvVar(t T, clusterName string) {
	err := os.Setenv("KUBECONFIG", clusterName+"/"+clusterName+"-eks-a-cluster.kubeconfig")
	if err != nil {
		t.Fatalf("Error setting KUBECONFIG env var: %v", err)
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
