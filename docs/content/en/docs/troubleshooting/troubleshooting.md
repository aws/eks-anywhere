---
title: "Cluster troubleshooting"
linkTitle: "Cluster troubleshooting"
weight: 20
description: >
  Troubleshooting EKS Anywhere clusters
aliases:
   - /docs/tasks/troubleshoot/_troubleshooting
   - /docs/tasks/troubleshoot/troubleshooting/
---

This guide covers EKS Anywhere troubleshooting. It is divided into the following sections:

* [General troubleshooting]({{< relref "#general-troubleshooting" >}})
* [Bare Metal Troubleshooting]({{< relref "#bare-metal-troubleshooting" >}})
* [vSphere Troubleshooting]({{< relref "#vsphere-troubleshooting" >}})
* [Snow Troubleshooting]({{< relref "#snow-troubleshooting" >}})
* [Nutanix Troubleshooting]({{< relref "#nutanix-troubleshooting" >}})

You may want to search this document for a fragment of the error you are seeing.

## Troubleshooting knowledge articles

More information on troubleshooting EKS Anywhere is available from the [AWS Knowledge Center](https://repost.aws/knowledge-center) including:

* [How do I check EKS Anywhere cluster component logs on primary and worker nodes for BottleRocket, Ubuntu, or Redhat?](https://repost.aws/knowledge-center/eks-anywhere-check-component-logs)
* [How can I troubleshoot etcdadm controller issues in EKS Anywhere?](https://repost.aws/knowledge-center/eks-anywhere-etcdadm-controller-issues)
* [How do I return an EKS Anywhere cluster to a working state when the cluster upgrade fails?](https://repost.aws/knowledge-center/eks-anywhere-return-cluster-upgrade-fail)
* [How do I clean up EKS Anywhere resources without eksctl when cluster creation fails?
](https://repost.aws/knowledge-center/eks-clean-up-resources-failed-cluster)
* [How does EKS Anywhere cluster bootstrapping work?](https://repost.aws/knowledge-center/eks-anywhere-bootstrapping-process)


## General troubleshooting

### Increase eksctl anywhere output

If you’re having trouble running `eksctl anywhere` you may get more verbose output with the `-v 6` option. The highest level of verbosity is `-v 9` and the default level of logging is level equivalent to `-v 0`.

### Cannot run Docker commands

The EKS Anywhere binary requires access to run Docker commands without using `sudo`.
If you're using a Linux distribution you will need to be using Docker 20.x.x and your user needs to be part of the Docker group.

To add your user to the Docker group you can use.

```bash
sudo usermod -a -G docker $USER
```

Now you need to log out and back in to get the new group permissions.

### Minimum requirements for Docker version have not been met

```
Error: failed to validate Docker: minimum requirements for Docker version have not been met. Install Docker version 20.x.x or above
```
Ensure you are running Docker 20.x.x for example:
```
% docker --version
Docker version 20.10.6, build 370c289
```
### Minimum requirements for Docker version have not been met on macOS
```
Error: EKS Anywhere does not support Docker Desktop versions between 4.3.0 and 4.4.1 on macOS
```
```
Error: EKS Anywhere requires Docker Desktop to be configured to use CGroups v1. Please  set `deprecatedCgroupv1:true` in your `~/Library/Group\\ Containers/group.com.docker/settings.json` file
```
Ensure you are running Docker Desktop 4.4.2 or newer and, if you are running EKS Anywhere v0.15 or earlier, have set `"deprecatedCgroupv1": true` in your settings.json file
```
% defaults read /Applications/Docker.app/Contents/Info.plist CFBundleShortVersionString
4.42
% docker info --format '{{json .CgroupVersion}}'
"1"
```

### For EKS Anywhere v0.15 and earlier, cgroups v2 is not supported in Ubuntu 21.10+ and 22.04
```
ERROR: failed to create cluster: could not find a log line that matches "Reached target .*Multi-User System.*|detected cgroup v1"
```
For EKS Anywhere v0.15 and earlier, if you are using Ubuntu it is recommended to use Ubuntu 20.04 for the Administrative Machine. This is because the EKS Anywhere Bootstrap cluster for those versions requires _cgroups v1_. Since Ubuntu 21.10 _cgroups v2_ is enabled by default. You can use Ubuntu 21.10 and 22.04 for the Administrative machine if you configure Ubuntu to use _cgroups v1_ instead. This is not an issue if you are using macOS for your Administrative machine.

To verify cgroups version
```
% docker info | grep Cgroup
 Cgroup Driver: cgroupfs
 Cgroup Version: 2
```
To use _cgroups v1_ you need to _sudo_ and edit _/etc/default/grub_ to set _GRUB_CMDLINE_LINUX_ to "systemd.unified_cgroup_hierarchy=0" and reboot.
```
%sudo <editor> /etc/default/grub
GRUB_CMDLINE_LINUX="systemd.unified_cgroup_hierarchy=0"
sudo update-grub
sudo reboot now
```
Then verify you are using _cgroups v1_.
```
% docker info | grep Cgroup
 Cgroup Driver: cgroupfs
 Cgroup Version: 1
```

### Pod errors due to `too many open files`

The bootstrap or EKS Anywhere Docker cluster pods show error: `too many open files`.

This may be caused by running out of [inotify](https://linux.die.net/man/7/inotify) resources. Resource limits are defined by `fs.inotify.max_user_watches` and `fs.inotify.max_user_instances` system variables. For example, in Ubuntu these default to 8192 and 128 respectively, which is not enough to create a cluster with many nodes.

To increase these limits temporarily run the following commands on the host:

```bash
sudo sysctl fs.inotify.max_user_watches=524288
sudo sysctl fs.inotify.max_user_instances=512
```

To make the changes persistent, edit the file `/etc/sysctl.conf` and add these lines:

```bash
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
```

Reference: https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files

### ECR access denied

```
Error: failed to create cluster: unable to initialize executables: failed to setup eks-a dependencies: Error response from daemon: pull access denied for public.ecr.aws/***/cli-tools, repository does not exist or may require 'docker login': denied: Your authorization token has expired. Reauthenticate and try again.
```

All images needed for EKS Anywhere are public and do not need authentication. Old cached credentials could trigger this error.
Remove cached credentials by running:
```sh
docker logout public.ecr.aws
```

### Error unmarshaling JSON: while decoding JSON: json: unknown field "spec"

```
Error: loading config file "cluster.yaml": error unmarshaling JSON: while decoding JSON: json: unknown field "spec"
```
Use `eksctl anywhere create cluster -f cluster.yaml` instead of `eksctl create cluster -f cluster.yaml` to create an EKS Anywhere cluster.

### Error: old cluster config file exists under my-cluster, please use a different clusterName to proceed

```
Error: old cluster config file exists under my-cluster, please use a different clusterName to proceed
```
The `my-cluster` directory already exists in the current directory.
Either use a different cluster name or move the directory.

### At least one WorkerNodeGroupConfiguration must not have NoExecute and/or NoSchedule taints

```
Error: the cluster config file provided is invalid: at least one WorkerNodeGroupConfiguration must not have NoExecute and/or NoSchedule taints
```
An EKS Anywhere management cluster requires at least one schedulable worker node group to run cluster administration components. Both `NoExecute` and `NoSchedule` taints must be absent from the workerNodeGroup for it to be considered schedulable. This validation was removed for workload clusters from v0.19.0 onwards, so now it only applies to management clusters.

To remedy, remove `NoExecute` and `NoSchedule` taints from at least one WorkerNodeGroupConfiguration on your management cluster.

Invalid configuration example:
```
# Invalid workerNodeGroupConfiguration
workerNodeGroupConfigurations:    # List of node groups you can define for workers
  - count: 1
    name: md-0
    taints:                       # NoSchedule taint applied to md-0, not schedulable
    - key: "key1"
      value: "value1"
      effect: "NoSchedule"
  - count: 1
    name: md-1
    taints:                       # NoExecute taint applied to md-1, not schedulable
    - key: "key2"
      value: "value2"
      effect: "NoExecute"
```

Valid configuration example:
```
# Valid workerNodeGroupConfiguration
workerNodeGroupConfigurations:    # List of node groups you can define for workers
- count: 1
  name: md-0
  taints:                         # NoSchedule taint applied to md-0, not schedulable
  - key: "key1"
    value: "value1"
    effect: "NoSchedule"
- count: 1
  name: md-1                      # md-1 has no NoSchedule/NoExecute taints applied, is schedulable
```


### Memory or disk resource problem

There are various disk and memory issues on the Admin machine that can cause problems. Make sure:
* Docker is configured with enough memory.
* The system-wide Docker memory configuration provides enough RAM for the bootstrap cluster.
* You do not have unneeded KinD clusters by running `kind get clusters`.
* You should delete unneeded clusters with `kind delete cluster --name <cluster-name>`. If you do not have `kind` installed, you can install it from https://kind.sigs.k8s.io/ or use `docker ps` to see the KinD clusters and run `docker stop` to stop the cluster.
* You do not have any unneeded Docker containers running with `docker ps`.
* Terminate any unneeded Docker containers.
* Make sure Docker isn't out of disk resources.
* If you don't have any other Docker containers running you may want to run `docker system prune` to clean up disk space.

You may want to restart Docker.
To restart Docker on Ubuntu `sudo systemctl restart docker`.

### Waiting for cert-manager to be available... Error: timed out waiting for the condition

```
Failed to create cluster {"error": "error initializing capi resources in cluster: error executing init: Fetching providers\nInstalling cert-manager Version=\"v1.1.0\"\nWaiting for cert-manager to be available...\nError: timed out waiting for the condition\n"}
```
This is likely a [memory or disk resource problem]({{< relref "#memory-or-disk-resource-problem" >}}).
You can also try using techniques from [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}).

### NTP Time sync issue

```
level=error msg=k8sError error="github.com/cilium/cilium/pkg/k8s/watchers/endpoint_slice.go:91: Failed to watch *v1beta1.EndpointSlice: failed to list *v1beta1.EndpointSlice: Unauthorized" subsys=k8s
```
You might notice authorization errors if the timestamps on your EKS Anywhere control plane nodes and worker nodes are out-of-sync. Please ensure that all the nodes are configured with the same healthy NTP servers to avoid out-of-sync issues.

```
Error running bootstrapper cmd: error joining as worker: Error waiting for worker join files: Kubeadm join kubelet-start killed after timeout
```
You might also notice that the joining of nodes will fail if your Admin machine differs in time compared to your nodes. Make sure to check the server time matches between the two as well.

### The connection to the server localhost:8080 was refused

```
Performing provider setup and validations
Creating new bootstrap cluster
Installing cluster-api providers on bootstrap cluster
Error initializing capi in bootstrap cluster	{"error": "error waiting for capi-kubeadm-control-plane-controller-manager in namespace capi-kubeadm-control-plane-system: error executing wait: The connection to the server localhost:8080 was refused - did you specify the right host or port?\n"}
Failed to create cluster	{"error": "error waiting for capi-kubeadm-control-plane-controller-manager in namespace capi-kubeadm-control-plane-system: error executing wait: The connection to the server localhost:8080 was refused - did you specify the right host or port?\n"}
```
Initializing Cluster API components on the bootstrap cluster fails. This is likely a [memory or disk resource problem]({{< relref "#memory-or-disk-resource-problem" >}}).

### Generic cluster unavailable

Troubleshoot more by inspecting bootstrap cluster or workload cluster (depending on the stage of the failure) using kubectl commands.
```
kubectl get pods -A --kubeconfig=<kubeconfig>
kubectl get nodes -A --kubeconfig=<kubeconfig>
kubectl logs <podname> -n <namespace> --kubeconfig=<kubeconfig>
....
```

### Bootstrap cluster fails to come up

If your bootstrap cluster has problems you may get detailed logs by looking at the files created under the `${CLUSTER_NAME}/logs` folder. The capv-controller-manager log file will surface issues with vsphere specific configuration while the capi-controller-manager log file might surface other generic issues with the cluster configuration passed in.

You may also access the logs from your bootstrap cluster directly as below:
```bash
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
kubectl logs -f -n capv-system -l control-plane="controller-manager" -c manager
```

It is also useful to start a shell session on the Docker container running the bootstrap cluster by running `docker ps` and then `docker exec -it <container-id> bash` the kind container.

### Bootstrap cluster fails to come up: node(s) already exist for a cluster with the name

During `create` and `delete` CLI, EKS Anywhere tries to create a temporary KinD bootstrap cluster with the name `${CLUSTER_NAME}-eks-a-cluster` on the Admin machine. This operation can fail with below error:

```
Error: creating bootstrap cluster: executing create/delete cluster: ERROR: failed to create/delete cluster: node(s) already exist for a cluster with the name \"cluster-name\"
```

This indicates that the cluster creation or deletion fails because a bootstrap cluster of the same name already exists. If you are sure the cluster is not being used, you can manually delete the old cluster:

```bash
docker ps | grep "${CLUSTER_NAME}-eks-a-cluster-control-plane" | awk '{ print $1 }' | xargs docker rm -f
```

Once the old KinD bootstrap cluster is deleted, you can rerun the `eksctl anywhere create` or `eksctl anywhere delete` command again.

### Cluster upgrade fails with management components on bootstrap cluster

{{% alert title="Important" color="warning" %}}

KinD cluster is no longer used during upgrade of management cluster from v0.18.0 onwards.

{{% /alert %}}

For `eksctl anywhere` version older than `v0.18.0`, if a cluster upgrade of a management (or self managed) cluster fails or is halted in the middle, you may be left in a
state where the management resources (CAPI) are still on the KinD bootstrap cluster on the Admin machine. Right now, you will have to
manually move the management resources from the KinD cluster back to the management cluster.

First create a backup:
```shell
CLUSTER_NAME=squid
KINDKUBE=${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
MGMTKUBE=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
DIRECTORY=backup
# Substitute the version with whatever version you are using
CONTAINER=public.ecr.aws/eks-anywhere/cli-tools:v0.12.0-eks-a-19

rm -rf ${DIRECTORY}
mkdir ${DIRECTORY}

docker run -i --network host -w $(pwd) -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd):/$(pwd) --entrypoint clusterctl ${CONTAINER} move \
        --namespace eksa-system \
        --kubeconfig $KINDKUBE \
        --to-directory ${DIRECTORY}

#After the backup, move the management cluster back
docker run -i --network host -w $(pwd) -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd):/$(pwd) --entrypoint clusterctl ${CONTAINER} move \
        --to-kubeconfig $MGMTKUBE \
        --namespace eksa-system \
        --kubeconfig $KINDKUBE
```

Before you delete your bootstrap KinD cluster, verify there are no import custom resources left on it:
```shell
kubectl get crds | grep eks | while read crd rol
do
  echo $crd
  kubectl get $crd -A
done
```

### Upgrade command stuck on `Waiting for external etcd to be ready`

There can be a few reasons if the upgrade command is stuck on `waiting for external etcd for workload cluster to be ready` and eventually times out.

First, check the underlying infrastructure to see if any new etcd machines are created.

#### No new etcd machines are created in the infrastrure provider

If no etcd machines are created, check the `etcdadm-bootstrap-provider-controller-manager`, `etcdadm-controller-controller-manager`, `capi-controller-manager`, `capi-kubeadm-control-plane-controller-manager`, and other Cluster API controller logs using the commands mentioned in [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}) section.

#### No IP assigned to the new etcd machine

Refer to [no IP assigned to a VM]({{< relref "#no-ip-assigned-to-a-vm" >}}) section.

### Machines are unhealthy after restoring CAPI control from the backup

When an EKS Anywhere management cluster loses its CAPI resources, you need use the backup stored in the <cluster-name> folder to regain the management control. It is possible that after recovering the cluster from CAPI backup, the machine objects are in unhealthy/not ready state. For example:

```sh
$ kubectl get machines -n eksa-system

NAME                                    PHASE
cluster1-etcd-2md6m                     Failed
cluster1-etcd-s8qs5                     Failed
cluster1-etcd-vqs6w                     Failed
cluster1-md-0-75584b4fccxfgs86-pfk22    Failed
cluster1-x6rzc                          Failed
```

This can happen when the new machines got rolled out after the backup was taken, thus the backup objects being applied do not have the latest machine configurations. When applying an outdated CAPI backup, the provider machine object is out of sync with the actual infrastructure VM.


#### etcdadm join failed in etcd machine bootstrap log

When an etcd machine object is not in ready state, you can `ssh` into the etcd node and check the etcd bootstrap log:

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
cat /var/log/cloud-init.log
{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
sudo sheltie
journalctl _COMM=host-ctr
{{< /tab >}}
{{< /tabpane >}}

If the bootstrap log indicates that the etcadm join operation fail, this can mean that the etcd cluster contains unhealthy member(s). Looking at the log details you can further troubleshoot from below sections.

#### New etcd machine cannot find the existing etcd cluster members

The edcdadm log shows error that the new etcd machine cannot connect to the existing etcd cluster memebers. This means the `etcdadm-init` secret is outdated. To update it, run

```sh
kubectl edit <cluster-name>-etcd-init -n eksa-system
```

and make sure the new etcd machine IP is included in the secret.

#### New etcd machine cannot join the cluster due to loss of quorum

An etcd cluster needs a majority of nodes, a quorum, to agree on updates to the cluster state. For a cluster with n members, quorum is (n/2)+1. If etcd could not automatically recover and restore quorum, it is possible that there was an unhealthy or broken VM (e.g. a VM without IP assigned) that is still a member of the etcd cluster. You need to find and manually remove the unhealthy member.

To achieve that, first `ssh` into the etcd node and use `etcdctl member list` command to detect the unhealthy member.

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
sudo etcdctl --cacert=/etc/etcd/pki/ca.crt --cert=/etc/etcd/pki/etcdctl-etcd-client.crt --key=/etc/etcd/pki/etcdctl-etcd-client.key member list
{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
ETCD_CONTAINER_ID=$(ctr -n k8s.io c ls | grep -w "etcd-io" | cut -d " " -f1)
ctr -n k8s.io t exec -t --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
     --cacert=/var/lib/etcd/pki/ca.crt \
     --cert=/var/lib/etcd/pki/server.crt \
     --key=/var/lib/etcd/pki/server.key \
     member list
{{< /tab >}}
{{< /tabpane >}}

After identifying the unhealthy member, use `etcdctl member remove` to remove it from the cluster.

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
sudo etcdctl --cacert=/etc/etcd/pki/ca.crt --cert=/etc/etcd/pki/etcdctl-etcd-client.crt --key=/etc/etcd/pki/etcdctl-etcd-client.key member remove ${UNHEALTHY_MEMBER_ID}
{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
ctr -n k8s.io t exec -t --exec-id etcd ${ETCD_CONTAINER_ID} etcdctl \
     --cacert=/var/lib/etcd/pki/ca.crt \
     --cert=/var/lib/etcd/pki/server.crt \
     --key=/var/lib/etcd/pki/server.key \
     member remove ${UNHEALTHY_MEMBER_ID}
{{< /tab >}}
{{< /tabpane >}}

#### Provider machine does not point to the underlying VMs

Follow the VM restore process in provider-specific section.
* [Restore VM for machine in vSphere]({{< relref "#restore-vm-for-machine-in-vsphere" >}})

#### Nodes cycling due to insufficient pod CIDRs

Kubernetes controller manager allocates a dedicated CIDR block per node for pod IPs from within `clusterNetwork.pods.cidrBlocks`. The size of this node CIDR block defaults to /24 and can be adjusted at cluster creation using the [optional cidrMaskSize field]({{< relref "../getting-started/optional/cni/#node-ips-configuration-option" >}}).

Since each node requires a CIDR block, the maximum number of nodes in a cluster is limited to the number of non-overlapping subnets of `cidrMaskSize` that fit in the pods CIDR block. For example, for a pod CIDR block mask of `/18` and a node CIDR mask size of `/22`, a maximum of 16 nodes can be proivisioned since there are 16 subnets of size `/22` in the overall `/18` block. If more nodes are created than the `clusterNetwork.pods.cidrBlocks` can accomodate, `kube-controller-manager` will not be able to allocate a CIDR block to the extra nodes.

This can cause nodes to become `NotReady` with the following sympotoms:

- `kubectl describe node <unhealthy node>` indicates `CIDRNotAvailable` events.
- kube-controller-manager log displays:
    ```
    Error while processing Node Add/Delete: failed to allocate cidr from cluster cidr at idx:0: CIDR allocation failed; there are no remaining CIDRs left to allocate in the accepted range
    ```
- Kubelet log on unhealthy node contains `NetworkPluginNotReady message:Network plugin returns error: cni plugin not initialized`

If more nodes need to be provisioned, either the `clusterNetwork.pods.cidrBlocks` must be expanded or the `node-cidr-mask-size` [should be reduced.]({{< relref "../getting-started/optional/cni/#node-ips-configuration-option" >}}).

### Machine health check shows "Remediation is not allowed"

Sometimes a cluster node is crashed but machine health check does not start the proper remediation process to recreate the failed machine. For example, if a worker node is crashed, and running `kubectl get mhc ${CLUSTER_NAME}-md-0-worker-unhealthy -n eksa-system -oyaml` shows status message below:

```
Remediation is not allowed, the number of not started or unhealthy machines exceeds maxUnhealthy
```

EKS Anywhere sets the machine health check's `MaxUnhealthy` of the workers in a worker node group to 40%. This means any further remediation is only allowed if at most 40% of the worker machines selected by "selector" are not healthy. If more than 40% of the worker machines are in unhealthy state, the remediation will not be triggered.

For example, if you create an EKS Anywhere cluster with 2 worker nodes in the same worker node group, and one of the worker node is down. The machine health check will not remediate the failed machine because the actual unhealthy machines (50%) in the worker node group already exceeds the maximum percentage of the unhealthy machine (40%) allowed. As a result, the failed machine will not be replaced with new healthy machine and your cluster will be left with single worker node. In this case, we recommend you to scale up the number of worker nodes, for example, to 4. Once the 2 more worker nodes are up and running, it brings the total unhealthy worker machines to 25% which is below the 40% limit. This will trigger the machine health check remediation which replace the unhealthy machine with new one.

### Etcd machines with false `NodeHealthy` condition due to `WaitingForNodeRef`

When inspecting the `Machine` CRs, etcd machines might appear as `Running` but containing a false `NodeHealthy` condition, with a `WaitingForNodeRef` reason. This is a purely cosmetic issue that has no impact in the health of your cluster. This has been fixed in more recent versions of EKS-A, so this condition won't be displayed anymore in etcd machines.

```yaml
Status:
  Addresses:
    Address:        144.47.85.93
    Type:           ExternalIP
  Bootstrap Ready:  true
  Conditions:
    Last Transition Time:  2023-05-15T23:13:01Z
    Status:                True
    Type:                  Ready
    Last Transition Time:  2023-05-15T23:12:09Z
    Status:                True
    Type:                  BootstrapReady
    Last Transition Time:  2023-05-15T23:13:01Z
    Status:                True
    Type:                  InfrastructureReady
    Last Transition Time:  2023-05-15T23:12:09Z
    Reason:                WaitingForNodeRef
    Severity:              Info
    Status:                False
    Type:                  NodeHealthy
  Infrastructure Ready:    true
  Last Updated:            2023-05-15T23:13:01Z
  Observed Generation:     3
  Phase:                   Running
```

### Machine gets stuck at `Provisioned` state without `providerID` set

The VM can be created in the provider infrastructure with proper IP assigned but running `kubectl get machines -n eksa-system` indicates that the machine is in `Provisioned` state and never gets to `Running`.

Check the CAPI controller manager log with `k logs -f -n capi-system -l control-plane="controller-manager" --tail -1`:

```sh
E0218 03:41:53.126751 1 machine_controller_noderef.go:152] controllers/Machine "msg"="Failed to parse ProviderID" "error"="providerID is empty" "providerID"={} "node"="test-cluster-6ffd74bd5b-khxzr"
E0218 03:42:09.155577 1 machine_controller.go:685] controllers/Machine "msg"="Unable to retrieve machine from node" "error"="no matching Machine" "node"="test-cluster-6ffd74bd5b-khxzr"
```

When inspecting the CAPI `machine` object, you may find out that the `Node.Spec.ProviderID` is not set.
This can happen when the workload environment does not have proper network access to the underlying provider infrastructure. For example in vSphere, without the network access to vCenter endpoint, the `vsphere-cloud-controller-manager` in the workload cluster cannot set node's providerID, thus the machine will never get to `Running` state, blocking cluster provisioning from continuing.

To fix it, make sure to validate the network/firewall settings from the workload cluster to the infrastructure provider environment. Read through the `Requirements` page, especially around the networking requirements in each provider before retrying the cluster provisioning:
* [Requirements for EKS Anywhere on VMware vSphere]({{< relref "../getting-started/vsphere/vsphere-prereq" >}})
* [Network Requirements for EKS Anywhere on Bare Metal]({{< relref "../getting-started/baremetal/bare-prereq" >}})
* [Requirements for EKS Anywhere on CloudStack]({{< relref "../getting-started/cloudstack/cloudstack-prereq" >}})
* [Prerequisite Checklist for EKS Anywhere on Snow]({{< relref "../getting-started/snow/snow-getstarted/#prerequisite-checklist" >}})
* [Requirements for EKS Anywhere on Nutanix Cloud Infrastructure]({{< relref "../getting-started/nutanix/nutanix-prereq" >}})

### Labeling nodes with reserved labels such as `node-role.kubernetes.io` fails with kubeadm error during bootstrap

If cluster creation or upgrade fails to complete successfully and kubelet throws an error similar to the one below, please refer to this section. The cluster spec for EKS Anywhere create or upgrade should look like:

```
.
.
   controlPlaneConfiguration:        
      count: 2                       
      endpoint:                      
         host: "192.168.x.x"
      labels:                        
        "node-role.kubernetes.io/control-plane": "cp"
   workerNodeGroupConfigurations: 
   - count: 2 
     labels:                        
        "node-role.kubernetes.io/worker": "worker"
.
.
```

If your cluster spec looks like the above one for either the control plane configuration and/or worker node configuration, you might run into the below kubelet error:
```
unknown 'kubernetes.io' or 'k8s.io' labels specified with --node-labels: [node-role.kubernetes.io/worker].
--node-labels in the 'kubernetes.io' namespace must begin with an allowed prefix (kubelet.kubernetes.io, node.kubernetes.io) or be in the specifically allowed set (beta.kubernetes.io/arch, beta.kubernetes.io/instance-type, beta.kubernetes.io/os, failure-domain.beta.kubernetes.io/region, failure-domain.beta.kubernetes.io/zone, kubernetes.io/arch, kubernetes.io/hostname, kubernetes.io/os, node.kubernetes.io/instance-type, topology.kubernetes.io/region, topology.kubernetes.io/zone)
```
Self-assigning node labels such as `node-role.kubernetes.io` using the kubelet `--node-labels` flag is not possible due to a security measure imposed by the NodeRestriction admission controller that kubeadm enables by default.

Assigning such labels to nodes can be done after the bootstrap process has completed:

```
kubectl label nodes <name> node-role.kubernetes.io/worker=""
```
For convenience, here are example one-liners to do this post-installation:

```
# For Kubernetes 1.19 (kubeadm 1.19 sets only the node-role.kubernetes.io/master label)
kubectl get nodes --no-headers -l '!node-role.kubernetes.io/master' -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}' | xargs -I{} kubectl label node {} node-role.kubernetes.io/worker=''

# For Kubernetes >= 1.20 (kubeadm >= 1.20 sets the node-role.kubernetes.io/control-plane label)
kubectl get nodes --no-headers -l '!node-role.kubernetes.io/control-plane' -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}' | xargs -I{} kubectl label node {} node-role.kubernetes.io/worker=''
```

## Bare Metal troubleshooting

### Creating new workload cluster hangs or fails

Cluster creation appears to be hung waiting for the Control Plane to be ready.
If the CLI is hung on this message for over 30 mins, something likely failed during the OS provisioning:

```
Waiting for Control Plane to be ready
```

Or if cluster creation times out on this step and fails with the following messages:

```
Support bundle archive created {"path": "support-bundle-2022-06-28T00_41_24.tar.gz"}
Analyzing support bundle {"bundle": "CLUSTER_NAME/generated/bootstrap-cluster-2022-06-28T00:41:24Z-bundle.yaml", "archive": "support-bundle-2022-06-28T00_41_24.tar.gz"}
Analysis output generated {"path": "CLUSTER_NAME/generated/bootstrap-cluster-2022-06-28T00:43:40Z-analysis.yaml"}
collecting workload cluster diagnostics
Error: waiting for workload cluster control plane to be ready: executing wait: error: timed out waiting for the condition on clusters/CLUSTER_NAME
```
In either of those cases, the following steps can help you determine the problem:

1. Export the kind cluster’s kubeconfig file:

    ```bash
    export KUBECONFIG=${PWD}/${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
    ```
1. If you have provided BMC information:

    * Check all of the machines that the EKS Anywhere CLI has picked up from the pool of hardware in the CSV file:

        ```bash
        kubectl get machines.bmc -A
        ```
    * Check if those nodes are powered on. If any of those nodes are not powered on after a while then it could be possible that BMC credentials are invalid. You can verify it by checking the logs:

        ```bash
        kubectl get tasks.bmc -n eksa-system
        kubectl get tasks.bmc <bmc-name> -n eksa-system -o yaml
        ```

    Validate BMC credentials are correct if a connection error is observed on the `tasks.bmc` resource. Note that "IPMI over LAN" or "Redfish" must be enabled in the BMC configuration for the `tasks.bmc` resource to communicate successfully.

1. If the machine is powered on but you see linuxkit is not running, then Tinkerbell failed to serve the node via iPXE. In this case, you would want to:

    * Check the Boots service logs from the machine where you are running the CLI to see if it received and/or responded to the request:

        ```bash
        docker logs boots
        ```
    * Confirm no other DHCP service responded to the request and check for any errors in the BMC console. Other DHCP servers on the network can result in race conditions and should be avoided by configuring the other server to block all MAC addresses and exclude all IP addresses used by EKS Anywhere.

1. If you see `Welcome to LinuxKit`, click enter in the BMC console to access the LinuxKit terminal. Run the following commands to check if the tink-worker container is running.

    ```bash
    docker ps -a
    docker logs <container-id>
    ```

1. If the machine has already started provisioning the OS and it’s in irrecoverable state, get the workflow of the provisioning/provisioned machine using:

    ```bash
    kubectl get workflows -n eksa-system
    kubectl describe workflow/<workflow-name> -n eksa-system
    ```

    Check all the actions and their status to determine if all actions have been executed successfully or not. If the *stream-image* has action failed, it’s likely due to a timeout or network related issue. You can also provide your own `image_url` by specifying `osImageURL` under datacenter spec.


## vSphere troubleshooting

### EKSA_VSPHERE_USERNAME is not set or is empty
```
❌ Validation failed	{"validation": "vsphere Provider setup is valid", "error": "failed setup and validations: EKSA_VSPHERE_USERNAME is not set or is empty", "remediation": ""}
```
Two environment variables need to be set and exported in your environment to create clusters successfully.
Be sure to use single quotes around your user name and password to avoid shell manipulation of these values.
```
export EKSA_VSPHERE_USERNAME='<vSphere-username>'
export EKSA_VSPHERE_PASSWORD='<vSphere-password>'
```

### vSphere authentication failed
```
❌ Validation failed	{"validation": "vsphere Provider setup is valid", "error": "error validating vCenter setup: vSphere authentication failed: govc: ServerFaultCode: Cannot complete login due to an incorrect user name or password.\n", "remediation": ""}
Error: failed to create cluster: validations failed
```
Two environment variables need to be set and exported in your environment to create clusters successfully.
Be sure to use single quotes around your user name and password to avoid shell manipulation of these values.
```
export EKSA_VSPHERE_USERNAME='<vSphere-username>'
export EKSA_VSPHERE_PASSWORD='<vSphere-password>'
```

### Issues detected with selected template
```
Issues detected with selected template. Details: - -1:-1:VALUE_ILLEGAL: No supported hardware versions among [vmx-15]; supported: [vmx-04, vmx-07, vmx-08, vmx-09, vmx-10, vmx-11, vmx-12, vmx-13].
```
Our upstream dependency on CAPV makes it a requirement that you use vSphere 6.7 update 3 or newer.
Make sure your ESXi hosts are also up to date.

### Waiting for external etcd to be ready
```
2022-01-19T15:56:57.734Z        V3      Waiting for external etcd to be ready   {"cluster": "mgmt"}
```
Debug this problem using techniques from [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}).

### Timed out waiting for the condition on deployments/capv-controller-manager
```
Failed to create cluster {"error": "error initializing capi in bootstrap cluster: error waiting for capv-controller-manager in namespace capv-system: error executing wait: error: timed out waiting for the condition on deployments/capv-controller-manager\n"}
```
Debug this problem using techniques from [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}).

### Timed out waiting for the condition on clusters/<your-cluster-name>
```
Failed to create cluster {"error": "error waiting for workload cluster control plane to be ready: error executing wait: error: timed out waiting for the condition on clusters/test-cluster\n"}
```
This can be an issue with the number of control plane and worker node replicas defined in your cluster yaml file.
Try to start off with a smaller number (3 or 5 is recommended for control plane) in order to bring up the cluster.

This error can also occur because your vCenter server is using self-signed certificates and you have `insecure` set to true in the generated cluster yaml.
To check if this is the case, run the commands below:
```bash
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
kubectl get machines
```
If all the machines are in `Provisioning` phase, this is most likely the issue.
To resolve the issue, set `insecure` to `false` and `thumbprint` to the TLS thumbprint of your vCenter server in the cluster yaml and try again.

```
"msg"="discovered IP address"
```
The aforementioned log message can also appear with an address value of the control plane in either of the ${CLUSTER_NAME}/logs/capv-controller-manager.log file
or the capv-controller-manager pod log which can be extracted with the following command,
```bash
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
kubectl logs -f -n capv-system -l control-plane="controller-manager" -c manager
```
Make sure you are choosing a control plane ip in your network range that does not conflict with other VMs.

### Generic cluster unavailable

The first thing to look at is: were virtual machines created on your target provider? In the case of vSphere, you should see some VMs in your folder and they should be up. Check the console and if you see:
```
[FAILED] Failed to start Wait for Network to be Configured.
```
Make sure your DHCP server is up and working.

For more troubleshooting tips. see the [CAPV Troubleshooting](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/blob/master/docs/troubleshooting.md#debugging-issues) guide.

### Workload VM is created on vSphere but can not power on
A similar issue is the VM does power on but does not show any logs on the console and does not have any IPs assigned.

This issue can occur if the `resourcePool` that the VM uses does not have enough CPU or memory resources to run a VM.
To resolve this issue, increase the CPU and/or memory reservations or limits for the resourcePool.

### Workload VMs start but Kubernetes not working properly
If the workload VMs start, but Kubernetes does not start or is not working properly, you may want to log onto the VMs and check the logs there.
If Kubernetes is at least partially working, you may use `kubectl` to get the IPs of the nodes:
```
kubectl get nodes -o=custom-columns="NAME:.metadata.name,IP:.status.addresses[2].address"
```
If Kubernetes is not working at all, you can get the IPs of the VMs from vCenter or using `govc`.

When you get the external IP you can `ssh` into the nodes using the private ssh key associated with the public ssh key you provided in your cluster configuration:
```
ssh -i <ssh-private-key> <ssh-username>@<external-IP>
```

### Create command stuck on `Creating new workload cluster`

There can be a few reasons that the create command is stuck on `Creating new workload cluster` for over 30 min.

First, check the vSphere UI to see if any workload VM are created.

#### No node VMs are created in vSphere

If no VMs are created, check the `capi-controller-manager`, `capv-controller-manager` and `capi-kubeadm-control-plane-controller-manager` logs using the commands mentioned in [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}) section.

#### No IP assigned to a VM

If a VM is created, check to see if it has an IPv4 IP assigned. For example, in BottleRocket machine boot logs, you might see `Failed to read current IP data`.

If there are no IPv4 IPs assigned to VMs, this is most likely because you don't have a DHCP server configured for the `network` configured in the cluster config yaml, OR there are not enough IP addresses available in the DHCP pool to assign to the VMs. Ensure that you have a DHCP server running with [enough IP addresses to create a cluster]({{< relref "../clustermgmt/cluster-upgrades/vsphere-and-cloudstack-upgrades/#prepare-dhcp-ip-addresses-pool" >}}) before running the create or upgrade command again.

To confirm this is a DHCP issue, you could create a new VM in the same network to validate if an IPv4 IP is assigned correctly.

#### Control Plane IP in clusterconfig is not present on any Control Plane VM

If there are any IPv4 IPs assigned, check if one of the VMs have the controlPlane IP specified in `Cluster.spec.controlPlaneConfiguration.endpoint.host` in the clusterconfig yaml.
If this IP is not present on any control plane VM, make sure the `network` has access to the following endpoints:

{{% content "../getting-started/vsphere/domains.md" %}}

#### `Failed to connect to <vSphere-FQDN>: connection refused` on vsphere-cloud-controller-manager

If the IPv4 IPs are assigned to the VM and you have the workload kubeconfig under `<cluster-name>/<cluster-name>-eks-a-cluster.kubeconfig`, you can use it to check `vsphere-cloud-controller-manager` logs.
```
kubectl logs -n kube-system vsphere-cloud-controller-manager-<xxxxx> --kubeconfig <cluster-name>/<cluster-name>-eks-a-cluster.kubeconfig
```

If you see the message below in the logs, it means your cluster nodes do not have access to vSphere, which is required for the cluster to get to a ready state.
```
Failed to connect to <vSphere-FQDN>: connection refused
```
In this case, you need to enable inbound traffic from your cluster nodes on your vCenter's management network.

### Cluster Deletion Fails
If cluster deletion fails, you may need to manually delete the VMs associated with the cluster.
The VMs should be named with the cluster name.
You can power off and delete from disk using the vCenter web user interface.
You may also use `govc`:
```
govc find -type VirtualMachine --name '<cluster-name>*'
```
This will give you a list of virtual machines that should be associated with your cluster.
For each of the VMs you want to delete run:
```
VM_NAME=vm-to-destroy
govc vm.power -off -force $VM_NAME
govc object.destroy $VM_NAME
```

### Restore VM for machine in vSphere

When a CAPI machine is in not ready state, it is possible that the machine object does not match the underlying VM in vSphere. Instead, the machine is pointing to a VM that does not exist anymore.

In order to solve this, you need to update the `vspheremachine` and `vspherevm` objects to point to the new VM running in vSphere. This requires getting the `UUID` for the new VM (using `govc` command), and using it to update the `biosUUID` and `providerID` fields. Notice that this would create a mismatch between the CAPI resources names (`vspheremachine` and `vspherevm`) and the actual VM in vSphere. However, that is only an aesthetic inconvenience, since the mapping is done based on `UUID` and not name. Once the objects are updated, a machine rollout would replace the VM with a new one, removing the name mismatch problem. You also need to cleanup the error message and error reason from the statuses to resume the reconciliation of these objects. With that done, the machine objects should move to `Running` state.

### vCenter VM Deletion Causes Cluster Stuck in Deleting State with Orphaned Objects

When a VM in vCenter has been deleted, however, your cluster is stuck in a deleting state with the corresponding the CAPI objects left behind, the user has to manually delete a few related resources to recover. The resources of interest related to the VM are the following: `machine`, `vspheremachine`, `vspherevm`, and `node`.

To clean up the resources in order to recover from this deleting state:
1. Delete the the corresponding CAPI `machine` object.
    ```
    kubectl delete machine -n eksa-system <machine-name>
    ```

2. If that is stuck, it means that the `capi-controller-manager` is unable to remove the finalizer from the object, so the user needs force delete the object.

    You can do this by editing the machine object using `kubectl` and removing the finalizers on it.
    ```
    kubectl edit machine -n eksa-system <machine-name>`
    ```

    Look for the finalizers field in the metadata section of the resource. remove the finalizers. Save your changes.
    ```
    metadata:
    finalizers:
    - finalizer.example.com
    ```

   Remove the finalizers field under metada and save your changes. After, the object should be deleted.

3. If the `capv-controller-manager` is unable to to clean up the `vspheremachine` and `vspherevm` objects, repeat the steps above for those orphaned objects that are related to the deleted VM.

4. You may also need to manually delete the corresponding orphaned `node` object using `kubectl`.
    ```
    kubectl delete node <node-name>
    ```

## Troubleshooting GitOps integration
### Cluster creation failure leaves outdated cluster configuration in GitHub.com repository
Failed cluster creation can sometimes leave behind cluster configuration files committed to your GitHub.com repository.
Make sure to delete these configuration files before you re-try `eksctl anywhere create cluster`.
If these configuration files are not deleted, GitOps installation will fail but cluster creation will continue.

They'll generally be located under the directory
`clusters/$CLUSTER_NAME` if you used the default path in your `flux` `gitops` config.
Delete the entire directory named $CLUSTER_NAME.

### Cluster creation failure leaves empty GitHub.com repository
Failed cluster creation can sometimes leave behind a completely empty GitHub.com repository.
This can cause the GitOps installation to fail if you re-try the creation of a cluster which uses this repository.
If cluster creation failure leaves behind an empty github repository, please manually delete the created GitHub.com repository before attempting cluster creation again.

### Changes not syncing to cluster
Please remember that the only fields currently supported for GitOps are:

Cluster

- `Cluster.workerNodeGroupConfigurations.count`
- `Cluster.workerNodeGroupConfigurations.machineGroupRef.name`

Worker Nodes

- `VsphereMachineConfig.diskGiB`
- `VsphereMachineConfig.numCPUs`
- `VsphereMachineConfig.memoryMiB`
- `VsphereMachineConfig.template`
- `VsphereMachineConfig.datastore`
- `VsphereMachineConfig.folder`
- `VsphereMachineConfig.resourcePool`

If you've changed these fields and they're not syncing to the cluster as you'd expect,
check out the logs of the pod in the `source-controller` deployment in the `flux-system` namespaces.
If `flux` is having a problem connecting to your GitHub repository the problem will be logged here.

```sh
$ kubectl get pods -n flux-system
NAME                                       READY   STATUS    RESTARTS   AGE
helm-controller-7d644b8547-k8wfs           1/1     Running   0          4h15m
kustomize-controller-7cf5875f54-hs2bt      1/1     Running   0          4h15m
notification-controller-776f7d68f4-v22kp   1/1     Running   0          4h15m
source-controller-7c4555748d-7c7zb         1/1     Running   0          4h15m
```
```sh
$ kubectl logs source-controller-7c4555748d-7c7zb -n flux-system
```
A well behaved flux pod will simply log the ongoing reconciliation process, like so:
```sh
{"level":"info","ts":"2021-07-01T19:58:51.076Z","logger":"controller.gitrepository","msg":"Reconciliation finished in 902.725344ms, next run in 1m0s","reconciler group":"source.toolkit.fluxcd.io","reconciler kind":"GitRepository","name":"flux-system","namespace":"flux-system"}
{"level":"info","ts":"2021-07-01T19:59:52.012Z","logger":"controller.gitrepository","msg":"Reconciliation finished in 935.016754ms, next run in 1m0s","reconciler group":"source.toolkit.fluxcd.io","reconciler kind":"GitRepository","name":"flux-system","namespace":"flux-system"}
{"level":"info","ts":"2021-07-01T20:00:52.982Z","logger":"controller.gitrepository","msg":"Reconciliation finished in 970.03174ms, next run in 1m0s","reconciler group":"source.toolkit.fluxcd.io","reconciler kind":"GitRepository","name":"flux-system","namespace":"flux-system"}
```

If there are issues connecting to GitHub, you'll instead see exceptions in the `source-controller` log stream.
For example, if the deploy key used by `flux` has been deleted, you'd see something like this:
```sh
{"level":"error","ts":"2021-07-01T20:04:56.335Z","logger":"controller.gitrepository","msg":"Reconciler error","reconciler group":"source.toolkit.fluxcd.io","reconciler kind":"GitRepository","name":"flux-system","namespace":"flux-system","error":"unable to clone 'ssh://git@github.com/youruser/gitops-vsphere-test', error: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain"}
```

### Other ways to troubleshoot GitOps integration
If you're still having problems after deleting any empty EKS Anywhere created GitHub repositories and looking at the `source-controller` logs.
You can look for additional issues by checking out the deployments in the `flux-system` and `eksa-system` namespaces and ensure they're running and their log streams are free from exceptions.

```sh
$ kubectl get deployments -n flux-system
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
helm-controller           1/1     1            1           4h13m
kustomize-controller      1/1     1            1           4h13m
notification-controller   1/1     1            1           4h13m
source-controller         1/1     1            1           4h13m
```

```sh
$ kubectl get deployments -n eksa-system
NAME                      READY   UP-TO-DATE   AVAILABLE   AGE
eksa-controller-manager   1/1     1            1           4h13m
```

## Snow troubleshooting

### Device outage

These are some conditions that can cause a device outage:

* Intentional outage (a planned power outage or an outage when moving devices, for example).
* Unintentional outage (a subset of devices or all devices are rebooted, or experiencing network disconnections from the LAN, which make device offline or isolated from the cluster).

>**_NOTE:_** If all Snowball Edge devices are moved to a different place and connected to a different local network, make sure you use the same subnet, netmask, and gateway for your network configuration. After moving, devices and all node instances need to maintain the original IP addresses. Then, follow the recover cluster procedure to get your cluster up and running again. Otherwise, it might be impossible to resume the cluster.

**To recover a cluster**

If there is a subset of devices or all devices experience an outage, see [Downloading and Installing the Snowball Edge client](https://docs.aws.amazon.com/snowball/latest/developer-guide/download-the-client.html) to get the Snowball Edge client and then follow these steps:

1. Reboot and unlock all affected devices manually.

    ```sh
    // use reboot-device command to reboot device, this may take several minutes
    $ path-to-snowballEdge_CLIENT reboot-device --endpoint https://snowball-ip --manifest-file path-to-manifest-file --unlock-code unlock-code

    // use describe-device command to check the status of device
    $ path-to-snowballEdge_CLIENT describe-device --endpoint https://snowball-ip --manifest-file path-to-manifest-file --unlock-code unlock-code

    // when the State in the output of describe-device is LOCKED, run unlock-device
    $ path-to-snowballEdge_CLIENT unlock-device --endpoint https://snowball-ip --manifest-file path-to-manifest-file --unlock-code unlock-code

    // use describe-device command to check the status of device until device is unlocked
    $ path-to-snowballEdge_CLIENT describe-device --endpoint https://snowball-ip --manifest-file path-to-manifest-file --unlock-code unlock-code
    ```

1. Get all instance IDs that were part of the cluster by looking up the impacted device IP in the PROVIDERID column.

    ```sh
    $ kubectl get machines -A --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig

    NAMESPACE     NAME            CLUSTER        NODENAME          PROVIDERID                                        PHASE     AGE   VERSION
    eksa-system   machine-name-1  cluster-name   node-name-1       aws-snow:///192.168.1.39/s.i-8319d8c75d54a32cc    Running   82s   v1.24.9-eks-1-24-7
    eksa-system   machine-name-2  cluster-name   node-name-2       aws-snow:///192.168.1.39/s.i-8d7d3679a1713e403    Running   82s   v1.24.9-eks-1-24-7
    eksa-system   machine-name-3  cluster-name   node-name-3       aws-snow:///192.168.1.231/s.i-8201c356fb369c37f   Running   81s   v1.24.9-eks-1-24-7
    eksa-system   machine-name-4  cluster-name   node-name-4       aws-snow:///192.168.1.39/s.i-88597731b5a4a9044    Running   81s   v1.24.9-eks-1-24-7
    eksa-system   machine-name-5  cluster-name   node-name-5       aws-snow:///192.168.1.77/s.i-822f0f46267ad4c6e    Running   81s   v1.24.9-eks-1-24-7
    ```

1. Start all instances on the impacted devices as soon as possible.

    ```sh
    $ aws ec2 start-instances --instance-id instance-id-1 instance-id-2 ... --endpoint http://snowball-ip:6078 --profile profile-name
    ```

1. Check the balance status of the current cluster after the cluster is ready again.

    ```sh
    $ kubectl get machines -A --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. Check if you have unstacked etcd machines.

    * If you have unstacked etcd machines, check the provision of unstacked etcd machines. You can find the device IP in the `PROVIDERID` column.
        * If there are more than 1 unstacked etcd machines provisioned on the same device and there are devices with no unstacked etcd machine, you need to rebalance unstacked etcd nodes. Follow the [rebalance nodes]({{< relref "./troubleshooting/#how-to-rebalance-nodes" >}}) procedure to rebalance your unstacked etcd nodes in order to recover high availability.
        * If you have your etcd nodes evenly distributed with 1 device having at most 1 etcd node, you are done with the recovery.

    * If you don’t have unstacked etcd machines, check the provision of control plane machines. You can find the device IP in `PROVIDERID` column.
        * If there are more than 1 control plan machines provisioned on the same device and there are devices with no control plane machine, you need to rebalance control plane nodes. Follow the [rebalance nodes]({{< relref "./troubleshooting/#how-to-rebalance-nodes" >}}) procedure to rebalance your control plane nodes in order to recover high availability.
        * If you have your control plane nodes evenly distributed with 1 device having at most 1 control plane node, you are done with the recovery.

#### How to rebalance nodes

1. Confirm the machines you want to delete and get their node name from the NODENAME column.

    You can determine which machines need to be deleted by referring to the AGE column. The newly-generated machines have short AGE. Delete those new etcd/control plane machine nodes which are not the only etcd/control plane machine nodes on their devices.

1. Cordon each node so no further workloads are scheduled to run on it.

    ```sh
    $ kubectl cordon node-name --ignore-daemonsets --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. Drain machine nodes of all current workloads.

    ```sh
    $ kubectl drain node-name --ignore-daemonsets --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. Delete machine node.

    ```sh
    $ kubectl delete node node-name --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. Repeat this process until etcd/control plane machine nodes are evenly provisioned.

### Device replacement

There might be some reasons which can require device replacement:

* When a subset of devices are determined to be broken and you want to join a new device into current cluster.
* When a subset of devices are offline and come back with a new device IP.

**To upgrade a cluster with new devices**:

1. Add new certificates to the certificate file and new credentials to the credential file.

1. Change the device list in your cluster yaml configuration file and use the `eksctl anywhere upgrade cluster` command.

    ```
    $ eksctl anywhere upgrade cluster -f eks-a-cluster.yaml
    ```

### Node outage

**Unintentional instance outage**

When an instance is in exception status (for example, terminated/stopped for some reason), it will be discovered automatically by Amazon EKS Anywhere and there will be a new replacement instance node created after 5 minutes. The new node will be provisioned to devices based on even provision strategy. In this case, the new node will be provisioned to a device with the fewest number of machines of the same type. Sometimes, more than one device will have the same number of machines of this type. Thus, we cannot guarantee it will be provisioned on the original device.

**Intentional node replacement**

If you want to replace an unhealthy node which didn't get detected by Amazon EKS Anywhere automatically, you can follow these steps.

>**_NOTE:_** Do not delete all worker machine nodes or control plane nodes or etcd nodes at the same time. Make sure you delete machine nodes one by one.

1. Cordon nodes so no further workloads are scheduled to run on it.

    ```sh
    $ kubectl cordon node-name --ignore-daemonsets --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. Drain machine nodes of all current workloads.

    ```sh
    $ kubectl drain node-name --ignore-daemonsets --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. Delete machine nodes.

    ```sh
    $ kubectl delete node node-name --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

1. New nodes will be provisioned automatically. You can check the provision result with the get machines command.

    ```sh
    $ kubectl get machines -A --kubeconfig=cluster-name/cluster-name-eks-a-cluster.kubeconfig
    ```

### Cluster Deletion Fails

If your Amazon EKS Anywhere cluster creation failed and the `eksctl anywhere delete cluster -f eksa-cluster.yaml` command cannot be run successfully, manually delete a few resources before trying the command again. Run the following commands from the computer on which you set up the AWS configuration and have the [Snowball Edge Client installed](https://docs.aws.amazon.com/snowball/latest/developer-guide/download-the-client.html). If you are using multiple Snowball Edge devices, run these commands on each.

```sh
// get the list of instance ids that are created for Amazon EKS Anywhere cluster,
// that can be identified by cluster name in the tag of the output
$ aws ec2 describe-instances --endpoint http://snowball-ip:8008 --profile profile-name

// the next two commands are for deleting DNI, this needs to be done before deleting instance
$ PATH_TO_Snowball_Edge_CLIENT/bin/snowballEdge describe-direct-network-interfaces --endpoint https://snowball-ip --manifest-file path-to-manifest-file --unlock-code unlock-code

// DNI arn can be found in the output of last command, which is associated with the specific instance id you get from describe-instances
$ PATH_TO_Snowball_Edge_CLIENT/bin/snowballEdge delete-direct-network-interface --direct-network-interface-arn DNI-ARN --endpoint https://snowball-ip --manifest-file path-to-manifest-file --unlock-code unlock-code

// delete instance
$ aws ec2 terminate-instances --instance-id instance-id-1,instance-id-2 --endpoint http://snowball-ip:8008 --profile profile-name
```

### Generate a log file from the Snowball Edge device

You can also generate a log file from the Snowball Edge device for AWS Support. See [AWS Snowball Edge Logs](https://docs.aws.amazon.com/snowball/latest/developer-guide/using-client-commands.html#logs) in this guide.

## Nutanix troubleshooting

### Error creating Nutanix client

```
Error: error creating nutanix client: username, password and endpoint are required
```

Verify if the required environment variables are set before creating the clusters:
```
export EKSA_NUTANIX_USERNAME="<Nutanix-username>"
export EKSA_NUTANIX_PASSWORD="<Nutanix-password>"
```

Also, make sure the `spec.endpoint` is correctly configured in the `NutanixDatacenterConfig`. The value of the `spec.endpoint` should be the IP or FQDN of Prism Central.

### x509: certificate signed by unknown authority

Failure of the `nutanix Provider setup is valid` validation with the `x509: certificate signed by unknown authority` message indicates the certificate of the Prism Central endpoint is not trusted.
In case Prism Central is configured with self-signed certificates, it is recommended to configure the `additionalTrustBundle` in the `NutanixDatacenterConfig`. More information can be found [here](https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/nutanix/#nutanixdatacenterconfig-fields).
