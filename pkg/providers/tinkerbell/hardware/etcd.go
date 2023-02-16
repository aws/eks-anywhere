package hardware

import (
	"context"
	"errors"
	"fmt"

	rufiov1alpha1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// OwnerNameLabel is the label set by CAPT to mark a hardware as part of a cluster.
const (
	OwnerNameLabel      = "v1alpha1.tinkerbell.org/ownerName"
	OwnerNamespaceLabel = "v1alpha1.tinkerbell.org/ownerNamespace"
)

// ETCDReader reads the tinkerbell hardware objects from the cluster.
// It holds the objects in a catalogue.
type ETCDReader struct {
	client             client.Client
	catalogue          *Catalogue
	hardwareCache      []tinkv1alpha1.Hardware
	hardwareCacheIndex int
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

// NewMachineCatalogueFromETCD fetches all the hardware, bmc and related secret objects and inserts in to ETCDReader catalogue.
func (er *ETCDReader) NewMachineCatalogueFromETCD(ctx context.Context) error {
	catalogueWriter := NewMachineCatalogueWriter(er.catalogue)
	hwList, err := er.getAllTinkerbellHardware(ctx)
	if err != nil {
		return fmt.Errorf("failed to build catalogue: %v", err)
	}

	for _, hw := range hwList {
		rufioMachine, err := er.getHardwareRufioMachine(ctx, &hw)
		if err != nil {
			return err
		}
		if rufioMachine == nil {
			continue
		}

		authSecret, err := er.getHardwareRufioAuthSecret(ctx, rufioMachine)
		if err != nil {
			return err
		}

		machine := NewMachineFromHardware(hw, rufioMachine, authSecret)
		catalogueWriter.Write(*machine)
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

	if err := er.client.List(ctx, &selectedHardware, &client.ListOptions{LabelSelector: selector}, client.InNamespace(constants.EksaSystemNamespace)); err != nil {
		return nil, fmt.Errorf("listing hardware without owner: %v", err)
	}

	return selectedHardware.Items, nil
}

// getAllTinkerbellHardware fetches all the tinkerbell hardware objects on the cluster.
func (er *ETCDReader) getAllTinkerbellHardware(ctx context.Context) ([]tinkv1alpha1.Hardware, error) {
	var allHardware tinkv1alpha1.HardwareList
	if err := er.client.List(ctx, &allHardware, &client.ListOptions{}, client.InNamespace(constants.EksaSystemNamespace)); err != nil {
		return nil, fmt.Errorf("listing hardware with owner: %v", err)
	}

	return allHardware.Items, nil
}

// getHardwareRufioMachine fetches the rufio machine with the Hardware if any.
func (er *ETCDReader) getHardwareRufioMachine(ctx context.Context, hw *tinkv1alpha1.Hardware) (*rufiov1alpha1.Machine, error) {
	if hw.Spec.BMCRef == nil {
		// Return nil, nil if there is no machine specified for this hardware.
		return nil, nil
	}

	rufioMachine := &rufiov1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hw.Spec.BMCRef.Name,
			Namespace: hw.Namespace,
		},
	}
	if err := er.client.Get(ctx, client.ObjectKeyFromObject(rufioMachine), rufioMachine); err != nil {
		return nil, fmt.Errorf("getting rufio machine for hardware %s: %v", hw.Name, err)
	}

	return rufioMachine, nil
}

// getHardwareRufioAuthSecret fetches the rufio machine auth secret associated with it.
func (er *ETCDReader) getHardwareRufioAuthSecret(ctx context.Context, r *rufiov1alpha1.Machine) (*corev1.Secret, error) {
	authSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Spec.Connection.AuthSecretRef.Name,
			Namespace: r.Spec.Connection.AuthSecretRef.Namespace,
		},
	}
	if err := er.client.Get(ctx, client.ObjectKeyFromObject(authSecret), authSecret); err != nil {
		return nil, fmt.Errorf("getting auth secret for rufio machine %s: %v", r.Name, err)
	}

	return authSecret, nil
}
