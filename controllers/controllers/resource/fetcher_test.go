package resource_test

import (
	"context"
	"reflect"
	"testing"

	"k8s.io/api/node/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vspherev3 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/controllers/controllers/resource"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestMapMachineTemplateToVSphereDatacenterConfigSpec(t *testing.T) {
	type args struct {
		vsMachineTemplate *vspherev3.VSphereMachineTemplate
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
				vsMachineTemplate: &vspherev3.VSphereMachineTemplate{
					Spec: vspherev3.VSphereMachineTemplateSpec{
						Template: vspherev3.VSphereMachineTemplateResource{
							Spec: vspherev3.VSphereMachineSpec{
								VirtualMachineCloneSpec: vspherev3.VirtualMachineCloneSpec{
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
									Network: vspherev3.NetworkSpec{
										Devices: []vspherev3.NetworkDeviceSpec{
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
				vsMachineTemplate: &vspherev3.VSphereMachineTemplate{
					Spec: vspherev3.VSphereMachineTemplateSpec{
						Template: vspherev3.VSphereMachineTemplateResource{
							Spec: vspherev3.VSphereMachineSpec{
								VirtualMachineCloneSpec: vspherev3.VirtualMachineCloneSpec{
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

func TestMapMachineTemplateToVSphereWorkerMachineConfigSpec(t *testing.T) {
	type args struct {
		vsMachineTemplate *vspherev3.VSphereMachineTemplate
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
				vsMachineTemplate: &vspherev3.VSphereMachineTemplate{
					Spec: vspherev3.VSphereMachineTemplateSpec{
						Template: vspherev3.VSphereMachineTemplateResource{
							Spec: vspherev3.VSphereMachineSpec{
								VirtualMachineCloneSpec: vspherev3.VirtualMachineCloneSpec{
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
									Network: vspherev3.NetworkSpec{
										Devices: []vspherev3.NetworkDeviceSpec{
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
			want: &anywherev1.VSphereMachineConfig{
				Spec: anywherev1.VSphereMachineConfigSpec{
					MemoryMiB:    64,
					DiskGiB:      100,
					NumCPUs:      3,
					Template:     "templateA",
					ResourcePool: "poolA",
					Datastore:    "ds-aaa",
					Folder:       "folder/A",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resource.MapMachineTemplateToVSphereWorkerMachineConfigSpec(tt.args.vsMachineTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapMachineTemplateToVSphereWorkerMachineConfigSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapMachineTemplateToVSphereWorkerMachineConfigSpec() got = %v, want %v", got, tt.want)
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
