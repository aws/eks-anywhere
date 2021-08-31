---
title: "Add an external load balancer"
linkTitle: "Add an external load balancer"
weight: 30
date: 2017-01-05
description: >
  How to deploy an external load balancer to expose a workload running in EKS-A
---

<!-- overview -->

A production-quality Kubernetes cluster requires planning and preparation for various networking features.

<!-- body -->


The purpose of this document is to walk you through getting set up with a recommended Kubernetes Load Balancer for EKS Anywhere (EKS-A).
Load Balancing is essential in order to maximize availability and scalability. It enables efficient distribution of incoming network traffic among multiple backend services.


## Current Recommendation: Kube-Vip

We currently recommend using Kube-Vip Kubernetes Service-type Load Balancer. Previously designed to support control-plane resiliency, it has since been expanded to provide load-balancing for applications and services within a Kubernetes cluster. Detailed information about Kube-Vip can be found [here](https://kube-vip.io/).

## Setting up Kube-Vip for Service-type Load Balancer

1. Enable strict ARP in kube-proxy as it's required for kube-vip
    ```bash
    kubectl get configmap kube-proxy -n kube-system -o yaml | \
    sed -e "s/strictARP: false/strictARP: true/" | \
    kubectl apply -f - -n kube-system
    ```

2. Create a configMap to specify the IP range for load balancer. You can use either a CIDR block or an IP range

    ```bash
    CIDR=192.168.0.0/24 # Use your CIDR range here
    kubectl create configmap --namespace kube-system kubevip --from-literal cidr-global=${CIDR}
    ```
    ```bash
    IP_START=192.168.0.0  # Use the starting IP in your range
    IP_END=192.168.0.255  # Use the ending IP in your range
    kubectl create configmap --namespace kube-system kubevip --from-literal range-global=${IP_START}-${IP_END}
    ```

3. Deploy kubevip-cloud-provider 

    ```bash
    kubectl apply -f https://kube-vip.io/manifests/controller.yaml
    ```

4. Create ClusterRoles and RoleBindings for kube-vip daemonset

    ```bash
    kubectl apply -f https://kube-vip.io/manifests/rbac.yaml
    ```

5. Create the kube-vip daemonset. An example manifest has been included at the end of this document which you can use in place of this step.

    ```bash
    alias kube-vip="docker run --network host --rm plndr/kube-vip:v0.3.5"
    kube-vip manifest daemonset --services --inCluster --arp --interface eth0 | kubectl apply -f -
    ```   
 
6. Deploy nginx 

    ```bash
    kubectl apply -f https://k8s.io/examples/application/deployment.yaml
    ```

7. Expose the nginx service

    ```bash
    kubectl expose deployment nginx-deployment --port=80 --type=LoadBalancer --name=nginx
    ```

8. Describe the service to get the IP. The external IP will be the one in CIDR range specified in step 4

    ```bash
    EXTERNAL_IP=$(kubectl get svc nginx -o jsonpath='{.spec.externalIP}')
    ```

9. Ensure the load balancer is working by curl'ing the IP you got in step 8

    ```bash
    curl ${EXTERNAL_IP}
    ```   
 
You should see something like this in the output

```html
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>

```

Here is an example manifest for kube-vip from step 5. Also available [here](https://raw.githubusercontent.com/kube-vip/kube-vip/f0f0ec3bc953d4b42c78f1b35ba944804a9e31aa/example/deploy/0.3.5.yaml)

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  creationTimestamp: null
  name: kube-vip-ds
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: kube-vip-ds
  template:
    metadata:
      creationTimestamp: null
      labels:
        name: kube-vip-ds
    spec:
      containers:
      - args:
        - manager
        env:
        - name: vip_arp
          value: "true"
        - name: vip_interface
          value: eth0
        - name: port
          value: "6443"
        - name: vip_cidr
          value: "32"
        - name: svc_enable
          value: "true"
        - name: vip_startleader
          value: "false"
        - name: vip_addpeerstolb
          value: "true"
        - name: vip_localpeer
          value: ip-172-20-40-207:172.20.40.207:10000
        - name: vip_address
        image: plndr/kube-vip:v0.3.5
        imagePullPolicy: Always
        name: kube-vip
        resources: {}
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - NET_RAW
            - SYS_TIME
      hostNetwork: true
      serviceAccountName: kube-vip
  updateStrategy: {}
status:
  currentNumberScheduled: 0
  desiredNumberScheduled: 0
  numberMisscheduled: 0
  numberReady: 0
```


## Alternatives

This is not the recommended choice but as an alternative, MetalLB Load Balancer can be set up. MetalLB is a native Kubernetes load balancing solution for bare-metal Kubernetes clusters. MetalLB installation is described [here](https://metallb.universe.tf/installation/)

### Prerequisites

You will need Helm installed on your system as this is the easiest way to deploy MetalLB. Helm can be installed from [here](https://helm.sh/docs/intro/install/)

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

3. Create an override file to specify LB IP range. LB-IP-RANGE can be a CIDR block like 198.18.210.0/24 or range like 198.18.210.0-198.18.210.10

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

5. Deploy nginx

    ```bash
    kubectl apply -f https://k8s.io/examples/application/deployment.yaml
    ```

6. Expose a LoadBalancer for nginx 

    ```bash
    kubectl expose deployment nginx-deployment --port=80 --type=LoadBalancer --name=nginx
    ```

7. Get the load balancer external IP

    ```
    EXTERNAL_IP=$(kubectl get svc nginx -o jsonpath='{.spec.externalIP}')
    ```

8. Hit the external ip

    ```bash
    curl ${EXTERNAL_IP}
    ```