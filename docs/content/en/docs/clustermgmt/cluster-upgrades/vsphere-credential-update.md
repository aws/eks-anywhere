---
title: "Update vSphere credentials"
linkTitle: "Update vSphere credentials"
weight: 22
date: 2017-01-05
description: >
  How to update vSphere credentials used by EKS Anywhere
---

EKS Anywhere does not currently support updating the vSphere credentials used by EKS Anywhere when upgrading clusters with the `eksctl anywhere upgrade` command. 

It is recommended to use the script maintained with EKS Anywhere to update your vSphere credentials, which automates the steps covered in the [Update vSphere credentials manually]({{< relref "./vsphere-credential-update/#update-vsphere-credentials-manually" >}}) section.

### Update vSphere credentials with script

You can update all vSphere credentials in related Secret objects used by EKS Anywhere with the [vSphere credential update script](https://github.com/aws/eks-anywhere/blob/main/scripts/update_vsphere_credential.sh) in EKS Anywhere GitHub repository. The following steps should be run from your admin machine or the local machine where you host the kubeconfig file for your EKS Anywhere management or standalone cluster.

1. Set environment variables on your local machine 

- Set the `KUBECONFIG` environment variable on your local machine to the kubeconfig file for your EKS Anywhere management or standalone cluster. For example `mgmt/mgmt-eks-a-cluster.kubeconfig`.
- Set the `EKSA_VSPHERE_USERNAME` and `EKSA_VSPHERE_PASSWORD` environment variables on your local machine with the new vSphere credentials.

```bash
export KUBECONFIG='<your-kubeconfig-file>'
export EKSA_VSPHERE_USERNAME='<your-vsphere-username>'
export EKSA_VSPHERE_PASSWORD='<your-vsphere-password>'
```

2. Download the script to your local machine

```bash
curl -OL https://raw.githubusercontent.com/aws/eks-anywhere/refs/heads/main/scripts/update_vsphere_credential.sh
```

3. Run the script from your local machine

- Replace `CLUSTER_NAME` with the name of your EKS Anywhere cluster and `VSPHERE_SERVER_NAME` with the name of the vSphere server.

```bash
./update_vsphere_credential.sh CLUSTER_NAME VSPHERE_SERVER_NAME
```

>**_NOTE:_** If you are using the vSphere CSI in your cluster, you must manually update the vSphere password in the `{CLUSTER_NAME}-csi-vsphere-config` Secret under the `eksa-system` namespace. If the annotation `kubectl.kubernetes.io/last-applied-configuration` exists on the secret object, update password in the `kubectl.kubernetes.io/last-applied-configuration` field.

### Update vSphere credentials manually

Follow the steps below to manually update the vSphere credentials used by EKS Anywhere.

- Update `EKSA_VSPHERE_PASSWORD` environment variable to the new password and get the base64 encoded string of the password using `echo -n "<YOUR_PASSWORD>" | base64`
- Update the following secrets in your vSphere cluster using `kubectl edit` command:
    - `{CLUSTER_NAME}-vsphere-credentials` under `eksa-system` namespace - Update `password` field under data.
    - `{CLUSTER_NAME}-cloud-provider-vsphere-credentials` under `eksa-system` namespace - Decode the string under data, in the decoded string (which is the template for Secret object `cloud-provider-vsphere-credential` under `kube-system` namespace), update the `{CLUSTER_NAME}.password` with the base64 encoding of new password, then encode the string and update data field with the encoded string.
    - `vsphere-credentials` under `eksa-system` namespace - Update `password`, `passwordCP`, `passwordCSI` field under data.
    - If annotation `kubectl.kubernetes.io/last-applied-configuration` exists on any of the above Secret object, update password in `kubectl.kubernetes.io/last-applied-configuration` field.
    -  `{CLUSTER_NAME}-csi-vsphere-config` under `eksa-system` namespace - If annotation `kubectl.kubernetes.io/last-applied-configuration` exists on the secret object, update password in `kubectl.kubernetes.io/last-applied-configuration` field.