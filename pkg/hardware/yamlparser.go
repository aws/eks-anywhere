package hardware

import (
	"fmt"
	"io/ioutil"
	"os"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	yamlPath      = "hardware-manifests/hardware.yaml"
	yamlSeparator = "---\n"
)

type YamlParser struct {
	file *os.File
}

func NewYamlParser() (*YamlParser, error) {
	err := ioutil.WriteFile(yamlPath, []byte{}, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error initializing YamlParser: %v", err)
	}

	file, err := os.OpenFile(yamlPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error initializing YamlParser: %v", err)
	}
	return &YamlParser{
		file: file,
	}, nil
}

func (y *YamlParser) Close() {
	y.file.Close()
}

func (y *YamlParser) WriteHardwareYaml(id, hostname, bmcIp, vendor, username, password string) error {
	bmcRef := fmt.Sprintf("bmc-%s", hostname)
	secretRef := fmt.Sprintf("%s-auth", bmcRef)
	hardware := tinkv1alpha1.Hardware{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Hardware",
			APIVersion: "tinkerbell.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostname,
			Namespace: eksaNamespace,
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			ID:     id,
			BmcRef: bmcRef,
		},
	}

	h, err := yaml.Marshal(hardware)
	if err != nil {
		return err
	}

	if _, err := y.file.Write(h); err != nil {
		return fmt.Errorf("error writing hardware object: %v", err)
	}

	if _, err := y.file.WriteString(yamlSeparator); err != nil {
		return fmt.Errorf("error writing yaml separator: %v", err)
	}

	bmc := pbnjv1alpha1.BMC{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BMC",
			APIVersion: "tinkerbell.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bmcRef,
			Namespace: eksaNamespace,
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
		},
		Spec: pbnjv1alpha1.BMCSpec{
			Host:   bmcIp,
			Vendor: vendor,
			AuthSecretRef: corev1.SecretReference{
				Name:      secretRef,
				Namespace: eksaNamespace,
			},
		},
	}

	b, err := yaml.Marshal(bmc)
	if err != nil {
		return err
	}

	if _, err := y.file.Write(b); err != nil {
		return fmt.Errorf("error writing bmc object: %v", err)
	}

	if _, err := y.file.WriteString(yamlSeparator); err != nil {
		return fmt.Errorf("error writing yaml separator: %v", err)
	}

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretRef,
			Namespace: eksaNamespace,
			Labels: map[string]string{
				"clusterctl.cluster.x-k8s.io/move": "true",
			},
		},
		Type: "kubernetes.io/basic-auth",
		Data: map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		},
	}

	s, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	if _, err := y.file.Write(s); err != nil {
		return fmt.Errorf("error writing secret object: %v", err)
	}

	if _, err := y.file.WriteString(yamlSeparator); err != nil {
		return fmt.Errorf("error writing yaml separator: %v", err)
	}

	return nil
}
