
---
title: "Reboot nodes"
linkTitle: "Reboot nodes"
weight: 85
aliases:
    /docs/tasks/cluster-rebootnode/
date: 2017-01-05
description: >
  How to properly reboot a node in an EKS Anywhere cluster
---

If you need to reboot a node in your cluster for maintenance or any other reason, performing the following steps will help prevent possible disruption of services on those nodes:

{{% alert title="Warning" color="primary" %}}
Rebooting a cluster node as described here is good for all nodes, but is critically important when rebooting a Bottlerocket node running the `boots` service on a Bare Metal cluster.
If it does go down while running the `boots` service, the Bottlerocket node will not be able to boot again until the `boots` service is restored on another machine. This is because Bottlerocket must get its address from a DHCP service.
{{% /alert %}}

1. Cordon the node so no further workloads are scheduled to run on it:

    ```bash
    kubectl cordon <node-name>
    ```

1. Drain the node of all current workloads:

   ```bash
   kubectl drain <node-name>
   ```

1. Shut down. Using the appropriate method for your provider, shut down the node.

1. Perform system maintenance or other task you need to do on the node and boot up the node.

1. Uncordon the node so that it can begin receiving workloads again.

   ```bash
   kubectl uncordon <node-name>
   ```

