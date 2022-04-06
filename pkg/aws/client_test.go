package aws_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type awsTest struct {
	*WithT
	ctx context.Context
}

func newAwsTest(t *testing.T) *awsTest {
	return &awsTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
	}
}

func TestLoadConfig(t *testing.T) {
	tt := newAwsTest(t)
	_, err := aws.LoadConfig(tt.ctx)
	tt.Expect(err).To(Succeed())
}
