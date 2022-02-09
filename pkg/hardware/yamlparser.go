package hardware

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	yamlDirectory        = "hardware-manifests"
	yamlFile             = "hardware.yaml"
	hardwareKind         = "Hardware"
	bmcKind              = "BMC"
	secretKind           = "Secret"
	tinkerbellApiVersion = "tinkerbell.org/v1alpha1"
	moveLabel            = "clusterctl.cluster.x-k8s.io/move"
)

var (
	yamlPath      = filepath.Join(yamlDirectory, yamlFile)
	yamlSeparator = []byte("---\n")
)

type YamlParser struct {
	file *os.File
}

func NewYamlParser() (*YamlParser, error) {
	if _, err := os.Stat(yamlDirectory); errors.Is(err, os.ErrNotExist) {
		logger.V(4).Info("Creating directory for YamlParser", "Directory", yamlDirectory)
		if err := os.Mkdir(yamlDirectory, os.ModePerm); err != nil {
			return nil, fmt.Errorf("error creating directory for YamlParser: %v", err)
		}
	}

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
			Kind:       hardwareKind,
			APIVersion: tinkerbellApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostname,
			Namespace: eksaNamespace,
			Labels: map[string]string{
				moveLabel: "true",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			ID:     id,
			BmcRef: bmcRef,
		},
	}

	bmc := pbnjv1alpha1.BMC{
		TypeMeta: metav1.TypeMeta{
			Kind:       bmcKind,
			APIVersion: tinkerbellApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bmcRef,
			Namespace: eksaNamespace,
			Labels: map[string]string{
				moveLabel: "true",
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

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       secretKind,
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretRef,
			Namespace: eksaNamespace,
			Labels: map[string]string{
				moveLabel: "true",
			},
		},
		Type: "kubernetes.io/basic-auth",
		Data: map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		},
	}

	h, err := yaml.Marshal(hardware)
	if err != nil {
		return fmt.Errorf("error marshalling Hardware for %s: %v", hostname, err)
	}

	b, err := yaml.Marshal(bmc)
	if err != nil {
		return fmt.Errorf("error marshalling BMC for %s: %v", hostname, err)
	}

	s, err := yaml.Marshal(secret)
	if err != nil {
		return fmt.Errorf("error marshalling Secret for %s: %v", hostname, err)
	}

	var hardwareYaml []byte
	for _, slice := range [][]byte{h, b, s} {
		hardwareYaml = append(hardwareYaml, slice...)
		hardwareYaml = append(hardwareYaml, yamlSeparator...)
	}

	logger.V(4).Info("Writing hardware yaml for hardware", "Hardware", hostname)
	if _, err := y.file.Write(hardwareYaml); err != nil {
		return fmt.Errorf("error writing hardware yaml for %s: %v", hostname, err)
	}

	return nil
}
