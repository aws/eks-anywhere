---
title: "FluentBit with Elastic Search"
linkTitle: "FluentBit with Elastic Search"
weight: 2
date: 2021-11-11
description: >  
---

## Overview

We will see how to setup FluentBit with Elastic Search to collect metrics and logs for EKS Anywhere cluster

**Fluent Bit** Fluent Bit is an open source Log Processor and Forwarder which allows you to collect any data like metrics and logs from different sources, enrich them with filters and send them to multiple destinations. It's the preferred choice for containerized environments like Kubernetes.

**Elastic search** is a distributed, RESTful search and analytics engine capable of addressing a growing number of use cases. As the heart of the Elastic Stack, it centrally stores your data for lightning fast search, fine‑tuned relevancy, and powerful analytics that scale with ease.

**Kibana** is a free and open user interface that lets you visualize your Elasticsearch data and navigate the Elastic Stack. Do anything from tracking query load to understanding the way requests flow through your apps.

### Step by step guide

1. Run the following command to create the storage class:

```bash
kubectl apply -f - <<EOF
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
provisioner: kubernetes.io/no-provisioner
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
EOF
```

**Output**

```
storageclass.storage.k8s.io/local-storage created
```

2. Since we are provisioning a local storage class, we need to create a persistent volume. Run the following command to create the persistent volume:

```bash
kubectl apply -f - <<EOF
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv1
spec:
  capacity:
    storage: 50Gi
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: standard
  local:
    path: /tmp
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - <<Node1>>
          - <<Node2>>
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv2
spec:
  capacity:
    storage: 50Gi
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: standard
  local:
    path: /tmp
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - <<Node1>>
          - <<Node2>>
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv3
spec:
  capacity:
    storage: 50Gi
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: standard
  local:
    path: /tmp
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - <<Node1>>
          - <<Node2>>
EOF
```

{{% alert title="Notes" color="warning" %}}
<<Node1>>` and `<<Node2>>` values needs to replaced with the actual node names when the cluster is created, which can be retrieved by running `kubectl get nodes` command.
{{% /alert %}}

**Output**

```
persistentvolume/local-pv1 created
persistentvolume/local-pv2 created
persistentvolume/local-pv3 created
```

3. Install Elastic Search and Kibana, by running the following command:

```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: es-cluster
spec:
  serviceName:  elasticsearch-entrypoint
  replicas: 1
  selector:
    matchLabels:
      name: elasticsearch
  template:
    metadata:
      labels:
        name: elasticsearch
    spec:
      containers:
        - name: elasticsearch
          image: elasticsearch:7.10.1
          env:
            - name: discovery.type
              value: single-node
          ports:
            - containerPort: 9200
              name: rest
            - containerPort: 9300
              name: inter-node            
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kibana
spec:
  replicas: 1
  selector:
    matchLabels:
      name: kibana
  template:
    metadata:
      labels:
        name: kibana
    spec:
      containers:
        - name: kibana
          image: kibana:7.10.1
          ports:
            - containerPort: 5601
          env:
            - name: ELASTICSEARCH_HOSTS
              value: http://es-cluster-0.elasticsearch-entrypoint.default.svc.cluster.local:9200
---
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch-entrypoint
  namespace: default
spec:
  clusterIP: None
  selector:
    name: elasticsearch
  ports:
  - port: 9200
    name: rest
  - port: 9300
    name: inter-node
---
apiVersion: v1
kind: Service
metadata:
  name: kibana-entrypoint
  namespace: default
spec:
  selector:
    name: kibana
  ports:
  - port: 5601
EOF
```

**Output**

```
statefulset.apps/es-cluster created
deployment.apps/kibana created
service/elasticsearch-entrypoint created
service/kibana-entrypoint created
```

4. Get the status of the Elastic Search and Kibana deployment by running the following command:

```bash
kubectl get pods -l name=elasticsearch
```

**Output**

```
NAME                      READY   STATUS    RESTARTS   AGE
es-cluster-0              1/1     Running   0          2m26s
kibana-5cbdfd8d86-6tcdh   1/1     Running   0          2m26s
```

5. Before installing the `fluentbit` deamonset, set up `namespace`,`serviceaccount`, and `rbac` as follows:

```bash
kubectl create namespace logging
kubectl create -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-service-account.yaml
kubectl create -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-role-1.22.yaml
kubectl create -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-role-binding-1.22.yaml
```

6. Run the following command to setup `configMap` for `fluent-bit`:

```bash
kubectl apply -f - <<EOF
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config  
  namespace: logging
  labels:
    k8s-app: fluent-bit
data:
  # Configuration files: server, input, filters and output
  # ======================================================
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020

    @INCLUDE input-kubernetes.conf
    @INCLUDE filter-kubernetes.conf
    @INCLUDE output-elasticsearch.conf

  input-kubernetes.conf: |
    [INPUT]
        Name              tail
        Tag               kube.*
        Path              /var/log/containers/*.log
        Parser            docker
        DB                /var/log/flb_kube.db
        Mem_Buf_Limit     5MB
        Skip_Long_Lines   On
        Refresh_Interval  10

  filter-kubernetes.conf: |
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Kube_Tag_Prefix     kube.var.log.containers.
        Merge_Log           On
        Merge_Log_Key       log_processed
        K8S-Logging.Parser  On
        K8S-Logging.Exclude Off

  output-elasticsearch.conf: |
    [OUTPUT]
        Name            es
        Match           *
        Host            es-cluster-0.elasticsearch-entrypoint.default.svc.cluster.local
        Port            9200
        Logstash_Format On
        Replace_Dots    On
        Retry_Limit     False

  parsers.conf: |
    [PARSER]
        Name   apache
        Format regex
        Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name   apache2
        Format regex
        Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name   apache_error
        Format regex
        Regex  ^\[[^ ]* (?<time>[^\]]*)\] \[(?<level>[^\]]*)\](?: \[pid (?<pid>[^\]]*)\])?( \[client (?<client>[^\]]*)\])? (?<message>.*)$

    [PARSER]
        Name   nginx
        Format regex
        Regex ^(?<remote>[^ ]*) (?<host>[^ ]*) (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name   json
        Format json
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z

    [PARSER]
        Name        docker
        Format      json
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L
        Time_Keep   On

    [PARSER]
        # http://rubular.com/r/tjUt3Awgg4
        Name cri
        Format regex
        Regex ^(?<time>[^ ]+) (?<stream>stdout|stderr) (?<logtag>[^ ]*) (?<message>.*)$
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L%z

    [PARSER]
        Name        syslog
        Format      regex
        Regex       ^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<host>[^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$
        Time_Key    time
        Time_Format %b %d %H:%M:%S
EOF
```

7. Run the following command to install `fluent-bit` DaemonSet:

```bash
kubectl apply -f - <<EOF
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: logging
  labels:
    k8s-app: fluent-bit-logging
    version: v1
    kubernetes.io/cluster-service: "true"
spec:
  selector:
    matchLabels:
      k8s-app: fluent-bit-logging
  template:
    metadata:
      labels:
        k8s-app: fluent-bit-logging
        version: v1
        kubernetes.io/cluster-service: "true"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "2020"
        prometheus.io/path: /api/v1/metrics/prometheus
    spec:
      containers:
      - name: fluent-bit
        image: fluent/fluent-bit:1.5
        imagePullPolicy: Always
        ports:
          - containerPort: 2020
        env:
        - name: FLUENT_ELASTICSEARCH_HOST
          value: "es-cluster-0.elasticsearch-entrypoint.default.svc.cluster.local"
        - name: FLUENT_ELASTICSEARCH_PORT
          value: "9200"
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: fluent-bit-config
          mountPath: /fluent-bit/etc/
        - name: mnt
          mountPath: /mnt
          readOnly: true
      terminationGracePeriodSeconds: 10
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: fluent-bit-config
        configMap:
          name: fluent-bit-config
      - name: mnt
        hostPath:
          path: /mnt
      serviceAccountName: fluent-bit
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      - operator: "Exists"
        effect: "NoExecute"
      - operator: "Exists"
        effect: "NoSchedule"
EOF
```

**Output**

```
daemonset.apps/fluent-bit created
```

8. Run the following command to check the status of `fluent-bit` pods:

```bash
kubectl get pods -n logging | grep fluent-bit
```

**Output**

```
fluent-bit-h4kcr          1/1     Running   0          21s
fluent-bit-s4c4t          1/1     Running   0          21s
```

{{% alert title="Notes" color="warning" %}}
All the pods should be up and running.
{{% /alert %}}

9. Deploy a sample workload to the cluster, by running the following command:

```bash
kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
```

**Output**

```
deployment.apps/hello-eks-a created
service/hello-eks-a created
```

## Visualize the metrics

1. Run the following command to port forward the Kibana service to localhost:

```bash
kubectl port-forward service/kibana-entrypoint 5601
```

2. Click on `Discover` menu on the left side of the Kibana UI. Click on **Create index button** and select the logstash index, click **Next**. In the next step select `@timestamp` as default **Time field**, click on **Create index pattern** to finish the setup.

3. After the setup, kibana will automatically direct you to the discover page. Type `kubernetes.container_name : "hello"` in the search box and press **Enter**. You should all the metrics collected for the sample `hello-eks-a` application that we deployed.

 ![kibana](../images/kibana.png)

## Cleanup

Run the following command to cleanup the resources:

```bash
kubectl delete configmap fluent-bit-config -n logging
kubectl delete daemonSet fluent-bit -n logging
kubectl delete statefulset es-cluster
kubectl delete deployment kibana
kubectl delete service elasticsearch-entrypoint kibana-entrypoint
kubectl delete -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-service-account.yaml
kubectl delete -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-role-1.22.yaml
kubectl delete -f https://raw.githubusercontent.com/fluent/fluent-bit-kubernetes-logging/master/fluent-bit-role-binding-1.22.yaml
kubectl delete namespace logging
kubectl delete -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
```
