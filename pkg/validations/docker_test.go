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

func TestCheckDockerDesktopVersion(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		wantErr              error
		dockerDesktopVersion string
		cgroupVersion        int
		needsCgroup          bool
	}{
		{
			name:                 "SuccessDockerDesktopVersion42",
			dockerDesktopVersion: "4.2.0",
			wantErr:              nil,
			cgroupVersion:        1,
			needsCgroup:          false,
		},
		{
			name:                 "FailureDockerDesktopVersion430",
			dockerDesktopVersion: "4.3.0",
			wantErr:              fmt.Errorf("EKS Anywhere does not support Docker desktop versions between 4.3.0 and 4.4.1 on macOS, please refer to https://github.com/aws/eks-anywhere/issues/789 for more information"),
			cgroupVersion:        2,
			needsCgroup:          false,
		},
		{
			name:                 "FailureDockerDesktopVersion440",
			dockerDesktopVersion: "4.4.0",
			wantErr:              fmt.Errorf("EKS Anywhere does not support Docker desktop versions between 4.3.0 and 4.4.1 on macOS, please refer to https://github.com/aws/eks-anywhere/issues/789 for more information"),
			cgroupVersion:        2,
			needsCgroup:          false,
		},
		{
			name:                 "FailureDockerDesktopVersion440",
			dockerDesktopVersion: "4.4.1",
			wantErr:              fmt.Errorf("EKS Anywhere does not support Docker desktop versions between 4.3.0 and 4.4.1 on macOS, please refer to https://github.com/aws/eks-anywhere/issues/789 for more information"),
			cgroupVersion:        2,
			needsCgroup:          false,
		},
		{
			name:                 "SuccessDockerDesktopVersion442",
			dockerDesktopVersion: "4.4.2",
			wantErr:              nil,
			cgroupVersion:        1,
			needsCgroup:          true,
		},
		{
			name:                 "FailureDockerDesktopVersion442",
			dockerDesktopVersion: "4.4.2",
			wantErr: fmt.Errorf("EKS Anywhere requires Docker desktop to be configured to use CGroups v1. " +
				"Please  set `deprecatedCgroupv1:true` in your `~/Library/Group\\ Containers/group.com.docker/settings.json` file"),
			cgroupVersion: 2,
			needsCgroup:   true,
		},
		{
			name:                 "SuccessDockerDesktopVersion450",
			dockerDesktopVersion: "4.5.0",
			wantErr:              nil,
			cgroupVersion:        1,
			needsCgroup:          true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			mockCtrl := gomock.NewController(t)
			dockerExecutableMock := mocks.NewMockDockerExecutable(mockCtrl)
			if tc.needsCgroup {
				dockerExecutableMock.EXPECT().CgroupVersion(ctx).Return(tc.cgroupVersion, tc.wantErr)
			}
			err := validations.ValidateDockerDesktopVersion(ctx, dockerExecutableMock, tc.dockerDesktopVersion)
			if err != nil && err.Error() != tc.wantErr.Error() {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateDockerExecutable(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		wantErr              error
		dockerVersion        int
		dockerDesktopVersion int
	}{
		{
			name:          "SuccessDockerExecutable",
			wantErr:       nil,
			dockerVersion: 21,
		},
		{
			name:          "FailureUnderMinDockerVersion",
			wantErr:       fmt.Errorf("failed to validate docker: minimum requirements for docker version have not been met. Install Docker version 20.x.x or above"),
			dockerVersion: 19,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			mockCtrl := gomock.NewController(t)
			dockerExecutableMock := mocks.NewMockDockerExecutable(mockCtrl)
			dockerExecutableMock.EXPECT().Version(ctx).Return(tc.dockerVersion, nil).AnyTimes()
			dockerExecutableMock.EXPECT().AllocatedMemory(ctx).Return(uint64(6200000001), nil).AnyTimes()
			err := validations.ValidateDockerExecutable(ctx, dockerExecutableMock, "linux")
			if err != nil && err.Error() != tc.wantErr.Error() {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
