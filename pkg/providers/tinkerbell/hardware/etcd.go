package hardware

import (
	"context"
	"errors"
	"fmt"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OwnerNameLabel is the label set by CAPT to mark a hardware as part of a cluster.
const OwnerNameLabel string = "v1alpha1.tinkerbell.org/ownerName"

// ETCDReader reads the tinkerbell hardware objects from the cluster.
// It holds the objects in a catalogue.
type ETCDReader struct {
	client    client.Client
	catalogue *Catalogue
}

// NewETCDReader returns a new instance of ETCDReader.
// Defines a new Catalogue for each ETCDReader instance.
func NewETCDReader(client client.Client) *ETCDReader {
	return &ETCDReader{
		client: client,
		catalogue: NewCatalogue(
			WithHardwareIDIndex(),
			WithHardwareBMCRefIndex(),
			WithBMCNameIndex(),
			WithSecretNameIndex(),
		),
	}
}

// NewCatalogueFromETCD fetches the unprovisioned tinkerbell hardware objects and inserts in to ETCDReader catalogue.
func (er *ETCDReader) NewCatalogueFromETCD(ctx context.Context) error {
	hwList, err := er.getUnprovisionedTinkerbellHardware(ctx)
	if err != nil {
		return fmt.Errorf("failed to build catalogue: %v", err)
	}

	if len(hwList) == 0 {
		return errors.New("no available hardware")
	}

	for i := range hwList {
		if err := er.catalogue.InsertHardware(&hwList[i]); err != nil {
			return err
		}
	}

	return nil
}

// GetCatalogue returns the ETCDReader catalogue.
func (er *ETCDReader) GetCatalogue() *Catalogue {
	return er.catalogue
}

// getUnprovisionedTinkerbellHardware fetches the tinkerbell hardware objects on the cluster which do not have an ownerName label.
func (er *ETCDReader) getUnprovisionedTinkerbellHardware(ctx context.Context) ([]tinkv1alpha1.Hardware, error) {
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

	if err := er.client.List(ctx, &selectedHardware, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, fmt.Errorf("listing hardware without owner: %v", err)
	}

	return selectedHardware.Items, nil
}
