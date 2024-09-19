---
toc_hide: true
---

Currently EKS Anywhere does not support updating vSphere credential when upgrading cluster with `eksctl anywhere upgrade` command. There are two ways to perform a vsphere credentials update:

1. Manually update all vSphere credentials in related Secret objects, follow the steps below:
- Update `EKSA_VSPHERE_PASSWORD` environment variable to the new password and get the base64 encoded string of the password using `echo -n "<YOUR_PASSWORD>" | base64`
- Update the following secrets in your vSphere cluster using `kubectl edit` command:
    - `{CLUSTER_NAME}-vsphere-credentials` under `eksa-system` namespace - Update `password` field under data.
    - `{CLUSTER_NAME}-cloud-provider-vsphere-credentials` under `eksa-system` namespace - Decode the string under data, in the decoded string (which is the template for Secret object `cloud-provider-vsphere-credential` under `kube-system` namespace), update the `{CLUSTER_NAME}.password` with the base64 encoding of new password, then encode the string and update data field with the encoded string.
    - `vsphere-credentials` under `eksa-system` namespace - Update `password`, `passwordCP`, `passwordCSI` field under data.
    - If annotation `kubectl.kubernetes.io/last-applied-configuration` exists on any of the above Secret object, update password in `kubectl.kubernetes.io/last-applied-configuration` field.
    -  `{CLUSTER_NAME}-csi-vsphere-config` under `eksa-system` namespace - If annotation `kubectl.kubernetes.io/last-applied-configuration` exists on the secret object, update password in `kubectl.kubernetes.io/last-applied-configuration` field.

2. Only update Secret `vsphere-credentials` under `eksa-system` namespace then trigger a full EKS-A CAPI cluster upgrade by modifying the cluster spec:
- Update `EKSA_VSPHERE_PASSWORD` environment variable to the new password and get the base64 encoded string of the password using `echo -n "<YOUR_PASSWORD>" | base64`
- Update secret `vsphere-credentials` under `eksa-system` namespace - Update `password`, `passwordCP`, `passwordCSI` field under data and in `kubectl.kubernetes.io/last-applied-configuration` if annotation exists.
- Modify any field in the cluster config file and then run `eksctl anywhere upgrade cluster -f <cluster-config-file>` to trigger a full cluster upgrade. This will automatically apply the new credentials to all related secrets.