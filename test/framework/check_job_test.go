package framework

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestJobResult(t *testing.T) {
	tests := map[string]struct {
		status              batchv1.JobStatus
		finished, succeeded bool
	}{
		"running":   {status: batchv1.JobStatus{Active: 1}},
		"succeeded": {status: batchv1.JobStatus{Succeeded: 1}, finished: true, succeeded: true},
		"failed":    {status: batchv1.JobStatus{Failed: 1}, finished: true},
		"deadline exceeded": {
			status:   batchv1.JobStatus{Conditions: []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue}}},
			finished: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			g := NewWithT(t)
			finished, succeeded := jobResult(&batchv1.Job{Status: tt.status})
			g.Expect(finished).To(Equal(tt.finished))
			g.Expect(succeeded).To(Equal(tt.succeeded))
		})
	}
}

func TestNewCheckJob(t *testing.T) {
	g := NewWithT(t)
	job := newCheckJob("harbor", "harbor-check-abc", corev1.Container{Name: "c", Image: "img"}, 15*time.Minute)
	g.Expect(job.Namespace).To(Equal("harbor"))
	g.Expect(*job.Spec.BackoffLimit).To(BeEquivalentTo(0))
	g.Expect(*job.Spec.ActiveDeadlineSeconds).To(BeEquivalentTo(900))
	g.Expect(job.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
	g.Expect(job.Spec.Template.Spec.Containers).To(HaveLen(1))
}
