---
title: "Customize OVAs: Ubuntu"
linkTitle: "Customize OVAs: Ubuntu"
weight: 30
description: >
  Customizing Imported Ubuntu OVAs
---

There may be a need to make specific configuration changes on the imported ova template before using it to create/update EKS-A clusters. 



## Set up SSH Access for Imported OVA


SSH user and key need to be configured in order to allow SSH login to the VM template


### Clone template to VM

Create an environment variable to hold the name of modified VM/template

```
export VM=<vm-name>
```

Clone the imported OVA template to create VM <vm-name>

```
govc vm.clone -on=false -vm=<full-path-to-imported-template> - folder=<full-path-to-folder-that-will-contain-the-VM> -ds=<datastore> $VM
```

### Configure VM with cloud-init and the VMX GuestInfo datasource

Create a metadata.yaml file

```
instance-id: cloud-vm
local-hostname: cloud-vm
network:
  version: 2
  ethernets:
    nics:
      match:
        name: ens*
      dhcp4: yes
```

Create a userdata.yaml file

```
#cloud-config

users:
  - default
  - name: <username>
    primary_group: <username>
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: sudo, wheel
    ssh_import_id: None
    lock_passwd: true
    ssh_authorized_keys:
    - <user's ssh public key>

```

Export environment variable containing the cloud-init metadata and userdata

```
export METADATA=$(gzip -c9 <metadata.yaml | { base64 -w0 2>/dev/null || base64; }) \
       USERDATA=$(gzip -c9 <userdata.yaml | { base64 -w0 2>/dev/null || base64; })

```

Assign metadata and userdata to VM's guestinfo

```
govc vm.change -vm "${VM}" \
  -e guestinfo.metadata="${METADATA}" \
  -e guestinfo.metadata.encoding="gzip+base64" \
  -e guestinfo.userdata="${USERDATA}" \
  -e guestinfo.userdata.encoding="gzip+base64"
```

Power the VM on

```
govc vm.power -on “$VM”
```


## Customize VM and Convert to Template

Once the VM is powered on and fetches an IP address, ssh into the VM using your private key corresponding to the public key specified in userdata.yaml

```
ssh -i <private-key-file> username@<VM-IP>
```

Make desired config changes on the VM 

Power the VM down

```
govc vm.power -off "$VM"
```

Take a snapshot of the VM (It is highly recommended that you snapshot the VM. This will reduce the time it takes to provision machines and cluster creation will be faster. If you prefer not to take snapshot, skip this step. If you do not snapshot the VM, you will not be able to customize the disk size on your cluster VMs)

```
govc snapshot.create -vm "$VM" root
```

Convert VM to template

```
govc vm.markastemplate $VM
```

Tag the template appropriately as described [here]({{< ref "/docs/reference/vsphere/vsphere-ovas#important-additional-steps-to-tag-the-ova" >}})

Use this customized template to create/upgrade EKS Anywhere clusters
