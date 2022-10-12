package common_test

import (
	"testing"

	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

func TestGetAuditPolicy(t *testing.T) {
	tests := []struct {
		testName       string
		kubeVersion    anywherev1.KubernetesVersion
		wantAPIVersion string
	}{
		{
			testName:       "kube 1.20",
			kubeVersion:    anywherev1.Kube120,
			wantAPIVersion: "audit.k8s.io/v1beta1",
		},
		{
			testName:       "kube 1.21",
			kubeVersion:    anywherev1.Kube121,
			wantAPIVersion: "audit.k8s.io/v1beta1",
		},
		{
			testName:       "kube 1.22",
			kubeVersion:    anywherev1.Kube122,
			wantAPIVersion: "audit.k8s.io/v1beta1",
		},
		{
			testName:       "kube 1.23",
			kubeVersion:    anywherev1.Kube123,
			wantAPIVersion: "audit.k8s.io/v1beta1",
		},
		{
			testName:       "kube 1.24",
			kubeVersion:    anywherev1.Kube124,
			wantAPIVersion: "audit.k8s.io/v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			a, _ := common.GetAuditPolicy(tt.kubeVersion)
			g.Expect(a.APIVersion).To(Equal(tt.wantAPIVersion))
		})
	}
}
