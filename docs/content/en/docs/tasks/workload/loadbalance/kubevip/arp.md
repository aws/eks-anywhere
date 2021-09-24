---
title: "Kube-Vip ARP Mode"
linkTitle: "Kube-Vip ARP Mode"
weight: 30
date: 2017-01-05
description: >
  How to set up kube-vip for Service-type Load Balancer in ARP mode
---

<!-- overview -->

In ARP mode, kube-vip will perform leader election and assign the virtual IP to the leader.
This node will inherit the VIP and become the load-balancing leader within the cluster.

<!-- body -->

## Setting up Kube-Vip for Service-type Load Balancer in ARP mode

1. Enable strict ARP in kube-proxy as it's required for kube-vip
    ```bash
    kubectl get configmap kube-proxy -n kube-system -o yaml | \
    sed -e "s/strictARP: false/strictARP: true/" | \
    kubectl apply -f - -n kube-system
    ```

1. Create a configMap to specify the IP range for load balancer.
You can use either a CIDR block or an IP range

    ```bash
    CIDR=192.168.0.0/24 # Use your CIDR range here
    kubectl create configmap --namespace kube-system kubevip --from-literal cidr-global=${CIDR}
    ```
    ```bash
    IP_START=192.168.0.0  # Use the starting IP in your range
    IP_END=192.168.0.255  # Use the ending IP in your range
    kubectl create configmap --namespace kube-system kubevip --from-literal range-global=${IP_START}-${IP_END}
    ```

1. Deploy kubevip-cloud-provider

    ```bash
    kubectl apply -f https://kube-vip.io/manifests/controller.yaml
    ```

1. Create ClusterRoles and RoleBindings for kube-vip daemonset

    ```bash
    kubectl apply -f https://kube-vip.io/manifests/rbac.yaml
    ```

1. Create the kube-vip daemonset

    An example manifest has been included at the end of this document which you can use in place of this step.

    ```bash
    alias kube-vip="docker run --network host --rm plndr/kube-vip:v0.3.5"
    kube-vip manifest daemonset --services --inCluster --arp --interface eth0 | kubectl apply -f -
    ```   
 
1. Deploy the [Hello EKS Anywhere]({{< ref "/docs/tasks/workload/test-app" >}}) test application.

    ```bash
    kubectl apply -f https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml
    ```

1. Expose the hello-eks-a service

    ```bash
    kubectl expose deployment hello-eks-a --port=80 --type=LoadBalancer --name=hello-eks-a-lb
    ```

1. Describe the service to get the IP.
The external IP will be the one in CIDR range specified in step 4

    ```bash
    EXTERNAL_IP=$(kubectl get svc hello-eks-a-lb -o jsonpath='{.spec.externalIP}')
    ```

1. Ensure the load balancer is working by curl'ing the IP you got in step 8

    ```bash
    curl ${EXTERNAL_IP}
    ```   
 
You should see something like this in the output

```
   ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡

   Thank you for using

   ███████╗██╗  ██╗███████╗                                             
   ██╔════╝██║ ██╔╝██╔════╝                                             
   █████╗  █████╔╝ ███████╗                                             
   ██╔══╝  ██╔═██╗ ╚════██║                                             
   ███████╗██║  ██╗███████║                                             
   ╚══════╝╚═╝  ╚═╝╚══════╝                                             
                                                                     
    █████╗ ███╗   ██╗██╗   ██╗██╗    ██╗██╗  ██╗███████╗██████╗ ███████╗
   ██╔══██╗████╗  ██║╚██╗ ██╔╝██║    ██║██║  ██║██╔════╝██╔══██╗██╔════╝
   ███████║██╔██╗ ██║ ╚████╔╝ ██║ █╗ ██║███████║█████╗  ██████╔╝█████╗  
   ██╔══██║██║╚██╗██║  ╚██╔╝  ██║███╗██║██╔══██║██╔══╝  ██╔══██╗██╔══╝  
   ██║  ██║██║ ╚████║   ██║   ╚███╔███╔╝██║  ██║███████╗██║  ██║███████╗
   ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝    ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝
                                                                     
   You have successfully deployed the hello-eks-a pod hello-eks-a-c5b9bc9d8-fx2fr

   For more information check out
   https://anywhere.eks.amazonaws.com

⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡

   ```

Here is an example manifest for kube-vip from step 5, which is also available [here](https://raw.githubusercontent.com/kube-vip/kube-vip/f0f0ec3bc953d4b42c78f1b35ba944804a9e31aa/example/deploy/0.3.5.yaml)

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
