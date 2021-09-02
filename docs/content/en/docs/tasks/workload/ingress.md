---
title: "Add an ingress controller"
linkTitle: "Add an ingress controller"
weight: 30
date: 2017-01-05
description: >
  How to deploy an ingress controller for simple host or URL based HTTP routing into workload running in EKS-A
---

<!-- overview -->

A production-quality Kubernetes cluster requires planning and preparation for various networking features.

<!-- body -->


The purpose of this document is to walk you through getting set up with a recommended Kubernetes Ingress Controller for EKS Anywhere (EKS-A).
Ingress Controller is essential in order to have routing rules that decide how external users access services running in a Kubernetes cluster. It enables efficient distribution of incoming network traffic among multiple backend services.


## Current Recommendation: Emissary-ingress 

We currently recommend using Emissary-ingress Kubernetes Ingress Controller by Ambassador. Emissary-ingress allows you to route and secure traffic to your cluster with an Open Source Kubernetes-native API Gateway. Detailed information about Emissary-ingress can be found [here](https://www.getambassador.io/docs/emissary/).

## Setting up Emissary-ingress for Ingress Controller

1. Set up a test web application in your cluster. You can use Ambassador's [Quote of the Moment service], as an example. Apply YAML for this application.
    ```bash
    kubectl apply -f https://app.getambassador.io/yaml/ambassador-docs/latest/quickstart/qotm.yaml
    ```

2. Set up kube-vip service type: Load Balancer in your cluster by following the instructions [here](https://eksanywhere.jgarr.net/docs/tasks/workload/loadbalance/#setting-up-kube-vip-for-service-type-load-balancer). 
Alternatively, you can set up MetalLB Load Balancer by following the instructions [here](https://eksanywhere.jgarr.net/docs/tasks/workload/loadbalance/#alternatives)

3. Install Ambassador CRDs and ClusterRoles and RoleBindings

    ```bash
    kubectl apply -f https://www.getambassador.io/yaml/ambassador/ambassador-crds.yaml
    kubectl apply -f https://www.getambassador.io/yaml/ambassador/ambassador-rbac.yaml
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
      name: quote-backend
    spec:
      prefix: /backend/
      service: quote
    EOF
    ```  
 
6. Store the Emissary-ingress load balancer IP address to a local environment variable. You will use this variable to test accessing your service.

    ```bash
    export EMISSARY_LB_ENDPOINT=$(kubectl get svc ambassador \ 
      -o "go-template={{range .status.loadBalancer.ingress}}{{or .ip .hostname}}{{end}}")
    ```   
 
7. Test the configuration by accessing the service through the Emissary-ingress load balancer.

    ```bash
    curl -Lk http://$EMISSARY_LB_ENDPOINT/backend/
    ```   
 
You should see something like this in the output

```html
  {
   "server": "idle-cranberry-8tbb6iks",
   "quote": "Non-locality is the driver of truth. By summoning, we vibrate.",
   "time": "2021-02-26T15:55:06.884798988Z"
  }

```
