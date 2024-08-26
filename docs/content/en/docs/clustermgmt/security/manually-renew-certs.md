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

Certificates for external etcd and control plane nodes expire after 1 year in EKS Anywhere. EKS Anywhere automatically rotates these certificates when new machines are rolled out in the cluster. New machines are rolled out during cluster lifecycle operations such as `upgrade`. If you upgrade your cluster at least once a year, as recommended, manual rotation of cluster certificates will not be necessary.

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

Commands below can be used for quickly checking your certificates expiring date:

```bash
# The expiry time of api-server certificate on you cp node
echo | openssl s_client -connect ${CONTROL_PLANE_IP}:6443 2>/dev/null | openssl x509 -noout -dates

# The expiry time of certificate used by your external etcd server, if you configured one
echo | openssl s_client -connect ${EXTERNAL_ETCD_IP}:2379 2>/dev/null | openssl x509 -noout -dates
```

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

- If the above command fails due to multiple etcd containers existing, then navigate to `/var/log/containers/etcd` and confirm which container was running during the issue timeframe (this container would be the 'stale' container). Delete this older etcd once you have renewed the certs and the new etcd container will be able to enter a functioning state. If you don’t do this, the two etcd containers will stay indefinitely and the etcd will not recover.

3. Repeat the above steps for all etcd nodes.

4. Save the `apiserver-etcd-client` `crt` and `key` file as a Secret from one of the etcd nodes, so the `key` can be picked up by new control plane nodes. You will also need them when renewing the certificates on control plane nodes. See the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/#edit-secret) for details on editing Secrets.
```bash
kubectl edit secret ${cluster-name}-apiserver-etcd-client -n eksa-system
```

{{% alert title="Note" color="primary" %}}
On Bottlerocket control plane nodes, the `certificate` filename of `apiserver-etcd-client` is `server-etcd.client.crt` instead of `apiserver-etcd-client.crt`.
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

3. If you have external etcd nodes, manually replace the `apiserver-etcd-client.crt` and `apiserver-etcd-client.key` file in `/etc/kubernetes/pki` (or `/var/lib/kubeadm/pki` in Bottlerocket) folder with the files you saved from any etcd node.

4. Restart static control plane pods.

   - **For Ubuntu and RHEL**: temporarily move all manifest files from `/etc/kubernetes/manifests/` and wait for 20 seconds, then move the manifests back to this file location.

   - **For Bottlerocket**: re-enable the static pods:
   ```
   apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=false 
   apiclient get | apiclient exec admin jq -r '.settings.kubernetes["static-pods"] | keys[]' | xargs -n 1 -I {} apiclient set settings.kubernetes.static-pods.{}.enabled=true
   ```

   You can verify Pods restarting by running `kubectl` from your Admin machine.

5. Repeat the above steps for all control plane nodes.

You can similarly use the above steps to rotate a single certificate instead of all certificates.

### Kubelet
If `kubeadm certs check-expiration` is happy, but kubectl commands against the cluster fail with `x509: certificate has expired or is not yet valid`, then it's likely that the kubelet certs did not rotate. To rotate them, SSH back into one of the control plane nodes and do the following.

```
# backup certs
cd /var/lib/kubelet
cp -r pki pki.bak
rm pki/*

systemctl restart kubelet
```

{{% alert title="Note" color="primary" %}}
When the control plane endpoint is unavailable because the API server pod is not running, the kubelet service may fail to start all static pods in the container runtime. Its logs may contain `failed to connect to apiserver`.

If this occurs, update `kubelet-client-current.pem` by running the following commands:

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
cat /var/lib/kubeadm/admin.conf | grep client-certificate-data: | sed 's/^.*: //' | base64 -d > /var/lib/kubelet/pki/kubelet-client-current.pem

cat /var/lib/kubeadm/admin.conf | grep client-key-data: | sed 's/^.*: //' | base64 -d >> /var/lib/kubelet/pki/kubelet-client-current.pem

systemctl restart kubelet

{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
cat /var/lib/kubeadm/admin.conf | grep client-certificate-data: | apiclient exec admin sed 's/^.*: //' | base64 -d > /var/lib/kubelet/pki/kubelet-client-current.pem

cat /var/lib/kubeadm/admin.conf | grep client-key-data: | apiclient exec admin sed 's/^.*: //' | base64 -d >> /var/lib/kubelet/pki/kubelet-client-current.pem

systemctl restart kubelet

{{< /tab >}}
{{< /tabpane >}}
{{% /alert %}}

#### Worker nodes
If worker nodes are in `Not Ready` state and the kubelet fails to bootstrap then it's likely that the kubelet client-cert `kubelet-client-current.pem` did not automatically rotate. If this rotation process fails you might see errors such as `x509: certificate has expired or is not yet valid` in kube-apiserver logs. To fix the issue, do the following:

1. Backup and delete `/etc/kubernetes/kubelet.conf` (ignore this file for BottleRocket) and `/var/lib/kubelet/pki/kubelet-client*` from the failed node.

2. From a working control plane node in the cluster that has /etc/kubernetes/pki/ca.key execute `kubeadm kubeconfig user --org system:nodes --client-name system:node:$NODE > kubelet.conf`. `$NODE` must be set to the name of the existing failed node in the cluster.  Modify the resulted kubelet.conf manually to adjust the cluster name and server endpoint, or pass `kubeconfig user --config` (modifying `kubelet.conf` file can be ignored for BottleRocket).

3. For Ubuntu or RHEL nodes, Copy this resulted `kubelet.conf` to `/etc/kubernetes/kubelet.conf` on the failed node. Restart the kubelet (`systemctl restart kubelet`) on the failed node and wait for `/var/lib/kubelet/pki/kubelet-client-current.pem` to be recreated. Manually edit the `kubelet.conf` to point to the rotated kubelet client certificates  by replacing client-certificate-data and client-key-data with `/var/lib/kubelet/pki/kubelet-client-current.pem` and `/var/lib/kubelet/pki/kubelet-client-current.pem`. For BottleRocket, manually copy over the base64 decoded values of `client-certificate-data` and `client-key-data` into the `kubelet-client-current.pem` on worker node. 

{{< tabpane >}}
{{< tab header="Ubuntu or RHEL" lang="bash" >}}
kubeadm kubeconfig user --org system:nodes --client-name system:node:$NODE > kubelet.conf (from control plane node with renewed `/etc/kubernetes/pki/ca.key`)
cp kubelet.conf /etc/kubernetes/kubelet.conf (on failed worker node)

{{< /tab >}}
{{< tab header="Bottlerocket" lang="bash" >}}
# From control plane node with renewed certs
# you would be in the admin container when you ssh to the Bottlerocket machine
# open root shell
sudo sheltie

# pull the image
IMAGE_ID=$(apiclient get | apiclient exec admin jq -r '.settings["host-containers"]["kubeadm-bootstrap"].source')
ctr image pull ${IMAGE_ID}

# set NODE value to the failed worker node name.
ctr run \
--mount type=bind,src=/var/lib/kubeadm,dst=/var/lib/kubeadm,options=rbind:rw \
--mount type=bind,src=/var/lib/kubeadm,dst=/etc/kubernetes,options=rbind:rw \
--rm \
${IMAGE_ID} tmp-cert-renew \
/opt/bin/kubeadm kubeconfig user --org system:nodes --client-name system:node:$NODE 

# from the stdout base64 decode `client-certificate-data` and `client-key-data`
# copy client-cert to kubelet-client-current.pem on worker node
echo -n `<base64 decoded client-certificate-data value>` > kubelet-client-current.pem

# append client key to kubelet-client-current.pem on worker node
echo -n `<base64 decoded client-key-data value>` >> kubelet-client-current.pem

{{< /tab >}}
{{< /tabpane >}}

4. Restart the kubelet. Make sure the node becomes `Ready`.

See the [Kubernetes documentation](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/troubleshooting-kubeadm/#kubelet-client-cert) for more details on manually updating kubelet client certificate.

### Post Renewal
Once all the certificates are valid, verify the kcp object on the affected cluster(s) is not paused by running `kubectl describe kcp -n eksa-system | grep cluster.x-k8s.io/paused`. If it is paused, then this usually indicates an issue with the etcd cluster. Check the logs for pods under the `etcdadm-controller-system` namespace for any errors. 
If the logs indicate an issue with the etcd endpoints, then you need to update `spec.clusterConfiguration.etcd.endpoints` in the cluster's `kubeadmconfig` resource: `kubectl edit kcp -n eksa-system`

Example:
```
etcd:
   external:
     caFile: /var/lib/kubeadm/pki/etcd/ca.crt
      certFile: /var/lib/kubeadm/pki/server-etcd-client.crt
      endpoints:
      - https://xxx.xxx.xxx.xxx:2379
      - https://xxx.xxx.xxx.xxx:2379
      - https://xxx.xxx.xxx.xxx:2379
```
