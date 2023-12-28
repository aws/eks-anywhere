---
title: "Using NVIDIA GPU Operator with EKS Anywhere"
linkTitle: "Using NVIDIA GPU Operator"
weight: 40
date: 2023-12-20
description: >
  How to use the NVIDIA GPU Operator with EKS Anywhere on bare metal
---

The [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html) allows GPUs to be exposed to applications in Kubernetes clusters much like CPUs. Instead of provisioning a special OS image for GPU nodes with the required drivers and dependencies, a standard OS image can be used for both CPU and GPU nodes. The NVIDIA GPU Operator can be used to provision the required software components for GPUs such as the NVIDIA drivers, Kubernetes device plugin for GPUs, and the NVIDIA Container Toolkit. See the [licensing section](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/index.html#licenses-and-contributing) of the NVIDIA GPU Operator documentation for information on the NVIDIA End User License Agreements.

In the example on this page, a single-node EKS Anywhere cluster on bare metal is used with an Ubuntu 20.04 image produced from image-builder without modifications and Kubernetes version 1.27.

### 1. Configure an EKS Anywhere cluster spec and hardware inventory

See the [Configure for Bare Metal]({{< relref "../getting-started/baremetal/bare-spec">}}) page and the [Prepare hardware inventory]({{< relref "../getting-started/baremetal/bare-preparation/#prepare-hardware-inventory">}})  page for details. If you use cluster spec sample below is used, your hardware inventory definition must have `type=cp` for the `labels` field in the hardware inventory for your server.

<details>
  <summary>Expand for a sample cluster spec</summary>
  <br /> 
  {{% content "gpu-sample-cluster.md" %}}
</details>


### 2. Create a single-node EKS Anywhere cluster
- Replace `hardware.csv` with the name of your hardware inventory file
- Replace `cluster.yaml` with the name of your cluster spec file

```bash
eksctl anywhere create cluster --hardware hardware.csv -f cluster.yaml
```

<details>
  <summary>Expand for sample output</summary>
  <br /> 
  {{% content "gpu-create-cluster-output.md" %}}
</details>

### 3. Install Helm

```bash
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 \ 
  && chmod 700 get_helm.sh \ 
  && ./get_helm.sh
```

### 4. Add NVIDIA Helm Repository

```bash
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia \ 
  && helm repo update
```

### 5. Configure kubectl to use EKS Anywhere cluster
- Replace `<path-to-cluster-folder>` with the directory location where your EKS Anywhere cluster folder is located. This is typically in the same directory in which the `eksctl anywhere` command was run.
- Replace `<cluster-name>` with the name of your cluster.

```bash
KUBECONFIG=<path-to-cluster-folder>/<cluster-name>-eks-a-cluster.kubeconfig
```

### 6. Install NVIDIA GPU Operator
  
```bash
helm install --wait --generate-name \ 
  -n gpu-operator --create-namespace \ 
  nvidia/gpu-operator
```

### 7. Validate the operator was installed successfully

```bash
kubectl get pods -n gpu-operator
```
```
NAME                                                              READY   STATUS      RESTARTS   AGE
gpu-feature-discovery-6djnw                                       1/1     Running     0          5m25s
gpu-operator-1691443998-node-feature-discovery-master-55cfkzbl5   1/1     Running     0          5m55s
gpu-operator-1691443998-node-feature-discovery-worker-dw8m7       1/1     Running     0          5m55s
gpu-operator-59f96d7646-7zcn4                                     1/1     Running     0          5m55s
nvidia-container-toolkit-daemonset-c2mdf                          1/1     Running     0          5m25s
nvidia-cuda-validator-6m4kg                                       0/1     Completed   0          3m41s
nvidia-dcgm-exporter-jw5wz                                        1/1     Running     0          5m25s
nvidia-device-plugin-daemonset-8vjrn                              1/1     Running     0          5m25s
nvidia-driver-daemonset-6hklg                                     1/1     Running     0          5m36s
nvidia-operator-validator-2pvzx                                   1/1     Running     0          5m25s
```

### 8. Validate GPU specs

```bash
kubectl get node -o json | jq '.items[].metadata.labels'
```
```
{
... 
  "nvidia.com/cuda.driver.major": "535",
  "nvidia.com/cuda.driver.minor": "86",
  "nvidia.com/cuda.driver.rev": "10",
  "nvidia.com/cuda.runtime.major": "12",
  "nvidia.com/cuda.runtime.minor": "2",
  "nvidia.com/gfd.timestamp": "1691444179",
  "nvidia.com/gpu-driver-upgrade-state": "upgrade-done",
  "nvidia.com/gpu.compute.major": "7",
  "nvidia.com/gpu.compute.minor": "5",
  "nvidia.com/gpu.count": "2",
  "nvidia.com/gpu.deploy.container-toolkit": "true",
  "nvidia.com/gpu.deploy.dcgm": "true",
  "nvidia.com/gpu.deploy.dcgm-exporter": "true",
  "nvidia.com/gpu.deploy.device-plugin": "true",
  "nvidia.com/gpu.deploy.driver": "true",
  "nvidia.com/gpu.deploy.gpu-feature-discovery": "true",
  "nvidia.com/gpu.deploy.node-status-exporter": "true",
  "nvidia.com/gpu.deploy.nvsm": "",
  "nvidia.com/gpu.deploy.operator-validator": "true",
  "nvidia.com/gpu.family": "turing",
  "nvidia.com/gpu.machine": "PowerEdge-R7525",
  "nvidia.com/gpu.memory": "15360",
  "nvidia.com/gpu.present": "true",
  "nvidia.com/gpu.product": "Tesla-T4",
  "nvidia.com/gpu.replicas": "1",
  "nvidia.com/mig.capable": "false",
  "nvidia.com/mig.strategy": "single"
}
```

### 9. Run Sample App

Create a `gpu-pod.yaml` file with the following and apply it to the cluster

```yaml
apiVersion: v1 
kind: Pod 
metadata: 
  name: gpu-pod 
spec: 
  restartPolicy: Never 
  containers: 
   - name: cuda-container 
     image: nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda10.2 
     resources: 
       limits: 
         nvidia.com/gpu: 1 # requesting 1 GPU 
  tolerations: 
    - key: nvidia.com/gpu operator: Exists 
      effect: NoSchedule
```

```bash
kubectl apply -f gpu-pod.yaml
```

### 10. Confirm Sample App Succeeded

```bash
kubectl logs gpu-pod
```
```
[Vector addition of 50000 elements]
Copy input data from the host memory to the CUDA device
CUDA kernel launch with 196 blocks of 256 threads
Copy output data from the CUDA device to the host memory
Test PASSED
Done
```