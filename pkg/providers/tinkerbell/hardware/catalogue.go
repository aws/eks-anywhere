package hardware

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	tinkhardware "github.com/tinkerbell/tink/protos/hardware"
	tinkworkflow "github.com/tinkerbell/tink/protos/workflow"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apimachineryvalidation "k8s.io/apimachinery/pkg/util/validation"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const Provisioning = "provisioning"

type Catalogue struct {
	Hardware []tinkv1alpha1.Hardware
	BMCs     []pbnjv1alpha1.BMC
	Secrets  []corev1.Secret
}

func (c *Catalogue) ValidateHardware(skipPowerActions, force bool, tinkHardwareMap map[string]*tinkhardware.Hardware, tinkWorkflowMap map[string]*tinkworkflow.Workflow) error {
	bmcRefMap := map[string]*tinkv1alpha1.Hardware{}
	if !skipPowerActions {
		bmcRefMap = c.initBmcRefMap()
	}

	// A database of observed hardware IDs so we can check for uniqueness.
	hardwareIdsDb := make(map[string]struct{}, len(c.Hardware))

	for _, hw := range c.Hardware {
		if hw.Name == "" {
			return fmt.Errorf("hardware name is required")
		}

		if errs := apimachineryvalidation.IsDNS1123Subdomain(hw.Name); len(errs) > 0 {
			return fmt.Errorf("invalid hardware name: %v: %v", hw.Name, errs)
		}

		if hw.Spec.ID == "" {
			return fmt.Errorf("hardware: %s ID is required", hw.Name)
		}

		if _, found := hardwareIdsDb[hw.Spec.ID]; found {
			return fmt.Errorf("duplicate hardware id: %v", hw.Spec.ID)
		}
		hardwareIdsDb[hw.Spec.ID] = struct{}{}

		if _, ok := tinkHardwareMap[hw.Spec.ID]; !ok {
			return fmt.Errorf("hardware id '%s' is not registered with tinkerbell stack", hw.Spec.ID)
		}

		hardwareInterface := tinkHardwareMap[hw.Spec.ID].GetNetwork().GetInterfaces()
		for _, interfaces := range hardwareInterface {
			mac := interfaces.GetDhcp()
			if _, ok := tinkWorkflowMap[mac.Mac]; ok {
				message := fmt.Sprintf("workflow %s already exixts for the hardware id %s", tinkWorkflowMap[mac.Mac].Id, hw.Spec.ID)

				// If the --force-cleanup flag was set we have to warn. This is beacuse we haven't separated static
				// and interactive validations so there's no opportunity, after performing static yaml validation, to
				// delete any workflows before we check if workflows alredy exist. To avoid erroring out before getting
				// the chance to delete workflows this code assumes workflows will be deleted at a later stage,
				// therefore warns only.
				if !force {
					return fmt.Errorf(message)
				}
				logger.V(2).Info(fmt.Sprintf("Warn: %v", message))
			}
		}

		if !force {
			hardwareMetadata := make(map[string]interface{})
			tinkHardware := tinkHardwareMap[hw.Spec.ID]

			if err := json.Unmarshal([]byte(tinkHardware.GetMetadata()), &hardwareMetadata); err != nil {
				return fmt.Errorf("unmarshaling hardware metadata: %v", err)
			}

			if hardwareMetadata["state"] != Provisioning {
				return fmt.Errorf("expecting hardware state to be '%s' but it is '%s'; use --force-cleanup flag to reset the state", "provisioning", hardwareMetadata["state"])
			}
		}

		if !skipPowerActions {
			if hw.Spec.BmcRef == "" {
				return fmt.Errorf("bmcRef not present in hardware %s", hw.Name)
			}

			h, ok := bmcRefMap[hw.Spec.BmcRef]
			if ok && h != nil {
				return fmt.Errorf("bmcRef %s present in both hardware %s and hardware %s", hw.Spec.BmcRef, hw.Name, h.Name)
			}
			if !ok {
				return fmt.Errorf("bmcRef %s not found in hardware config", hw.Spec.BmcRef)
			}

			bmcRefMap[hw.Spec.BmcRef] = &hw
		}
	}

	return nil
}

func (c *Catalogue) ValidateBMC() error {
	secretRefMap := c.initSecretRefMap()
	bmcIpMap := make(map[string]struct{}, len(c.BMCs))
	for _, bmc := range c.BMCs {
		if bmc.Name == "" {
			return fmt.Errorf("bmc name is required")
		}

		if bmc.Spec.AuthSecretRef.Name == "" {
			return fmt.Errorf("authSecretRef name required for bmc %s", bmc.Name)
		}

		if bmc.Spec.AuthSecretRef.Namespace != constants.EksaSystemNamespace {
			return fmt.Errorf("invalid authSecretRef namespace: %s for bmc %s", bmc.Spec.AuthSecretRef.Namespace, bmc.Name)
		}

		if _, ok := secretRefMap[bmc.Spec.AuthSecretRef.Name]; !ok {
			return fmt.Errorf("bmc authSecretRef: %s not present in hardware config", bmc.Spec.AuthSecretRef.String())
		}

		if _, ok := bmcIpMap[bmc.Spec.Host]; ok {
			return fmt.Errorf("duplicate host IP: %s for bmc %s", bmc.Spec.Host, bmc.Name)
		} else {
			bmcIpMap[bmc.Spec.Host] = struct{}{}
		}

		if err := networkutils.ValidateIP(bmc.Spec.Host); err != nil {
			return fmt.Errorf("bmc host IP: %v", err)
		}

		if bmc.Spec.Vendor == "" {
			return fmt.Errorf("bmc: %s vendor is required", bmc.Name)
		}
	}

	return nil
}

func (c *Catalogue) ValidateBmcSecretRefs() error {
	for _, s := range c.Secrets {
		if s.Name == "" {
			return fmt.Errorf("secret name is required")
		}
		if s.Namespace != constants.EksaSystemNamespace {
			return fmt.Errorf("invalid secret namespace: %s for secret: %s expected: %s", s.Namespace, s.Name, constants.EksaSystemNamespace)
		}
		dUsr, dOk := s.Data["username"]
		sdUsr, sdOk := s.StringData["username"]
		if !dOk && !sdOk {
			return fmt.Errorf("secret: %s must contain key username", s.Name)
		}
		if (dOk && len(dUsr) == 0) || (sdOk && sdUsr == "") {
			return fmt.Errorf("username can not be empty for secret: %s", s.Name)
		}

		dPwd, dOk := s.Data["password"]
		sdPwd, sdOk := s.StringData["password"]
		if !dOk && !sdOk {
			return fmt.Errorf("secret: %s must contain key password", s.Name)
		}
		if (dOk && len(dPwd) == 0) || (sdOk && sdPwd == "") {
			return fmt.Errorf("password can not be empty for secret: %s", s.Name)
		}
	}

	return nil
}

func (c *Catalogue) initBmcRefMap() map[string]*tinkv1alpha1.Hardware {
	bmcRefMap := make(map[string]*tinkv1alpha1.Hardware, len(c.BMCs))
	for _, bmc := range c.BMCs {
		bmcRefMap[bmc.Name] = nil
	}

	return bmcRefMap
}

func (c *Catalogue) initSecretRefMap() map[string]struct{} {
	secretRefMap := make(map[string]struct{}, len(c.Secrets))
	for _, s := range c.Secrets {
		secretRefMap[s.Name] = struct{}{}
	}

	return secretRefMap
}

func (c *Catalogue) HardwareSpecMarshallable() ([]byte, error) {
	var marshallables []v1alpha1.Marshallable

	for _, hw := range c.Hardware {
		marshallables = append(marshallables, hardwareMarshallable(hw))
	}
	for _, bmc := range c.BMCs {
		marshallables = append(marshallables, bmcMarshallable(bmc))
	}
	for _, secret := range c.Secrets {
		marshallables = append(marshallables, secretsMarshallable(secret))
	}

	resources := make([][]byte, 0, len(marshallables))
	for _, marshallable := range marshallables {
		resource, err := yaml.Marshal(marshallable)
		if err != nil {
			return nil, fmt.Errorf("failed marshalling resource for hardware spec: %v", err)
		}
		resources = append(resources, resource)
	}
	return templater.AppendYamlResources(resources...), nil
}

func hardwareMarshallable(hw tinkv1alpha1.Hardware) *tinkv1alpha1.Hardware {
	config := &tinkv1alpha1.Hardware{
		TypeMeta: hw.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        hw.Name,
			Annotations: hw.Annotations,
			Namespace:   hw.Namespace,
			Labels:      hw.Labels,
		},
		Spec: hw.Spec,
	}

	return config
}

func bmcMarshallable(bmc pbnjv1alpha1.BMC) *pbnjv1alpha1.BMC {
	config := &pbnjv1alpha1.BMC{
		TypeMeta: bmc.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        bmc.Name,
			Annotations: bmc.Annotations,
			Namespace:   bmc.Namespace,
			Labels:      bmc.Labels,
		},
		Spec: bmc.Spec,
	}

	return config
}

func secretsMarshallable(secret corev1.Secret) *corev1.Secret {
	config := &corev1.Secret{
		TypeMeta: secret.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        secret.Name,
			Annotations: secret.Annotations,
			Namespace:   secret.Namespace,
			Labels:      secret.Labels,
		},
		Data: secret.Data,
	}

	return config
}

// ParseYAMLCatalogueFromFile parses filename, a YAML document, using ParseYamlCatalogue.
func ParseYAMLCatalogueFromFile(config *Catalogue, filename string) error {
	fh, err := os.Open(filename)
	if err != nil {
		return err
	}

	return ParseYAMLCatalogue(config, fh)
}

// ParseCatalogue parses a YAML document, r, that represents a set of Kubernetes manifests.
// Manifests parsed include CAPT Hardware, PBnJ BMCs and associated Core API Secret.
func ParseYAMLCatalogue(catalogue *Catalogue, r io.Reader) error {
	document := yamlutil.NewYAMLReader(bufio.NewReader(r))
	for {
		manifest, err := document.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		var resource unstructured.Unstructured
		if err = yaml.Unmarshal(manifest, &resource); err != nil {
			return err
		}

		switch resource.GetKind() {
		case "Hardware":
			var hardware tinkv1alpha1.Hardware
			err = yaml.UnmarshalStrict(manifest, &hardware)
			if err != nil {
				return fmt.Errorf("unable to parse hardware manifest: %v", err)
			}
			catalogue.Hardware = append(catalogue.Hardware, hardware)
		case "BMC":
			var bmc pbnjv1alpha1.BMC
			err = yaml.UnmarshalStrict(manifest, &bmc)
			if err != nil {
				return fmt.Errorf("unable to parse bmc manifest: %v", err)
			}
			catalogue.BMCs = append(catalogue.BMCs, bmc)
		case "Secret":
			var secret corev1.Secret
			err = yaml.UnmarshalStrict(manifest, &secret)
			if err != nil {
				return fmt.Errorf("unable to parse secret manifest: %v", err)
			}
			catalogue.Secrets = append(catalogue.Secrets, secret)
		}
	}
}
