package hardware

import (
	"context"
	"fmt"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rufiov1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// OwnerNameLabel is the label set by CAPT to mark a hardware as part of a cluster.
const OwnerNameLabel string = "v1alpha1.tinkerbell.org/ownerName"

// KubeReader reads the tinkerbell hardware objects from the cluster.
// It holds the objects in a catalogue.
type KubeReader struct {
	client    client.Client
	catalogue *Catalogue
}

// NewKubeReader returns a new instance of KubeReader.
// Defines a new Catalogue for each KubeReader instance.
func NewKubeReader(client client.Client) *KubeReader {
	return &KubeReader{
		client: client,
		catalogue: NewCatalogue(
			WithHardwareIDIndex(),
			WithHardwareBMCRefIndex(),
			WithBMCNameIndex(),
			WithSecretNameIndex(),
		),
	}
}

// LoadHardware fetches the unprovisioned tinkerbell hardware objects and inserts in to KubeReader catalogue.
func (kr *KubeReader) LoadHardware(ctx context.Context) error {
	hwList, err := kr.getUnprovisionedTinkerbellHardware(ctx)
	if err != nil {
		return fmt.Errorf("failed to build catalogue: %v", err)
	}

	for i := range hwList {
		if err := kr.catalogue.InsertHardware(&hwList[i]); err != nil {
			return err
		}
	}

	return nil
}

// GetCatalogue returns the KubeReader catalogue.
func (kr *KubeReader) GetCatalogue() *Catalogue {
	return kr.catalogue
}

// getUnprovisionedTinkerbellHardware fetches the tinkerbell hardware objects on the cluster which do not have an ownerName label.
func (kr *KubeReader) getUnprovisionedTinkerbellHardware(ctx context.Context) ([]tinkv1alpha1.Hardware, error) {
	var selectedHardware tinkv1alpha1.HardwareList
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      OwnerNameLabel,
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("converting label selector: %w", err)
	}

	if err := kr.client.List(ctx, &selectedHardware, &client.ListOptions{LabelSelector: selector}, client.InNamespace(constants.EksaSystemNamespace)); err != nil {
		return nil, fmt.Errorf("listing hardware without owner: %v", err)
	}

	return selectedHardware.Items, nil
}

// LoadRufioMachines fetches rufio machine objects from the cluster and inserts into KubeReader catalogue.
func (kr *KubeReader) LoadRufioMachines(ctx context.Context) error {
	var rufioMachines rufiov1alpha1.MachineList
	if err := kr.client.List(ctx, &rufioMachines, &client.ListOptions{Namespace: constants.EksaSystemNamespace}); err != nil {
		return fmt.Errorf("listing rufio machines: %v", err)
	}

	for i := range rufioMachines.Items {
		if err := kr.catalogue.InsertBMC(&rufioMachines.Items[i]); err != nil {
			return err
		}
	}

	return nil
}
