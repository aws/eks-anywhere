---
title: "User permissions"
weight: 20
description: >
  The permissions needed by the EKS Anywhere vSphere user
---

## Role Creation and Assignment

In order to assign specific permissions to an EKS Anywhere vSphere user account, you will first need to create a role using the vCenter console. Alternatively, you may decide to use an existing role and set the necessary permissions on it.

### Create Role

Go to Administration->Access Control->Roles. Click on '+' to add a new role. Select the desired permissions category from the left panel, and specific permissions within the category using the checkbox items on the right panel.

### Modify Role Permissions

In order to modify permissions on an existing role, 
Go to Administration->Access Control->Roles. Select the desired role and click on the 'Edit role action' icon that resembles a pencil. Select the desired permissions category from the left panel, and specific permissions within the category using the checkbox items on the right panel.

### Assign Role to User/Group

Select Administration->Access Control->Global Permissions. Select the desired User/Group and click on the 'Change Role' icon that resembles a pencil. Select the desired role from the Roles drop-down to assign to the User/Group.


## EKS Anywhere vSphere User Permissions

Below are the permissions needed by the role assigned to an EKS Anywhere vSphere user to be able to successfully create/update/delete EKS-A clusters.

### Content Library

* Add library item
* Check in a template
* Check out a template
* Create local library

### Datastore

* Allocate space
* Browse datastore
* Low level file operations
 
### Folder

* Create folder

### vSphere Tagging

* Assign or Unassign vSphere Tag
* Assign or Unassign vSphere Tag on Object
* Create vSphere Tag
* Create vSphere Tag Category

### Network

* Assign network

### Resource

* Assign virtual machine to resource pool

### Scheduled task

* Create tasks
* Modify task
* Remove task
* Run task

### Profile-driven storage

* Profile-driven storage view

### Storage views

* View

### vApp

* Import

### Virtual machine

##### Change Configuration

* Add existing disk
* Add new disk
* Add or remove device
* Advanced configuration
* Change CPU count
* Change Memory
* Change Settings
* Change resource
* Configure Raw device
* Extend virtual disk
* Modify device settings
* Remove disk
 
##### Edit Inventory

* Create from existing
* Create new
* Remove

##### Interaction

* Power off
* Power on
 
##### Provisioning

* Clone template
* Clone virtual machine
* Create template from virtual machine
* Customize guest
* Deploy template
* Mark as template
* Read customization specifications
 
##### Snapshot management

* Create snapshot
