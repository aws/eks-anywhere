package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/tinkerbell/tink/workflow"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	TinkerbellMachineConfigKind              = "TinkerbellMachineConfig"
	TinkerbellDefaultMachineTemplateHegelURL = "http://tinkerbell-example:50061"
	TinkerbellDefaultMachineTemplateImageURL = "http://tinkerbell-example:8080/ubuntu-2004-kube-v1.21.5.gz"
)

// Used for generating yaml for generate clusterconfig command
func NewTinkerbellMachineConfigGenerate(name string) (*TinkerbellMachineConfigGenerate, error) {
	templateOverride, err := GetTinkerbellDefaultTemplateOverride(name).ToString()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TinkerbellMachineConfig: %v", err)
	}

	return &TinkerbellMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: TinkerbellMachineConfigSpec{
			OSFamily: Ubuntu,
			Users: []UserConfiguration{{
				Name:              "ec2-user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
			}},
			TemplateOverride: templateOverride,
		},
	}, nil
}

func GetTinkerbellDefaultTemplateOverride(name string) *TinkerbellTemplate {
	return &TinkerbellTemplate{
		Version:       "0.1",
		Name:          name,
		GlobalTimeout: 6000,
		Tasks: []workflow.Task{{
			Name:       name,
			WorkerAddr: "{{.device_1}}",
			Volumes: []string{
				"/dev:/dev",
				"/dev/console:/dev/console",
				"/lib/firmware:/lib/firmware:ro",
			},
			Actions: []workflow.Action{
				{
					Name:    "stream-image",
					Image:   "image2disk:v1.0.0",
					Timeout: 360,
					Environment: map[string]string{
						"IMG_URL":    TinkerbellDefaultMachineTemplateImageURL,
						"DEST_DISK":  "/dev/sda",
						"COMPRESSED": "true",
					},
				},
				{
					Name:    "install-openssl",
					Image:   "cexec:v1.0.0",
					Timeout: 90,
					Environment: map[string]string{
						"BLOCK_DEVICE":        "/dev/sda1",
						"FS_TYPE":             "ext4",
						"CHROOT":              "y",
						"DEFAULT_INTERPRETER": "/bin/sh -c",
						"CMD_LINE":            "apt -y update && apt -y install openssl",
					},
				},
				{
					Name:    "write-netplan",
					Image:   "writefile:v1.0.0",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "/dev/sda1",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/netplan/config.yaml",
						"CONTENTS": `network:
  version: 2
  renderer: networkd
  ethernets:
    eno1:
        dhcp4: true
    eno2:
        dhcp4: true
    eno3:
        dhcp4: true
    eno4:
        dhcp4: true`,
						"UID":     "0",
						"GID":     "0",
						"MODE":    "0644",
						"DIRMODE": "0755",
					},
				},
				{
					Name:    "add-tink-cloud-init-config",
					Image:   "writefile:v1.0.0",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "/dev/sda1",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/cloud.cfg.d/10_tinkerbell.cfg",
						"CONTENTS": fmt.Sprintf(`datasource:
  Ec2:
    metadata_urls: ["%s"]
    strict_id: false
system_info:
  default_user:
    name: tink
    groups: [wheel, adm]
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    shell: /bin/bash
manage_etc_hosts: localhost
warnings:
  dsid_missing_source: off`, TinkerbellDefaultMachineTemplateHegelURL),
						"UID":     "0",
						"GID":     "0",
						"MODE":    "0600",
						"DIRMODE": "0700",
					},
				},
				{
					Name:    "add-tink-cloud-init-ds-config",
					Image:   "writefile:v1.0.0",
					Timeout: 90,
					Environment: map[string]string{
						"DEST_DISK": "/dev/sda1",
						"FS_TYPE":   "ext4",
						"DEST_PATH": "/etc/cloud/ds-identify.cfg",
						"CONTENTS":  `datasource: Ec2`,
						"UID":       "0",
						"GID":       "0",
						"MODE":      "0600",
						"DIRMODE":   "0700",
					},
				},
				{
					Name:    "kexec-image",
					Image:   "kexec:v1.0.0",
					Timeout: 90,
					Pid:     "host",
					Environment: map[string]string{
						"BLOCK_DEVICE": "/dev/sda1",
						"FS_TYPE":      "ext4",
					},
				},
			},
		}},
	}
}

func (c *TinkerbellMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *TinkerbellMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *TinkerbellMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetTinkerbellMachineConfigs(fileName string) (map[string]*TinkerbellMachineConfig, error) {
	configs := make(map[string]*TinkerbellMachineConfig)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), YamlSeparator) {
		var config TinkerbellMachineConfig
		if err = yaml.UnmarshalStrict([]byte(c), &config); err == nil {
			if config.Kind == TinkerbellMachineConfigKind {
				configs[config.Name] = &config
				continue
			}
		}
		_ = yaml.Unmarshal([]byte(c), &config) // this is to check if there is a bad spec in the file
		if config.Kind == TinkerbellMachineConfigKind {
			return nil, fmt.Errorf("unable to unmarshall content from file due to: %v", err)
		}
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("unable to find kind %v in file", TinkerbellMachineConfigKind)
	}
	return configs, nil
}
