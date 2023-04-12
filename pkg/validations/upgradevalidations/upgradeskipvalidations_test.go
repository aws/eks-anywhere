package upgradevalidations_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

func TestValidateSkippableUpgradeValidation(t *testing.T) {
	tests := []struct {
		name               string
		want               map[string]bool
		wantErr            error
		skippedValidations []string
	}{
		{
			name:               "invalid upgrade validation param",
			want:               nil,
			wantErr:            fmt.Errorf("invalid validation name to be skipped. The supported upgrade validations that can be skipped using --skip-validations are %s", strings.Join(upgradevalidations.SkippableValidations[:], ",")),
			skippedValidations: []string{"test"},
		},
		{
			name: "valid upgrade validation param",
			want: map[string]bool{
				upgradevalidations.PDB: true,
			},
			wantErr:            nil,
			skippedValidations: []string{upgradevalidations.PDB},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := upgradevalidations.ValidateSkippableUpgradeValidation(tt.skippedValidations)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidateSkippableUpgradeValidation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateSkippableUpgradeValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}
