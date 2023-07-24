package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestParseConfig(t *testing.T) {
	kube120 := anywherev1.KubernetesVersion("1.20")
	kube119 := anywherev1.KubernetesVersion("1.19")
	tests := []struct {
		name                         string
		yamlManifest                 []byte
		wantCluster                  *anywherev1.Cluster
		wantCloudStackDatacenter     *anywherev1.CloudStackDatacenterConfig
		wantVsphereDatacenter        *anywherev1.VSphereDatacenterConfig
		wantDockerDatacenter         *anywherev1.DockerDatacenterConfig
		wantVsphereMachineConfigs    []*anywherev1.VSphereMachineConfig
		wantCloudStackMachineConfigs []*anywherev1.CloudStackMachineConfig
		wantOIDCConfigs              []*anywherev1.OIDCConfig
		wantAWSIamConfigs            []*anywherev1.AWSIamConfig
		wantGitOpsConfig             *anywherev1.GitOpsConfig
		wantFluxConfig               *anywherev1.FluxConfig
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
							Name:              "workers-1",
							KubernetesVersion: &kube119,
							Count:             ptr.Int(1),
							MachineGroupRef: &anywherev1.Ref{
								Kind: "VSphereMachineConfig",
								Name: "eksa-unit-test",
							},
						},
						{
							Name:              "workers-2",
							KubernetesVersion: &kube120,
							Count:             ptr.Int(1),
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
					Network:    "/myDatacenter/network-1",
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
						Datastore:    "myDatastore",
						DiskGiB:      25,
						MemoryMiB:    8192,
						NumCPUs:      2,
						OSFamily:     anywherev1.Ubuntu,
						ResourcePool: "myResourcePool",
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
						Datastore:    "myDatastore",
						DiskGiB:      25,
						MemoryMiB:    8192,
						NumCPUs:      2,
						OSFamily:     anywherev1.Ubuntu,
						ResourcePool: "myResourcePool",
						Users: []anywherev1.UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
					},
				},
			},
		},
		{
			name:         "cloudstack cluster",
			yamlManifest: []byte(test.ReadFile(t, "testdata/cluster_1_20_cloudstack.yaml")),
			wantCluster: &anywherev1.Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.ClusterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.20",
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count:    3,
						Endpoint: &anywherev1.Endpoint{Host: "test-ip"},
						MachineGroupRef: &anywherev1.Ref{
							Kind: "CloudStackMachineConfig",
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Count: ptr.Int(3),
							MachineGroupRef: &anywherev1.Ref{
								Kind: "CloudStackMachineConfig",
								Name: "eksa-unit-test",
							},
						},
					},
					DatacenterRef: anywherev1.Ref{
						Kind: "CloudStackDatacenterConfig",
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
			wantCloudStackDatacenter: &anywherev1.CloudStackDatacenterConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       anywherev1.CloudStackDatacenterKind,
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: anywherev1.CloudStackDatacenterConfigSpec{
					Account: "admin",
					Domain:  "domain1",
					Zones: []anywherev1.CloudStackZone{
						{
							Name: "zone1",
							Network: anywherev1.CloudStackResourceIdentifier{
								Name: "net1",
							},
						},
					},
					ManagementApiEndpoint: "https://127.0.0.1:8080/client/api",
				},
			},
			wantCloudStackMachineConfigs: []*anywherev1.CloudStackMachineConfig{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.CloudStackMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.CloudStackMachineConfigSpec{
						ComputeOffering: anywherev1.CloudStackResourceIdentifier{
							Name: "m4-large",
						},
						Template: anywherev1.CloudStackResourceIdentifier{
							Name: "centos7-k8s-120",
						},
						Users: []anywherev1.UserConfiguration{{
							Name:              "mySshUsername",
							SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
						}},
						Affinity:          "pro",
						UserCustomDetails: map[string]string{"foo": "bar"},
						Symlinks:          map[string]string{"/var/log/kubernetes": "/data/var/log/kubernetes"},
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
							Count: ptr.Int(1),
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
							Name:              "workers-1",
							KubernetesVersion: &kube120,
							Count:             ptr.Int(1),
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
			g.Expect(len(got.VSphereMachineConfigs)).To(Equal(len(tt.wantVsphereMachineConfigs)), "it should return the right number of VSphereMachineConfigs")
			for _, m := range tt.wantVsphereMachineConfigs {
				g.Expect(got.VsphereMachineConfig(m.Name)).To(Equal(m))
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

func TestParseConfigForSnow(t *testing.T) {
	tests := []struct {
		name                   string
		yamlManifest           []byte
		wantCluster            *anywherev1.Cluster
		wantSnowDatacenter     *anywherev1.SnowDatacenterConfig
		wantSnowMachineConfigs []*anywherev1.SnowMachineConfig
		wantSnowIPPools        []*anywherev1.SnowIPPool
	}{
		{
			name:         "snow cluster basic",
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
							Count: ptr.Int(1),
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
			name:         "snow cluster with ip pool",
			yamlManifest: []byte(test.ReadFile(t, "testdata/cluster_snow_with_ip_pool.yaml")),
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
							Count: ptr.Int(1),
							MachineGroupRef: &anywherev1.Ref{
								Kind: "SnowMachineConfig",
								Name: "eksa-unit-test-worker-1",
							},
						},
						{
							Name:  "workers-2",
							Count: ptr.Int(1),
							MachineGroupRef: &anywherev1.Ref{
								Kind: "SnowMachineConfig",
								Name: "eksa-unit-test-worker-2",
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
						Network: anywherev1.SnowNetwork{
							DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
								{
									Index: 1,
									IPPoolRef: &anywherev1.Ref{
										Kind: anywherev1.SnowIPPoolKind,
										Name: "ip-pool-1",
									},
									Primary: true,
								},
								{
									Index: 2,
									IPPoolRef: &anywherev1.Ref{
										Kind: anywherev1.SnowIPPoolKind,
										Name: "ip-pool-2",
									},
									Primary: false,
								},
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.SnowMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-worker-1",
					},
					Spec: anywherev1.SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-ami",
						InstanceType: "sbe-c.xlarge",
						SshKeyName:   "default",
						Network: anywherev1.SnowNetwork{
							DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
								{
									Index: 1,
									IPPoolRef: &anywherev1.Ref{
										Kind: anywherev1.SnowIPPoolKind,
										Name: "ip-pool-1",
									},
									Primary: true,
								},
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.SnowMachineConfigKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-worker-2",
					},
					Spec: anywherev1.SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-ami",
						InstanceType: "sbe-c.xlarge",
						SshKeyName:   "default",
						Network: anywherev1.SnowNetwork{
							DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
								{
									Index:   1,
									DHCP:    true,
									Primary: true,
								},
							},
						},
					},
				},
			},
			wantSnowIPPools: []*anywherev1.SnowIPPool{
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.SnowIPPoolKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "ip-pool-1",
					},
					Spec: anywherev1.SnowIPPoolSpec{
						Pools: []anywherev1.IPPool{
							{
								IPStart: "start-1",
								IPEnd:   "end-1",
								Subnet:  "subnet-1",
								Gateway: "gateway-1",
							},
						},
					},
				},
				{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.SnowIPPoolKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "ip-pool-2",
					},
					Spec: anywherev1.SnowIPPoolSpec{
						Pools: []anywherev1.IPPool{
							{
								IPStart: "start-2",
								IPEnd:   "end-2",
								Subnet:  "subnet-2",
								Gateway: "gateway-2",
							},
						},
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
			g.Expect(got.SnowDatacenter).To(Equal(tt.wantSnowDatacenter))
			g.Expect(len(got.SnowMachineConfigs)).To(Equal(len(tt.wantSnowMachineConfigs)), "it should return the right number of SnowMachineConfigs")
			for _, m := range tt.wantSnowMachineConfigs {
				g.Expect(got.SnowMachineConfig(m.Name)).To(Equal(m))
			}
			for _, p := range tt.wantSnowIPPools {
				g.Expect(got.SnowIPPool(p.Name)).To(Equal(p))
			}
		})
	}
}
