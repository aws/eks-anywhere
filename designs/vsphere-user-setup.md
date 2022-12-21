# vSphere User Creation

## Introduction

**Problem:** the process of configuring a vSphere user to create an EKS-A management cluster requires a significant number of permissions on specific objects, adding friction for the customer in adopting EKS-A.

The current process for provisioning a user and configuring their permissions is entirely manual and involves a non-trivial amount of pointing and clicking.

Specifically, the vSphere admin needs to create three roles [per our docs](https://anywhere.eks.amazonaws.com/docs/reference/vsphere/vsphere-preparation/#create-and-define-user-roles):

* A “Global” role applied to all vSphere objects
* A “User” role to apply to the Network, Datastore, and ResourcePool objects that EKS-A will use
* An “Admin” role applied to the vSphere EKS-A template and VM folders

The Kubernetes admin often does not have admin rights to the vSphere cluster they will create a cluster on, and has to ask an administrator to perform the user configuration on their behalf.

If the vSphere user is mis-configured due to human error, this introduces back and forth communication for the customer with the vSphere admin.

The net effect of all this is more friction in the onboarding process.

### Goals and Objectives

As a customer trying to create an EKS-A management cluster on vSphere, I want to:

* configure a vSphere user for the cluster with appropriate permissions and minimal hassle

### Statement of Scope

#### In Scope

* Support provisioning and configuring a vSphere user’s permissions via a eksctl anywhere command

#### Future Scope

* Support provisioning separate primary, vSphere CSI Driver, and Cluster Provider users as specified [here](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/vsphere/#optional-vsphere-credentials)
* Support an EKS-A CLI wizard to generate config file

## Overview of Solution

We propose solving this problem with an `eksctl anywhere` command that vSphere admins can run after provisioning the Network, ResourcePool, Datastore, and VM Folder objects.

The command will associate the appropriate permissions to all the provisioned vSphere objects.

#### vSphere Permission Model

vSphere uses an object-level permissioning model allowing users or groups to be assigned roles on objects within vSphere. Each role contains a set of permissions. Our command will create a user-group-object-role-permission mapping by associating a user to a group and assigning roles to that group for particular objects.

#### Necessary Config Information

In order to generate appropriate user-group-object-role-permission mappings, we need to the following config information:

* admin username
* admin password
* username (optional - default eksa)
* password (required if user does not exist yet)
* group (optional - default EKSAUsers)
* “Global” role name (optional - default EKSAGlobalRole)
* “User” role name (optional - default EKSAUserRole)
* “Admin” role name (optional - default EKSACloudAdminRole)
* vSphere domain
* vSphere server
* vSphere datacenter
* vSphere Network object references
* vSphere Datastore object references
* vSphere ResourcePool object references
* vSphere VirtualMachine folder object references
* vSphere Template folder object references


**Command Business Logic**

When run, the command would then execute the following actions:

* Create the specified user with password
* Create the specified group
* Create the specified roles
* Associate appropriate permissions to each role
* Create appropriate group-object-role mappings for the vSphere objects specified in the config

#### Implementation

We can provide the necessary additional information to generate user-group-object-role-permission mappings via a vSphereUser kubernetes-style yaml config.

A minimal invocation of the command could look like this:

```
eksctl anywhere exp vsphere setup user --password NewUserPassword --filename user.yaml
```

Together with the following config file:

```
apiVersion: "eks-anywhere.amazon.com/v1"
kind: vSphereUser
spec:
  username: !!str "eksa"
  datacenter: "MyDatacenter"
  vSphereDomain: "vsphere.local"
  connection:
    server: "https://my-vsphere.internal.acme.com"
    insecure: false
  objects:
    networks:
      - !!str "/MyDatacenter/network/My Network"
    datastores:
      - !!str "/MyDatacenter/datastore/MyDatastore2"
    resourcePools:
      - !!str "/MyDatacenter/host/Cluster-03/MyResourcePool"
    folders:
      - !!str "/MyDatacenter/vm/OrgDirectory/MyVMs"
    templates:
      - !!str "/MyDatacenter/vm/Templates/MyTemplates"
```

Deriving admin credentials from environment variables:
```
EKSA_VSPHERE_USERNAME
EKSA_VSPHERE_PASSWORD
```

We feel this is the best choice for a few reasons:

1. it decouples the user setup command from the cluster config file. This allows the vSphere admin to execute it immediately after provisioning the objects, before the Kubernetes admin has started to build the cluster config. Per Lichun dependency on the cluster config file would make this command useless from a customer perspective.
2. Creating a kubernetes-style yaml config provides a standardized API for interacting with our CLI
3. Creating a kubernetes-style yaml config gives us a clear versioning convention so we can make non-breaking changes in future if necessary



##### Optional —force Flag

By default, the command will only create and operate on new group and role objects to protect the user from shooting themselves in the foot by accidentally destroying an existing group-object-role mapping.

However, one can imagine a scenario where a vSphere admin has a pre-existing group or roles that they would like to re-use in our vSphere setup command. To support this flow, we will provide an `--force` boolean flag that allows them to execute the setup user command with pre-existing objects.

We propose that the `--force ` flag should *not* require re-entering a new password if the user already exists.

For example:

```
eksctl anywhere exp vsphere setup user --force -f user.yaml
```

```
apiVersion: "eks-anywhere.amazon.com/v1"
kind: vSphereUser
spec:
  username: "eksa"
  group: "MyExistingGroup"
  globalRole: "MyExistingGlobalRole"
  userRole: "MyExistingUserRole"
  adminRole: "MyExistingEKSAAdminRole"
  datacenter: "MyDatacenter"
  vSphereDomain: "vsphere.local"
  connection:
    server: "https://my-vsphere.internal.acme.com"
    insecure: false
  objects:
    networks:
      - !!str "/MyDatacenter/network/My Network"
    datastores:
      - !!str "/MyDatacenter/datastore/MyDatastore2"
    resourcePools:
      - !!str "/MyDatacenter/host/Cluster-03/MyResourcePool"
    folders:
      - !!str "/MyDatacenter/vm/OrgDirectory/MyVMs"
    templates:
      - !!str "/MyDatacenter/vm/Templates/MyTemplates"
```

##### Optional —password Flag

The `--password` flag allows the user to provide a password when creating a new user. Its functionality is defined as follows:

* If user with username does not exist, require `--password` and create new user
* If user with username exists, throw error when `--password` flag is present



## UX

### Flows

A. Existing Flow

1. Kubernetes admin asks vSphere admin for VM folder, Network, ResourcePool, etc
2. vSphere admin provisions VM folder, Network, ResourcePool, etc
3. vSphere admin configures user with appropriate permissions by pointing and clicking at the vSphere UI to create and associate user, group, roles, and objects
4. Kubernetes admin receives vSphere object and user information from vSphere admin and builds cluster config file

B. Proposed Flow:

1. Kubernetes admin asks vSphere admin for VM folder, Network, ResourcePool, etc
2. vSphere admin provisions VM folder, Network, ResourcePool, etc
3. vSphere admin configures user with appropriate permissions by running `eksctl anywhere exp vsphere setup user --filename user.yaml`
4. Kubernetes admin receives vSphere object and user information from vSphere admin and builds cluster config file
