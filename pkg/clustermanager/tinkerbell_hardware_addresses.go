package clustermanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const tinkerbellHardwareResource = "hardware.tinkerbell.org"

// HardwareAddressSnapshot keeps observed addresses tied to the same hardware
// when clusterctl recreates the object without status.
type HardwareAddressSnapshot struct {
	Name      string               `json:"name"`
	Namespace string               `json:"namespace"`
	MACs      []string             `json:"macs"`
	Addresses []corev1.NodeAddress `json:"addresses"`
}

// HardwareAddressClient reads Hardware objects and updates their status.
type HardwareAddressClient interface {
	ListObjects(context.Context, string, string, string, kubernetes.ObjectList) error
	MergePatchResourceStatus(context.Context, string, string, string, string, string) error
}

// SnapshotTinkerbellHardwareAddresses reads unstructured objects because the
// pinned Tinkerbell Go types do not include the observed address field.
func SnapshotTinkerbellHardwareAddresses(ctx context.Context, kubeconfig string, client HardwareAddressClient) ([]HardwareAddressSnapshot, error) {
	var list unstructured.UnstructuredList
	if err := client.ListObjects(ctx, tinkerbellHardwareResource, constants.EksaSystemNamespace, kubeconfig, &list); err != nil {
		return nil, fmt.Errorf("reading Tinkerbell hardware addresses: %w", err)
	}

	var snapshots []HardwareAddressSnapshot
	for _, hw := range list.Items {
		addresses, err := hardwareAddresses(hw)
		if err != nil {
			return nil, err
		}
		if len(addresses) == 0 {
			continue
		}
		macs, err := hardwareMACs(hw)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, HardwareAddressSnapshot{
			Name: hw.GetName(), Namespace: hw.GetNamespace(), MACs: macs, Addresses: addresses,
		})
	}
	return snapshots, nil
}

// RestoreTinkerbellHardwareAddresses restores only missing observations.
// CAPT reconciles TinkerbellMachine addresses from the updated Hardware status.
func RestoreTinkerbellHardwareAddresses(ctx context.Context, kubeconfig string, snapshots []HardwareAddressSnapshot, client HardwareAddressClient) error {
	if len(snapshots) == 0 {
		return nil
	}
	var list unstructured.UnstructuredList
	if err := client.ListObjects(ctx, tinkerbellHardwareResource, constants.EksaSystemNamespace, kubeconfig, &list); err != nil {
		return fmt.Errorf("reading moved Tinkerbell hardware: %w", err)
	}
	hardware := make(map[string]unstructured.Unstructured, len(list.Items))
	for _, hw := range list.Items {
		hardware[hw.GetNamespace()+"/"+hw.GetName()] = hw
	}

	for _, snapshot := range snapshots {
		key := snapshot.Namespace + "/" + snapshot.Name
		hw, found := hardware[key]
		if !found {
			return fmt.Errorf("moved Hardware %s not found", key)
		}
		if err := restoreHardwareAddresses(ctx, kubeconfig, snapshot, hw, client); err != nil {
			return err
		}
	}
	return nil
}

func restoreHardwareAddresses(ctx context.Context, kubeconfig string, snapshot HardwareAddressSnapshot, hw unstructured.Unstructured, client HardwareAddressClient) error {
	addresses, err := hardwareAddresses(hw)
	if err != nil {
		return err
	}
	if len(addresses) != 0 {
		return nil
	}
	macs, err := hardwareMACs(hw)
	if err != nil {
		return err
	}
	key := snapshot.Namespace + "/" + snapshot.Name
	if !reflect.DeepEqual(macs, snapshot.MACs) {
		return fmt.Errorf("Hardware %s MAC addresses changed during move", key)
	}
	patch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]string{"resourceVersion": hw.GetResourceVersion()},
		"status":   map[string]interface{}{"addresses": snapshot.Addresses},
	})
	if err != nil {
		return fmt.Errorf("encoding Hardware %s address status: %w", key, err)
	}
	if err := client.MergePatchResourceStatus(ctx, tinkerbellHardwareResource, snapshot.Name, string(patch), kubeconfig, snapshot.Namespace); err != nil {
		return fmt.Errorf("restoring Hardware %s addresses: %w", key, err)
	}
	return nil
}

func hardwareAddresses(hw unstructured.Unstructured) ([]corev1.NodeAddress, error) {
	raw, _, err := unstructured.NestedSlice(hw.Object, "status", "addresses")
	if err != nil {
		return nil, fmt.Errorf("reading Hardware %s/%s addresses: %w", hw.GetNamespace(), hw.GetName(), err)
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var addresses []corev1.NodeAddress
	if err := json.Unmarshal(data, &addresses); err != nil {
		return nil, fmt.Errorf("decoding Hardware %s/%s addresses: %w", hw.GetNamespace(), hw.GetName(), err)
	}
	return addresses, nil
}

func hardwareMACs(hw unstructured.Unstructured) ([]string, error) {
	interfaces, _, err := unstructured.NestedSlice(hw.Object, "spec", "interfaces")
	if err != nil {
		return nil, err
	}
	var macs []string
	for _, iface := range interfaces {
		fields, ok := iface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Hardware %s/%s has an invalid interface", hw.GetNamespace(), hw.GetName())
		}
		value, found, err := unstructured.NestedString(fields, "dhcp", "mac")
		if err != nil {
			return nil, err
		}
		if !found || value == "" {
			continue
		}
		mac, err := net.ParseMAC(value)
		if err != nil {
			return nil, fmt.Errorf("Hardware %s/%s has an invalid MAC: %w", hw.GetNamespace(), hw.GetName(), err)
		}
		macs = append(macs, mac.String())
	}
	if len(macs) == 0 {
		return nil, fmt.Errorf("Hardware %s/%s has no MAC to verify its identity", hw.GetNamespace(), hw.GetName())
	}
	sort.Strings(macs)
	return macs, nil
}
