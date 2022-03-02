---
title: "Preparing a VMware vSphere environment for EKS Anywhere"
weight: 20
description: >
  The steps of setting up your VSphere cluster to make it ready to deploy EKS Anywhere on it.
---

## Create a VM and Template Folder (Optional):
You may need to restrict the user that will be used to create the cluster to only seeing and creating the EKS Anywhere resources to a particular folder and its nested folders instead of giving the previlge to this user to have full visibility and control over the whole vSphere cluster domain and its sub-child objects (data center, resource pools and other folders). In this case you may not be allowed to create a VM and template folder under your datacenter, and then you would ask the VSphere administrator to create this folder for you at whatevel level and set your user permission with the permissions needed ONLY on this folder level, instead of giving you the permission needed on the domain level and its sub-children objects.

When it comes to create your cluster you will reference a folder path under this folder created by the VSphere administrator for you. You would have all the prevleges needed on this folder created by the VSphere administrator including creating nested folders under this folder but not under the datacenter name or other folders, for Security team reasons. 

You would have a nested folder for the management cluster and another one for each workload cluster you are adding. Each folder will host the VMs of the Controlplane and Data plane nodes of each cluster. Each cluster VMs will be hosted into their own nested folder under this folder. 

In your EKS Anywhere configuration file you will reference a path under this folder created by the Vsphere administrator, and it will create it for you if it's not exist, as long as your VSphere user has the permission to do so.

Ask your vSphere administrator to create this VM and template folder for you. Each folder will host the VMs of the Controlplane and Data plane nodes of each cluster.

> Procedure to add a folder:
    1. In the vSphere Client, hit the VM and Template tab. Select either a data center or another folder as a parent object for the folder that you want to create.
    2. Right-click the parent object and click New Folder.
    3. Enter a name for the folder and click OK.

    For more details, check the following [link](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vcenterhost.doc/GUID-031BDB12-D3B2-4E2D-80E6-604F304B4D0C.html).

## vSphere Roles and User Permission Setup:
You would need to get a VSphere username with the right privleges to enabke you creating EKS Anywhere clusters on top of your VSphere cluster. Then you would need to import the latest release of EKS Anywhere OVA template to your VSphere cluster to use it to provision your Cluster nodes.

### Adding vCenter User:
Ask you VSphere administrator to add a vCenter user that will be used during the provisioning of the EKS Anywhere cluster in VMware vSphere.
1.	Log in with the vSphere Client to the vCenter Server.
2.	Specify the user name and password for a member of the vCenter Single Sign-On Administrators group.
3.	Navigate to the vCenter Single Sign-On user configuration UI.
    1.	From the Home menu, select Administration.
    2.	Under Single Sign On, click Users and Groups.
4.	If vsphere.local is not the currently selected domain, select it from the drop-down menu.
5.	You cannot add users to other domains.
6.	On the Users tab, click Add.
7.	Enter a user name and password for the new user.
8.	The maximum number of characters allowed for the user name is 300.
9.	You cannot change the user name after you create a user. The password must meet the password policy requirements for the system.
10.	Click Add.

For more details, check the following [link](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.authentication.doc/GUID-72BFF98C-C530-4C50-BF31-B5779D2A4BBB.html).

### Creating and Defining Roles to our User:
When you add a user, that user initially has no privileges to perform management operations. So you have to add this user to a group/groups with the required permissions, or to assign a role/roles with the required permission to this user. Three roles needed to be able creating the EKS Anywhere cluster:

1. The first role is a custom role. Name it for example EKS Anywhere Global. We define it to our user on the vCenter domain level and its children objects. We create this role with the following privileges:

> Content Library
* Add library item
* Check in a template
* Check out a template
* Create local library

> vSphere Tagging
* Assign or Unassign vSphere Tag
* Assign or Unassign vSphere Tag on Object
* Create vSphere Tag
* Create vSphere Tag Category
* Delete vSphere Tag
* Delete vSphere Tag Category
* Edit vSphere Tag
* Edit vSphere Tag Category
* Modify UsedBy Field For Category
* Modify UsedBy Field For Tag

2. The second role is also a custom role that we named it for example EKS Anywhere User, that we define it to our user on the following objects and their children objects. 
    1. The pool resource level and its children objects. This resource pool that our EKS Anywhere VMs will be part of.
    2. The storage object level and its children objects. This storage that will be used to store the cluster VMs.
    3. The network VLAN object level and its children objects. This network that will host the cluster VMs.
    4. The VM and Template folder level and its children objects.
We create this role with the following privileges:
> Content Library
* Add library item
* Check in a template
* Check out a template
* Create local library
> Datastore
* Allocate space
* Browse datastore
* Low level file operations
> Folder
* Create folder
> vSphere Tagging
* Assign or Unassign vSphere Tag
* Assign or Unassign vSphere Tag on Object
* Create vSphere Tag
* Create vSphere Tag Category
* Delete vSphere Tag
* Delete vSphere Tag Category
* Edit vSphere Tag
* Edit vSphere Tag Category
* Modify UsedBy Field For Category
* Modify UsedBy Field For Tag
> Network
* Assign network
> Resource
* Assign virtual machine to resource pool
> Scheduled task
* Create tasks
* Modify task
* Remove task
* Run task
> Profile-driven storage
* Profile-driven storage view
> Storage views
* View
> vApp
* Import
> Virtual machine
* Change Configuration
  - Add existing disk
  - Add new disk
  - Add or remove device
  - Advanced configuration
  - Change CPU count
  - Change Memory
  - Change Settings
  - Configure Raw device
  - Extend virtual disk
  - Modify device settings
  - Remove disk
* Edit Inventory
  - Create from existing
  - Create new
  - Remove
* Interaction
  - Power off
  - Power on
* Provisioning
  - Clone template
  - Clone virtual machine
  - Create template from virtual machine
  - Customize guest
  - Deploy template
  - Mark as template
  - Read customization specifications
* Snapshot management
  - Create snapshot
  - Remove snapshot
  - Revert to snapshot

3. The third role is the default system role Administrator that we define it to our user on the folder level that was created by the VSphere admistrator for us, and its children objects (VMs and OVA templates). 

To create a role and define privileges check [Create a vCenter Server Custom Role](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.security.doc/GUID-41E5E52E-A95B-4E81-9724-6AD6800BEF78.html) and [Defined Privileges](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.security.doc/GUID-ED56F3C4-77D0-49E3-88B6-B99B8B437B62.html#GUID-ED56F3C4-77D0-49E3-88B6-B99B8B437B62) pages.

## Deploy an OVA Template
This OVA template contains desired OS (Ubuntu or Bottlerocket), desired EKS-D Kubernetes release, and EKS-A version. In our scenario, the OS we went with is Ubuntu.
> Procedure to deploy the Ubuntu OVA:
1.	Go to the artifacts page and download the OVA template with the newest EKS-D Kubernetes release to your computer.
2.	Log in with the vSphere Client to the vCenter Server.
3.	Right-click the folder you created above, and select Deploy OVF Template. The Deploy OVF Template wizard opens.
4.	On the Select an OVF template page, hit the Local file option, specify the location of the OVA template you downloaded to your computer and click Next.
5.	On the Select a name and folder page, enter a unique name for the virtual machine or leave the default generated name if you do not have other templates with same name within your vCenter Server virtual machine folder. The default deployment location for the virtual machine is the inventory object where you started the wizard, which is the folder we created above. Click Next.
6.	On the Select a compute resource page, select the resource pool where to run the deployed VM template, and click Next. 
7.	On the Review details page, verify the OVF or OVA template details and click Next.
8.	On the Select storage page, select a datastore to store the deployed OVF or OVA template, and click Next.
9.	On the Select networks page, select a source network and map it to a destination network. Click Next.
10.	On the Ready to complete page, review the page and click Finish.

For more details, check the following [link](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-17BEDA21-43F6-41F4-8FB2-E01D275FE9B4.html)

To build your own Ubuntu OVA template check the Building your own Ubuntu OVA section in the following [link]({{< relref "../reference/artifacts/" >}}).


## Tagging the Deployed OVA Template
To be able using the deployed OVA template to create the VMs of the EKS Anywhere cluster, you have to tag it with specific values for the keys os and eksdRelease. The value of the OS key is the operating system of the deployed OVA template, which is ubuntu in our scenario. The value of the eksdRelease holds the the kubernetes, and the EKS-D release used in the deployed OVA template. Check the following [page](https://anywhere.eks.amazonaws.com/docs/reference/vsphere/customize-ovas/)({{< relref "../reference/vsphere/customize-ovas/" >}}). for more details.

> Procedure to tag the deployed OVA template:
1.	Go to the [artifacts]({{< relref "../reference/artifacts/" >}}) page and take notes of the tags and values associated with the OVA template you deployed in the previous step.
2.	In the vSphere Client, select Menu > Tags & Custom Attributes.
3.	Select the Tags tab and click Tags.
4.	Click New.
5.	In the Create Tag dialog box, copy the os tag name associated with your OVA that you took notes of, which in our case is os:ubuntu and paste it as the name for the first tag required.
6.	Specify the tag category os if exist. If not exists, create it. 
7.	Click Create.
8.	Repeat steps 2-4.
9.	In the Create Tag dialog box, copy the os tag name associated with your OVA that you took notes of, which in our case is eksdRelease:kubernetes-1-21-eks-8 and paste it as the name for the second tag required.
10.	Specify the tag category eksdRelease if exist. If not exists, create it. 
11.	Click Create.
12.	Navigate to the VM and Template tab. 
13.	Select the folder was created.
14.	Select deployed template and click Actions.
15.	From the drop-down menu, select Tags and Custom Attributes > Assign Tag.
16.	Select the tags we created from the list and confirm the operation.
