package cmd

const (
	defaultTinkerbellTemplateConfigTemplateUbuntu = `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellTemplateConfig
metadata:
  name: $$NAME
spec:
  template:
    global_timeout: 6000
    id: ""
    name: $$NAME
    tasks:
    - actions:
      - environment:
          COMPRESSED: "true"
          DEST_DISK: /dev/sda
          IMG_URL: $$IMG_URL
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/image2disk:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: stream-image
        timeout: 720
      - environment:
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/netplan/config.yaml
          STATIC_NETPLAN: true
          DIRMODE: "0755"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-netplan
        timeout: 90
      - environment:
          CONTENTS: |
            datasource:
              Ec2:
                metadata_urls: [http://$$ADMIN_IP:50061,http://$$TINKERBELL_IP:50061]
                strict_id: false
            manage_etc_hosts: localhost
            warnings:
              dsid_missing_source: off
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/cloud/cloud.cfg.d/10_tinkerbell.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: add-tink-cloud-init-config
        timeout: 90
      - environment:
          CONTENTS: |
            network:
              config: disabled
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: disable-cloud-init-network-capabilities
        timeout: 90
      - environment:
          CONTENTS: |
            datasource: Ec2
          DEST_DISK: /dev/sda2
          DEST_PATH: /etc/cloud/ds-identify.cfg
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0600"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: add-tink-cloud-init-ds-config
        timeout: 90
      - environment:
          BLOCK_DEVICE: /dev/sda2
          CHROOT: "y"
          CMD_LINE: apt -y update && apt -y install openssl
          DEFAULT_INTERPRETER: /bin/sh -c
          FS_TYPE: ext4
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/cexec:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: install-openssl
        timeout: 90
      - environment:
          BLOCK_DEVICE: /dev/sda2
          FS_TYPE: ext4
          CHROOT: "y"
          DEFAULT_INTERPRETER: "/bin/sh -c"
          CMD_LINE: "useradd --password $(openssl passwd -1 tinkerbell) --shell /bin/bash --create-home --groups sudo tinkerbell"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/cexec:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: "create-user"
        timeout: 90
      - name: "reboot"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/reboot:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        timeout: 90
        volumes:
          - /worker:/worker
      name: $$NAME
      volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
      worker: '{{.device_1}}'
    version: "0.1"
`
)

// GetDefaultTinkerbellTemplateConfigTemplateUbuntu returns the default TinkerbellTemplateConfigTemplate for Ubuntu.
func GetDefaultTinkerbellTemplateConfigTemplateUbuntu() string {
	return string(defaultTinkerbellTemplateConfigTemplateUbuntu)
}
