# Required configuration for EKS Anywhere Tinkerbell provider

## Introduction

The purpose of this document is to explain the different input files required to create an EKS Anywhere Tinkerbell cluster and how to generate them.

### Goals and objectives

As an EKS Anywhere Tinkerbell provider user, I want to:

* Generate the required cluster configuration file with minimal customizations
* Generate the required hardware files using CSV files that contain hardware information

### Statement of Scope

The scope of this document is mainly for beta release of Tinkerbell provider. Some of the details in the design will change once Tinkerbell is [migrated completely to Kubernetes](https://github.com/tinkerbell/proposals/tree/main/proposals/0026).

**In Scope**
* Explain the user configuration required for setting up an EKS Anywhere cluster using Tinkerbell
* Provide users a way to easily generate the required cluster config and hardware manifests
* Push the hardware directly to the Tinkerbell stack so users don't have to do it manually

## Solution details

### Cluster Configuration 

The first file users will need is the clusterconfig file, which they can use to configure their EKS Anywhere cluster.

To generate the clusterconfig for Tinkerbell provider, we use the existing `generate` command and specify `tinkerbell` as the provider
```bash
eksctl anywhere generate clusterconfig <cluster-name> -p tinkerbell
```

For Tinkerbell provider, we have introduced two new CRDs:
* TinkerbellDatacenterConfig
* TinkerbellMachineConfig

**TinkerbellDatacenterConfig**

TinkerbellDatacenterConfig describes the configuration details for Tinkerbell stack.

In this config, users must specify the following fields:
* spec.tinkerbellIP: defines the IP of the machine running the Tinkerbell stack.
* spec.tinkerbellCertURL: defines the endpoint where the Tinkerbell stack hosts its certificate. This is used to communicate with the stack.
* spec.tinkerbellGRPCAuth: defines the GRPC endpoint used by Tinkerbell stack. This is also used to communicate with the stack.
* spec.tinkerbellPBnJGRPCAuth: defines the GRPC endpoint used by PNBJ for hardware remote management. This is used to change the power state and the boot order of the servers.

Here's an example `TinkerbellDatacenterConfig`

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellDatacenterConfig
metadata:
  name: eks-a-tinkerbell-cluster
spec:
  tinkerbellIP: "192.168.0.20"
  tinkerbellCertURL: "http://192.168.0.20:42114/cert"
  tinkerbellGRPCAuth: "192.168.0.20:42113"
  tinkerbellPBnJGRPCAuth: "192.168.0.20:50051"
```

**TinkerbellMachineConfig**

TinkerbellMachineConfig describes the configuration for machines used for creating EKS Anywhere cluster on Tinkerbell.

This config supports the following fields
* spec.osFamily (required): defines the operating system to be installed on this machine. We currently only support ubuntu for this but this list is expected to grow in the future.
* spec.users (required): defines the users to be created on this machine along with the ssh keys.
  * spec.users[].name (required): defines the name of the user.
  * spec.users[].sshAuthorizedKeys (required): defines the SSH authorized keys to be installed on this machine to be used for SSH'ing into it.
* spec.templateOverride (optional): defines the [Tinkerbell template](https://docs.tinkerbell.org/templates/) to be used to provision the OS on the machine. This field is passed as a YAML encoded as string.

Here is an example `TinkerbellMachineConfig`
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: TinkerbellMachineConfig
metadata:
  name: eks-a-tinkerbell-cluster-cp
spec:
  osFamily: ubuntu
  users:
  - name: ec2-user
    sshAuthorizedKeys:
    - ssh-rsa AAAA...
  templateOverride: |
    <>
```

Here is the default `templateOverride` that is generated in the clusterconfig
```yaml
version: "0.1"
name: eks-a-tinkerbell-cluster-cp-test
global_timeout: 6000
tasks:
  - name: "eks-a-tinkerbell-cluster-cp-ndrl8"
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
          IMG_URL: "https://<some-s3-endpoint>/ubuntu-2004-kube-v1.21.5.gz"
          DEST_DISK: /dev/sda
          COMPRESSED: true
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
                metadata_urls: ["http://<tinkerbellIP>:50061"]
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
```

### Hardware manifests

Tinkerbell provider also requires hardware manifest files along with the clusterconfig to provide configuration details about the hardware. There are two hardware manifest files involved: a hardware JSON and a hardware YAML. 

>**_Note_**: There is some ongoing work here and this may change post-beta

**Hardware JSON**

This file contains details about the hardware like the storage and network configurations.
For every hardware that the user wants to provision and manage using Tinkerbell, they need this file.

Here is an example hardware JSON for one hardware
```json
{
  "id": "a238d17d-0964-4655-9d18-7789fc823d32",
  "metadata": {
    "facility": {
      "facility_code": "onprem",
      "plan_slug": "c2.medium.x86",
      "plan_version_slug": ""
    },
    "instance": {
      "id": "a238d17d-0964-4655-9d18-7789fc823d32",
      "hostname": "worker1",
      "storage": {
        "disks": [
          {
            "device": "/dev/sda"
          }
        ]
      }
    },
    "state": "provisioning"
  },
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "mac": "00:00:00:00:00:00",
          "nameservers": [
            "8.8.8.8",
            "8.8.4.4"
          ],
          "uefi": true,
          "ip": {
            "address": "192.168.1.10",
            "gateway": "192.168.0.1",
            "netmask": "255.255.255.0"
          }
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}
```

**Hardware YAML**

The hardware is passed into our cli and informs CAPT which machines to use for creating clusters.
For every hardware that the user wants to add to the cluster, they must specify these three sections:

* Hardware: Used to specify which hardware from the Tinkerbell stack should be used to create the cluster
  * Hardware.spec.id: ID of the hardware specified in the hardware JSON to tell CAPT to query that hardware from Tinkerbell.
  * Hardware.spec.bmcRef: Reference to the BMC object for this hardware

* BMC: Bareboard Management Controller related information for the hardware, used to perform actions like power on/power off and changing the boot order
  * BMC.spec.host: The endpoint where the BMC for this hardware is running
  * BMC.spec.vendor: Name of BMC's vendor
  * BMC.spec.authSecretRef: Reference to BMC's user credentials

* Secret: Used to provide credentials for BMC
  * Secret.data.username: Base64 encoded username for BMC
  * Secret.data.password: Base64 encoded password for BMC
  * Secret.type: Type of secret which should be set to `kubernetes.io/basic-auth`

Here is an example Hardware YAML for one hardware
```yaml
apiVersion: tinkerbell.org/v1alpha1
kind: Hardware
metadata:
  creationTimestamp: null
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: worker1
  namespace: eksa-system
spec:
  bmcRef: bmc-worker1
  id: 16035822-7813-42b2-ae9a-04eae8574a73
status: {}
---
apiVersion: tinkerbell.org/v1alpha1
kind: BMC
metadata:
  creationTimestamp: null
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-worker1
  namespace: eksa-system
spec:
  authSecretRef:
    name: bmc-worker1-auth
    namespace: eksa-system
  host: 192.168.0.10
  vendor: supermicro
status: {}
---
apiVersion: v1
data:
  password: QWRtaW4=
  username: YWRtaW4=
kind: Secret
metadata:
  creationTimestamp: null
  labels:
    clusterctl.cluster.x-k8s.io/move: "true"
  name: bmc-worker1-auth
  namespace: eksa-system
type: kubernetes.io/basic-auth
---
```

**Generating Hardware Manifests**

Users need to generate and maintain one copy of hardware JSON and hardware YAML for each hardware they want to use with EKS Anywhere. This can get tedious. So, we are adding a new subcommand under `eksctl anywhere generate` to make this easier.

This new command will take a CSV file as input with configurations for all the hardware, parse it and output the required hardware manifests. This will also by default push the hardware to the Tinkerbell stack.

The CSV file will need these headers to identify which field corresponds to which value:
* ip_address: IP address to configure on the machine
* gateway: Gateway to configure on the machine
* nameservers: List of nameservers to be specified on the machine, delimited by `|`
* netmask: Netmask for the IP
* mac: MAC address of the interface to assign the IP to
* hostname: Hostname of the machine
* vendor: Name of the machine's vendor
* bmc_ip: IP address of the machine's BMC interface
* bmc_username: Username for the machine's BMC
* bmc_password: Password for machine's BMC

All of these fields are required and can be placed in any order in the CSV file.

Here is an example hardware CSV file with the required information for two hardwares:
```csv
ip_address,gateway,nameservers,netmask,mac,hostname,vendor,bmc_ip,bmc_username,bmc_password
192.168.1.10,192.168.0.1,8.8.8.8|8.8.4.4,255.255.255.0,00:00:00:00:00:00,worker1,supermicro,192.168.0.10,admin,Admin
192.168.1.11,192.168.0.1,8.8.8.8|8.8.4.4,255.255.255.0,00:00:00:00:00:01,worker2,supermicro,192.168.0.11,admin,Admin
```

Once the CSV file is set up, users can use the `setup hardware` command to generate and push the hardware manifests

(Recommended) Command to generate the hardware manifests and push it to the Tinkerbell stack
```bash
eksctl anywhere setup hardware -f <path-to-csv-file> --tinkerbell-ip <tinkerbell-stack-ip>
```

(NOT Recommended) Command to just generate the hardware manifests without pushing
```bash
eksctl anywhere generate hardware -f <path-to-csv-file> --dry-run
```

Running the above commands is going a create a new directory called `hardware-manifests` with the manifests. This is what the directory structure will look like:
```
hardware-manifests
├── hardware.yaml
└── json
    ├── worker1.json
    └── worker2.json
```

### Using the generated files to create a cluster

Once the users have created the necessary files, they can use them to create an EKS Anywhere Tinkerbell cluster.

```bash
eksctl anywhere create cluster -f <clusterconfig> -w <hardware YAML>
```

## Future changes

Most of the design details described above are only valid for beta release of Tinkerbell provider as for beta, we are using the default installation of Tinkerbell. This default installation runs Tinkerbell components as microservices on docker containers. We have plans to eliminate Tinkerbell's dependency on docker and instead run it on Kubernetes as controllers. You can read the proposal for this [here](https://github.com/tinkerbell/proposals/tree/main/proposals/0026).

Since this is a work-in-progress, we still have not finalized how our design will look like when we leverage Kubernetes backend for Tinkerbell.
