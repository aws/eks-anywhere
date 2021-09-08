---
title: "Alternative: MetalLB Service-type Load Balancer"
linkTitle: "MetalLB Service-type Load Balancer Setup"
weight: 30
date: 2017-01-05
description: >
  How to set up MetalLB for Service-type Load Balancer
---

<!-- overview -->

The purpose of this document is to walk you through getting set up with MetalLB Kubernetes Load Balancer for your cluster.
This is suggested as an alternative if your networking requirements do not allow you to use [Kube-vip]({{< ref "/docs/tasks/workload/loadbalance/kubevip" >}}).

<!-- body -->

MetalLB is a native Kubernetes load balancing solution for bare-metal Kubernetes clusters.
Detailed information about MetalLB can be found [here](https://metallb.universe.tf/).

### Prerequisites

You will need Helm installed on your system as this is the easiest way to deploy MetalLB.
Helm can be installed from [here](https://helm.sh/docs/intro/install/).
MetalLB installation is described [here](https://metallb.universe.tf/installation/)

### Steps

1. Enable strict ARP as it's required for MetalLB

    ```bash
    kubectl get configmap kube-proxy -n kube-system -o yaml | \
    sed -e "s/strictARP: false/strictARP: true/" | \
    kubectl apply -f - -n kube-system
    ```

2. Pull helm repo for metalLB

    ```bash
    helm repo add metallb https://metallb.github.io/metallb
    ```

3. Create an override file to specify LB IP range

    LB-IP-RANGE can be a CIDR block like 198.18.210.0/24 or range like 198.18.210.0-198.18.210.10

    ```bash
    cat << 'EOF' >> values.yaml
    configInline:
      address-pools:
        - name: default
          protocol: layer2
          addresses:
          - <LB-IP-range>
    EOF
    ```

4. Install metalLB on your cluster

    ```bash
    helm install metallb metallb/metallb -f values.yaml
    ```

5. Deploy the [Hello EKS Anywhere]({{< ref "/docs/tasks/workload/test-app" >}}) test application.

    ```bash
    kubectl apply -f https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml
    ```

6. Expose the hello-eks-a deployment

    ```bash
    kubectl expose deployment hello-eks-a --port=80 --type=LoadBalancer --name=hello-eks-a-lb
    ```

7. Get the load balancer external IP

    ```
    EXTERNAL_IP=$(kubectl get svc hello-eks-a-lb -o jsonpath='{.spec.externalIP}')
    ```

8. Hit the external ip

    ```bash
    curl ${EXTERNAL_IP}
    ```
