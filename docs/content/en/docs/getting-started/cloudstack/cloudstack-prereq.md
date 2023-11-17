---
title: "Requirements for EKS Anywhere on CloudStack"
linkTitle: "1. Requirements"
weight: 10
aliases:
    /docs/reference/cloudstack/cloudstack-prereq/
description: >
  CloudStack provider requirements for EKS Anywhere
---

To run EKS Anywhere, you will need:

## Prepare Administrative machine
Set up an Administrative machine as described in [Install EKS Anywhere ]({{< relref "../../getting-started/install/" >}}).

## Prepare a CloudStack environment

To prepare a CloudStack environment to run EKS Anywhere, you need the following:

* A CloudStack 4.14 or later environment. CloudStack 4.16 is used for examples in these docs.
* Capacity to deploy 6-10 VMs.
* One shared network in CloudStack to use for the cluster. EKS Anywhere clusters need access to CloudStack through the network to enable self-managing and storage capabilities.
* A Red Hat Enterprise Linux qcow2 image built using the `image-builder` tool as described in [artifacts]({{< relref "../../osmgmt/artifacts/" >}}).
* User credentials (CloudStack API key and Secret key) to create VMs and attach networks in CloudStack.
* [Prepare DHCP IP addresses pool]({{< relref "../../clustermgmt/cluster-upgrades/vsphere-and-cloudstack-upgrades.md/#prepare-dhcp-ip-addresses-pool" >}})
* One IP address routable from the cluster but excluded from DHCP offering. This IP address is to be used as the Control Plane Endpoint IP. Below are some suggestions to ensure that this IP address is never handed out by your DHCP server. You may need to contact your network engineer.

    * Pick an IP address reachable from the cluster subnet which is excluded from DHCP range OR
    * Alter DHCP ranges to leave out an IP address(s) at the top and/or the bottom of the range OR
    * Create an IP reservation for this IP on your DHCP server. This is usually accomplished by adding a dummy mapping of this IP address to a non-existent mac address.

Each VM will require:

* 2 vCPUs
* 8GB RAM
* 25GB Disk

The administrative machine and the target workload environment will need network access (TCP/443) to:

* CloudStack endpoint (must be accessible to EKS Anywhere clusters)
* public.ecr.aws
* anywhere-assets.eks.amazonaws.com (http://anywhere-assets.eks.amazonaws.com/) (to download the EKS Anywhere binaries and manifests)
* distro.eks.amazonaws.com (http://distro.eks.amazonaws.com/) (to download EKS Distro binaries and manifests)
* d2glxqk2uabbnd.cloudfront.net (http://d2glxqk2uabbnd.cloudfront.net/) (for EKS Anywhere and EKS Distro ECR container images)
* api.ecr.us-west-2.amazonaws.com (http://api.ecr.us-west-2.amazonaws.com/) (for EKS Anywhere package authentication matching your region)
* d5l0dvt14r5h8.cloudfront.net (http://d5l0dvt14r5h8.cloudfront.net/) (for EKS Anywhere package ECR container images)
* api.github.com (http://api.github.com/) (only if GitOps is enabled)

## CloudStack information needed before creating the cluster

You need at least the following information before creating the cluster.
See [CloudStack configuration]({{< relref "./cloud-spec/" >}}) for a complete list of options and [Preparing CloudStack]({{< relref "./cloudstack-preparation/" >}}) for instructions on creating the assets.

* *Static IP Addresses*: You will need one IP address for the management cluster control plane endpoint, and a separate one for the controlplane of each workload cluster you add.

Let’s say you are going to have the management cluster and two workload clusters. For those, you would need three IP addresses, one for each. All of those addresses will be configured the same way in the configuration file you will generate for each cluster.

A static IP address will be used for each control plane VM in your EKS Anywhere cluster. Choose IP addresses in your network range that do not conflict with other VMs and make sure they are excluded from your DHCP offering.
An IP address will be the value of the property controlPlaneConfiguration.endpoint.host in the config file of the management cluster. A separate IP address must be assigned for each workload cluster.

* CloudStack datacenter: You need the name of the CloudStack Datacenter plus the following for each Availability Zone (availabilityZones). Most items can be represented by name or ID:
    * Account (account): Account with permission to create a cluster (optional, admin by default).
    * Credentials (credentialsRef): Credentials provided in an ini file used to access the CloudStack API endpoint. See [CloudStack Getting started]({{< relref "../../getting-started/cloudstack/" >}}) for details.
    * Domain (domain):  The CloudStack domain in which to deploy the cluster (optional, ROOT by default)
    * Management endpoint (managementApiEndpoint): Endpoint for a cloudstack client to make API calls to client.
    * Zone network (zone.network): Either name or ID of the network.
* CloudStack machine configuration: For each set of machines (for example, you could configure separate set of machines for control plane, worker, and etcd nodes), obtain the following information. This must be predefined in the cloudStack instance and identified by name or ID:
    * Compute offering (computeOffering): Choose an existing compute offering (such as `large-instance`), reflecting the amount of resources to apply to each VM.
    * Operating system (template): Identifies the operating system image to use (such as rhel8-k8s-118).
    * Users (users.name): Identifies users and SSH keys needed to access the VMs.
