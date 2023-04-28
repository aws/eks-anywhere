package cloudstack

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/collection"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

// GetCloudstackExecConfig gets cloudstack exec config from secrets.
func GetCloudstackExecConfig(ctx context.Context, cli client.Client, datacenterConfig *v1alpha1.CloudStackDatacenterConfig) (*decoder.CloudStackExecConfig, error) {
	var profiles []decoder.CloudStackProfileConfig
	credRefs := collection.NewSet[string]()
	for _, zone := range datacenterConfig.Spec.AvailabilityZones {
		credRefs.Add(zone.CredentialsRef)
	}
	for _, profileName := range credRefs.ToSlice() {
		secret := &apiv1.Secret{}
		secretKey := client.ObjectKey{
			Namespace: constants.EksaSystemNamespace,
			Name:      profileName,
		}
		if err := cli.Get(ctx, secretKey, secret); err != nil {
			return nil, err
		}
		profiles = append(profiles, decoder.CloudStackProfileConfig{
			Name:          profileName,
			ApiKey:        string(secret.Data[decoder.APIKeyKey]),
			SecretKey:     string(secret.Data[decoder.SecretKeyKey]),
			ManagementUrl: string(secret.Data[decoder.APIUrlKey]),
			VerifySsl:     string(secret.Data[decoder.VerifySslKey]),
		})
	}

	return &decoder.CloudStackExecConfig{
		Profiles: profiles,
	}, nil
}
