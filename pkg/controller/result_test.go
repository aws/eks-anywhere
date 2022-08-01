package controller_test

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aws/eks-anywhere/pkg/controller"
)

func TestResultToCtrlResult(t *testing.T) {
	tests := []struct {
		name string
		in   controller.Result
		want ctrl.Result
	}{
		{
			name: "no result",
			in:   controller.Result{},
			want: ctrl.Result{},
		},
		{
			name: "requeue result",
			in: controller.Result{
				Result: &ctrl.Result{
					Requeue: true,
				},
			},
			want: ctrl.Result{
				Requeue: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.in.ToCtrlResult()).To(Equal(tt.want))
		})
	}
}

func TestResultToCtrlReturn(t *testing.T) {
	tests := []struct {
		name string
		in   controller.Result
		want bool
	}{
		{
			name: "no return",
			in:   controller.Result{},
			want: false,
		},
		{
			name: "return",
			in: controller.Result{
				Result: &ctrl.Result{
					Requeue: true,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.in.Return()).To(Equal(tt.want))
		})
	}
}

func TestResultWithReturn(t *testing.T) {
	g := NewWithT(t)
	r := controller.ResultWithReturn()
	g.Expect(r.Return()).To(BeTrue())
	g.Expect(r.ToCtrlResult().Requeue).To(BeFalse())
}

func ResultWithRequeue(t *testing.T) {
	g := NewWithT(t)
	r := controller.ResultWithRequeue(2 * time.Second)
	g.Expect(r.Return()).To(BeTrue())
	g.Expect(r.ToCtrlResult().RequeueAfter).To(Equal(2 * time.Second))
}
