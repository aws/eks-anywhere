package framework

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

//go:embed testdata/endpoint_content_check.sh
var endpointContentCheckScript string

// VerifyEndpointServesContent checks from inside the cluster, retrying for up to 10 minutes, that the
// response body of url contains expected. The check job runs in namespace.
func (e *ClusterE2ETest) VerifyEndpointServesContent(namespace, url, expected string) {
	e.T.Logf("Verifying %s returns %q", url, expected)
	container := corev1.Container{
		Name:    "endpoint-check",
		Image:   e.GetDiagnosticCollectorImage(),
		Command: []string{"bash", "-c", endpointContentCheckScript, "endpoint-check", url, expected},
	}
	e.runCheckJob(namespace, "endpoint-check", container, 12*time.Minute)
}

// runCheckJob runs container as a single-attempt Job in the cluster under test, logs its output,
// deletes it, and fails the test if it doesn't succeed within timeout.
func (e *ClusterE2ETest) runCheckJob(namespace, namePrefix string, container corev1.Container, timeout time.Duration) {
	ctx := context.Background()
	kubeconfig := e.KubeconfigFilePath()
	name := fmt.Sprintf("%s-%s", namePrefix, utilrand.String(7))

	manifest, err := yaml.Marshal(newCheckJob(namespace, name, container, timeout))
	if err != nil {
		e.T.Fatalf("Marshalling job %s: %v", name, err)
	}

	e.T.Logf("Running check job %s/%s", namespace, name)
	if err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, e.Cluster(), manifest, namespace); err != nil {
		e.T.Fatalf("Creating job %s/%s: %v", namespace, name, err)
	}
	defer func() {
		if _, err := e.KubectlClient.ExecuteCommand(ctx, "delete", "job", name, "--namespace", namespace,
			"--kubeconfig", kubeconfig, "--ignore-not-found", "--wait=false"); err != nil {
			e.T.Logf("Deleting job %s/%s: %v", namespace, name, err)
		}
	}()

	succeeded := false
	waitErr := retrier.New(timeout+time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(10*time.Second))).Retry(func() error {
		j := &batchv1.Job{}
		if err := e.KubectlClient.GetObject(ctx, "jobs.batch", name, namespace, kubeconfig, j); err != nil {
			return err
		}
		var finished bool
		if finished, succeeded = jobResult(j); !finished {
			return fmt.Errorf("job %s/%s has not finished", namespace, name)
		}
		return nil
	})

	logs, err := e.KubectlClient.ExecuteCommand(ctx, "logs", "job/"+name, "--namespace", namespace, "--kubeconfig", kubeconfig)
	if err != nil {
		e.T.Logf("Getting logs of job %s/%s: %v", namespace, name, err)
	}
	e.T.Logf("Logs of job %s/%s:\n%s", namespace, name, logs.String())

	if waitErr != nil {
		e.T.Fatalf("Waiting for job %s/%s: %v", namespace, name, waitErr)
	}
	if !succeeded {
		e.T.Fatalf("Job %s/%s failed", namespace, name)
	}
}

func newCheckJob(namespace, name string, container corev1.Container, timeout time.Duration) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta:   metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: batchv1.JobSpec{
			BackoffLimit:          ptr.Int32(0),
			ActiveDeadlineSeconds: ptr.Int64(int64(timeout.Seconds())),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    []corev1.Container{container},
				},
			},
		},
	}
}

// jobResult reports whether a single-attempt Job finished and whether it succeeded.
func jobResult(j *batchv1.Job) (finished, succeeded bool) {
	if j.Status.Succeeded > 0 {
		return true, true
	}
	if j.Status.Failed > 0 {
		return true, false
	}
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return true, false
		}
	}
	return false, false
}
