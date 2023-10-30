---
title: "Encrypting Confidential Data at Rest"
linkTitle: "Encrypting Confidential Data at Rest"
weight: 10
description: >
  EKS Anywhere cluster specification for encryption of etcd data at-rest 
---

You can configure EKS Anywhere clusters to encrypt confidential API resource data, such as `secrets`, at-rest in etcd using a KMS encryption provider.
EKS Anywhere supports a hybrid model for configuring etcd encryption where cluster admins are responsible for deploying and maintaining
the KMS provider on the cluster and EKS Anywhere handles configuring `kube-apiserver` with the KMS properties.

Because of this model, etcd encryption can only be enabled on **_cluster upgrades_** after the KMS provider has been deployed on the cluster.

{{% alert title="Note" color="warning" %}}
Currently, etcd encryption is only supported for CloudStack and vSphere.
Support for other providers will be added in a future release.
{{% /alert %}}

## Before you begin
Before enabling etcd encryption, make sure you have done the following:
- Learn how encrypting confidential data works in Kubernetes by reading [Encrypting Secret Data at Rest.](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)
- Learn how Kubernetes uses KMS encryption providers to encrypt data by reading [Using a KMS Provider for data encryption.](https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/)
- Learn how you can use AWS KMS keys to encrypt data using [AWS Encryption Provider.](https://github.com/kubernetes-sigs/aws-encryption-provider#aws-encryption-provider)
- Create an EKS Anywhere cluster since encryption can only be enabled on already created clusters.
- Deploy a KMS provider on the cluster that the `kube-apiserver` can use to encrypt/decrypt secrets.<br>
[AWS Encryption Provider](https://github.com/kubernetes-sigs/aws-encryption-provider#aws-encryption-provider) is the recommended KMS provider for EKS Anywhere.

## Example etcd encryption configuration

The following cluster spec enables etcd encryption configuration:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
  namespace: default
spec:
  ...
  etcdEncryption:
  - providers:
    - kms:
        cachesize: 1000
        name: example-kms-config
        socketListenAddress: unix:///var/run/kmsplugin/socket.sock
        timeout: 3s
    resources:
    - secrets
```

## Description of etcd encryption fields

#### `etcdEncryption`
Key used to specify etcd encryption configuration for a cluster. This field is only supported on cluster upgrades.

  * #### `providers`
    Key used to specify which encryption provider to use. Currently, only one provider can be configured.

    * #### `kms`
      Key used to configure [KMS encryption provider.](https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/)

      * ##### `name`
        Key used to set the name of the KMS plugin. This cannot be changed once set.

      * ##### `endpoint`
        Key used to specify the listen address of the gRPC server (KMS plugin). The endpoint is a UNIX domain socket.
        
      * ##### `cachesize`
        Number of data encryption keys (DEKs) to be cached in the clear. 
        When cached, DEKs can be used without another call to the KMS; whereas DEKs that are not cached require a call to the KMS to unwrap.
        If `cachesize` isn't specified, a default of `1000` is used.

      * ##### `timeout`
        How long should kube-apiserver wait for kms-plugin to respond before returning an error. If a timeout isn't specified, a default timeout of `3s` is used.

  * #### `resources`
    Key used to specify a list of resources that should be encrypted using the corresponding encryption provider.
    These can be native Kubernetes resources such as `secrets` and `configmaps` or custom resource definitions such as `clusters.anywhere.eks.amazonaws.com`.

## Example AWS Encryption Provider DaemonSet
Here's a sample AWS encryption provider daemonset configuration. 

{{< alert title="Note" color="warning" >}}
This example doesn't include any configuration for AWS credentials to call the KMS APIs.<br>
You can configure this daemonset with static creds for an IAM user or use [IRSA]({{< relref "irsa" >}}) for a more secure way to authenticate to AWS.
{{< /alert >}}

{{< details "Expand" >}}
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: aws-encryption-provider
  name: aws-encryption-provider
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: aws-encryption-provider
  template:
    metadata:
      labels:
        app: aws-encryption-provider
    spec:
      containers:
      - image: <AWS_ENCRYPTION_PROVIDER_IMAGE>    # Specify the AWS KMS encryption provider image 
        name: aws-encryption-provider
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        command:
        - /aws-encryption-provider
        - --key=<KEY_ARN>                         # Specify the arn of KMS key to be used for encryption/decryption
        - --region=<AWS_REGION>                   # Specify the region in which the KMS key exists
        - --listen=<KMS_SOCKET_LISTEN_ADDRESS>    # Specify a socket listen address for the KMS provider. Example: /var/run/kmsplugin/socket.sock
        ports:
        - containerPort: 8080
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        volumeMounts:
          - mountPath: /var/run/kmsplugin
            name: var-run-kmsplugin
          - mountPath: /root/.aws
            name: aws-credentials
      tolerations:
      - key: "node-role.kubernetes.io/master"
        effect: "NoSchedule"
      - key: "node-role.kubernetes.io/control-plane"
        effect: "NoSchedule"
      volumes:
      - hostPath:
          path: /var/run/kmsplugin
          type: DirectoryOrCreate
        name: var-run-kmsplugin
      - hostPath:
          path: /etc/kubernetes/aws
          type: DirectoryOrCreate
        name: aws-credentials
```
{{< /details >}}
