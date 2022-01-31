---
title: "Verification and cleanup"
linkTitle: "Verification and cleanup"
weight: 5
date: 2021-11-11
description: >  
---

## Verification

After a couple of minutes, you should see the following in the Amazon EKS console:

![register](../images/register.png)


Here we can see an  overview of the workloads in our cluster, showing metadata information such as `Name`, `Namespace`, and `Status`. 

We also the ability to filter by propeties such as `Namespace` and other searchable values such as `capi` to bring up all of our Cluster API workloads.

Additionally if we wanted more information of a particular workload we can click on the name of the deployment to get a more detailed view.

**Workload**

![workloads](../images/workloads.png)

## Deregistration

You can deregister the EKS cluster by clicking on the **Deregister** button in AWS console.

![deregister](../images/deregister.png)