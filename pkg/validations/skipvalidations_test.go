package validations_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

func TestValidateSkippableValidation(t *testing.T) {
	tests := []struct {
		name                 string
		want                 map[string]bool
		wantErr              error
		skippedValidations   []string
		skippableValidations []string
	}{
		{
			name:                 "invalid upgrade validation param",
			want:                 nil,
			wantErr:              fmt.Errorf("invalid validation name to be skipped. The supported validations that can be skipped using --skip-validations are %s", strings.Join(upgradevalidations.SkippableValidations[:], ",")),
			skippedValidations:   []string{"test"},
			skippableValidations: upgradevalidations.SkippableValidations,
		},
		{
			name: "valid upgrade validation param",
			want: map[string]bool{
				validations.PDB:             true,
				validations.VSphereUserPriv: false,
				validations.EksaVersionSkew: false,
			},
			wantErr:              nil,
			skippedValidations:   []string{validations.PDB},
			skippableValidations: upgradevalidations.SkippableValidations,
		},
		{
			name: "valid create validation param",
			want: map[string]bool{
				validations.VSphereUserPriv: true,
			},
			wantErr:              nil,
			skippedValidations:   []string{validations.VSphereUserPriv},
			skippableValidations: createvalidations.SkippableValidations,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validations.ValidateSkippableValidation(tt.skippedValidations, tt.skippableValidations)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidateSkippableValidation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateSkippableValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}
