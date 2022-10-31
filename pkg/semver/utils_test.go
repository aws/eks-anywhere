package semver_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/semver"
)

func TestKubeVersionToValidSemver(t *testing.T) {
	type args struct {
		kubeVersion v1alpha1.KubernetesVersion
	}
	tests := []struct {
		name    string
		args    args
		want    *semver.Version
		wantErr error
	}{
		{
			name: "convert kube 1.22",
			args: args{
				kubeVersion: v1alpha1.Kube122,
			},
			want: &semver.Version{
				Major: 1,
				Minor: 22,
				Patch: 0,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semver.KubeVersionToValidSemver(tt.args.kubeVersion)
			if err != tt.wantErr {
				t.Errorf("KubeVersionToValidSemver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KubeVersionToValidSemver() = %v, want %v", got, tt.want)
			}
		})
	}
}
