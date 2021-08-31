package v1alpha1_test

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestGetAWSDatacenterConfig(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    *v1alpha1.AWSDatacenterConfig
		wantErr bool
	}{
		{
			name: "Good AWS cluster config parse",
			args: args{
				fileName: "testdata/cluster_aws.yaml",
			},
			wantErr: false,
			want: &v1alpha1.AWSDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       v1alpha1.AWSDatacenterKind,
					APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: v1alpha1.AWSDatacenterConfigSpec{
					AmiID:  "my-image",
					Region: "us-west",
				},
			},
		},
		{
			name: "Non existent AWS file",
			args: args{
				fileName: "testdata/cluster_nonexistent.yaml",
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "Bad AWS cluster config",
			args: args{
				fileName: "testdata/cluster_vsphere.yaml",
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v1alpha1.GetAWSDatacenterConfig(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAWSDatacenterConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAWSDatacenterConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
