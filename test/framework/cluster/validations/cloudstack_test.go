package validations_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/test/framework/cluster/validations"
)

const (
	azName1          = "test-az-1"
	azAccount        = "test-account-1"
	azDomain         = "test-domain-1"
	azCredentialsRef = "global"
)

func TestValidateAvailabilityZones(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	tests := []struct {
		name                 string
		availabilityZones    []v1alpha1.CloudStackAvailabilityZone
		failureDomainObjects []client.Object
		wantErr              string
	}{
		{
			name: "found az failure domain",
			availabilityZones: []v1alpha1.CloudStackAvailabilityZone{
				{
					Name:                  azName1,
					CredentialsRef:        azCredentialsRef,
					Domain:                azDomain,
					Account:               azAccount,
					ManagementApiEndpoint: "test-api-endpoint",
					Zone:                  v1alpha1.CloudStackZone{},
				},
			},
			failureDomainObjects: []client.Object{
				&cloudstackv1.CloudStackFailureDomain{
					TypeMeta: metav1.TypeMeta{
						Kind:       "CloudStackFailureDomain",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta3",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: constants.EksaSystemNamespace,
						Name:      cloudstackv1.FailureDomainHashedMetaName(azName1, clusterName),
					},
					Spec: cloudstackv1.CloudStackFailureDomainSpec{
						Name:    azName1,
						Zone:    cloudstackv1.CloudStackZoneSpec{},
						Account: azAccount,
						Domain:  azDomain,
						ACSEndpoint: corev1.SecretReference{
							Name:      azCredentialsRef,
							Namespace: constants.EksaSystemNamespace,
						},
					},
				},
			},
			wantErr: "",
		},

		{
			name: "missing az failure domain",
			availabilityZones: []v1alpha1.CloudStackAvailabilityZone{
				{
					Name:                  azName1,
					CredentialsRef:        azCredentialsRef,
					Domain:                azDomain,
					Account:               azAccount,
					ManagementApiEndpoint: "test-api-endpoint",
					Zone:                  v1alpha1.CloudStackZone{},
				},
			},
			failureDomainObjects: []client.Object{},
			wantErr:              "failed to find failure domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.CloudStackDatacenter = &v1alpha1.CloudStackDatacenterConfig{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: clusterNamespace,
						Name:      clusterName,
					},
					Spec: v1alpha1.CloudStackDatacenterConfigSpec{
						AvailabilityZones: tt.availabilityZones,
					},
				}
			})

			vt := newStateValidatorTest(t, spec)
			vt.createTestObjects(ctx)
			vt.createManagementClusterObjects(ctx, tt.failureDomainObjects...)
			err := validations.ValidateAvailabilityZones(ctx, vt.config)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
