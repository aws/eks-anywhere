---
toc_hide: true
---
  ```
  apiVersion: anywhere.eks.amazonaws.com/v1alpha1
  kind: Cluster
  metadata:
    name: gpu-test
  spec:
    clusterNetwork:
      cniConfig:
        cilium: {}
      pods:
        cidrBlocks:
        - 192.168.0.0/16
      services:
        cidrBlocks:
        - 10.96.0.0/12
    controlPlaneConfiguration:
      count: 1
      endpoint:
        host: "<my-cp-ip>"
      machineGroupRef:
        kind: TinkerbellMachineConfig
        name: gpu-test-cp
    datacenterRef:
      kind: TinkerbellDatacenterConfig
      name: gpu-test
    kubernetesVersion: "1.27"
  ---
  apiVersion: anywhere.eks.amazonaws.com/v1alpha1
  kind: TinkerbellDatacenterConfig
  metadata:
    name: gpu-test
  spec:
    tinkerbellIP: "<my-tb-ip>"
    osImageURL: "https://<url-for-image>/ubuntu.gz"
  ---
  apiVersion: anywhere.eks.amazonaws.com/v1alpha1
  kind: TinkerbellMachineConfig
  metadata:
    name: gpu-test-cp
  spec:
    hardwareSelector: {type: "cp"}
    osFamily: ubuntu
    templateRef: {}
  ```