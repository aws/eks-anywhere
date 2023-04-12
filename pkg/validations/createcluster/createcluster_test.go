package createcluster_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createcluster"
	createmocks "github.com/aws/eks-anywhere/pkg/validations/createcluster/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

type createClusterValidationTest struct {
	clusterSpec       *cluster.Spec
	ctx               context.Context
	kubectl           *mocks.MockKubectlClient
	provider          *providermocks.MockProvider
	flux              *flux.Flux
	docker            *mocks.MockDockerExecutable
	createValidations *createmocks.MockValidator
}

func newValidateTest(t *testing.T) *createClusterValidationTest {
	mockCtrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectlClient(mockCtrl)
	provider := providermocks.NewMockProvider(mockCtrl)
	docker := mocks.NewMockDockerExecutable(mockCtrl)
	createValidations := createmocks.NewMockValidator(mockCtrl)
	return &createClusterValidationTest{
		ctx:               context.Background(),
		kubectl:           kubectl,
		provider:          provider,
		docker:            docker,
		createValidations: createValidations,
	}
}

func (c *createClusterValidationTest) expectValidDockerClusterSpec() {
	s := test.NewClusterSpec()
	s.Cluster = &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test",
		},
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{},
				},
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: v1alpha1.Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
			},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 1,
			},
			DatacenterRef: v1alpha1.Ref{
				Kind: "DockerDatacenterConfig",
				Name: "eksa-unit-test",
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
			},
			KubernetesVersion: v1alpha1.GetClusterDefaultKubernetesVersion(),
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: "dev-cluster",
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(1),
			}},
		},
	}
	s.DockerDatacenter = &v1alpha1.DockerDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test",
		},
		Spec: v1alpha1.DockerDatacenterConfigSpec{},
	}
	c.clusterSpec = s
}

func (c *createClusterValidationTest) expectValidProvider() {
	c.provider.EXPECT().SetupAndValidateCreateCluster(c.ctx, c.clusterSpec).Return(nil).AnyTimes()
	c.provider.EXPECT().Name().Return("docker").AnyTimes()
}

func (c *createClusterValidationTest) expectValidDockerExec() {
	c.docker.EXPECT().Version(c.ctx).Return(21, nil).AnyTimes()
	c.docker.EXPECT().AllocatedMemory(c.ctx).Return(uint64(6200000001), nil).AnyTimes()
}

func (c *createClusterValidationTest) expectEmptyFlux() {
	c.flux = flux.NewFluxFromGitOpsFluxClient(nil, nil, nil, nil)
}

type validation struct {
	run bool
}

func (c *createClusterValidationTest) expectBuildValidations() *validation {
	v := &validation{}
	c.createValidations.EXPECT().PreflightValidations(c.ctx).Return(
		[]validations.Validation{
			func() *validations.ValidationResult {
				v.run = true
				return &validations.ValidationResult{
					Err: nil,
				}
			},
		},
	)

	return v
}

func TestCreateClusterValidationsSuccess(t *testing.T) {
	g := NewWithT(t)
	test := newValidateTest(t)
	test.expectValidDockerClusterSpec()
	test.expectValidProvider()
	test.expectEmptyFlux()
	test.expectValidDockerExec()
	validationFromBuild := test.expectBuildValidations()

	commandVal := createcluster.NewValidations(test.clusterSpec, test.provider, test.flux, test.createValidations, test.docker)

	g.Expect(commandVal.Validate(test.ctx)).To(Succeed())
	g.Expect(validationFromBuild.run).To(BeTrue(), "validation coming from BuildValidations should be run")
}
