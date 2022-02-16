package hardware

import (
	"fmt"
	"io/ioutil"
	"strings"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type HardwareConfig struct {
	hardwareList []tinkv1alpha1.Hardware
	bmcList      []pbnjv1alpha1.BMC
	secretList   []corev1.Secret
}

func (hc *HardwareConfig) ParseHardwareConfig(hardwareFileName string) error {
	err := hc.setHardwareConfigFromFile(hardwareFileName)
	if err != nil {
		return fmt.Errorf("unable to parse hardware file %s: %v", hardwareFileName, err)
	}
	return nil
}

func (hc *HardwareConfig) setHardwareConfigFromFile(hardwareFileName string) error {
	content, err := ioutil.ReadFile(hardwareFileName)
	if err != nil {
		return fmt.Errorf("unable to read file due to: %v", err)
	}

	for _, c := range strings.Split(string(content), v1alpha1.YamlSeparator) {
		var resource unstructured.Unstructured
		if err = yaml.Unmarshal([]byte(c), &resource); err != nil {
			return fmt.Errorf("unable to parse %s\nyaml: %s\n %v", hardwareFileName, c, err)
		}
		switch resource.GetKind() {
		case "Hardware":
			var hardware tinkv1alpha1.Hardware
			err = yaml.UnmarshalStrict([]byte(c), &hardware)
			if err != nil {
				return fmt.Errorf("unable to parse hardware CRD\n%s \n%v", c, err)
			}
			hc.hardwareList = append(hc.hardwareList, hardware)
		case "BMC":
			var bmc pbnjv1alpha1.BMC
			err = yaml.UnmarshalStrict([]byte(c), &bmc)
			if err != nil {
				return fmt.Errorf("unable to parse bmc CRD\n%s \n%v", c, err)
			}
			hc.bmcList = append(hc.bmcList, bmc)
		case "Secret":
			var secret corev1.Secret
			err = yaml.UnmarshalStrict([]byte(c), &secret)
			if err != nil {
				return fmt.Errorf("unable to parse k8s secret\n%s \n%v", c, err)
			}
			hc.secretList = append(hc.secretList, secret)
		}
	}

	return nil
}

func (hc *HardwareConfig) ValidateBmcRefMapping() error {
	bmcRefMap := hc.initBmcRefMap()
	for _, hw := range hc.hardwareList {
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

	return nil
}

func (hc *HardwareConfig) ValidateBMC() error {
	secretRefMap := hc.initSecretRefMap()
	bmcIpMap := make(map[string]struct{}, len(hc.bmcList))
	for _, bmc := range hc.bmcList {
		if bmc.Spec.AuthSecretRef.Name == "" {
			return fmt.Errorf("invalid authSecretRef name %s for bmc %s", bmc.Spec.AuthSecretRef.Name, bmc.Name)
		}

		if bmc.Spec.AuthSecretRef.Namespace != constants.EksaSystemNamespace {
			return fmt.Errorf("invalid authSecretRef namespace %s for bmc %s", bmc.Spec.AuthSecretRef.Namespace, bmc.Name)
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
	}

	return nil
}

func (hc *HardwareConfig) initBmcRefMap() map[string]*tinkv1alpha1.Hardware {
	bmcRefMap := make(map[string]*tinkv1alpha1.Hardware, len(hc.bmcList))
	for _, bmc := range hc.bmcList {
		bmcRefMap[bmc.Name] = nil
	}

	return bmcRefMap
}

func (hc *HardwareConfig) initSecretRefMap() map[string]struct{} {
	secretRefMap := make(map[string]struct{}, len(hc.secretList))
	for _, s := range hc.secretList {
		secretRefMap[s.Name] = struct{}{}
	}

	return secretRefMap
}
