---
title: "Troubleshooting"
linkTitle: "Troubleshooting"
weight: 40
description: >
  Troubleshooting EKS Anywhere clusters
aliases:
   - /docs/tasks/troubleshoot/_troubleshooting
---

This guide covers some generic troubleshooting techniques and then cover more detailed examples. You may want to search this document for a fragment of the error you are seeing.

### Increase eksctl anywhere output

If you’re having trouble running `eksctl anywhere` you may get more verbose output with the `-v 6` option. The highest level of verbosity is `-v 9` and the default level of logging is level equivalent to `-v 0`.

### Cannot run docker commands

The EKS Anywhere binary requires access to run docker commands without using `sudo`.
If you're using a Linux distribution you will need to be using Docker 20.x.x add your user needs to be part of the docker group.

To add your user to the docker group you can use.

```bash
sudo usermod -a -G docker $USER
```

Now you need to log out and back in to get the new group permissions.

### Minimum requirements for docker version have not been met
```
Error: failed to validate docker: minimum requirements for docker version have not been met. Install Docker version 20.x.x or above
```
Ensure you are running Docker 20.x.x for example:
```
% docker --version
Docker version 20.10.6, build 370c289
```

### Minimum requirements for docker version have not been met on Mac OS
```
Error: EKS Anywhere does not support Docker desktop versions between 4.3.0 and 4.4.1 on macOS
```
```
Error: EKS Anywhere requires Docker desktop to be configured to use CGroups v1. Please  set `deprecatedCgroupv1:true` in your `~/Library/Group\\ Containers/group.com.docker/settings.json` file
```
Ensure you are running Docker Desktop 4.4.2 or newer and have set `"deprecatedCgroupv1": true` in your settings.json file
```
% defaults read /Applications/Docker.app/Contents/Info.plist CFBundleShortVersionString
4.42
% docker info --format '{{json .CgroupVersion}}' 
"1"
```

### ECR access denied

```
Error: failed to create cluster: unable to initialize executables: failed to setup eks-a dependencies: Error response from daemon: pull access denied for public.ecr.aws/***/cli-tools, repository does not exist or may require 'docker login': denied: Your authorization token has expired. Reauthenticate and try again.
```

All images needed for EKS Anywhere are public and do not need authentication. Old cached credentials could trigger this error.
Remove cached credentials by running:
```sh
docker logout public.ecr.aws
```

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

### error unmarshaling JSON: while decoding JSON: json: unknown field "spec"
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

### failed to create cluster: node(s) already exist for a cluster with the name
```
Performing provider setup and validations
Creating new bootstrap cluster
Error create bootstrapcluster	{"error": "error creating bootstrap cluster: error executing create cluster: ERROR: failed to create cluster: node(s) already exist for a cluster with the name \"cluster-name\"\n, try rerunning with --force-cleanup to force delete previously created bootstrap cluster"}
Failed to create cluster	{"error": "error creating bootstrap cluster: error executing create cluster: ERROR: failed to create cluster: node(s) already exist for a cluster with the name \"cluster-name\"\n, try rerunning with --force-cleanup to force delete previously created bootstrap cluster"}ry rerunning with --force-cleanup to force delete previously created bootstrap cluster"}
```
A bootstrap cluster already exists with the same name. If you are sure the cluster is not being used, you may use the `--force-cleanup` option to `eksctl anywhere` to delete the cluster or you may delete the cluster with `kind delete cluster --name <cluster-name>`. If you do not have `kind` installed, you may use `docker stop` to stop the docker container running the KinD cluster.

### Bootstrap cluster fails to come up
If your bootstrap cluster has problems you may get detailed logs by looking at the files created under the `${CLUSTER_NAME}/logs` folder. The capv-controller-manager log file will surface issues with vsphere specific configuration while the capi-controller-manager log file might surface other generic issues with the cluster configuration passed in.

You may also access the logs from your bootstrap cluster directly as below:
```bash
export KUBECONFIG=${PWD}/${CLUSTER_NAME}/generated/${CLUSTER_NAME}.kind.kubeconfig
kubectl logs -f -n capv-system -l control-plane="controller-manager" -c manager
```

It also might be useful to start a shell session on the docker container running the bootstrap cluster by running `docker ps` and then `docker exec -it <container-id> bash` the kind container.

### Memory or disk resource problem
There are various disk and memory issues that can cause problems.
Make sure docker is configured with enough memory.
Make sure the system wide Docker memory configuration provides enough RAM for the bootstrap cluster.

Make sure you do not have unneeded KinD clusters running `kind get clusters`.
You may want to delete unneeded clusters with `kind delete cluster --name <cluster-name>`.
If you do not have kind installed, you may install it from https://kind.sigs.k8s.io/ or use `docker ps` to see the KinD clusters and `docker stop` to stop the cluster.
 
Make sure you do not have any unneeded Docker containers running with `docker ps`.
Terminate any unneeded Docker containers.
   
Make sure Docker isn't out of disk resources.
If you don't have any other docker containers running you may want to run `docker system prune` to clean up disk space.

You may want to restart Docker.
To restart Docker on Ubuntu `sudo systemctl restart docker`.

### Issues detected with selected template
```
Issues detected with selected template. Details: - -1:-1:VALUE_ILLEGAL: No supported hardware versions among [vmx-15]; supported: [vmx-04, vmx-07, vmx-08, vmx-09, vmx-10, vmx-11, vmx-12, vmx-13].
```
Our upstream dependency on CAPV makes it a requirement that you use vSphere 6.7 update 3 or newer.
Make sure your ESXi hosts are also up to date.

### Waiting for cert-manager to be available... Error: timed out waiting for the condition
```
Failed to create cluster {"error": "error initializing capi resources in cluster: error executing init: Fetching providers\nInstalling cert-manager Version=\"v1.1.0\"\nWaiting for cert-manager to be available...\nError: timed out waiting for the condition\n"}
```
This is likely a [Memory or disk resource problem]({{< relref "#memory-or-disk-resource-problem" >}}).
You can also try using techniques from [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}).

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
Make sure you are choosing an ip in your network range that does not conflict with other VMs.
https://anywhere.eks.amazonaws.com/docs/reference/clusterspec/vsphere/#controlplaneconfigurationendpointhost-required


### The connection to the server localhost:8080 was refused 
```
Performing provider setup and validations
Creating new bootstrap cluster
Installing cluster-api providers on bootstrap cluster
Error initializing capi in bootstrap cluster	{"error": "error waiting for capi-kubeadm-control-plane-controller-manager in namespace capi-kubeadm-control-plane-system: error executing wait: The connection to the server localhost:8080 was refused - did you specify the right host or port?\n"}
Failed to create cluster	{"error": "error waiting for capi-kubeadm-control-plane-controller-manager in namespace capi-kubeadm-control-plane-system: error executing wait: The connection to the server localhost:8080 was refused - did you specify the right host or port?\n"}
```
This is likely a [Memory or disk resource problem]({{< relref "#memory-or-disk-resource-problem" >}}).

### Generic cluster unavailable
Troubleshoot more by inspecting bootstrap cluster or workload cluster (depending on the stage of failure) using kubectl commands. 
```
kubectl get pods -A --kubeconfig=<kubeconfig>
kubectl get nodes -A --kubeconfig=<kubeconfig>
kubectl get logs <podname> -n <namespace> --kubeconfig=<kubeconfig>
....
```
Capv troubleshooting guide: https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/blob/master/docs/troubleshooting.md#debugging-issues

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

When you get the external IP you can ssh into the nodes using the private ssh key associated with the public ssh key you provided in your cluster configuration:
```
ssh -i <ssh-private-key> <ssh-username>@<external-IP>
```

### create command stuck on `Creating new workload cluster`
There can we a few reasons if the create command is stuck on `Creating new workload cluster` for over 30 min.
First, check the vSphere UI to see if any workload VM are created.

If any VMs are created, check to see if they have any IPv4 IPs assigned to them.

If there are no IPv4 IPs assigned to them, this is most likely because you don't have a DHCP server configured for the `network` configured in the cluster config yaml.
Ensure that you have DHCP running and run the create command again.

If there are any IPv4 IPs assigned, check if one of the VMs have the controlPlane IP specified in `Cluster.spec.controlPlaneConfiguration.endpoint.host` in the clusterconfig yaml.
If this IP is not present on any control plane VM, make sure the `network` has access to the following endpoints:

{{% content "../../reference/vsphere/domains.md" %}}

If the IPv4 IPs are assigned to the VM and you have the workload kubeconfig under `<cluster-name>/<cluster-name>-eks-a-cluster.kubeconfig`, you can use it to check `vsphere-cloud-controller-manager` logs.
```
kubectl logs -n kube-system vsphere-cloud-controller-manager-<xxxxx> --kubeconfig <cluster-name>/<cluster-name>-eks-a-cluster.kubeconfig
```

If you see this message in the logs, it means your cluster nodes do not have access to vSphere, which is required for cluster to get to a ready state.
```
Failed to connect to <vSphere-FQDN>: connection refused
```
In this case, you need to enable inbound traffic from your cluster nodes on your vCenter's management network.

If VMs are created, but they do not get a network connection and DHCP is not configured for your vSphere deployment, you may need to [create your own DHCP server]({{< relref "../../reference/vsphere/vsphere-dhcp" >}}).
If no VMs are created, check the `capi-controller-manager`, `capv-controller-manager` and `capi-kubeadm-control-plane-controller-manager` logs using the commands mentioned in [Generic cluster unavailable]({{< relref "#generic-cluster-unavailable" >}}) section.

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

### Troubleshooting GitOps integration
#### Cluster creation failure leaves outdated cluster configuration in GitHub.com repository
Failed cluster creation can sometimes leave behind cluster configuration files committed to your GitHub.com repository.
Make sure to delete these configuration files before you re-try `eksctl anywhere create cluster`.
If these configuration files are not deleted, GitOps installation will fail but cluster creation will continue.

They'll generally be located under the directory
`clusters/$CLUSTER_NAME` if you used the default path in your `flux` `gitops` config.
Delete the entire directory named $CLUSTER_NAME.

#### Cluster creation failure leaves empty GitHub.com repository
Failed cluster creation can sometimes leave behind a completely empty GitHub.com repository.
This can cause the GitOps installation to fail if you re-try the creation of a cluster which uses this repository.
If cluster creation failure leaves behind an empty github repository, please manually delete the created GitHub.com repository before attempting cluster creation again.

#### Changes not syncing to cluster
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

#### Other ways to troubleshoot GitOps integration
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
