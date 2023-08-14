---
title: "Certificate rotation"
linkTitle: "Certificate rotation"
weight: 31
aliases:
    /docs/tasks/cluster/manually-renew-certs/
date: 2021-11-04
description: >
  How to rotate certificates for etcd and control plane nodes
---

Certificates for external etcd and control plane nodes expire after 1 year in EKS Anywhere. EKS Anywhere automatically rotates these certificates when new machines are rolled out in the cluster. New machines are rolled out during cluster lifecycle operations such as `upgrade`. If you upgrade your cluster at least once a year, you do not have to manually rotate the cluster certificates. 

This page shows the process for manually rotating certificates if you have not upgraded your cluster in 1 year.

The following table lists the cluster certificate files:

| etcd node             | control plane node       |
|-----------------------|--------------------------|
| apiserver-etcd-client | apiserver-etcd-client    |
| ca                    | ca                       |
| etcdctl-etcd-client   | front-proxy-ca           |
| peer                  | sa                       |
| server                | etcd/ca.crt              |
|                       | apiserver-kubelet-client |
|                       | apiserver                |
|                       | front-proxy-client       |

You can rotate certificates by following the steps given below. You cannot rotate the `ca` certificate because it is the root certificate. Note that the commands used for Bottlerocket nodes are different than those for Ubuntu and RHEL nodes.

#### External etcd nodes

If your cluster is using external etcd nodes, you need to renew the etcd node certificates first. 

{{% alert title="Note" color="primary" %}}
You can check for external etcd nodes by running the following command:
```bash
kubectl get etcdadmcluster -A
```
{{% /alert %}}

1. SSH into each etcd node and run the following commands. Etcd automatically detects the new certificates and deprecates its old certificates.

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
# backup certs
cd /etc/etcd
sudo cp -r pki pki.bak
sudo rm pki/*
sudo cp pki.bak/ca.* pki

# run certificates join phase to regenerate the deleted certificates
sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url
{{< /tab >}}

{{< tab header="Bottlerocket" lang="bash" >}}
# you would be in the admin container when you ssh to the Bottlerocket machine
# open a root shell
sudo sheltie

# pull the image
IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}

# backup certs
cd /var/lib/etcd
cp -r pki pki.bak
rm pki/*
cp pki.bak/ca.* pki

# recreate certificates
ctr run \
--mount type=bind,src=/var/lib/etcd/pki,dst=/etc/etcd/pki,options=rbind:rw \
--net-host \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/etcdadm join phase certificates http://eks-a-etcd-dumb-url --init-system kubelet
{{< /tab >}}
{{< /tabpane >}}

2. Verify your etcd node is running correctly

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

3. Repeat the above steps for all etcd nodes.

4. Save the `api-server-etcd-client` `crt` and `key` file as a Secret from one of the etcd nodes, so the `key` can be picked up by new control plane nodes. You will also need them when renewing the certificates on control plane nodes. See the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/#edit-secret) for details on editing Secrets.
```bash
kubectl edit secret ${cluster-name}-api-server-etcd-client -n eksa-system
```

{{% alert title="Note" color="primary" %}}
For Bottlerocket nodes, the `key` of `api-server-etcd-client` is `server-etcd.client.crt` instead of `api-server-etcd-client.crt`.
{{% /alert %}}

#### Control plane nodes
When there are no external etcd nodes, you only need to rotate the certificates for control plane nodes, as etcd certificates are managed by `kubeadm` when there are no external etcd nodes. 

1. SSH into each control plane node and run the following commands.

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
sudo kubeadm certs renew all
{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
# you would be in the admin container when you ssh to the Bottlerocket machine
# open root shell
sudo sheltie

# pull the image
IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}

# renew certs
# you may see missing etcd certs error, which is expected if you have external etcd nodes
ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs renew all
{{< /tab >}}
{{< /tabpane >}}

2. Verify the certificates have been rotated.

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
sudo kubeadm certs check-expiration
{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
# you may see missing etcd certs error, which is expected if you have external etcd nodes
ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm certs check-expiration
{{< /tab >}}
{{< /tabpane >}}

3. If you have external etcd nodes, manually replace the `api-server-etcd-client.crt` and `api-server-etcd-client.key` file in `/etc/kubernetes/pki` (or `/var/lib/kubeadm/pki` in Bottlerocket) folder with the files you saved from any etcd node.

4. Restart static control plane pods.

   - **For Ubuntu and RHEL**: temporarily move all manifest files from `/etc/kubernetes/manifests/` and wait for 20 seconds, then move the manifests back to this file location.

   - **For Bottlerocket**: re-enable the static pods:
   ```
   apiclient get | jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false 
   apiclient get | jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true`
   ```

   You can verify Pods restarting by running `kubectl` from your Admin machine.

5. Repeat the above steps for all control plane nodes.

You can similarly use the above steps to rotate a single certificate instead of all certificates.
