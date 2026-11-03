package clustermanager_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	mocksmanager "github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func addressHardware(address string) unstructured.Unstructured {
	hw := unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "tinkerbell.org/v1alpha1",
		"kind":       "Hardware",
		"metadata": map[string]interface{}{
			"name": "test-hardware", "namespace": constants.EksaSystemNamespace, "resourceVersion": "42",
		},
		"spec": map[string]interface{}{
			"interfaces": []interface{}{
				map[string]interface{}{"dhcp": map[string]interface{}{"mac": "02:00:00:00:00:01"}},
			},
		},
	}}
	if address != "" {
		hw.Object["status"] = map[string]interface{}{
			"addresses": []interface{}{
				map[string]interface{}{"address": address, "type": "InternalIP"},
			},
		}
	}
	return hw
}

func TestSnapshotTinkerbellHardwareAddresses(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	client := mocksmanager.NewMockClusterClient(gomock.NewController(t))
	client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, "source.kubeconfig", gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
			list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{
				addressHardware("2001:db8::10"), addressHardware(""),
			}
			return nil
		})

	snapshots, err := clustermanager.SnapshotTinkerbellHardwareAddresses(ctx, "source.kubeconfig", client)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(snapshots).To(Equal([]clustermanager.HardwareAddressSnapshot{{
		Name: "test-hardware", Namespace: constants.EksaSystemNamespace, MACs: []string{"02:00:00:00:00:01"},
		Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "2001:db8::10"}},
	}}))
}

func TestSnapshotTinkerbellHardwareAddressesReadError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	client := mocksmanager.NewMockClusterClient(gomock.NewController(t))
	client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, "source.kubeconfig", gomock.Any()).
		Return(errors.New("API unavailable"))
	_, err := clustermanager.SnapshotTinkerbellHardwareAddresses(ctx, "source.kubeconfig", client)
	g.Expect(err).To(MatchError(ContainSubstring("API unavailable")))
}

func TestRestoreTinkerbellHardwareAddresses(t *testing.T) {
	for _, tc := range []struct {
		name       string
		address    string
		changedMAC bool
		wantPatch  bool
		wantError  bool
	}{
		{name: "status cleared by move", wantPatch: true},
		{name: "new destination observation preserved", address: "2001:db8::20"},
		{name: "different hardware rejected", changedMAC: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			client := mocksmanager.NewMockClusterClient(gomock.NewController(t))
			hw := addressHardware(tc.address)
			if tc.changedMAC {
				hw.Object["spec"] = map[string]interface{}{
					"interfaces": []interface{}{
						map[string]interface{}{"dhcp": map[string]interface{}{"mac": "02:00:00:00:00:02"}},
					},
				}
			}
			client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, "target.kubeconfig", gomock.Any()).
				DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
					list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{hw}
					return nil
				})
			if tc.wantPatch {
				client.EXPECT().MergePatchResourceStatus(ctx, "hardware.tinkerbell.org", "test-hardware", gomock.Any(), "target.kubeconfig", constants.EksaSystemNamespace).
					DoAndReturn(func(_ context.Context, _, _ string, patch string, _, _ string) error {
						g.Expect(patch).To(MatchJSON(`{"metadata":{"resourceVersion":"42"},"status":{"addresses":[{"type":"InternalIP","address":"2001:db8::10"}]}}`))
						return nil
					})
			}
			err := clustermanager.RestoreTinkerbellHardwareAddresses(ctx, "target.kubeconfig", []clustermanager.HardwareAddressSnapshot{{
				Name: "test-hardware", Namespace: constants.EksaSystemNamespace, MACs: []string{"02:00:00:00:00:01"},
				Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "2001:db8::10"}},
			}}, client)
			if tc.wantError {
				g.Expect(err).To(MatchError(ContainSubstring("MAC addresses changed")))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestRestoreTinkerbellHardwareAddressesPatchError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	client := mocksmanager.NewMockClusterClient(gomock.NewController(t))
	client.EXPECT().ListObjects(ctx, "hardware.tinkerbell.org", constants.EksaSystemNamespace, "target.kubeconfig", gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _, _ string, list kubernetes.ObjectList) error {
			list.(*unstructured.UnstructuredList).Items = []unstructured.Unstructured{addressHardware("")}
			return nil
		})
	client.EXPECT().MergePatchResourceStatus(ctx, "hardware.tinkerbell.org", "test-hardware", gomock.Any(), "target.kubeconfig", constants.EksaSystemNamespace).
		Return(errors.New("status update failed"))
	err := clustermanager.RestoreTinkerbellHardwareAddresses(ctx, "target.kubeconfig", []clustermanager.HardwareAddressSnapshot{{
		Name: "test-hardware", Namespace: constants.EksaSystemNamespace, MACs: []string{"02:00:00:00:00:01"},
		Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "2001:db8::10"}},
	}}, client)
	g.Expect(err).To(MatchError(ContainSubstring("status update failed")))
}

func TestRestoreTinkerbellHardwareAddressesEmptySnapshot(t *testing.T) {
	g := NewWithT(t)
	client := mocksmanager.NewMockClusterClient(gomock.NewController(t))
	g.Expect(clustermanager.RestoreTinkerbellHardwareAddresses(context.Background(), "target.kubeconfig", nil, client)).To(Succeed())
}
