---
title: "Local Storage"
linkTitle: "Local Storage"
weight: 1
date: 2021-11-11
description: >  
---

In this chapter we will see how to set up storage for workloads running on local/dev EKS Anywhere environments.

{{% alert title="Notes" color="primary" %}}
An EKS Anywhere cluster install by default doesn't provision any storage classes. So to enable persistent storage, we need to create a storage class and provision a persistent volume.
{{% /alert %}}

### Step by step guide

  1. Run the following command to create the storage class:

  ```bash
  kubectl apply -f - <<EOF
  ---
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: local-storage  
  provisioner: kubernetes.io/no-provisioner
  allowVolumeExpansion: true
  volumeBindingMode: WaitForFirstConsumer
  EOF
  ```

  **Output**

  ```
  storageclass.storage.k8s.io/local-storage created
  ```

  2. Since we are provisioning a local storage class, we need to create a persistent volume. Run the following command to create the persistent volume:

  ```bash
  kubectl apply -f - <<EOF
  ---
  apiVersion: v1
  kind: PersistentVolume
  metadata:
    name: local-pv
  spec:
    capacity:
      storage: 50Gi
    accessModes:
    - ReadWriteOnce
    persistentVolumeReclaimPolicy: Retain
    storageClassName: local-storage
    local:
      path: /tmp
    nodeAffinity:
      required:
        nodeSelectorTerms:
        - matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
            - <<Node1>>
            - <<Node2>>
  EOF
  ```

{{% alert title="Notes" color="warning" %}}
`<<Node1>>` and `<<Node2>>` values need to be replaced with the actual node names when the cluster is created, which can be retrieved by running `kubectl get nodes` command.
{{% /alert %}}

  **Output**

  ```
  persistentvolume/local-pv created
  ```

  3. Lets create a `PersistentVolumeClaim` to mount the persistent volume to a pod. Run the following command to create the persistent volume claim::

  ```bash
  kubectl apply -f - <<EOF
  ---
  apiVersion: v1
  kind: PersistentVolumeClaim
  metadata:
      name: local-claim
  spec:
      storageClassName: "local-storage"
      accessModes:
          - ReadWriteOnce
      resources:
          requests:
              storage: 1Gi
  EOF
  ```

  **Output**

  ```
  persistentvolumeclaim/local-claim created
  ```

  4. Lets create a new deployment with the persistent volume claim. Run the following command to create the deployment:

  ```bash
  kubectl apply -f - <<EOF
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: hello-eks-a
    labels:
      app: hello-eks-a
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: hello-eks-a
    template:
      metadata:
        labels:
          app: hello-eks-a
      spec:
        volumes:
        - name: mypd
          persistentVolumeClaim:
          claimName: local-claim
        containers:
        - name: hello
          image: public.ecr.aws/aws-containers/hello-eks-anywhere:latest
          volumeMounts:
          - mountPath: "/external"
            name: mypd
          ports:
          - containerPort: 80
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          resources:
            requests:
              memory: "64Mi"
              cpu: "250m"
            limits:
              memory: "128Mi"
              cpu: "500m"
  ---
  apiVersion: v1
  kind: Service
  metadata:
    name: hello-eks-a
  spec:
    type: NodePort
    selector:
      app: hello-eks-a
    ports:
      - port: 80
  EOF
  ```

  5. Run the following command to get the pod status, wait for the pod to get to a running state:

  ```bash
  kubectl get pods
  ```

  **Output**

  ```
  NAME                           READY   STATUS    RESTARTS   AGE
  hello-eks-a-77dcff877b-sngtd   1/1     Running   0          25m
  ```

  6. Let check whether volume is mounted by run the following command:

  ```bash
  kubectl exec -it <<podname>> -- df -h
  ```

{{% alert title="Notes" color="warning" %}}
Replace `podname` with the actual pod name.
{{% /alert %}}

  **Output**

  <pre><code>
  Filesystem                Size      Used Available Use% Mounted on
  overlay                  58.4G     34.6G     20.8G  62% /
  tmpfs                    64.0M         0     64.0M   0% /dev
  tmpfs                     2.9G         0      2.9G   0% /sys/fs/cgroup
  <strong>tmpfs                     2.9G      5.4M      2.9G   0% /external</strong>
  /dev/vda1                58.4G     34.6G     20.8G  62% /etc/hosts
  /dev/vda1                58.4G     34.6G     20.8G  62% /dev/termination-log
  /dev/vda1                58.4G     34.6G     20.8G  62% /etc/hostname
  /dev/vda1                58.4G     34.6G     20.8G  62% /etc/resolv.conf
  shm                      64.0M         0     64.0M   0% /dev/shm
  tmpfs                     2.9G     12.0K      2.9G   0% /run/secrets/kubernetes.io/serviceaccount
  tmpfs                     2.9G         0      2.9G   0% /proc/acpi
  tmpfs                    64.0M         0     64.0M   0% /proc/kcore
  tmpfs                    64.0M         0     64.0M   0% /proc/keys
  tmpfs                    64.0M         0     64.0M   0% /proc/timer_list
  tmpfs                    64.0M         0     64.0M   0% /proc/sched_debug
  tmpfs                     2.9G         0      2.9G   0% /sys/firmware
  </code></pre>

{{% alert title="Notes" color="warning" %}}
The mounted volume is highlighted.
{{% /alert %}}

### Cleanup

Run the following command to clean up:

```bash
kubectl delete deployment hello-eks-a
kubectl delete service hello-eks-a
kubectl delete persistentvolumeclaim local-claim
kubectl delete persistentvolume local-pv
kubectl delete storageclass local-storage
```

**Output**

```
deployment.apps "hello-eks-a" deleted
service "hello-eks-a" deleted
persistentvolumeclaim "local-claim" deleted
persistentvolume "local-pv" deleted
storageclass.storage.k8s.io "local-storage" deleted
```
