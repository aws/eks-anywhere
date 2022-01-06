package v1alpha1

import (
	"fmt"
	"io/ioutil"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	TinkerbellMachineConfigKind = "TinkerbellMachineConfig"
	TinkerbellDefaultTemplate   = `version: "0.1"
name: %s-test
global_timeout: 6000
tasks:
  - name: "%s-ndrl8"
    worker: "{{.device_1}}"
    volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
    actions:
      - name: "stream-image"
        image: image2disk:v1.0.0
        timeout: 360
        environment:
          IMG_URL: "http://10.62.2.49:8080/ubuntu-2004-kube-v1.20.11.gz"
          DEST_DISK: /dev/sda
          COMPRESSED: true
      - name: "install-openssl"
        image: cexec:v1.0.0
        timeout: 90
        environment:
          BLOCK_DEVICE: /dev/sda1
          FS_TYPE: ext4
          CHROOT: y
          DEFAULT_INTERPRETER: "/bin/sh -c"
          CMD_LINE: "apt -y update && apt -y install openssl"
      - name: "create-user"
        image: cexec:v1.0.0
        timeout: 90
        environment:
          BLOCK_DEVICE: /dev/sda1
          FS_TYPE: ext4
          CHROOT: y
          DEFAULT_INTERPRETER: "/bin/sh -c"
          CMD_LINE: "useradd -p $(openssl passwd -1 tink) -s /bin/bash -d /home/tink/ -m -G sudo tink"
      - name: "write-netplan"
        image: writefile:v1.0.0
        timeout: 90
        environment:
          DEST_DISK: /dev/sda1
          FS_TYPE: ext4
          DEST_PATH: /etc/netplan/config.yaml
          CONTENTS: |
            network:
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
                        dhcp4: true
          UID: 0
          GID: 0
          MODE: 0644
          DIRMODE: 0755
      - name: "add-tink-cloud-init-config"
        image: writefile:v1.0.0
        timeout: 90
        environment:
          DEST_DISK: /dev/sda1
          FS_TYPE: ext4
          DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg
          UID: 0
          GID: 0
          MODE: 0600
          DIRMODE: 0700
          CONTENTS: |
            datasource:
              Ec2:
                metadata_urls: ["http://10.62.2.49:50061"]
                strict_id: false
            system_info:
              default_user:
                name: tink
                groups: [wheel, adm]
                sudo: ["ALL=(ALL) NOPASSWD:ALL"]
                shell: /bin/bash
            manage_etc_hosts: localhost
            warnings:
              dsid_missing_source: off
      - name: "add-tink-cloud-init-ds-config"
        image: writefile:v1.0.0
        timeout: 90
        environment:
          DEST_DISK: /dev/sda1
          FS_TYPE: ext4
          DEST_PATH: /etc/cloud/ds-identify.cfg
          UID: 0
          GID: 0
          MODE: 0600
          DIRMODE: 0700
          CONTENTS: |
            datasource: Ec2
      - name: "kexec-image"
        image: kexec:v1.0.0
        timeout: 90
        pid: host
        environment:
          BLOCK_DEVICE: /dev/sda1
          FS_TYPE: ext4
`
)

// Used for generating yaml for generate clusterconfig command
func NewTinkerbellMachineConfigGenerate(name string) *TinkerbellMachineConfigGenerate {
	return &TinkerbellMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: TinkerbellMachineConfigSpec{
			OSFamily: Bottlerocket,
			Users: []UserConfiguration{{
				Name:              "ec2-user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
			}},
			TemplateOverride: fmt.Sprintf(TinkerbellDefaultTemplate, name, name),
		},
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
