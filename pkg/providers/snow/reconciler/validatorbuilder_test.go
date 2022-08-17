package reconciler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
	awsMocks "github.com/aws/eks-anywhere/pkg/aws/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler/mocks"
)

func TestBuildSuccess(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	ctrl := gomock.NewController(t)

	mockaws := aws.NewClientFromEC2(awsMocks.NewMockEC2Client(ctrl))
	awsClients := aws.Clients{
		"device-1": mockaws,
		"device-2": mockaws,
	}
	clientBuilder := mocks.NewMockClientRegistry(ctrl)
	clientBuilder.EXPECT().Get(ctx).Return(awsClients, nil)
	validatorBuilder := reconciler.NewValidatorBuilder(clientBuilder)

	validator, err := validatorBuilder.Build(ctx)
	g.Expect(validator).NotTo(BeNil())
	g.Expect(err).To(BeNil())
}

func TestBuildFailure(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	ctrl := gomock.NewController(t)

	clientBuilder := mocks.NewMockClientRegistry(ctrl)
	clientBuilder.EXPECT().Get(ctx).Return(nil, errors.New("test error"))
	validatorBuilder := reconciler.NewValidatorBuilder(clientBuilder)

	validator, err := validatorBuilder.Build(ctx)
	g.Expect(validator).To(BeNil())
	g.Expect(err).NotTo(BeNil())
}
