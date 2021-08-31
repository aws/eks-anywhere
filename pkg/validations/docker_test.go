package validations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

const (
	requiredMajorVersion = 20
)

func TestValidateDockerVersion(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		wantErr       error
		dockerVersion int
	}{
		{
			name:          "FailureDockerVersion10",
			dockerVersion: 19,
			wantErr:       fmt.Errorf("minimum requirements for docker version have not been met. Install Docker version %d.x.x or above", requiredMajorVersion),
		},
		{
			name:          "SuccessDockerVersion20",
			dockerVersion: 20,
			wantErr:       nil,
		},
		{
			name:          "SuccessDockerVersion22",
			dockerVersion: 22,
			wantErr:       nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			mockCtrl := gomock.NewController(t)
			dockerExecutableMock := mocks.NewMockDockerExecutable(mockCtrl)
			dockerExecutableMock.EXPECT().Version(ctx).Return(tc.dockerVersion, tc.wantErr)
			err := validations.CheckMinimumDockerVersion(ctx, dockerExecutableMock)
			if err != tc.wantErr {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
