---
title: "Import OVAs"
linkTitle: "Import OVAs"
weight: 10
aliases:
    /docs/reference/vsphere/vsphere-ovas/
description: >
  Importing EKS Anywhere OVAs to vSphere
---

If you want to specify an OVA template, you will need to import OVA files into vSphere before you can use it in your EKS Anywhere cluster.
This guide was written using VMware Cloud on AWS,
but the [VMware OVA import guide can be found here.](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-17BEDA21-43F6-41F4-8FB2-E01D275FE9B4.html)

{{% alert title="Note" color="primary" %}}
If you don't specify a template in the cluster spec file, EKS Anywhere will use the proper default one for the Kubernetes minor version and OS family you specified in the spec file.
If the template doesn't exist, it will import the appropriate OVA into vSphere and add the necessary tags.

The default OVA for a Kubernetes minor version + OS family will change over time, for example, when a new EKS Distro version is released. In that case, new clusters will use the new OVA (EKS Anywhere will import it automatically).
{{% /alert %}}

{{% alert title="Warning" color="warning" %}}
Do not power on the imported OVA directly as it can cause some undesired configurations on the OS template and affect cluster creation. If you want to explore or modify the OS, please follow the instructions to [customize the OVA.]({{< relref "../customize/customize-ovas/" >}})
{{% /alert %}}

EKS Anywhere supports the following operating system families

* Bottlerocket (default)
* Ubuntu
* RHEL

A list of OVAs for this release can be found on the [artifacts page.]({{< relref "../../../osmgmt/artifacts" >}})

## Using vCenter Web User Interface

1. Right click on your Datacenter, select *Deploy OVF Template*
   ![Import ova drop down](/images/ss1.jpg)
1. Select an OVF template using URL or selecting a local OVF file and click on *Next*. If you are not able to select an
   OVF template using URL, download the file and use Local file option.
   
   Note: If you are using Bottlerocket OVAs, please select local file option.
   ![Import ova wizard](/images/ss2.jpg)
1. Select a folder where you want to deploy your OVF package (most of our OVF templates are under SDDC-Datacenter
   directory) and click on *Next*. You cannot have an OVF template with the same name in one directory. For workload
   VM templates, leave the Kubernetes version in the template name for reference. A workload VM template will
   support at least one prior Kubernetes major versions.
   ![Import ova wizard](/images/ss3.jpg)
1. Select any compute resource to run (from cluster-1, 10.2.34.5, etc..) the deployed VM and click on *Next*
   ![Import ova wizard](/images/ss4.jpg)
1. Review the details and click *Next*.
1. Accept the agreement and click *Next*.
1. Select the appropriate storage (e.g. “WorkloadDatastore“) and click *Next*.
1. Select destination network (e.g. “sddc-cgw-network-1”) and click *Next*.
1. Finish. 
1. Snapshot the VM. Right click on the imported VM and select Snapshots -> Take Snapshot... 
   (It is highly recommended that you snapshot the VM. This will reduce the time it takes to provision
   machines and cluster creation will be faster. If you prefer not to take snapshot, skip to step 13)
![Import ova wizard](/images/ss6.jpg)
1. Name your template (e.g. "root") and click *Create*.
![Import ova wizard](/images/ss7.jpg)
1. Snapshots for the imported VM should now show up under the *Snapshots* tab for the VM.
![Import ova wizard](/images/ss8.jpg)
1. Right click on the imported VM and select Template and Convert to Template
![Import ova wizard](/images/ss9.jpg)

## Steps to deploy a template using GOVC (CLI)

To deploy a template using `govc`, you must first ensure that you have
[GOVC installed](https://github.com/vmware/govmomi/blob/master/govc/README.md). You need to set and export three
environment variables to run `govc` GOVC_USERNAME, GOVC_PASSWORD and GOVC_URL.

1. Import the template to a content library in vCenter using URL or selecting a local OVA file

    Using URL:

    ```
    govc library.import -k -pull <library name> <URL for the OVA file>
    ```
    
    Using a file from the local machine:

    ```
    govc library.import <library name> <path to OVA file on local machine>
    ```

2. Deploy the template

    ```
    govc library.deploy -pool <resource pool> -folder <folder location to deploy template> /<library name>/<template name> <name of new VM>
    ```
   2a. If using Bottlerocket template for newer Kubernetes version than 1.21, resize disk 1 to 22G
   ```
   govc vm.disk.change -vm <template name> -disk.label "Hard disk 1" -size 22G
   ```
   2b. If using Bottlerocket template for Kubernetes version 1.21, resize disk 2 to 20G
      ```
      govc vm.disk.change -vm <template name> -disk.label "Hard disk 2" -size 20G
      ```


3. Take a snapshot of the VM (It is highly recommended that you snapshot the VM. This will reduce the time it takes to provision machines
   and cluster creation will be faster. If you prefer not to take snapshot, skip this step)

    ```
    govc snapshot.create -vm ubuntu-2004-kube-v1.25.6 root
    ```

4. Mark the new VM as a template

    ```
    govc vm.markastemplate <name of new VM>
    ```


## Important Additional Steps to Tag the OVA

### Using vCenter UI

#### Tag to indicate OS family

1. Select the template that was newly created in the steps above and navigate to *Summary* -> *Tags*.
   ![Import ova wizard](/images/ss10.jpg)
1. Click *Assign* -> *Add Tag* to create a new tag and attach it
   ![Import ova wizard](/images/ss11.jpg)
1. Name the tag *os:ubuntu* or *os:bottlerocket*
   ![Import ova wizard](/images/ss12.jpg)

#### Tag to indicate eksd release
1. Select the template that was newly created in the steps above and navigate to *Summary* -> *Tags*.
   ![Import ova wizard](/images/ss10.jpg)
1. Click *Assign* -> *Add Tag* to create a new tag and attach it
   ![Import ova wizard](/images/ss11.jpg)
1. Name the tag *eksdRelease:{eksd release for the selected ova}*, for example *eksdRelease:kubernetes-1-25-eks-5* for the 1.25 ova. You can find the rest of eksd releases in the previous [section]({{< relref "../vsphere-preparation#deploy-an-ova-template" >}}). If it's the first time you add an `eksdRelease` tag, you would need to create the category first. Click on "Create New Category" and name it `eksdRelease`.
   ![Import ova wizard](/images/ss13.png)

### Using govc

#### Tag to indicate OS family

1. Create tag category

```
govc tags.category.create -t VirtualMachine os
```
1. Create tags os:ubuntu and os:bottlerocket

```
govc tags.create -c os os:bottlerocket
govc tags.create -c os os:ubuntu
```
1. Attach newly created tag to the template

```
govc tags.attach os:bottlerocket <Template Path>
govc tags.attach os:ubuntu <Template Path>
```
1. Verify tag is attached to the template

```
govc tags.ls <Template Path> 
```

#### Tag to indicate eksd release
1. Create tag category
```
govc tags.category.create -t VirtualMachine eksdRelease
```
2. Create the proper eksd release Tag, depending on your template. You can find the eksd releases in the previous [section]({{< relref "../vsphere-preparation#deploy-an-ova-template" >}}). For example *eksdRelease:kubernetes-1-25-eks-5* for the 1.25 template.
```
govc tags.create -c eksdRelease eksdRelease:kubernetes-1-25-eks-5
```
3. Attach newly created tag to the template
```
govc tags.attach eksdRelease:kubernetes-1-25-eks-5 <Template Path>
```
4. Verify tag is attached to the template 

```
govc tags.ls <Template Path> 
```
{{% alert title="Note" color="primary" %}}
If the tags above are not applied as shown exactly, eks-a template validations will fail and CLI will abort
{{% /alert %}}

After you are done you can use the template for your workload cluster.
