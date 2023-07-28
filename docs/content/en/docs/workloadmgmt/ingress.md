---
title: "Add an ingress controller"
linkTitle: "Add an ingress controller"
weight: 30
aliases:
    /docs/tasks/workload/ingress/
date: 2017-01-05
description: >
  How to deploy an ingress controller for simple host or URL-based HTTP routing into workload running in EKS-A
---

While you are free to use any Ingress Controller you like with your EKS Anywhere cluster, AWS currently only supports Emissary Ingress.
For information on how to configure a Emissary Ingress curated package for EKS Anywhere, see the [Add Emissary Ingress]({{< relref "../packages/emissary" >}}) page.
You may also reference the official [emissary documentation](https://www.getambassador.io/docs/emissary) for further configuration details. Operators can also leverage the CNI chaining feature from Isovalent where in both Cilium as the CNI and another CNI can work in a [chain mode]({{< cilium "gettingstarted/cni-chaining" >}}).

## Setting up Emissary-ingress for Ingress Controller

1. Deploy the [Hello EKS Anywhere]({{< relref "./test-app" >}}) test application.
    ```bash
    kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
    ```

2. Set up a load balancer: Set up MetalLB Load Balancer by following the instructions [here]({{< relref "../packages/metallb" >}})

3. Install Emissary Ingress: Follow the instructions here [Add Emissary Ingress]({{< relref "../packages/emissary" >}})

4. Create Emissary Listeners on your cluster (This is a one time setup).
   
    ```bash
    kubectl apply -f - <<EOF
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Listener
    metadata:
      name: http-listener
      namespace: default
    spec:
      port: 8080
      protocol: HTTP
      securityModel: XFP
      hostBinding:
        namespace:
          from: ALL
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Listener
    metadata:
      name: https-listener
      namespace: default
    spec:
      port: 8443
      protocol: HTTPS
      securityModel: XFP
      hostBinding:
        namespace:
          from: ALL
    EOF
    ```

5. Create a Mapping, and Host for your cluster. This Mapping tells Emissary-ingress to route all traffic inbound to the /hello/ path to the Hello EKS Anywhere Service. The name of your hello-eks-anywhere service will be the same as the package name.

    ```bash
    kubectl apply -f - <<EOF
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: hello-backend
      labels:
        examplehost: host 
    spec:
      prefix: /hello/
      service: hello-eks-a
      hostname: "*"
    EOF
    ```  
 
6. Store the Emissary-ingress load balancer IP address to a local environment variable. You will use this variable to test accessing your service. You can find this if you're using a setup with MetalLB by finding the namespace you launched your emissary service in, and finding the external IP from the service.

    ```bash
    emissary-cluster        LoadBalancer   10.100.71.222   195.16.99.64   80:31794/TCP,443:31200/TCP

    export EMISSARY_LB_ENDPOINT=195.16.99.64
    ```   
 
1. Test the configuration by accessing the service through the Emissary-ingress load balancer.

    ```bash
    curl -Lk http://$EMISSARY_LB_ENDPOINT/hello/
    ```    
    ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

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

    You have successfully deployed the hello-eks-a pod hello-eks-anywhere-95fb65657-vk9rz

    For more information check out
    https://anywhere.eks.amazonaws.com

    Amazon EKS Anywhere
    Run EKS in your datacenter
    version: v0.1.2-11d92fc1e01c17601e81c7c29ea4a3db232068a8

    ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

   ```
