---
title: "eksctl anywhere CLI reference"
linkTitle: "eksctl command"
weight: 65
description: >
  Details on the options and parameters for eksctl anywhere CLI
---

The `eksctl` CLI, with the EKS Anywhere plugin added, lets you create and manage EKS Anywhere clusters.
While a cluster is running, most EKS Anywhere administration can be done using `kubectl` or other native Kubernetes tools.

Use this page as a reference to useful `eksctl anywhere` command examples for working with EKS Anywhere clusters.
Available `eksctl anywhere` commands include:

* `create cluster` To create an EKS Anywhere cluster
* `delete cluster`  To delete an EKS Anywhere cluster
* `generate` [`clusterconfig` | `support-bundle` | `support-bundle-config`] To generate cluster and support configs
* `help`  To get help information
* `upgrade` To upgrade a workload cluster
* `version` To get the EKS Anywhere version

Options used with multiple commands include:

* `-h` or `--help` To get help for a command or subcommand
* `-v int` or `--verbosity int` To set log level verbosity from 0-9
* `-f `filename` or `--filename filename` To identify the filename containing the cluster config
* `--force-cleanup` To force deletion of previously created bootstrap cluster
* `-w string` or `--w-config string` To identify the kubeconfig file when needed to create a support bundle or upgrade a cluster

Other available options and arguments are listed with the command examples that follow.

## `eksctl anywhere generate`

With `eksctl anywhere generate`, you can output sets of cluster resources to create a new cluster
or troubleshoot an existing cluster.
Here are some examples.

### `eksctl anywhere generate clusterconfig`

Using `eksctl anywhere generate clusterconfig` you can generate a cluster configuration
for a specific provider (`-p` or `--provider`*provider_name*). Here are examples:

Generate a configuration file to create an EKS Anywhere cluster for a `vsphere` provider:

```
export CLUSTER_NAME=vsphere01
eksctl anywhere generate clusterconfig ${CLUSTER_NAME} -p vsphere > ${CLUSTER_NAME}.yaml
```
Generate a configuration file to create an EKS Anywhere cluster for a Docker provider:

```
export CLUSTER_NAME=docker01
eksctl anywhere generate clusterconfig ${CLUSTER_NAME} -p docker > ${CLUSTER_NAME}.yaml
```
Once you have generated the yaml configuration file, edit that file to add configuration information before you use the file to create your cluster.
See [local](../../getting-started/local-environment/) and [production](../../getting-started/production-environment/) cluster creation procedures for details.

### `eksctl anywhere generate support-bundle-config`

If you would like to customize your support bundle, you can generate a support bundle configuration file (`support-bundle-config`),
edit that file to choose the data you want to gather,
then gather the selected data into a support bundle (`support-bundle`).

Generate a support bundle config file (then edit that file to select the log data you want to gather):

```
export CLUSTER_NAME=vsphere01
eksctl anywhere generate support-bundle-config > ${CLUSTER_NAME}_bundle_config.yaml 
```
### `eksctl anywhere generate support-bundle`

Once you have a bundle config file, generate a support bundle from an existing EKS Anywhere cluster.
Additional options available for this command include:

* `--bundle-config string` To identify the bundle config file to use to generate the support bundle
* `--since string` To collect pod logs in the latest duration like 5s, 2m, or 3h.
* `--since-time string` To collect pod logs after a specific datetime(RFC3339) like 2021-06-28T15:04:05Z

Here is an example:

```
export CLUSTER_NAME=vsphere01
eksctl anywhere generate support-bundle --bundle-config ${CLUSTER_NAME}_bundle_config.yaml \
   -w ${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig \
   --since 2h -f ${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.yaml
```

The example just shown:

* Uses `${CLUSTER_NAME}_bundle.yaml` as the file to hold the results
* Collects pod logs for the past two hours (2h)
* Identifies the bundle config file to use (`${CLUSTER_NAME}_bundle_config.yaml`)
* Identifies the `.kubeconfig` file to use for a workload cluster

To change the command to generate a support bundle that gathers pod logs starting from a specific date (September 8, 2021) and time (1:27 PM):

```
export CLUSTER_NAME=vsphere01
eksctl anywhere generate support-bundle --bundle-config ${CLUSTER_NAME}_bundle_config.yaml \
   -w KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig \
   --since-time 2021-09-8T13:27:00Z 2h -f ${CLUSTER_NAME}_bundle.yaml
```

## `eksctl anywhere create cluster`

Create an EKS Anywhere cluster from a cluster configuration file you generated (and modified) earlier.
This example sets verbosity to most verbose (`-v 9`):

```
export CLUSTER_NAME=vsphere01
eksctl anywhere create cluster -v 9 -f ${CLUSTER_NAME}.yaml
```

See [local](../../getting-started/local-environment/) and [production](../../getting-started/production-environment/) cluster creation procedures for details.

## `eksctl anywhere upgrade cluster`

Upgrade an existing EKS Anywhere cluster.
This example uses maximum verbosity and forces a cleanup of the previously created bootstrap cluster:

```
export CLUSTER_NAME=vsphere01
eksctl anywhere upgrade cluster -f ${CLUSTER_NAME}.yaml --force-cleanup -v9 \
   -w KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig 
```
For more information on this and other ways to upgrade a cluster, see [Upgrade cluster](../../tasks/cluster/cluster-upgrades/).

## `eksctl anywhere delete cluster`

Delete an existing EKS Anywhere cluster.
This example deletes all VMs and the forces the deletion of the previously created bootstrap cluster:

```
export CLUSTER_NAME=vsphere01
eksctl anywhere delete cluster -f ${CLUSTER_NAME}.yaml \
   --force-cleanup \
   -w KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig 
```
For more information on deleting a cluster, see [Delete cluster](../../tasks/cluster/cluster-delete/).

## `eksctl anywhere version`

View the version of `eksctl anywhere`:

```
eksctl anywhere version
v0.5.0
```
## `eksctl anywhere help`

Use `eksctl anywhere help` or the `-h` option to see general options or options specific to a particular set of commands.

View general help information using `help`:

```
eksctl anywhere help

Use eksctl anywhere to build your own self-managing cluster on your hardware with the best of Amazon EKS

Usage:
  eksctl anywhere [command]

Available Commands:
  create      Create resources
  delete      Delete resources
  generate    Generate resources
  help        Help about any command
  upgrade     Upgrade resources
  version     Get the eksctl version

Flags:
  -h, --help            help for eksctl
  -v, --verbosity int   Set the log level verbosity

Use "eksctl [command] --help" for more information about a command.
...
```

Display help options for generating a support bundle:

```
eksctl anywhere generate support-bundle -h

This command is used to create a support bundle to troubleshoot a cluster

Usage:
  eksctl anywhere generate support-bundle -f my-cluster.yaml [flags]

Flags:
      --bundle-config string   Bundle Config file to use when generating support bundle
  -f, --filename string        Filename that contains EKS-A cluster configuration
  -h, --help                   help for support-bundle
      --since string           Collect pod logs in the latest duration like 5s, 2m, or 3h.
      --since-time string      Collect pod logs after a specific datetime(RFC3339) like 2021-06-28T15:04:05Z
  -w, --w-config string        Kubeconfig file to use when creating support bundle for a workload cluster

Global Flags:
  -v, --verbosity int   Set the log level verbosity

```
Display options for creating a cluster:

```
eksctl anywhere create cluster -h
This command is used to create workload clusters

Usage:
  eksctl anywhere create cluster [flags]

Flags:
  -f, --filename string   Filename that contains EKS-A cluster configuration
      --force-cleanup     Force deletion of previously created bootstrap cluster
  -h, --help              help for cluster

Global Flags:
  -v, --verbosity int   Set the log level verbosity
```
