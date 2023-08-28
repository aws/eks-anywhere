---
title: "Tinkerbell Concepts"
linkTitle: "Tinkerbell Concepts"
weight: 10
description: >
  Overview of Tinkerbell and network booting for EKS Anywhere on Bare Metal
---

EKS Anywhere uses [Tinkerbell](https://docs.tinkerbell.org/) to provision machines for a Bare Metal cluster.
Understanding what Tinkerbell is and how it works with EKS Anywhere can help you take advantage of advanced provisioning features or overcome provisioning problems you encounter.

As someone deploying an EKS Anywhere cluster on Bare Metal, you have several opportunities to interact with Tinkerbell:

* **Create a hardware CSV file**: You are required to create a [hardware CSV file]({{< relref "bare-preparation/#prepare-hardware-inventory" >}}) that contains an entry for every physical machine you want to add at cluster creation time. 
* **Create an EKS Anywhere cluster**: By modifying the [Bare Metal configuration file]({{< relref "./bare-spec" >}}) used to create a cluster, you can change some Tinkerbell settings or add actions to define how the operating system on each machine is configured.
* **Monitor provisioning**: You can follow along with the Tinkerbell Overview in this page to monitor the progress of your hardware provisioning, as Tinkerbell finds machines and attempts to network boot, configure, and restart them.

When you run the command to create an EKS Anywhere Bare Metal cluster, a set of Tinkerbell components start up on the Admin machine. One of these components runs in a container on Docker (Boots), while other components run as either controllers or services in pods on the Kubernetes [kind](https://kind.sigs.k8s.io/) cluster that is started up on the Admin machine. Tinkerbell components include Boots, Hegel, Rufio, and Tink.

### Tinkerbell Boots service

The Boots service runs in a single container to handle the DHCP service and network booting activities.
In particular, Boots hands out IP addresses, serves iPXE binaries via HTTP and TFTP, delivers an iPXE script to the provisioned machines, and runs a syslog server.

Boots is different from the other Tinkerbell services because the DHCP service it runs must listen directly to layer 2 traffic.
(The kind cluster running on the Admin machine doesn’t have the ability to have pods listening on layer 2 networks, which is why Boots is run directly on Docker instead, with host networking enabled.)

Because Boots is running as a container in Docker, you can see the output in the logs for the Boots container by running:

```bash
docker logs boots
```

From the logs output, you will see iPXE try to network boot each machine.
If the process doesn’t get all the information it wants from the DHCP server, it will time out.
You can see iPXE loading variables, loading a kernel and initramfs (via DHCP), then booting into that kernel and initramfs: in other words, you will see everything that happens with iPXE before it switches over to the kernel and initramfs.
The kernel, initramfs, and all images retrieved later are obtained remotely over HTTP and HTTPS.

### Tinkerbell Hegel, Rufio, and Tink components

After Boots comes up on Docker, a small Kubernetes kind cluster starts up on the Admin machine.
Other Tinkerbell components run as pods on that kind cluster. Those components include:

* **Hegel**: Manages Tinkerbell’s metadata service.
The Hegel service gets its metadata from the hardware specification stored in Kubernetes in the form of custom resources.
The format that it serves is similar to an Ec2 metadata format.
* **Rufio**: Handles talking to BMCs (which manages things like starting and stopping systems with IPMI or Redfish).
The Rufio Kubernetes controller sets things such as power state, persistent boot order.
BMC authentication is managed with Kubernetes secrets.
* **Tink**: The Tink service consists of three components: Tink server, Tink controller, and Tink worker.
The Tink controller manages hardware data, templates you want to execute, and the workflows that each target specific hardware you are provisioning.
The Tink worker is a small binary that runs inside of HookOS and talks to the Tink server.
The worker sends the Tink server its MAC address and asks the server for workflows to run.
The Tink worker will then go through each action, one-by-one, and try to execute it.

To see those services and controllers running on the kind bootstrap cluster, type:

```bash
kubectl get pods -n eksa-system
```
```
NAME                                      READY STATUS    RESTARTS AGE
hegel-sbchp                               1/1   Running   0        3d
rufio-controller-manager-5dcc568c79-9kllz 1/1   Running   0        3d
tink-controller-manager-54dc786db6-tm2c5  1/1   Running   0        3d
tink-server-5c494445bc-986sl              1/1   Running   0        3d
```

### Provisioning hardware with Tinkerbell

After you start up the cluster create process, the following is the general workflow that Tinkerbell performs to begin provisioning the bare metal machines and prepare them to become part of the EKS Anywhere target cluster.
You can set up kubectl on the Admin machine to access the bootstrap cluster and follow along:

```bash
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
```

### Power up the nodes

Tinkerbell starts by finding a node from the hardware list (based on MAC address) and contacting it to identify a baseboard management job (`job.bmc`) that runs a set of baseboard management tasks (`task.bmc`).
To see that information, type:

```bash
kubectl get job.bmc -A
```
```
NAMESPACE    NAME                                           AGE
eksa-system  mycluster-md-0-1656099863422-vxvh2-provision   12m
```

```bash
kubectl get tasks.bmc -A
```
```
NAMESPACE    NAME                                                AGE
eksa-system  mycluster-md-0-1656099863422-vxh2-provision-task-0  55s
eksa-system  mycluster-md-0-1656099863422-vxh2-provision-task-1  51s
eksa-system  mycluster-md-0-1656099863422-vxh2-provision-task-2  47s
```

The following shows snippets from the `tasks.bmc` output that represent the three tasks: Power Off, enable network boot, and Power On.

```bash
kubectl describe tasks.bmc -n eksa-system eksa-system mycluster-md-0-1656099863422-vxh2-provision-task-0
```
```
...
  Task:
    Power Action:  Off
Status:
  Completion Time:   2022-06-27T20:32:59Z
  Conditions:
    Status:    True
    Type:      Completed 
```
        
```bash
kubectl describe tasks.bmc -n eksa-system eksa-system mycluster-md-0-1656099863422-vxh2-provision-task-1
```
```
...
  Task:
    One Time Boot Device Action:
      Device:
        pxe
      Efi Boot:  true
Status:
  Completion Time:   2022-06-27T20:33:04Z
  Conditions:
    Status:    True
    Type:      Completed   
```

```bash
kubectl describe tasks.bmc -n eksa-system eksa-system mycluster-md-0-1656099863422-vxh2-provision-task-2
```
```
  Task:
    Power Action:  on
Status:
  Completion Time:   2022-06-27T20:33:10Z
  Conditions:
    Status:    True
    Type:      Completed   
```

Rufio converts the baseboard management jobs into task objects, then goes ahead and executes each task. To see Rufio logs, type:

```bash
kubectl logs -n eksa-system rufio-controller-manager-5dcc568c79-9kllz | less
```

### Network booting the nodes

Next the Boots service netboots the machine and begins streaming the HookOS (`vmlinuz` and `initramfs`) to the machine.
HookOS runs in memory and provides the installation environment.
To watch the Boots log messages as each node powers up, type:

```bash
docker logs boots 
```

You can search the output for `vmlinuz` and `initramfs` to watch as the HookOS is downloaded and booted from memory on each machine.

### Running workflows

Once the HookOS is up, Tinkerbell begins running the tasks and actions contained in the workflows.
This is coordinated between the Tink worker, running in memory within the HookOS on the machine, and the Tink server on the kind cluster.
To see the workflows being run, type the following:

```bash
kubectl get workflows.tinkerbell.org -n eksa-system
```
```
NAME                                TEMPLATE                            STATE
mycluster-md-0-1656099863422-vxh2   mycluster-md-0-1656099863422-vxh2   STATE_RUNNING
```

This shows the workflow for the first machine that is being provisioned.
Add `-o yaml` to see details of that workflow template:

```bash
kubectl get workflows.tinkerbell.org -n eksa-system -o yaml
```
```
...
status:
  state: STATE_RUNNING
  tasks:
  - actions
    - environment:
        COMPRESSED: "true"
        DEST_DISK: /dev/sda
        IMG_URL: https://anywhere-assets.eks.amazonaws.com/releases/bundles/11/artifacts/raw/1-22/bottlerocket-v1.22.10-eks-d-1-22-8-eks-a-11-amd64.img.gz
      image: public.ecr.aws/eks-anywhere/tinkerbell/hub/image2disk:6c0f0d437bde2c836d90b000312c8b25fa1b65e1-eks-a-15
      name: stream-image
      seconds: 35
      startedAt: "2022-06-27T20:37:39Z"
      status: STATE_SUCCESS
...
```

You can see that the first action in the workflow is to stream (`stream-image`) the operating system to the destination disk (`DEST_DISK`) on the machine.
In this example, the Bottlerocket operating system that will be copied to disk (`/dev/sda`) is being served from the location specified by IMG_URL.
The action was successful (STATE_SUCCESS) and it took 35 seconds.

Each action and its status is shown in this output for the whole workflow.
To see details of the default actions for each supported operating system, see the Ubuntu TinkerbellTemplateConfig example and Bottlerocket TinkerbellTemplateConfig example.

In general, the actions include:

* Streaming the operating system image to disk on each machine.
* Configuring the network interfaces on each machine.
* Setting up the cloud-init or similar service to add users and otherwise configure the system.
* Identifying the data source to add to the system.
* Setting the kernel to pivot to the installed system (using kexec) or having the system reboot to bring up the installed system from disk.

If all goes well, you will see all actions set to STATE_SUCCESS, except for the kexec-image action. That should show as STATE_RUNNING for as long as the machine is running.

You can review the CAPT logs to see provisioning activity.
For example, at the start of a new provisioning event, you would see something like the following:

```bash
kubectl logs -n capt-system capt-controller-manager-9f8b95b-frbq | less
```
```
..."Created BMCJob to get hardware ready for provisioning"...
```

You can follow this output to see the machine as it goes through the provisioning process.

After the node is initialized, completes all the Tinkerbell actions, and is booted into the installed operating system (Ubuntu or Bottlerocket), the new system starts cloud-init to do further configuration.
At this point, the system will reach out to the Tinkerbell Hegel service to get its metadata.

If something goes wrong, viewing Hegel files can help you understand why a stuck system that has booted into Ubuntu or Bottlerocket has not joined the cluster yet.
To see the Hegel logs, get the internal IP address for one of the new nodes. Then check for the names of Hegel logs and display the contents of one of those logs, searching for the IP address of the node:

```bash
kubectl get nodes -o wide
```
```
NAME        STATUS   ROLES                 AGE    VERSION               INTERNAL-IP    ...
eksa-da04   Ready    control-plane,master  9m5s   v1.22.10-eks-7dc61e8  10.80.30.23
```
```bash
kubectl get logs -n eksa-system | grep hegel
```
```
hegel-n7ngs
```
```bash
kubectl logs -n eksa-system hegel-n7ngs
```
```
..."Retrieved IP peer IP..."userIP":"10.80.30.23...
```

If the log shows you are getting requests from the node, the problem is not a cloud-init issue.

After the first machine successfully completes the workflow, each other machine repeats the same process until the initial set of machines is all up and running.

### Tinkerbell moves to target cluster

Once the initial set of machines is up and the EKS Anywhere cluster is running, all the Tinkerbell services and components (including Boots) are moved to the new target cluster and run as pods on that cluster.
Those services are deleted on the kind cluster on the Admin machine.

### Reviewing the status

At this point, you can change your kubectl credentials to point at the new target cluster to get information about Tinkerbell services on the new cluster. For example:

```bash
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
```

First check that the Tinkerbell pods are all running by listing pods from the eksa-system namespace:

```bash
kubectl get pods -n eksa-system
```
```
NAME                                        READY   STATUS    RESTARTS   AGE
boots-5dc66b5d4-klhmj                       1/1     Running   0          3d
hegel-sbchp                                 1/1     Running   0          3d
rufio-controller-manager-5dcc568c79-9kllz   1/1     Running   0          3d
tink-controller-manager-54dc786db6-tm2c5    1/1     Running   0          3d
tink-server-5c494445bc-986sl                1/1     Running   0          3d
```

Next, check the list of Tinkerbell machines.
If all of the machines were provisioned successfully, you should see `true` under the READY column for each one.

```bash
kubectl get tinkerbellmachine -A
```
```
NAMESPACE    NAME                                                   CLUSTER    STATE  READY  INSTANCEID                          MACHINE
eksa-system  mycluster-control-plane-template-1656099863422-pqq2q   mycluster         true   tinkerbell://eksa-system/eksa-da04  mycluster-72p72
```

You can also check the machines themselves.
Watch the PHASE change from Provisioning to Provisioned to Running.
The Running phase indicates that the machine is now running as a node on the new cluster:

```bash
kubectl get machines -n eksa-system
```
```
NAME              CLUSTER    NODENAME    PROVIDERID                         PHASE    AGE  VERSION
mycluster-72p72   mycluster  eksa-da04   tinkerbell://eksa-system/eksa-da04 Running  7m25s   v1.22.10-eks-1-22-8
```
Once you have confirmed that all your machines are successfully running as nodes on the target cluster, there is not much for Tinkerbell to do.
It stays around to continue running the DHCP service and to be available to add more machines to the cluster.
