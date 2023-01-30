package validations_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/test/framework/cluster/validations"
)

func csiDeployment() *v1.Deployment {
	return &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSystemNamespace,
			Name:      "vsphere-csi-controller",
		},
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vsphere-csi-controller",
					Namespace: constants.KubeSystemNamespace,
					Labels: map[string]string{
						"vsphere-csi": "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "vsphere-csi-controller-pod-container",
							Image: "vsphere-csi-controller-pod-image",
						},
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"vsphere-csi": "",
				},
			},
		},
	}
}

func csiDaemonSet() *v1.DaemonSet {
	return &v1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubeSystemNamespace,
			Name:      "vsphere-csi-node",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vsphere-csi-node",
					Namespace: constants.KubeSystemNamespace,
					Labels: map[string]string{
						"vsphere-csi": "",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "vsphere-csi-node-pod-container",
							Image: "vsphere-csi-node-pod-image",
						},
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"vsphere-csi": "",
				},
			},
		},
	}
}

func TestValidatateCSI(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	tests := []struct {
		name       string
		csiObjects []client.Object
		disableCSI bool
		wantErr    string
	}{
		{
			name: "CSI enabled valid",
			csiObjects: []client.Object{
				csiDeployment(),
				csiDaemonSet(),
			},
			disableCSI: false,
			wantErr:    "",
		},
		{
			name: "CSI enabled missing one csi object",
			csiObjects: []client.Object{
				csiDeployment(),
			},
			disableCSI: false,
			wantErr:    "CSI state does not match disableCSI false, daemonsets.apps \"vsphere-csi-node\" not found",
		},
		{
			name:       "CSI enabled missing all csi objects",
			csiObjects: []client.Object{},
			disableCSI: false,
			wantErr:    "CSI state does not match disableCSI false, deployments.apps \"vsphere-csi-controller\" not found",
		},
		{
			name:       "CSI disabled, valid stat - no CSI objects",
			csiObjects: []client.Object{},
			disableCSI: true,
			wantErr:    "",
		},
		{
			name: "CSI disabled, invalid state - has CSI objects",
			csiObjects: []client.Object{
				csiDeployment(),
			},
			disableCSI: true,
			wantErr:    "CSI state does not match disableCSI true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := cluster.NewSpec(func(s *cluster.Spec) {
				s.Cluster = testCluster()
				s.VSphereDatacenter = &v1alpha1.VSphereDatacenterConfig{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: clusterNamespace,
						Name:      clusterName,
					},
					Spec: v1alpha1.VSphereDatacenterConfigSpec{
						DisableCSI: tt.disableCSI,
					},
				}
			})

			vt := newStateValidatorTest(t, spec)
			vt.createTestObjects(ctx)
			vt.createClusterObjects(ctx, tt.csiObjects...)
			err := validations.ValidateCSI(ctx, vt.config)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
