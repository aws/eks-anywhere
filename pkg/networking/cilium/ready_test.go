package cilium

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckDaemonSetReady(t *testing.T) {
	tests := []struct {
		name      string
		daemonSet *v1.DaemonSet
		wantErr   error
	}{
		{
			name: "ready",
			daemonSet: &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ds",
				},
				Status: v1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            5,
				},
			},
		},
		{
			name: "not ready",
			daemonSet: &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ds",
				},
				Status: v1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            4,
				},
			},
			wantErr: errors.New("daemonSet ds is not ready: 4/5 ready"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := checkDaemonSetReady(tt.daemonSet)
			if tt.wantErr != nil {
				g.Expect(err).To(MatchError(tt.wantErr))
				return
			}
			g.Expect(err).To(Succeed())
		})
	}
}

func TestCheckPreflightDaemonSetReady(t *testing.T) {
	tests := []struct {
		name              string
		cilium, preflight *v1.DaemonSet
		wantErr           error
	}{
		{
			name: "ready",
			cilium: &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ds",
				},
				Status: v1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            5,
				},
			},
			preflight: &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ds-pre",
				},
				Status: v1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            5,
				},
			},
		},
		{
			name: "not ready",
			cilium: &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ds",
				},
				Status: v1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            5,
				},
			},
			preflight: &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ds-pre",
				},
				Status: v1.DaemonSetStatus{
					DesiredNumberScheduled: 5,
					NumberReady:            4,
				},
			},
			wantErr: errors.New("cilium preflight check DS is not ready: 5 want and 4 ready"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := checkPreflightDaemonSetReady(tt.cilium, tt.preflight)
			if tt.wantErr != nil {
				g.Expect(err).To(MatchError(tt.wantErr))
				return
			}
			g.Expect(err).To(Succeed())
		})
	}
}

func TestCheckDeploymentReady(t *testing.T) {
	tests := []struct {
		name       string
		deployment *v1.Deployment
		wantErr    error
	}{
		{
			name: "ready",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dep",
				},
				Status: v1.DeploymentStatus{
					Replicas:      5,
					ReadyReplicas: 5,
				},
			},
		},
		{
			name: "not ready",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dep",
				},
				Status: v1.DeploymentStatus{
					Replicas:      5,
					ReadyReplicas: 4,
				},
			},
			wantErr: errors.New("deployment dep is not ready: 4/5 ready"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := checkDeploymentReady(tt.deployment)
			if tt.wantErr != nil {
				g.Expect(err).To(MatchError(tt.wantErr))
				return
			}
			g.Expect(err).To(Succeed())
		})
	}
}
