package awsiam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
)

const kubectl = "kubectl"

// GetAllPods uses the kubectl installed on the host machine
// Modifies the command context to add ./bin to path
func GetAllPods(ctx context.Context, opts ...string) ([]corev1.Pod, error) {
	var stdout, stderr bytes.Buffer
	args := []string{"get", "po", "-o", "json", "-A"}
	cmd := exec.CommandContext(ctx, kubectl, append(args, opts...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	envPath := os.Getenv("PATH")
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error finding current working directory: %v", err)
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s/bin:%s", workDir, envPath))

	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("error running %s command: %v", kubectl, stderr.String())
		}
	}
	response := &corev1.PodList{}
	err = json.Unmarshal(stdout.Bytes(), response)
	if err != nil {
		return nil, fmt.Errorf("error parsing get pods response: %v", err)
	}
	return response.Items, nil
}
