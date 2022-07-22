---
title: "Add an ingress controller"
linkTitle: "Add an ingress controller"
weight: 30
date: 2017-01-05
description: >
  How to deploy an ingress controller for simple host or URL-based HTTP routing into workload running in EKS-A
---

<!-- overview -->

A production-quality Kubernetes cluster requires planning and preparation for various networking features.

<!-- body -->


The purpose of this document is to walk you through getting set up with a recommended Kubernetes Ingress Controller for EKS Anywhere.
Ingress Controller is essential in order to have routing rules that decide how external users access services running in a Kubernetes cluster. It enables efficient distribution of incoming network traffic among multiple backend services.


## Current Recommendation: Emissary-ingress 

We currently recommend using Emissary-ingress Kubernetes Ingress Controller by Ambassador. Emissary-ingress allows you to route and secure traffic to your cluster with an Open Source Kubernetes-native API Gateway. Detailed information about Emissary-ingress can be found [here](https://www.getambassador.io/docs/emissary/).

## Setting up Emissary-ingress for Ingress Controller

1. Deploy the [Hello EKS Anywhere]({{< relref "./test-app" >}}) test application.
    ```bash
    kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
    ```

2. Set up a load balancer: Set up MetalLB Load Balancer by following the instructions [here]({{< relref "./loadbalance" >}})

3. Install Ambassador CRDs and ClusterRoles and RoleBindings

    ```bash
    kubectl apply -f "https://www.getambassador.io/yaml/ambassador/ambassador-crds.yaml"
    kubectl apply -f "https://www.getambassador.io/yaml/ambassador/ambassador-rbac.yaml"
    ```

4. Create Ambassador Service with Type LoadBalancer.

    ```bash
    kubectl apply -f - <<EOF
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: ambassador
    spec:
      type: LoadBalancer
      externalTrafficPolicy: Local
      ports:
      - port: 80
        targetPort: 8080
      selector:
        service: ambassador
    EOF
    ```

5. Create a Mapping on your cluster. This Mapping tells Emissary-ingress to route all traffic inbound to the /backend/ path to the quote Service.

    ```bash
    kubectl apply -f - <<EOF
    ---
    apiVersion: getambassador.io/v2
    kind: Mapping
    metadata:
      name: hello-backend
    spec:
      prefix: /backend/
      service: hello-eks-a
    EOF
    ```  
 
6. Store the Emissary-ingress load balancer IP address to a local environment variable. You will use this variable to test accessing your service.

    ```bash
    export EMISSARY_LB_ENDPOINT=$(kubectl get svc ambassador -o "go-template={{range .status.loadBalancer.ingress}}{{or .ip .hostname}}{{end}}")
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
