package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name                      string
		yamlManifest              []byte
		wantCluster               *anywherev1.Cluster
		wantVsphereDatacenter     *anywherev1.VSphereDatacenterConfig
		wantDockerDatacenter      *anywherev1.DockerDatacenterConfig
		wantSnowDatacenter        *anywherev1.SnowDatacenterConfig
		wantVsphereMachineConfigs []*anywherev1.VSphereMachineConfig
		wantSnowMachineConfigs    []*anywherev1.SnowMachineConfig
		wantOIDCConfigs           []*anywherev1.OIDCConfig
		wantAWSIamConfigs         []*anywherev1.AWSIamConfig
		wantGitOpsConfig          *anywherev1.GitOpsConfig
		wantFluxConfig            *anywherev1.FluxConfig
	}{
		{
			name:         "vsphere cluster",
			yamlManifest: []byte(test.ReadFile(t, "testdata/cluster_1_19.yaml")),
			wantCluster: &anywherev1.Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.ClusterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.19",
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count:    1,
						Endpoint: &anywherev1.Endpoint{Host: "myHostIp"},
						MachineGroupRef: &anywherev1.Ref{
							Kind: "VSphereMachineConfig",
							Name: "eksa-unit-test-cp",
						},
					},
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:  "workers-1",
							Count: 1,
							MachineGroupRef: &anywherev1.Ref{
								Kind: "VSphereMachineConfig",
								Name: "eksa-unit-test",
							},
						},
					},
					DatacenterRef: anywherev1.Ref{
						Kind: "VSphereDatacenterConfig",
						Name: "eksa-unit-test",
					},
					ClusterNetwork: anywherev1.ClusterNetwork{
						Pods: anywherev1.Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: anywherev1.Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
						CNI: "cilium",
					},
				},
			},
			wantVsphereDatacenter: &anywherev1.VSphereDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.VSphereDatacenterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.VSphereDatacenterConfigSpec{
					Datacenter: "myDatacenter",
					Network:    "myNetwork",
					Server:     "myServer",
					Thumbprint: "myTlsThumbprint",
					Insecure:   false,
				},
			},
			wantVsphereMachineConfigs: []*anywherev1.VSphereMachineConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.VSphereMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-cp",
					},
					Spec: anywherev1.VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  anywherev1.Ubuntu,
						Users: []anywherev1.UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.VSphereMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.VSphereMachineConfigSpec{
						DiskGiB:   25,
						MemoryMiB: 8192,
						NumCPUs:   2,
						OSFamily:  anywherev1.Ubuntu,
						Users: []anywherev1.UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
		},
		{
			name:         "snow cluster",
			yamlManifest: []byte(test.ReadFile(t, "testdata/cluster_snow_1_21.yaml")),
			wantCluster: &anywherev1.Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.ClusterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.21",
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count:    1,
						Endpoint: &anywherev1.Endpoint{Host: "myHostIp"},
						MachineGroupRef: &anywherev1.Ref{
							Kind: "SnowMachineConfig",
							Name: "eksa-unit-test-cp",
						},
					},
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:  "workers-1",
							Count: 1,
							MachineGroupRef: &anywherev1.Ref{
								Kind: "SnowMachineConfig",
								Name: "eksa-unit-test",
							},
						},
					},
					DatacenterRef: anywherev1.Ref{
						Kind: "SnowDatacenterConfig",
						Name: "eksa-unit-test",
					},
					ClusterNetwork: anywherev1.ClusterNetwork{
						Pods: anywherev1.Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: anywherev1.Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
						CNI: "cilium",
					},
				},
			},
			wantSnowDatacenter: &anywherev1.SnowDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.SnowDatacenterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.SnowDatacenterConfigSpec{},
			},
			wantSnowMachineConfigs: []*anywherev1.SnowMachineConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.SnowMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-cp",
					},
					Spec: anywherev1.SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-ami",
						InstanceType: "sbe-c.large",
						SshKeyName:   "default",
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.SnowMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-ami",
						InstanceType: "sbe-c.xlarge",
						SshKeyName:   "default",
					},
				},
			},
		},
		{
			name:         "docker cluster with oidc, awsiam and gitops",
			yamlManifest: []byte(test.ReadFile(t, "testdata/docker_cluster_oidc_awsiam_gitops.yaml")),
			wantCluster: &anywherev1.Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.ClusterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "m-docker",
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.21",
					ManagementCluster: anywherev1.ManagementCluster{
						Name: "m-docker",
					},
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:  "workers-1",
							Count: 1,
						},
					},
					DatacenterRef: anywherev1.Ref{
						Kind: anywherev1.DockerDatacenterKind,
						Name: "m-docker",
					},
					ClusterNetwork: anywherev1.ClusterNetwork{
						Pods: anywherev1.Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: anywherev1.Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
						CNI: "cilium",
					},
					IdentityProviderRefs: []anywherev1.Ref{
						{
							Kind: anywherev1.OIDCConfigKind,
							Name: "eksa-unit-test",
						},
						{
							Kind: anywherev1.AWSIamConfigKind,
							Name: "eksa-unit-test",
						},
					},
					GitOpsRef: &anywherev1.Ref{
						Kind: anywherev1.GitOpsConfigKind,
						Name: "eksa-unit-test",
					},
				},
			},
			wantDockerDatacenter: &anywherev1.DockerDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.DockerDatacenterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "m-docker",
				},
			},
			wantGitOpsConfig: &anywherev1.GitOpsConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "GitOpsConfig",
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.GitOpsConfigSpec{
					Flux: anywherev1.Flux{
						Github: anywherev1.Github{
							Owner:      "janedoe",
							Repository: "flux-fleet",
						},
					},
				},
			},
			wantOIDCConfigs: []*anywherev1.OIDCConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "OIDCConfig",
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.OIDCConfigSpec{
						ClientId:     "id12",
						GroupsClaim:  "claim1",
						GroupsPrefix: "prefix-for-groups",
						IssuerUrl:    "https://mydomain.com/issuer",
						RequiredClaims: []anywherev1.OIDCConfigRequiredClaim{
							{
								Claim: "sub",
								Value: "test",
							},
						},
						UsernameClaim:  "username-claim",
						UsernamePrefix: "username-prefix",
					},
				},
			},
			wantAWSIamConfigs: []*anywherev1.AWSIamConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.AWSIamConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.AWSIamConfigSpec{
						AWSRegion:   "test-region",
						BackendMode: []string{"mode1", "mode2"},
						MapRoles: []anywherev1.MapRoles{
							{
								RoleARN:  "test-role-arn",
								Username: "test",
								Groups:   []string{"group1", "group2"},
							},
						},
						MapUsers: []anywherev1.MapUsers{
							{
								UserARN:  "test-user-arn",
								Username: "test",
								Groups:   []string{"group1", "group2"},
							},
						},
						Partition: "aws",
					},
				},
			},
		},
		{
			name:         "docker cluster with oidc, awsiam and flux",
			yamlManifest: []byte(test.ReadFile(t, "testdata/docker_cluster_oidc_awsiam_flux.yaml")),
			wantCluster: &anywherev1.Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.ClusterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "m-docker",
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.21",
					ManagementCluster: anywherev1.ManagementCluster{
						Name: "m-docker",
					},
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:  "workers-1",
							Count: 1,
						},
					},
					DatacenterRef: anywherev1.Ref{
						Kind: anywherev1.DockerDatacenterKind,
						Name: "m-docker",
					},
					ClusterNetwork: anywherev1.ClusterNetwork{
						Pods: anywherev1.Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: anywherev1.Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
						CNI: "cilium",
					},
					IdentityProviderRefs: []anywherev1.Ref{
						{
							Kind: anywherev1.OIDCConfigKind,
							Name: "eksa-unit-test",
						},
						{
							Kind: anywherev1.AWSIamConfigKind,
							Name: "eksa-unit-test",
						},
					},
					GitOpsRef: &anywherev1.Ref{
						Kind: anywherev1.FluxConfigKind,
						Name: "eksa-unit-test",
					},
				},
			},
			wantDockerDatacenter: &anywherev1.DockerDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.DockerDatacenterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "m-docker",
				},
			},
			wantFluxConfig: &anywherev1.FluxConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "FluxConfig",
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.FluxConfigSpec{
					Github: &anywherev1.GithubProviderConfig{
						Owner:      "janedoe",
						Repository: "flux-fleet",
					},
				},
			},
			wantOIDCConfigs: []*anywherev1.OIDCConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       "OIDCConfig",
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.OIDCConfigSpec{
						ClientId:     "id12",
						GroupsClaim:  "claim1",
						GroupsPrefix: "prefix-for-groups",
						IssuerUrl:    "https://mydomain.com/issuer",
						RequiredClaims: []anywherev1.OIDCConfigRequiredClaim{
							{
								Claim: "sub",
								Value: "test",
							},
						},
						UsernameClaim:  "username-claim",
						UsernamePrefix: "username-prefix",
					},
				},
			},
			wantAWSIamConfigs: []*anywherev1.AWSIamConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.AWSIamConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.AWSIamConfigSpec{
						AWSRegion:   "test-region",
						BackendMode: []string{"mode1", "mode2"},
						MapRoles: []anywherev1.MapRoles{
							{
								RoleARN:  "test-role-arn",
								Username: "test",
								Groups:   []string{"group1", "group2"},
							},
						},
						MapUsers: []anywherev1.MapUsers{
							{
								UserARN:  "test-user-arn",
								Username: "test",
								Groups:   []string{"group1", "group2"},
							},
						},
						Partition: "aws",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := cluster.ParseConfig(tt.yamlManifest)

			g.Expect(err).To(Not(HaveOccurred()))

			g.Expect(got.Cluster).To(Equal(tt.wantCluster))
			g.Expect(got.VSphereDatacenter).To(Equal(tt.wantVsphereDatacenter))
			g.Expect(got.DockerDatacenter).To(Equal(tt.wantDockerDatacenter))
			g.Expect(got.SnowDatacenter).To(Equal(tt.wantSnowDatacenter))
			g.Expect(len(got.VSphereMachineConfigs)).To(Equal(len(tt.wantVsphereMachineConfigs)), "it should return the right number of VSphereMachineConfigs")
			g.Expect(len(got.SnowMachineConfigs)).To(Equal(len(tt.wantSnowMachineConfigs)), "it should return the right number of SnowMachineConfigs")
			for _, m := range tt.wantVsphereMachineConfigs {
				g.Expect(got.VsphereMachineConfig(m.Name)).To(Equal(m))
			}
			for _, m := range tt.wantSnowMachineConfigs {
				g.Expect(got.SnowMachineConfig(m.Name)).To(Equal(m))
			}
			g.Expect(len(got.OIDCConfigs)).To(Equal(len(tt.wantOIDCConfigs)), "it should return the right number of OIDCConfigs")
			for _, o := range tt.wantOIDCConfigs {
				g.Expect(got.OIDCConfig(o.Name)).To(Equal(o))
			}
			g.Expect(len(got.AWSIAMConfigs)).To(Equal(len(tt.wantAWSIamConfigs)), "it should return the right number of AWSIAMConfigs")
			for _, a := range tt.wantAWSIamConfigs {
				g.Expect(got.AWSIamConfig(a.Name)).To(Equal(a))
			}
			g.Expect(got.GitOpsConfig).To(Equal(tt.wantGitOpsConfig))
			g.Expect(got.FluxConfig).To(Equal(tt.wantFluxConfig))
		})
	}
}
