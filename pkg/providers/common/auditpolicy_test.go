package common

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	. "github.com/onsi/gomega"
	"reflect"
	"testing"
)

func TestGetAuditPolicy(t *testing.T) {
	tests := []struct {
		testName string
		k8sVersion v1alpha1.KubernetesVersion
		wantPolicy string
		wantErr error
	}{
		{
			testName: "k8s version < 1.24 uses v1beta1",
			k8sVersion: v1alpha1.Kube123,
			wantPolicy: "audit.k8s.io/v1beta1",
			wantErr: nil,
		},
		{
			testName: "k8s version = 1.24 uses v1",
			k8sVersion: v1alpha1.Kube124,
			wantPolicy: "audit.k8s.io/v1\n",
			wantErr: nil,
		},
		{
			testName: "k8s version > 1.24 uses v1",
			k8sVersion: v1alpha1.Kube125,
			wantPolicy: "audit.k8s.io/v1\n",
			wantErr: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			g := NewWithT(t)
			gotPolicy, gotErr := GetAuditPolicy(tc.k8sVersion)
			g.Expect(gotPolicy).To(ContainSubstring(tc.wantPolicy))
			if !reflect.DeepEqual(tc.wantErr, gotErr) {
				t.Errorf("%v got = %v, want %v", tc.testName, gotErr, tc.wantErr)
			}
		})
	}
}
