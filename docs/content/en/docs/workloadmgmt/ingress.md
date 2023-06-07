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

5. Create a Mapping, and Host for your cluster. This Mapping tells Emissary-ingress to route all traffic inbound to the /backend/ path to the Hello EKS Anywhere Service.

    ```bash
    kubectl apply -f - <<EOF
    apiVersion: getambassador.io/v3alpha1
    kind: Host
    metadata:
      name: hello-host
    spec:
      hostname: hello.example.com
      mappingSelector:
        matchLabels:
          examplehost: host
    ---
    apiVersion: getambassador.io/v3alpha1
    kind: Mapping
    metadata:
      name: hello-backend
      labels:
        examplehost: host 
    spec:
      prefix: /backend/
      service: hello-eks-a
      hostname: hello.example.com
    EOF
    ```  
 
6. Store the Emissary-ingress load balancer IP address to a local environment variable. You will use this variable to test accessing your service.

    ```bash
    export EMISSARY_LB_ENDPOINT=$(kubectl get svc emissary-$CLUSTER_NAME -o "go-template={{range .status.loadBalancer.ingress}}{{or .ip .hostname}}{{end}}")
    ```   
 
7. Test the configuration by accessing the service through the Emissary-ingress load balancer.

    ```bash
    curl -Lk http://$EMISSARY_LB_ENDPOINT/backend/
    ```   

   NOTE: URL base path will need to match what is specified in the prefix exactly, including the trailing '/'
 


   You should see something like this in the output

   ```
   ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

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

   ⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢⬡⬢

   ```
