package cmd

const (
	defaultTinkerbellTemplateConfigTemplateBottlerocket = `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
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
          # An example console declaration that will send all kernel output to both consoles, and systemd output to ttyS0.
          # kernel {
          #     console = "tty0", "ttyS0,115200n8"
          # }
          BOOTCONFIG_CONTENTS: |
                        kernel {}
          DEST_DISK: /dev/sda12
          DEST_PATH: /bootconfig.data
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-bootconfig
        timeout: 90
      - environment:
          CONTENTS: |
            # Version is required, it will change as we support
            # additional settings
            version = 1
            # "eno1" is the interface name
            # Users may turn on dhcp4 and dhcp6 via boolean
            [eno1]
            dhcp4 = true
            # Define this interface as the "primary" interface
            # for the system.  This IP is what kubelet will use
            # as the node IP.  If none of the interfaces has
            # "primary" set, we choose the first interface in
            # the file
            primary = true
          DEST_DISK: /dev/sda12
          DEST_PATH: /net.toml
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-netconfig
        timeout: 90
      - environment:
          HEGEL_URLS: [http://$$ADMIN_IP:50061,http://$$TINKERBELL_IP:50061]
          DEST_DISK: /dev/sda12
          DEST_PATH: /user-data.toml
          DIRMODE: "0700"
          FS_TYPE: ext4
          GID: "0"
          MODE: "0644"
          UID: "0"
        image: public.ecr.aws/eks-anywhere/tinkerbell/hub/writefile:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
        name: write-user-data
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

// GetDefaultTinkerbellTemplateConfigTemplateBottlerocket returns the default TinkerbellTemplateConfigTemplate for Bottlerocket.
func GetDefaultTinkerbellTemplateConfigTemplateBottlerocket() string {
	return string(defaultTinkerbellTemplateConfigTemplateBottlerocket)
}
