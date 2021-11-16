---
title: "VSphere Storage"
linkTitle: "VSphere Storage"
weight: 1
date: 2021-11-11
description: >  
---

This chapter will guide you through setting up storage for workloads running on production/VSphere EKS Anywhere environments.

{{% alert title="Notes" color="primary" %}}
This guide is only applicable for vSphere based environments.
{{% /alert %}}

EKS Anywhere supports adding any compatible Container Storage Interface (CSI). EKS Anywhere with vSphere provider by default installs [vSphere CSI Driver](https://vsphere-csi-driver.sigs.k8s.io/) components in the workload cluster with a default storage class. For more information on supported features refer [vSphere CSI Driver - Supported Features Matrix](https://vsphere-csi-driver.sigs.k8s.io/supported_features_matrix.html) and for more information on CSI refer [Kubernetes Container Storage Interface (CSI) Documentation](https://kubernetes-csi.github.io/docs/introduction.html)


### Step by step guide

1. **Verify that all CSI components are up and running**

    After you have populated your `eks-cluster.yaml` and set your vSphere credentials as environment variables, you will be ready to create a cluster:

    ```bash
    kubectl get pods -n kube-system -l app=vsphere-csi-node
    ```

    **Output**

    ```bash
    vsphere-csi-controller-68c5bb89d8-r4qkb               5/5     Running
    vsphere-csi-node-29kbd                                3/3     Running
    vsphere-csi-node-f522j                                3/3     Running
    vsphere-csi-node-gbnvs                                3/3     Running
    vsphere-csi-node-nnggv                                3/3     Running
    vsphere-csi-node-ztwm8                                3/3     Running
    ```

2. **Create a StorageClass**

   {{% alert title="Notes" color="primary" %}}
   A Storage class named standard is created by default and annotated as default StorageClass. This will serve all dynamic PersitentVolume creation requests. The steps below walk you through creation of a StorageClass and using it to create a PV and use it as a volume in a sample deployment.
   {{% /alert %}}

   ```yaml
   kubectl apply -f - <<EOF
   ---
   kind: StorageClass
   apiVersion: storage.k8s.io/v1
   metadata:
     name: example-vanilla-rwo-filesystem-sc
   provisioner: csi.vsphere.vmware.com
   parameters:
     storagepolicyname: "vSAN Default Storage Policy"
     csi.storage.k8s.io/fstype: "ext4"
   EOF
   ```

   Verification of StorageClass creation:

   ```bash
   kubectl get storageclass
   ```

   **Output**

   ```bash
   NAME                                          PROVISIONER              RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION
   example-vanilla-rwo-filesystem-sc             csi.vsphere.vmware.com   Delete          Immediate           true
   standard (default)                            csi.vsphere.vmware.com   Delete          Immediate           false
   ```

3. **Create a PersistentVolumeClaim**

   ```yaml
   kubectl apply -f - <<EOF
   ---
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: example-vanilla-rwo-pvc
     labels:
       appname: eksanywhere-sample
   spec:
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 5Gi
     storageClassName: example-vanilla-rwo-filesystem-sc
   EOF
   ```

   Verification of PersistentVolumeClaim (PVC) creation:

   ```bash
   kubectl get persistentvolumeclaims -o wide
   ```

   **Output**

   ```bash
   NAME                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS                        AGE
   example-vanilla-rwo-pvc   Bound    pvc-51e643b0-678e-4a75-b3b0-89c6e370489e   5Gi        RWO            example-vanilla-rwo-filesystem-sc   Filesystem
   ```

   Describe the PersistentVolumeClaim to verify the dynamic PersistentVolume creation

   ```bash
   kubectl describe pvc example-vanilla-rwo-pvc
   ```

   **Output**

   ```bash
   Name:          example-vanilla-rwo-pvc
   Namespace:     default
   StorageClass:  example-vanilla-rwo-filesystem-sc
   Status:        Bound
   Volume:        pvc-51e643b0-678e-4a75-b3b0-89c6e370489e
   Labels:        appname=eksanywhere-sample
   Annotations:   pv.kubernetes.io/bind-completed: yes
                  pv.kubernetes.io/bound-by-controller: yes
                  volume.beta.kubernetes.io/storage-provisioner: csi.vsphere.vmware.com
   Finalizers:    [kubernetes.io/pvc-protection]
   Capacity:      5Gi
   Access Modes:  RWO
   VolumeMode:    Filesystem
   Used By:       <none>
   Events:
     Type    Reason                 Age   From                                                                                                 Message
     ----    ------                 ----  ----                                                                                                 -------
     Normal  Provisioning           2m4s  csi.vsphere.vmware.com_vsphere-csi-controller-68c5bb89d8-xl9kr_15c06902-1bb5-404c-a081-3632d7e03bb7  External provisioner is provisioning volume for claim "default/example-vanilla-rwo-pvc"
     Normal  ExternalProvisioning   2m4s  persistentvolume-controller                                                                          waiting for a volume to be created, either by external provisioner "csi.vsphere.vmware.com" or manually created by system administrator
     Normal  ProvisioningSucceeded  2m3s  csi.vsphere.vmware.com_vsphere-csi-controller-68c5bb89d8-xl9kr_15c06902-1bb5-404c-a081-3632d7e03bb7  Successfully provisioned volume pvc-51e643b0-678e-4a75-b3b0-89c6e370489e
   ```

   Verification of dynamic PersistentVolume (PV) creation:

   ```bash
   kubectl get persistentvolumes -o wide
   ```

   **Output**

   ```bash
   NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                             STORAGECLASS                        REASON   AGE
   pvc-51e643b0-678e-4a75-b3b0-89c6e370489e   5Gi        RWO            Delete           Bound    default/example-vanilla-rwo-pvc   example-vanilla-rwo-filesystem-sc   Filesystem
   ```

   Describe the PersistentVolume to verify the properties like Size, VolumeMode, etc.

   ```bash
   kubectl describe pv pvc-51e643b0-678e-4a75-b3b0-89c6e370489e
   ```

   **Output**

   ```bash
   Name:            pvc-51e643b0-678e-4a75-b3b0-89c6e370489e
   Labels:          <none>
   Annotations:     pv.kubernetes.io/provisioned-by: csi.vsphere.vmware.com
   Finalizers:      [kubernetes.io/pv-protection]
   StorageClass:    example-vanilla-rwo-filesystem-sc
   Status:          Bound
   Claim:           default/example-vanilla-rwo-pvc
   Reclaim Policy:  Delete
   Access Modes:    RWO
   VolumeMode:      Filesystem
   Capacity:        5Gi
   Node Affinity:   <none>
   Message:
   Source:
       Type:              CSI (a Container Storage Interface (CSI) volume source)
       Driver:            csi.vsphere.vmware.com
       FSType:            ext4
       VolumeHandle:      f50546c3-dddb-41be-877a-8cedeae1dd4b
       ReadOnly:          false
       VolumeAttributes:      storage.kubernetes.io/csiProvisionerIdentity=1633392647145-8081-csi.vsphere.vmware.com
                              type=vSphere CNS Block Volume
   Events:                <none> 
   ```

   You can also check if the above PV is created under cluster > Monitor Tab > Cloud Native Storage in vSphere client. The PV is labelled with the metadata.lables provided in the PVC configuration.

   ![PVC - Container Volumes - Kubernetes Objects](../../images/vSphere-csi-storage-1.png)
   ![PVC - Physical Placement](../../images/vSphere-csi-storage-2.png)

4. **Create a sample Deployment with volume using the PVC created above**

   ```yaml
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
           containers:
           - name: hello
             image: public.ecr.aws/aws-containers/hello-eks-anywhere:latest
             volumeMounts:
             - mountPath: "/external-csi"
               name: data
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
         volumes:
            - name: data
              persistentVolumeClaim:
                claimName: example-vanilla-rwo-pvc
   EOF
   ```

5. **Verify if the pod has the volume attached**

   ```bash
   kubectl get pods
   ```

   **Output**

   ```bash
   NAME                          READY   STATUS    RESTARTS   AGE
   hello-eks-a-bf9c9cbb5-zphf4   1/1     Running   0          55s
   ```

   Check if the assocaited volume is mounted as a Filesystem in the pod

   ```bash
   kubectl exec -it hello-eks-a-bf9c9cbb5-zphf4 -- df -h
   ```

   {{% alert title="Notes" color="primary" %}}
   The **df** command output from the pod should show the Filesystem path, Mount (external-csi) and Size
   {{% /alert %}}

   **Output**

   ```bash
   Filesystem                Size      Used Available Use% Mounted on
   overlay                  19.6G      4.9G     13.7G  26% /
   tmpfs                    64.0M         0     64.0M   0% /dev
   tmpfs                     3.9G         0      3.9G   0% /sys/fs/cgroup
   <strong>/dev/sdb                  4.8G     20.0M      4.5G   0% /external-csi</strong>
   /dev/sda1                19.6G      4.9G     13.7G  26% /etc/hosts
   /dev/sda1                19.6G      4.9G     13.7G  26% /dev/termination-log
   /dev/sda1                19.6G      4.9G     13.7G  26% /etc/hostname
   /dev/sda1                19.6G      4.9G     13.7G  26% /etc/resolv.conf
   shm                      64.0M         0     64.0M   0% /dev/shm
   tmpfs                     3.9G     12.0K      3.9G   0% /run/secrets/kubernetes.io/serviceaccount
   tmpfs                     3.9G         0      3.9G   0% /proc/acpi
   tmpfs                    64.0M         0     64.0M   0% /proc/kcore
   tmpfs                    64.0M         0     64.0M   0% /proc/keys
   tmpfs                    64.0M         0     64.0M   0% /proc/timer_list
   tmpfs                    64.0M         0     64.0M   0% /proc/sched_debug
   tmpfs                     3.9G         0      3.9G   0% /proc/scsi
   tmpfs                     3.9G         0      3.9G   0% /sys/firmware
   ```

   {{% alert title="Notes" color="primary" %}}
   Refer [vsphere-csi-driver-examples](https://github.com/kubernetes-sigs/vsphere-csi-driver/tree/master/example) for information on creating other advanced storage constructs like Block Volumes, Zone based storage, Volumes with Read Write Many access, etc.
   {{% /alert %}}