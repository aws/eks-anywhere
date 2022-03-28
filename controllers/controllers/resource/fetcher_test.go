package resource_test

import (
	"context"
	"reflect"
	"testing"

	cloudstackv1 "github.com/aws/cluster-api-provider-cloudstack/api/v1beta1"
	"github.com/golang/mock/gomock"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/node/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/controllers/controllers/resource"
	"github.com/aws/eks-anywhere/controllers/controllers/resource/mocks"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

var etcdadmCluster = &etcdv1.EtcdadmCluster{
	Spec: etcdv1.EtcdadmClusterSpec{
		InfrastructureTemplate: corev1.ObjectReference{
			Name: "etcdCluster",
		},
	},
}

var capiCluster = &clusterv1.Cluster{
	Spec: clusterv1.ClusterSpec{
		ControlPlaneRef: &corev1.ObjectReference{
			Name:      "capiClusterName",
			Namespace: "test-ns",
		},
	},
}

var kubeadmControlPlane = controlplanev1.KubeadmControlPlane{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "kubeAdmControlPlane",
		Namespace: "test-ns",
	},
	Spec: controlplanev1.KubeadmControlPlaneSpec{
		MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
			InfrastructureRef: corev1.ObjectReference{
				Name: "kubeadmControlPlane",
			},
		},
	},
}

var machineDeployments = &clusterv1.MachineDeploymentList{
	Items: []clusterv1.MachineDeployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testCluster-test",
				Namespace: "test-ns",
			},
			Spec: clusterv1.MachineDeploymentSpec{
				Template: clusterv1.MachineTemplateSpec{
					Spec: clusterv1.MachineSpec{
						InfrastructureRef: corev1.ObjectReference{
							Name: "testCluster-test",
						},
					},
				},
			},
		},
	},
}

var cloudstackCluster = &cloudstackv1.CloudStackCluster{
	Spec: cloudstackv1.CloudStackClusterSpec{},
}

func TestMapMachineTemplateToVSphereDatacenterConfigSpec(t *testing.T) {
	type args struct {
		vsMachineTemplate *vspherev1.VSphereMachineTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    *anywherev1.VSphereDatacenterConfig
		wantErr bool
	}{
		{
			name:    "All path are available",
			wantErr: false,
			args: args{
				vsMachineTemplate: &vspherev1.VSphereMachineTemplate{
					Spec: vspherev1.VSphereMachineTemplateSpec{
						Template: vspherev1.VSphereMachineTemplateResource{
							Spec: vspherev1.VSphereMachineSpec{
								VirtualMachineCloneSpec: vspherev1.VirtualMachineCloneSpec{
									MemoryMiB:    int64(64),
									DiskGiB:      int32(100),
									NumCPUs:      int32(3),
									Template:     "templateA",
									Thumbprint:   "aaa",
									Server:       "ssss",
									ResourcePool: "poolA",
									Datacenter:   "daaa",
									Datastore:    "ds-aaa",
									Folder:       "folder/A",
									Network: vspherev1.NetworkSpec{
										Devices: []vspherev1.NetworkDeviceSpec{
											{
												NetworkName: "networkA",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: &anywherev1.VSphereDatacenterConfig{
				Spec: anywherev1.VSphereDatacenterConfigSpec{
					Thumbprint: "aaa",
					Server:     "ssss",
					Datacenter: "daaa",
					Network:    "networkA",
				},
			},
		},
		{
			name:    "NetworkName missing, throw error",
			wantErr: true,
			args: args{
				vsMachineTemplate: &vspherev1.VSphereMachineTemplate{
					Spec: vspherev1.VSphereMachineTemplateSpec{
						Template: vspherev1.VSphereMachineTemplateResource{
							Spec: vspherev1.VSphereMachineSpec{
								VirtualMachineCloneSpec: vspherev1.VirtualMachineCloneSpec{
									MemoryMiB:    int64(64),
									DiskGiB:      int32(100),
									NumCPUs:      int32(3),
									Template:     "templateA",
									Thumbprint:   "aaa",
									Server:       "ssss",
									ResourcePool: "poolA",
									Datacenter:   "daaa",
									Datastore:    "ds-aaa",
									Folder:       "folder/A",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resource.MapMachineTemplateToVSphereDatacenterConfigSpec(tt.args.vsMachineTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapMachineTemplateToVSphereDatacenterConfigSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapMachineTemplateToVSphereDatacenterConfigSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapClusterToCloudStackDatacenterConfigSpec(t *testing.T) {
	type args struct {
		csCluster *cloudstackv1.CloudStackCluster
	}
	tests := []struct {
		name string
		args args
		want *anywherev1.CloudStackDatacenterConfig
	}{
		{
			name: "All path are available",
			args: args{
				csCluster: &cloudstackv1.CloudStackCluster{
					Spec: cloudstackv1.CloudStackClusterSpec{
						Zones: []cloudstackv1.Zone{
							{
								Name: "zone",
								Network: cloudstackv1.Network{
									Name: "network",
								},
							},
						},
						Account: "account",
						Domain:  "domain",
					},
				},
			},
			want: &anywherev1.CloudStackDatacenterConfig{
				Spec: anywherev1.CloudStackDatacenterConfigSpec{
					Zones: []anywherev1.CloudStackZone{
						{
							Name: "zone",
							Network: anywherev1.CloudStackResourceIdentifier{
								Name: "network",
							},
						},
					},
					Account: "account",
					Domain:  "domain",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resource.MapClusterToCloudStackDatacenterConfigSpec(tt.args.csCluster)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapMachineTemplateToCloudStackDatacenterConfigSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapMachineTemplateToVSphereWorkerMachineConfigSpec(t *testing.T) {
	type args struct {
		vsMachineTemplate *vspherev1.VSphereMachineTemplate
		users             []kubeadmv1.User
	}
	tests := []struct {
		name    string
		args    args
		want    *anywherev1.VSphereMachineConfig
		wantErr bool
	}{
		{
			name:    "All path are available",
			wantErr: false,
			args: args{
				vsMachineTemplate: &vspherev1.VSphereMachineTemplate{
					Spec: vspherev1.VSphereMachineTemplateSpec{
						Template: vspherev1.VSphereMachineTemplateResource{
							Spec: vspherev1.VSphereMachineSpec{
								VirtualMachineCloneSpec: vspherev1.VirtualMachineCloneSpec{
									MemoryMiB:    int64(64),
									DiskGiB:      int32(100),
									NumCPUs:      int32(3),
									Template:     "templateA",
									Thumbprint:   "aaa",
									Server:       "ssss",
									ResourcePool: "poolA",
									Datacenter:   "daaa",
									Datastore:    "ds-aaa",
									Folder:       "folder/A",
									Network: vspherev1.NetworkSpec{
										Devices: []vspherev1.NetworkDeviceSpec{
											{
												NetworkName: "networkA",
											},
										},
									},
								},
							},
						},
					},
				},
				users: []kubeadmv1.User{
					{
						Name: "test",
						SSHAuthorizedKeys: []string{
							"ssh_rsa",
						},
					},
				},
			},
			want: &anywherev1.VSphereMachineConfig{
				Spec: anywherev1.VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Datastore:    "ds-aaa",
					Folder:       "folder/A",
					Users: []anywherev1.UserConfiguration{
						{
							Name: "test",
							SshAuthorizedKeys: []string{
								"ssh_rsa",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resource.MapMachineTemplateToVSphereMachineConfigSpec(tt.args.vsMachineTemplate, tt.args.users)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapMachineTemplateToVSphereWorkerMachineConfigSpec() error = %v, \n wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapMachineTemplateToVSphereWorkerMachineConfigSpec()\n got = %v, \n want %v", got, tt.want)
			}
		})
	}
}

func TestMapMachineTemplateToCloudStackWorkerMachineConfigSpec(t *testing.T) {
	type args struct {
		csMachineTemplate *cloudstackv1.CloudStackMachineTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    *anywherev1.CloudStackMachineConfig
		wantErr bool
	}{
		{
			name:    "All path are available",
			wantErr: false,
			args: args{
				csMachineTemplate: &cloudstackv1.CloudStackMachineTemplate{
					Spec: cloudstackv1.CloudStackMachineTemplateSpec{
						Spec: cloudstackv1.CloudStackMachineTemplateResource{
							Spec: cloudstackv1.CloudStackMachineSpec{
								Offering:         cloudstackv1.CloudStackResourceIdentifier{Name: "large"},
								Template:         cloudstackv1.CloudStackResourceIdentifier{Name: "rhel8-1.20"},
								AffinityGroupIDs: []string{"c", "d"},
							},
						},
					},
				},
			},
			want: &anywherev1.CloudStackMachineConfig{
				Spec: anywherev1.CloudStackMachineConfigSpec{
					Template:         anywherev1.CloudStackResourceIdentifier{Name: "rhel8-1.20"},
					ComputeOffering:  anywherev1.CloudStackResourceIdentifier{Name: "large"},
					AffinityGroupIds: []string{"c", "d"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resource.MapMachineTemplateToCloudStackMachineConfigSpec(tt.args.csMachineTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapMachineTemplateToCloudStackWorkerMachineConfigSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapMachineTemplateToCloudStackWorkerMachineConfigSpec() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchCloudStackCluster(t *testing.T) {
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
	}{
		{
			name: "Fetch CloudStackCluster",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			reader := mocks.NewMockReader(mockCtrl)
			logger := log.NullLogger{}
			capiResourceFetcher := resource.NewCAPIResourceFetcher(reader, logger)
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: tt.cluster.Name}, gomock.Any()).Do(
				func(ctx context.Context, arg1 types.NamespacedName, arg2 *cloudstackv1.CloudStackCluster) {
					cloudstackCluster.DeepCopyInto(arg2)
				})
			_, err := capiResourceFetcher.CloudStackCluster(ctx, tt.cluster, anywherev1.WorkerNodeGroupConfiguration{Name: "test"})
			if err != nil {
				t.Errorf("CloudStackEtcdMachineTemplate(): %v", err)
			}
		})
	}
}

func TestFetchCloudStackEtcdMachineTemplate(t *testing.T) {
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
	}{
		{
			name: "Etcd",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			reader := mocks.NewMockReader(mockCtrl)
			logger := log.NullLogger{}
			capiResourceFetcher := resource.NewCAPIResourceFetcher(reader, logger)
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: "testCluster-etcd"}, gomock.Any()).Do(
				func(ctx context.Context, arg1 types.NamespacedName, arg2 *etcdv1.EtcdadmCluster) {
					etcdadmCluster.DeepCopyInto(arg2)
				})
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: etcdadmCluster.Spec.InfrastructureTemplate.Name},
				gomock.Any())
			_, err := capiResourceFetcher.CloudStackEtcdMachineTemplate(ctx, tt.cluster)
			if err != nil {
				t.Errorf("CloudStackEtcdMachineTemplate(): %v", err)
			}
		})
	}
}

func TestFetchCloudStackCPMachineTemplate(t *testing.T) {
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
	}{
		{
			name: "Fetch CP Machine Template",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			reader := mocks.NewMockReader(mockCtrl)
			logger := log.NullLogger{}
			capiResourceFetcher := resource.NewCAPIResourceFetcher(reader, logger)
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: tt.cluster.Name}, gomock.Any()).Do(
				func(ctx context.Context, arg1 types.NamespacedName, arg2 *clusterv1.Cluster) {
					capiCluster.DeepCopyInto(arg2)
				})
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: capiCluster.Spec.ControlPlaneRef.Namespace, Name: capiCluster.Spec.ControlPlaneRef.Name}, gomock.Any()).Do(
				func(ctx context.Context, arg1 types.NamespacedName, arg2 *controlplanev1.KubeadmControlPlane) {
					kubeadmControlPlane.DeepCopyInto(arg2)
				})
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: kubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name},
				gomock.Any())
			_, err := capiResourceFetcher.CloudStackControlPlaneMachineTemplate(ctx, tt.cluster)
			if err != nil {
				t.Errorf("CloudStackControlPlaneMachineTemplate(): %v", err)
			}
		})
	}
}

func TestFetchCloudStackMDMachineTemplate(t *testing.T) {
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
	}{
		{
			name: "Fetch MD Machine Template",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockCtrl := gomock.NewController(t)
			reader := mocks.NewMockReader(mockCtrl)
			logger := log.NullLogger{}
			capiResourceFetcher := resource.NewCAPIResourceFetcher(reader, logger)
			reader.EXPECT().List(ctx, gomock.Any(), gomock.Any()).Do(
				func(ctx context.Context, arg1 *clusterv1.MachineDeploymentList, arg2 *client.ListOptions) {
					machineDeployments.DeepCopyInto(arg1)
				})
			reader.EXPECT().Get(ctx, types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: machineDeployments.Items[0].Spec.Template.Spec.InfrastructureRef.Name},
				gomock.Any())
			_, err := capiResourceFetcher.CloudStackWorkerMachineTemplate(ctx, tt.cluster, anywherev1.WorkerNodeGroupConfiguration{Name: "test"})
			if err != nil {
				t.Errorf("CloudStackWorkerMachineTemplate(): %v", err)
			}
		})
	}
}

func TestCAPIResourceFetcherFetchCluster(t *testing.T) {
	type fields struct {
		client client.Reader
	}
	type args struct {
		objectKey types.NamespacedName
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *anywherev1.Cluster
		wantErr bool
	}{
		{
			name: "fetch cluster from VSphereDatacenterKind",
			fields: fields{
				client: &stubbedReader{
					clusterName: "testCluster",
					kind:        anywherev1.VSphereDatacenterKind,
					cluster: anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testCluster",
						},
						Spec: anywherev1.ClusterSpec{
							DatacenterRef: anywherev1.Ref{
								Name: "testVSphereDatacenter",
								Kind: anywherev1.VSphereDatacenterKind,
							},
						},
					},
				},
			},
			args: args{
				objectKey: types.NamespacedName{Name: "testVSphereDatacenter", Namespace: "default"},
			},
			want: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
				Spec: anywherev1.ClusterSpec{
					DatacenterRef: anywherev1.Ref{
						Name: "testVSphereDatacenter",
						Kind: anywherev1.VSphereDatacenterKind,
					},
				},
			},
		},
		{
			name: "fetch cluster from CloudStackDatacenterKind",
			fields: fields{
				client: &stubbedReader{
					clusterName: "testCluster",
					kind:        anywherev1.CloudStackDatacenterKind,
					cluster: anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testCluster",
						},
						Spec: anywherev1.ClusterSpec{
							DatacenterRef: anywherev1.Ref{
								Name: "testCloudStackDatacenter",
								Kind: anywherev1.CloudStackDatacenterKind,
							},
						},
					},
				},
			},
			args: args{
				objectKey: types.NamespacedName{Name: "testCloudStackDatacenter", Namespace: "default"},
			},
			want: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
				Spec: anywherev1.ClusterSpec{
					DatacenterRef: anywherev1.Ref{
						Name: "testCloudStackDatacenter",
						Kind: anywherev1.CloudStackDatacenterKind,
					},
				},
			},
		},
		{
			name: "fetch cluster from DockerDatacenterKind",
			fields: fields{
				client: &stubbedReader{
					clusterName: "testCluster",
					kind:        anywherev1.DockerDatacenterKind,
					cluster: anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testCluster",
						},
						Spec: anywherev1.ClusterSpec{
							DatacenterRef: anywherev1.Ref{
								Name: "testDockerDatacenter",
								Kind: anywherev1.DockerDatacenterKind,
							},
						},
					},
				},
			},
			args: args{
				objectKey: types.NamespacedName{Name: "testDockerDatacenter", Namespace: "default"},
			},
			want: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
				Spec: anywherev1.ClusterSpec{
					DatacenterRef: anywherev1.Ref{
						Name: "testDockerDatacenter",
						Kind: anywherev1.DockerDatacenterKind,
					},
				},
			},
		},
		{
			name: "fetch cluster from VSphereMachineConfigKind",
			fields: fields{
				client: &stubbedReader{
					clusterName: "testCluster",
					kind:        anywherev1.VSphereMachineConfigKind,
					cluster: anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testCluster",
						},
						Spec: anywherev1.ClusterSpec{
							WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
								{
									MachineGroupRef: &anywherev1.Ref{
										Name: "testVSphereMachineConfig",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				objectKey: types.NamespacedName{Name: "testVSphereMachineConfig", Namespace: "default"},
			},
			want: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testCluster",
				},
				Spec: anywherev1.ClusterSpec{
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							MachineGroupRef: &anywherev1.Ref{
								Name: "testVSphereMachineConfig",
							},
						},
					},
				},
			},
		},
		{
			name: "fetch cluster from VSphereMachineConfigKind, external etcd field empty",
			fields: fields{
				client: &stubbedReader{
					clusterName: "testCluster",
					kind:        anywherev1.VSphereMachineConfigKind,
					cluster: anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "testCluster",
						},
						Spec: anywherev1.ClusterSpec{},
					},
				},
			},
			args: args{
				objectKey: types.NamespacedName{Name: "testVSphereMachineConfig", Namespace: "default"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := resource.NewCAPIResourceFetcher(tt.fields.client, log.NullLogger{})
			got, err := r.FetchCluster(context.Background(), tt.args.objectKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type stubbedReader struct {
	kind        string
	cluster     anywherev1.Cluster
	clusterName string
}

func (s *stubbedReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	if s.kind == obj.GetObjectKind().GroupVersionKind().Kind {
		return nil
	}
	if key.Name == s.clusterName {
		return nil
	}
	return errors.NewNotFound(v1alpha1.Resource("foo"), "kind not found")
}

func (s *stubbedReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	clusterList := list.(*anywherev1.ClusterList)
	clusterList.Items = []anywherev1.Cluster{s.cluster}
	return nil
}
