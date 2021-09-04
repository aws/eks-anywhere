---
title: "User permissions"
weight: 20
description: >
  Permissions needed by the EKS Anywhere vSphere user
---

The permissions needed by the EKS Anywhere vSphere user are just short of full administrative access.
Further information about vSphere permissions can be found
[here](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.security.doc/GUID-5372F580-5C23-4E9C-8A4E-EF1B4DD9033E.html).
Here are the recommended user permissions for running an EKS Anywhere cluster on vSphere:
 
## Virtual machine

### Configuration

* Change configuration
* Add existing disk
* Add new disk
* Add or remove device
 
### Advanced configuration

* Change CPU count
* Change memory
* Change settings
* Configure raw devices
* Extend virtual disk
* Modify device settings
* Remove disk
* Create from existing
* Remove
 
### Interaction

* Power off
* Power on
 
### Provisioning

* Deploy template
 
## Datastore

* Allocate space
* List datastore
* Low level file operations
 
## Network

* Assign network
 
## Resource Pool

* Assign a virtual machine to resource pool

