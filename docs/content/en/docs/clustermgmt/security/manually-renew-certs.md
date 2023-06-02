---
title: "Certificate rotation"
linkTitle: "Certificate rotation"
weight: 31
aliases:
    /docs/tasks/cluster/manually-renew-certs/
date: 2021-11-04
description: >
  How to renew certificates in etcd and control plane machines
---

Certificates in external etcd and control plane nodes expire after 1 year. The following table shows all the certificate files:


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

You can renew all certificates in the table except `ca` by following the steps given below. Note that the commands used in Bottlerocket nodes are more sophisticated.

### External etcd node

If the cluster has external etcd, you need to renew the certificates in etcd nodes first.

{{% alert title="Note" color="primary" %}}
You can find out whether any external etcd node exists by running the following command:
```
kubectl get etcdadmcluster -A
```
{{% /alert %}}

1. ssh into the etcd node and run the following commands:

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
# backup certs
cd /etc/etcd
sudo cp -r pki pki.bak
sudo rm pki/*
sudo cp pki.bak/ca.* pki

# run certificates join phase to regenerate the deleted certificates
sudo etcdadm join phase certificates http://eks-a-etcd-dumb-url

# etcd will auto pick the new certificates
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

# etcd will auto pick the new certificates
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

3. Save api-server-etcd-client crt and key file as secret from one of the etcd nodes, so the key can be picked up by new control plane node. You will also need them when renewing certs in CP node.
```
# refer https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/#edit-secret
kubectl edit secret ${cluster-name}-api-server-etcd-client -n eksa-system
```

{{% alert title="Note" color="primary" %}}
In bottlerocket, the key of api-server-etcd-client is "server-etcd.client.crt" instead of "api-server-etcd-client.crt".
{{% /alert %}}

### Control plane node
When there is no external etcd nodes, you only need to renew the certificates in control plane nodes, as etcd certificates are managed by kubeadm. 

1. ssh into the control plane node and run the following commands:

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

2. Verify your certs have been renewed.

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

3. If you have external etcd node, manually replace the api-server-etcd-client.crt and api-server-etcd-client.key file in /etc/kubernetes/pki (or /var/lib/kubeadm/pki in Bottlerocket) folder with the one you saved from any etcd node.

4. Restart static control plane pods.

   - For non-bottlerocket OS: temporarily move out all manifest files from /etc/kubernetes/manifests/ and wait for 20 second. Then move back Move all manifests back.

   - For Bottlerocket: re-enable the static pods:
   ```
   apiclient get | jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false 
   apiclient get | jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true`
   ```

   You can verify pods restarting by running kubectl in your admin machine.

5. Repeat the above steps for all control plane nodes.

It's worth noting that you can use the above steps to renew a single certificate with small modification.
