package reconciler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	addonsv1 "sigs.k8s.io/cluster-api/api/addons/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
)

func TestToClientObjects(t *testing.T) {
	t.Run("convert ConfigMaps", func(t *testing.T) {
		input := []*apiv1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm-1",
					Namespace: constants.EksaSystemNamespace,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm-2",
					Namespace: constants.EksaSystemNamespace,
				},
			},
		}
		result := toClientObjects(input)
		assert.Equal(t, 2, len(result))
		for _, obj := range result {
			assert.NotNil(t, obj)
		}
	})

	t.Run("convert Secrets", func(t *testing.T) {
		input := []*apiv1.Secret{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret-1",
					Namespace: constants.EksaSystemNamespace,
				},
			},
		}
		result := toClientObjects(input)
		assert.Equal(t, 1, len(result))
		for _, obj := range result {
			assert.NotNil(t, obj)
		}
	})

	t.Run("convert ClusterResourceSets", func(t *testing.T) {
		input := []*addonsv1.ClusterResourceSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-crs-1",
					Namespace: constants.EksaSystemNamespace,
				},
			},
		}
		result := toClientObjects(input)
		assert.Equal(t, 1, len(result))
		for _, obj := range result {
			assert.NotNil(t, obj)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []*apiv1.ConfigMap{}
		result := toClientObjects(input)
		assert.Equal(t, 0, len(result))
	})
}

func TestSetOwnerReferencesOnObjects(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, apiv1.AddToScheme(scheme))
	require.NoError(t, clusterv1.AddToScheme(scheme))

	owner := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: constants.EksaSystemNamespace,
			UID:       types.UID("test-uid-123"),
		},
	}

	tests := []struct {
		name           string
		objects        []client.Object
		existingObjs   []client.Object
		wantErr        bool
		wantOwnerRefs  bool
		expectedErrMsg string
	}{
		{
			name: "set owner reference on ConfigMap",
			objects: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			existingObjs: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			wantErr:       false,
			wantOwnerRefs: true,
		},
		{
			name: "set owner reference on Secret",
			objects: []client.Object{
				&apiv1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			existingObjs: []client.Object{
				&apiv1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			wantErr:       false,
			wantOwnerRefs: true,
		},
		{
			name: "skip non-existent object",
			objects: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "non-existent-cm",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			existingObjs:  []client.Object{},
			wantErr:       false,
			wantOwnerRefs: false,
		},
		{
			name: "skip object with existing owner reference",
			objects: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm-with-owner",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			existingObjs: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm-with-owner",
						Namespace: constants.EksaSystemNamespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "cluster.x-k8s.io/v1beta1",
								Kind:       "Cluster",
								Name:       "test-cluster",
								UID:        types.UID("test-uid-123"),
							},
						},
					},
				},
			},
			wantErr:       false,
			wantOwnerRefs: true,
		},
		{
			name: "handle multiple objects",
			objects: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm-1",
						Namespace: constants.EksaSystemNamespace,
					},
				},
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm-2",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			existingObjs: []client.Object{
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm-1",
						Namespace: constants.EksaSystemNamespace,
					},
				},
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm-2",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			wantErr:       false,
			wantOwnerRefs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize fake client with existing objects
			initObjs := append([]client.Object{owner}, tt.existingObjs...)
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(initObjs...).
				Build()

			r := &Reconciler{
				client: fakeClient,
			}

			logger := test.NewNullLogger()
			err := r.setOwnerReferencesOnObjects(context.TODO(), logger, owner, tt.objects)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)

				// If we expect owner refs and have existing objects, verify them
				if tt.wantOwnerRefs && len(tt.existingObjs) > 0 {
					for _, obj := range tt.objects {
						// Get the updated object
						key := client.ObjectKey{
							Name:      obj.GetName(),
							Namespace: obj.GetNamespace(),
						}

						// Create a new object of the same type to fetch into
						var fetchedObj client.Object
						switch obj.(type) {
						case *apiv1.ConfigMap:
							fetchedObj = &apiv1.ConfigMap{}
						case *apiv1.Secret:
							fetchedObj = &apiv1.Secret{}
						}

						err := fakeClient.Get(context.TODO(), key, fetchedObj)
						if err == nil {
							// Verify owner reference exists
							ownerRefs := fetchedObj.GetOwnerReferences()
							hasOwnerRef := false
							for _, ref := range ownerRefs {
								if ref.UID == owner.UID {
									hasOwnerRef = true
									break
								}
							}
							assert.True(t, hasOwnerRef, "Expected owner reference on object %s", obj.GetName())
						}
					}
				}
			}
		})
	}
}

func TestEnsureOwnerReferences(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, apiv1.AddToScheme(scheme))
	require.NoError(t, clusterv1.AddToScheme(scheme))
	require.NoError(t, addonsv1.AddToScheme(scheme))
	require.NoError(t, anywherev1.AddToScheme(scheme))

	capiCluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: constants.EksaSystemNamespace,
			UID:       types.UID("test-uid-123"),
		},
	}

	tests := []struct {
		name                     string
		clusterSpec              *cluster.Spec
		controlPlane             *nutanix.ControlPlane
		existingObjs             []client.Object
		wantErr                  bool
		expectedErrMsg           string
		verifyOwnerReferences    bool
		verifyNoDuplicateOnRetry bool
	}{
		{
			name: "successfully set owner references on all resources",
			clusterSpec: &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-cluster",
							Namespace: constants.EksaSystemNamespace,
						},
					},
				},
			},
			controlPlane: &nutanix.ControlPlane{
				ConfigMaps: []*apiv1.ConfigMap{
					{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-cm",
							Namespace: constants.EksaSystemNamespace,
						},
					},
				},
				Secrets: []*apiv1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "Secret",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: constants.EksaSystemNamespace,
						},
					},
				},
				ClusterResourceSets: []*addonsv1.ClusterResourceSet{},
			},
			existingObjs: []client.Object{
				capiCluster,
				&apiv1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: constants.EksaSystemNamespace,
					},
				},
				&apiv1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
			wantErr:                  false,
			verifyOwnerReferences:    true,
			verifyNoDuplicateOnRetry: true,
		},
		{
			name: "skip when CAPI cluster not found",
			clusterSpec: &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "non-existent-cluster",
							Namespace: constants.EksaSystemNamespace,
						},
					},
				},
			},
			controlPlane: &nutanix.ControlPlane{
				ConfigMaps:          []*apiv1.ConfigMap{},
				Secrets:             []*apiv1.Secret{},
				ClusterResourceSets: []*addonsv1.ClusterResourceSet{},
			},
			existingObjs:          []client.Object{},
			wantErr:               false,
			verifyOwnerReferences: false,
		},
		{
			name: "handle empty control plane resources",
			clusterSpec: &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-cluster",
							Namespace: constants.EksaSystemNamespace,
						},
					},
				},
			},
			controlPlane: &nutanix.ControlPlane{
				ConfigMaps:          []*apiv1.ConfigMap{},
				Secrets:             []*apiv1.Secret{},
				ClusterResourceSets: []*addonsv1.ClusterResourceSet{},
			},
			existingObjs:          []client.Object{capiCluster},
			wantErr:               false,
			verifyOwnerReferences: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.existingObjs...).
				Build()

			r := &Reconciler{
				client: fakeClient,
			}

			logger := test.NewNullLogger()
			err := r.ensureOwnerReferences(context.TODO(), logger, tt.clusterSpec, tt.controlPlane)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify owner references if requested
			if tt.verifyOwnerReferences {
				// Verify ConfigMaps have owner references
				for _, cm := range tt.controlPlane.ConfigMaps {
					updatedCM := &apiv1.ConfigMap{}
					err := fakeClient.Get(context.TODO(), client.ObjectKey{
						Name:      cm.Name,
						Namespace: cm.Namespace,
					}, updatedCM)
					require.NoError(t, err)
					assert.Len(t, updatedCM.OwnerReferences, 1)
					assert.Equal(t, capiCluster.UID, updatedCM.OwnerReferences[0].UID)
				}

				// Verify Secrets have owner references
				for _, secret := range tt.controlPlane.Secrets {
					updatedSecret := &apiv1.Secret{}
					err := fakeClient.Get(context.TODO(), client.ObjectKey{
						Name:      secret.Name,
						Namespace: secret.Namespace,
					}, updatedSecret)
					require.NoError(t, err)
					assert.Len(t, updatedSecret.OwnerReferences, 1)
					assert.Equal(t, capiCluster.UID, updatedSecret.OwnerReferences[0].UID)
				}
			}

			// Test for no duplicate owner references on retry
			if tt.verifyNoDuplicateOnRetry {
				err = r.ensureOwnerReferences(context.TODO(), logger, tt.clusterSpec, tt.controlPlane)
				require.NoError(t, err)

				// Verify no duplicate owner references on ConfigMaps
				for _, cm := range tt.controlPlane.ConfigMaps {
					updatedCM := &apiv1.ConfigMap{}
					err := fakeClient.Get(context.TODO(), client.ObjectKey{
						Name:      cm.Name,
						Namespace: cm.Namespace,
					}, updatedCM)
					require.NoError(t, err)
					assert.Len(t, updatedCM.OwnerReferences, 1, "Should not add duplicate owner references")
				}

				// Verify no duplicate owner references on Secrets
				for _, secret := range tt.controlPlane.Secrets {
					updatedSecret := &apiv1.Secret{}
					err := fakeClient.Get(context.TODO(), client.ObjectKey{
						Name:      secret.Name,
						Namespace: secret.Namespace,
					}, updatedSecret)
					require.NoError(t, err)
					assert.Len(t, updatedSecret.OwnerReferences, 1, "Should not add duplicate owner references")
				}
			}
		})
	}
}
