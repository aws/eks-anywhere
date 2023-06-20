package v1alpha1_test

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/version"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestValidateVersionSkew(t *testing.T) {
	v122, _ := version.ParseGeneric(string(v1alpha1.Kube122))
	v123, _ := version.ParseGeneric(string(v1alpha1.Kube123))
	v124, _ := version.ParseGeneric(string(v1alpha1.Kube124))

	tests := []struct {
		name       string
		oldVersion *version.Version
		newVersion *version.Version
		wantErr    error
	}{
		{
			name:       "No upgrade",
			oldVersion: v122,
			newVersion: v122,
			wantErr:    nil,
		},
		{
			name:       "Minor version increment success",
			oldVersion: v122,
			newVersion: v123,
			wantErr:    nil,
		},
		{
			name:       "Minor version invalid, failure",
			oldVersion: v122,
			newVersion: v124,
			wantErr:    fmt.Errorf("only +%d minor version skew is supported, minor version skew detected 2", v1alpha1.SupportedMinorVersionIncrement),
		},
		{
			name:       "Minor version downgrade, failure",
			oldVersion: v124,
			newVersion: v123,
			wantErr:    fmt.Errorf("kubernetes version downgrade is not supported (%s) -> (%s)", v124, v123),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v1alpha1.ValidateVersionSkew(tt.oldVersion, tt.newVersion)
			if err != nil && !reflect.DeepEqual(err.Error(), tt.wantErr.Error()) {
				t.Errorf("ValidateVersionSkew() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
