---
title: "Generating a Support Bundle"
linkTitle: "Generating a Support Bundle"
weight: 30
description: >
    Using the Support Bundle with your EKS Anywhere Cluster
aliases:
   - /docs/tasks/troubleshoot/_supportbundle
---

This guide covers the use of the EKS Anywhere Support Bundle for troubleshooting and support.
This allows you to gather cluster information, save it to your administrative machine, and perform analysis of the results.

EKS Anywhere leverages [troubleshoot.sh](https://troubleshoot.sh/) to [collect](https://troubleshoot.sh/docs/collect/) and [analyze](https://troubleshoot.sh/docs/analyze/) kubernetes cluster logs, 
cluster resource information, and other relevant debugging information. 

EKS Anywhere has two Support Bundle commands:

`eksctl anywhere generate support-bundle` will execute a support bundle on your cluster, 
collecting relevant information, archiving it locally, and performing analysis of the results.

`eksctl anywhere generate support-bundle-config` will generate a Support Bundle config yaml file for you to customize.

Do not add personally identifiable information (PII) or other confidential or sensitive information to your support bundle.
If you provide the support bundle to get support from AWS, it will be accessible to other AWS services, including AWS Support.

### Collecting a Support Bundle and running analyzers
```
eksctl anywhere generate support-bundle
```

`generate support-bundle` will allow you to quickly collect relevant logs and cluster resources and save them locally in an archive file.
This archive can then be used to aid in further troubleshooting and debugging.

If you provide a cluster configuration file containing your cluster spec using the `-f` flag,
`generate support-bundle` will customize the auto-generated support bundle collectors and analyzers 
to match the state of your cluster.

If you provide a support bundle configuration file using the `--bundle-config` flag, 
for example one generated with `generate support-bundle-config`, 
`generate support-bundle` will use the provided configuration when collecting information from your cluster and analyzing the results.

```
Flags:
      --bundle-config string   Bundle Config file to use when generating support bundle
  -f, --filename string        Filename that contains EKS-A cluster configuration
  -h, --help                   help for support-bundle
      --since string           Collect pod logs in the latest duration like 5s, 2m, or 3h.
      --since-time string      Collect pod logs after a specific datetime(RFC3339) like 2021-06-28T15:04:05Z
  -w, --w-config string        Kubeconfig file to use when creating support bundle for a workload cluster
```

### Collecting and analyzing a bundle
You only need to run a single command to generate a support bundle, collect information and analyze the output:
`eksctl anywhere generate support-bundle -f myCluster.yaml`

This command will collect the information from your cluster
and run an analysis of the collected information.

The collected information will be saved to your local disk in an archive which can be used for 
debugging and obtaining additional in-depth support.

The analysis will be printed to your console.

#### Collect phase:
```
$ ./bin/eksctl anywhere generate support-bundle -f ./testcluster100.yaml
 Collecting support bundle cluster-info
 Collecting support bundle cluster-resources
 Collecting support bundle secret
 Collecting support bundle logs
 Analyzing support bundle
```

#### Analysis phase:
```
 Analyze Results
------------
Check PASS
Title: gitopsconfigs.anywhere.eks.amazonaws.com
Message: gitopsconfigs.anywhere.eks.amazonaws.com is present on the cluster

------------
Check PASS
Title: vspheredatacenterconfigs.anywhere.eks.amazonaws.com
Message: vspheredatacenterconfigs.anywhere.eks.amazonaws.com is present on the cluster

------------
Check PASS
Title: vspheremachineconfigs.anywhere.eks.amazonaws.com
Message: vspheremachineconfigs.anywhere.eks.amazonaws.com is present on the cluster

------------
Check PASS
Title: capv-controller-manager Status
Message: capv-controller-manager is running.

------------
Check PASS
Title: capv-controller-manager Status
Message: capv-controller-manager is running.

------------
Check PASS
Title: coredns Status
Message: coredns is running.

------------
Check PASS
Title: cert-manager-webhook Status
Message: cert-manager-webhook is running.

------------
Check PASS
Title: cert-manager-cainjector Status
Message: cert-manager-cainjector is running.

------------
Check PASS
Title: cert-manager Status
Message: cert-manager is running.

------------
Check PASS
Title: capi-kubeadm-control-plane-controller-manager Status
Message: capi-kubeadm-control-plane-controller-manager is running.

------------
Check PASS
Title: capi-kubeadm-bootstrap-controller-manager Status
Message: capi-kubeadm-bootstrap-controller-manager is running.

------------
Check PASS
Title: capi-controller-manager Status
Message: capi-controller-manager is running.

------------
Check PASS
Title: capi-controller-manager Status
Message: capi-controller-manager is running.

------------
Check PASS
Title: capi-kubeadm-control-plane-controller-manager Status
Message: capi-kubeadm-control-plane-controller-manager is running.

------------
Check PASS
Title: capi-kubeadm-control-plane-controller-manager Status
Message: capi-kubeadm-control-plane-controller-manager is running.

------------
Check PASS
Title: capi-kubeadm-bootstrap-controller-manager Status
Message: capi-kubeadm-bootstrap-controller-manager is running.

------------
Check PASS
Title: clusters.anywhere.eks.amazonaws.com
Message: clusters.anywhere.eks.amazonaws.com is present on the cluster

------------
Check PASS
Title: bundles.anywhere.eks.amazonaws.com
Message: bundles.anywhere.eks.amazonaws.com is present on the cluster

------------
```

#### Archive phase:
``` 
a support bundle has been created in the current directory:	{"path": "support-bundle-2021-09-02T19_29_41.tar.gz"}
```

### Generating a custom Support Bundle configuration for your EKS Anywhere Cluster
EKS Anywhere will automatically generate a support bundle based on your cluster configuration;
however, if you'd like to customize the support bundle to collect specific information,
you can generate your own support bundle configuration yaml for EKS Anywhere to run on your cluster.

`eksctl anywhere generate support-bundle-config` will generate a default support bundle configuration and print it  as yaml.

`eksctl anywhere generate support-bundle-config -f myCluster.yaml` will generate a support bundle configuration customized to your cluster and print it as yaml.

To run a customized support bundle configuration yaml file on your cluster,
save this output to a file and run the command `eksctl anywhere generate support-bundle` using the flag `--bundle-config`.

```
eksctl anywhere generate support-bundle-config
Flags:
  -f, --filename string   Filename that contains EKS-A cluster configuration
  -h, --help              help for support-bundle-config
```