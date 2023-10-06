package cluster_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

func TestWaitForCondition(t *testing.T) {
	testCases := []struct {
		name                         string
		clusterInput, currentCluster *anywherev1.Cluster
		condition                    anywherev1.ConditionType
		retrier                      *retrier.Retrier
		wantErr                      string
	}{
		{
			name: "cluster does not exist",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-cluster",
					Namespace: "my-n",
				},
			},
			retrier: retrier.NewWithMaxRetries(1, 0),
			wantErr: "clusters.anywhere.eks.amazonaws.com \"my-c\" not found",
		},
		{
			name: "observed generation not updated",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-c",
					Namespace:  "my-n",
					Generation: 5,
				},
				Status: anywherev1.ClusterStatus{
					ObservedGeneration: 4,
				},
			},
			retrier: retrier.NewWithMaxRetries(1, 0),
			wantErr: "cluster generation (5) and observedGeneration (4) differ",
		},
		{
			name: "no condition",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-c",
					Namespace:  "my-n",
					Generation: 5,
				},
				Status: anywherev1.ClusterStatus{
					ObservedGeneration: 5,
				},
			},
			retrier:   retrier.NewWithMaxRetries(1, 0),
			condition: anywherev1.ControlPlaneReadyCondition,
			wantErr:   "cluster doesn't yet have condition ControlPlaneReady",
		},
		{
			name: "condition is False",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-c",
					Namespace:  "my-n",
					Generation: 5,
				},
				Status: anywherev1.ClusterStatus{
					ObservedGeneration: 5,
					Conditions: []anywherev1.Condition{
						{
							Type:    anywherev1.ControlPlaneReadyCondition,
							Status:  "False",
							Message: "CP is being rolled out",
						},
					},
				},
			},
			retrier:   retrier.NewWithMaxRetries(1, 0),
			condition: anywherev1.ControlPlaneReadyCondition,
			wantErr:   "cluster condition ControlPlaneReady is False: CP is being rolled out",
		},
		{
			name: "condition is True",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-c",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-c",
					Namespace:  "default",
					Generation: 5,
				},
				Status: anywherev1.ClusterStatus{
					ObservedGeneration: 5,
					Conditions: []anywherev1.Condition{
						{
							Type:   anywherev1.ControlPlaneReadyCondition,
							Status: "True",
						},
					},
				},
			},
			retrier:   retrier.NewWithMaxRetries(2, 0),
			condition: anywherev1.ControlPlaneReadyCondition,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			g := NewWithT(t)
			client := test.NewFakeKubeClient(tt.currentCluster)

			gotErr := cluster.WaitForCondition(ctx, test.NewNullLogger(), client, tt.clusterInput, 2, tt.retrier, tt.condition)
			if tt.wantErr != "" {
				g.Expect(gotErr).To(MatchError(tt.wantErr))
			} else {
				g.Expect(gotErr).NotTo(HaveOccurred())
			}
		})
	}
}

func TestWaitFor(t *testing.T) {
	testCases := []struct {
		name                         string
		clusterInput, currentCluster *anywherev1.Cluster
		retrier                      *retrier.Retrier
		matcher                      func(_ *anywherev1.Cluster) error
		wantErr                      string
	}{
		{
			name: "cluster does not exist",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "default",
				},
			},
			retrier: retrier.NewWithMaxRetries(1, 0),
			matcher: func(_ *anywherev1.Cluster) error {
				return nil
			},
			wantErr: "clusters.anywhere.eks.amazonaws.com \"my-c\" not found",
		},
		{
			name: "cluster namespace not provided",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-c",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "eksa-namespace",
				},
			},
			retrier: retrier.NewWithMaxRetries(1, 0),
			matcher: func(_ *anywherev1.Cluster) error {
				return nil
			},
			wantErr: "clusters.anywhere.eks.amazonaws.com \"my-c\" not found",
		},
		{
			name: "observed generation not updated",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "my-c",
					Namespace:  "my-n",
					Generation: 5,
				},
				Status: anywherev1.ClusterStatus{
					ObservedGeneration: 4,
				},
			},
			retrier: retrier.NewWithMaxRetries(1, 0),
			matcher: func(_ *anywherev1.Cluster) error {
				return nil
			},
			wantErr: "cluster generation (5) and observedGeneration (4) differ",
		},
		{
			name: "matcher return error",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			retrier: retrier.NewWithMaxRetries(1, 0),
			matcher: func(_ *anywherev1.Cluster) error {
				return fmt.Errorf("error")
			},
			wantErr: "error",
		},
		{
			name: "condition is met, retry not enough",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			retrier: retrier.NewWithMaxRetries(3, 0),
			matcher: func(_ *anywherev1.Cluster) error {
				return nil
			},
			wantErr: "cluster has reached to expected condition in 3/5 times",
		},
		{
			name: "condition is met, consistency checked",
			clusterInput: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			currentCluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-c",
					Namespace: "my-n",
				},
			},
			retrier: retrier.NewWithMaxRetries(5, 0),
			matcher: func(_ *anywherev1.Cluster) error {
				return nil
			},
			wantErr: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			g := NewWithT(t)
			client := test.NewFakeKubeClient(tt.currentCluster)

			gotErr := cluster.WaitFor(ctx, test.NewNullLogger(), client, tt.clusterInput, 5, tt.retrier, tt.matcher)
			if tt.wantErr != "" {
				g.Expect(gotErr).To(MatchError(tt.wantErr))
			} else {
				g.Expect(gotErr).NotTo(HaveOccurred())
			}
		})
	}
}
