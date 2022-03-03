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
		wantVsphereMachineConfigs []*anywherev1.VSphereMachineConfig
		wantOIDCConfigs           []*anywherev1.OIDCConfig
		wantAWSIamConfigs         []*anywherev1.AWSIamConfig
		wantGitOpsConfig          *anywherev1.GitOpsConfig
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
			name:         "docker cluster with oidc, awsaim and flux",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := cluster.ParseConfig(tt.yamlManifest)

			g.Expect(err).To(Not(HaveOccurred()))

			g.Expect(got.Cluster).To(Equal(tt.wantCluster))
			g.Expect(got.VSphereDatacenter).To(Equal(tt.wantVsphereDatacenter))
			g.Expect(got.DockerDatacenter).To(Equal(tt.wantDockerDatacenter))
			g.Expect(len(got.VSphereMachineConfigs)).To(Equal(len(tt.wantVsphereMachineConfigs)))
			for _, m := range tt.wantVsphereMachineConfigs {
				g.Expect(got.VsphereMachineConfig(m.Name)).To(Equal(m))
			}
			g.Expect(len(got.OIDCConfigs)).To(Equal(len(tt.wantOIDCConfigs)))
			for _, o := range tt.wantOIDCConfigs {
				g.Expect(got.OIDCConfig(o.Name)).To(Equal(o))
			}
			g.Expect(len(got.AWSIAMConfigs)).To(Equal(len(tt.wantAWSIamConfigs)))
			for _, a := range tt.wantAWSIamConfigs {
				g.Expect(got.AWSIamConfig(a.Name)).To(Equal(a))
			}
			g.Expect(got.GitOpsConfig).To(Equal(tt.wantGitOpsConfig))
		})
	}
}

func TestParseConfigMissingCluster(t *testing.T) {
	g := NewWithT(t)
	_, err := cluster.ParseConfig([]byte{})
	g.Expect(err).To(MatchError(ContainSubstring("no Cluster found in manifest")))
}

func TestParseConfigTwoClusters(t *testing.T) {
	g := NewWithT(t)
	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-unit-test
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: eksa-unit-test-2
`
	_, err := cluster.ParseConfig([]byte(manifest))
	g.Expect(err).To(MatchError(ContainSubstring("only one Cluster per yaml manifest is allowed")))
}

func TestParseConfigUnknownKind(t *testing.T) {
	g := NewWithT(t)
	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: MysteryCRD
`
	_, err := cluster.ParseConfig([]byte(manifest))
	g.Expect(err).To(MatchError(ContainSubstring("invalid object with kind MysteryCRD found on manifest")))
}
