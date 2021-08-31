package v1alpha1_test

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestGetDockerDatacenterConfig(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    *v1alpha1.DockerDatacenterConfig
		wantErr bool
	}{
		{
			name: "Good Docker cluster config parse",
			args: args{
				fileName: "testdata/cluster_docker.yaml",
			},
			wantErr: false,
			want: &v1alpha1.DockerDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.DockerDatacenterKind,
					APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
			},
		},
		{
			name: "Non existent Docker file",
			args: args{
				fileName: "testdata/cluster_docker_nonexistent.yaml",
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "Bad Docker cluster config",
			args: args{
				fileName: "testdata/cluster_vsphere.yaml",
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v1alpha1.GetDockerDatacenterConfig(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDockerDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDockerDatacenterConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
