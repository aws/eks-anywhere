---
title: "Istio"
linkTitle: "Istio"
weight: 1
date: 2021-11-11
description: >  
---

## Overview

Part of this article, we will setup istio, deploy a microservice, visualize the topology and traffic with Kiali

**Istio** extends Kubernetes to establish a programmable, application-aware network using the powerful Envoy service proxy. Working with both Kubernetes and traditional workloads, Istio brings standard, universal traffic management, telemetry, and security to complex deployments.

**Kiali** is a management console for Istio service mesh. It provides robust observability for your running mesh, letting you quickly identify issues and then troubleshoot those issues. Kiali offers in-depth traffic topology, health grades, powerful dashboards, and lets you drill into component detail.

## Step by step guide

### Install Istio

1. Download `Istio`

    ```bash
    curl -L https://github.com/istio/istio/releases/download/1.11.3/istio-1.11.3-osx.tar.gz | tar xzvf -
    cd istio-1.11.3
    export PATH=$PWD/bin:$PATH
    ```

2. Install `Istio`

    ```bash
    istioctl install
    ```

    **Output**

    ```bash
    This will install the Istio 1.11.3 default profile with ["Istio core" "Istiod" "Ingress gateways"] components into the cluster. Proceed? (y/N) y
    ✔ Istio core installed
    ✔ Istiod installed
    ✔ Ingress gateways installed
    ✔ Installation complete
    Thank you for installing Istio 1.11.  Please take a few minutes to tell us about your install/upgrade experience!  https://forms.gle/kWULBRjUv7hHci7T6
    ```

3. Verify services:

    ```bash
    kubectl get pods -n istio-system
    ```

4. Enable Istio on default namespace:

    ```bash
    kubectl label namespace default istio-injection=enabled
    ```

### Install Addons (with Kiali)

1. Run the following command to deploy `Kiali` application, since Kiali relies on Prometheus and few other services, we will deploy the the `addons` folder inside `istio-1.11.3`

    ```bash
    kubectl create -f samples/addons
    ```

    **Output**

    ```bash
    serviceaccount/grafana created
    configmap/grafana created
    service/grafana created
    deployment.apps/grafana created
    configmap/istio-grafana-dashboards created
    configmap/istio-services-grafana-dashboards created
    deployment.apps/jaeger created
    service/tracing created
    service/zipkin created
    service/jaeger-collector created
    deployment.apps/kiali configured
    serviceaccount/prometheus created
    configmap/prometheus created
    clusterrole.rbac.authorization.k8s.io/prometheus created
    clusterrolebinding.rbac.authorization.k8s.io/prometheus created
    service/prometheus created
    deployment.apps/prometheus created
    ```

2. Verify `Kiali` is up and running by execution the following command:

    ```bash
    kubectl get pods -n istio-system | grep kiali
    ```

    **Output**

    ```bash
    kiali-fd9f88575-q4664                  1/1     Running   0          108s
    ```

### Install BookInfo application

#### Deploy application

1. Deploy the application with automatic sidecar injection for each pod:

    ```bash
    kubectl apply -f samples/bookinfo/platform/kube/bookinfo.yaml
    ```

    **Output**

    ```bash
    service/details created
    serviceaccount/bookinfo-details created
    deployment.apps/details-v1 created
    service/ratings created
    serviceaccount/bookinfo-ratings created
    deployment.apps/ratings-v1 created
    service/reviews created
    serviceaccount/bookinfo-reviews created
    deployment.apps/reviews-v1 created
    deployment.apps/reviews-v2 created
    deployment.apps/reviews-v3 created
    service/productpage created
    serviceaccount/bookinfo-productpage created
    deployment.apps/productpage-v1 created
    ```

2. Verify services

    ```bash
    kubectl get svc
    ```

    **Output**

    ```bash
    details       ClusterIP   10.139.35.132   <none>        9080/TCP   28s
    kubernetes    ClusterIP   10.128.0.1      <none>        443/TCP    15h
    productpage   ClusterIP   10.132.32.112   <none>        9080/TCP   28s
    ratings       ClusterIP   10.137.89.129   <none>        9080/TCP   28s
    reviews       ClusterIP   10.128.45.75    <none>        9080/TCP   28s
    ```

3. Verify pods:

    ```bash
    kubectl get pods
    ```

    **Output**

    ```bash
    NAME                              READY   STATUS    RESTARTS   AGE
    details-v1-79f774bdb9-k65jh       1/1     Running   0          2m38s
    productpage-v1-6b746f74dc-m58m9   1/1     Running   0          2m37s
    ratings-v1-b6994bb9-qqdqt         1/1     Running   0          2m38s
    reviews-v1-545db77b95-cn7wh       1/1     Running   0          2m38s
    reviews-v2-7bf8c9648f-mjx8x       1/1     Running   0          2m38s
    reviews-v3-84779c7bbc-tjk25       1/1     Running   0          2m38s
    ```

#### Access Application (vSphere)

{{% alert title="Notes" color="warning" %}}
This section is only applicable for EKS Anywhere cluster running in vSphere environments.
{{% /alert %}}

1. Define the ingress gateway for the application:

    ```bash
    kubectl apply -f samples/bookinfo/networking/bookinfo-gateway.yaml
    ```

    **Output**

    ```bash
    gateway.networking.istio.io/bookinfo-gateway created
    virtualservice.networking.istio.io/bookinfo created
    ```

2. Verify gateway, by running the following command:

    ```bash
    kubectl get gateway
    ```

    **Output**

    ```bash
    NAME               AGE
    bookinfo-gateway   6s
    ```

3. Define environment variables:

    ```bash
    export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http")].port}')
    export SECURE_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
    export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT
    ```

4. Access the application:

    ```bash
    echo "Accessing bookinfo at http://${GATEWAY_URL}/productpage"
    open http://${GATEWAY_URL}/productpage    # Open browser, OS/X only.
    ```

    **Output**
    ![bookinfo](../images/bookinfo.png)

#### Access application (Local)

{{% alert title="Notes" color="warning" %}}
This section is only applicable for Local/Dev cluster running in docker environment.
{{% /alert %}}

1. Enable port forwarding on `productpage` service, by running the following command:

    ```bash
    kubectl port-forward svc/productpage 8080:9080
    ```

    **Output**

    ```bash
    Forwarding from 127.0.0.1:8080 -> 9080
    Forwarding from [::1]:8080 -> 9080
    ```

2. Open the browser and navigate to `http://localhost:8080/productpage`

### Visualize topology and traffic

1. Enable port forwarding on `Kiali` service, by running the following command:

    ```bash
    kubectl -n istio-system port-forward svc/kiali 20001:20001
    ```

    **Output**

    ```bash
    Forwarding from 127.0.0.1:20001 -> 20001
    Forwarding from [::1]:20001 -> 20001
    ```

2. Open the browser and navigate to `http://localhost:20001`. Click on **Graph** side menu and filter `default` **Namespace**, since the sample application is deployed to this namespace. We should see the following topology:

    ![topology](../images/topology.png)

## Cleanup

1. Run the following command, from `istio` root directory to delete the addons:

    ```bash
    kubectl delete -f samples/addons/
    ```

    **Output**

    ```bash
    serviceaccount "grafana" deleted
    configmap "grafana" deleted
    service "grafana" deleted
    deployment.apps "grafana" deleted
    configmap "istio-grafana-dashboards" deleted
    configmap "istio-services-grafana-dashboards" deleted
    deployment.apps "jaeger" deleted
    service "tracing" deleted
    service "zipkin" deleted
    service "jaeger-collector" deleted
    serviceaccount "kiali" deleted
    configmap "kiali" deleted
    clusterrole.rbac.authorization.k8s.io "kiali-viewer" deleted
    clusterrole.rbac.authorization.k8s.io "kiali" deleted
    clusterrolebinding.rbac.authorization.k8s.io "kiali" deleted
    role.rbac.authorization.k8s.io "kiali-controlplane" deleted
    rolebinding.rbac.authorization.k8s.io "kiali-controlplane" deleted
    service "kiali" deleted
    deployment.apps "kiali" deleted
    serviceaccount "prometheus" deleted
    configmap "prometheus" deleted
    clusterrole.rbac.authorization.k8s.io "prometheus" deleted
    clusterrolebinding.rbac.authorization.k8s.io "prometheus" deleted
    service "prometheus" deleted
    deployment.apps "prometheus" deleted
    ```

2. Run the following command, from `istio` root directory to delete the sample application:

    ```bash
    kubectl delete -f samples/bookinfo/platform/kube/bookinfo.yaml
    ```

    **Output**

    ```bash
    service "details" deleted
    serviceaccount "bookinfo-details" deleted
    deployment.apps "details-v1" deleted
    service "ratings" deleted
    serviceaccount "bookinfo-ratings" deleted
    deployment.apps "ratings-v1" deleted
    service "reviews" deleted
    serviceaccount "bookinfo-reviews" deleted
    deployment.apps "reviews-v1" deleted
    deployment.apps "reviews-v2" deleted
    deployment.apps "reviews-v3" deleted
    service "productpage" deleted
    serviceaccount "bookinfo-productpage" deleted
    deployment.apps "productpage-v1" deleted
    ```
