---
title: "Expose metrics for EKS Anywhere components"
linkTitle: "Expose metrics"
weight: 100
date: 2024-04-06
description: >
  Expose metrics for EKS Anywhere components
---

Some Kubernetes system components like kube-controller-manager, kube-scheduler, kube-proxy and etcd (Stacked) expose metrics only on the localhost by default. In order to expose metrics for these components so that other monitoring systems like Prometheus can scrape them, you can deploy a proxy as a Daemonset on the host network of the nodes. The proxy pods also need to be configured with control plane tolerations so that they can be scheduled on the control plane nodes. For Unstacked/External etcd, metrics are already exposed on `https://<etcd-machine-ip>:2379/metrics` endpoint and can be scraped by Prometheus directly without deploying anything.

### Configure Proxy

To configure a proxy for exposing metrics on an EKS Anywhere cluster, you can perform the following steps:

1.  Create a config map to store the proxy configuration.
    
    Below is an example ConfigMap if you use HAProxy as the proxy server.
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: metrics-proxy
    data:
      haproxy.cfg: |
        defaults
          mode http
          timeout connect 5000ms
          timeout client 5000ms
          timeout server 5000ms
          default-server maxconn 10

        frontend kube-proxy
          bind \${NODE_IP}:10249
          http-request deny if !{ path /metrics }
          default_backend kube-proxy
        backend kube-proxy
          server kube-proxy 127.0.0.1:10249 check

        frontend kube-controller-manager
          bind \${NODE_IP}:10257
          http-request deny if !{ path /metrics }
          default_backend kube-controller-manager
        backend kube-controller-manager
          server kube-controller-manager 127.0.0.1:10257 ssl verify none check

        frontend kube-scheduler
          bind \${NODE_IP}:10259
          http-request deny if !{ path /metrics }
          default_backend kube-scheduler
        backend kube-scheduler
          server kube-scheduler 127.0.0.1:10259 ssl verify none check

        frontend etcd
          bind \${NODE_IP}:2381
          http-request deny if !{ path /metrics }
          default_backend etcd
        backend etcd
          server etcd 127.0.0.1:2381 check
    EOF
    ```

2.  Create a daemonset for the proxy and mount the config map volume onto the proxy pods.
    
    Below is an example configuration for the HAProxy daemonset.
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: apps/v1
    kind: DaemonSet
    metadata:
      name: metrics-proxy
    spec:
      selector:
        matchLabels:
          app: metrics-proxy
      template:
        metadata:
          labels:
            app: metrics-proxy
        spec:
          tolerations:
          - key: node-role.kubernetes.io/control-plane
            operator: Exists
            effect: NoSchedule
          hostNetwork: true
          containers:
            - name: haproxy
              image: public.ecr.aws/eks-anywhere/kubernetes-sigs/kind/haproxy:v0.20.0-eks-a-54
              env:
                - name: NODE_IP
                  valueFrom:
                    fieldRef:
                      apiVersion: v1
                      fieldPath: status.hostIP
              ports:
                - name: kube-proxy
                  containerPort: 10249
                - name: kube-ctrl-mgr
                  containerPort: 10257
                - name: kube-scheduler
                  containerPort: 10259
                - name: etcd
                  containerPort: 2381
              volumeMounts:
                - mountPath: "/usr/local/etc/haproxy"
                  name: haproxy-config
          volumes:
            - configMap:
                name: metrics-proxy
              name: haproxy-config
    EOF
    ```

### Configure Client Permissions

1.  Create a new cluster role for the client to access the metrics endpoint of the components.
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: metrics-reader
    rules:
      - nonResourceURLs:
          - "/metrics"
        verbs:
          - get
    EOF
    ```

2.  Create a new cluster role binding to bind the above cluster role to the client pod's service account.
    
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: metrics-reader-binding
    subjects:
    - kind: ServiceAccount
      name: default
      namespace: default
    roleRef:
      kind: ClusterRole
      name: metrics-reader
      apiGroup: rbac.authorization.k8s.io
    EOF
    ```

3.  Verify that the metrics are exposed to the client pods by running the following commands:
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: v1
    kind: Pod
    metadata:
      name: test-pod
    spec:
      tolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
      containers:
      - command:
        - /bin/sleep
        - infinity
        image: curlimages/curl:latest
        name: test-container
        env:
        - name: NODE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
    EOF
    ```

    ```bash
    kubectl exec -it test-pod -- sh
    export TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
    curl -H "Authorization: Bearer ${TOKEN}" "http://${NODE_IP}:10257/metrics"
    curl -H "Authorization: Bearer ${TOKEN}" "http://${NODE_IP}:10259/metrics"
    curl -H "Authorization: Bearer ${TOKEN}" "http://${NODE_IP}:10249/metrics"
    curl -H "Authorization: Bearer ${TOKEN}" "http://${NODE_IP}:2381/metrics"
    ```