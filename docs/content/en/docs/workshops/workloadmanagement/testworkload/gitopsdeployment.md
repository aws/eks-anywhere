---
title: "GitOps Deployment"
linkTitle: "GitOps Deployment"
weight: 2
date: 2021-11-11
description: >  
---

Part of this chapter we will see how to manage and deploy workloads to cluster using GitOps.

## Manage cluster

After your cluster is created you can test the GitOps controller by modifying the cluster specification. In this section we will increase the workernode count from `1` to `2`.

1. Run the following command to get the list of worker nodes currently in the cluster:

    ```bash
    kubectl get nodes
    ```

    **Output**
    {{< highlight bash "hl_lines=3" >}}
    NAME                           STATUS     ROLES                  AGE   VERSION
    gitops-lt5qd                   Ready      control-plane,master   42m   v1.21.2-eks-1-21-4
    gitops-md-0-64568d878d-c7mpq   Ready      <none>                 41m   v1.21.2-eks-1-21-4
    {{< / highlight >}}

    > Currently EKS Anywhere cluster has only one worker node running under the name `gitops-md-0-64568d878d-c7mpq`

2. Clone your git repository:

    ```bash
    mkdir gitrepo
    cd gitrepo
    git clone https://github.com/${repositoryownername}/eks-gitops-cluster.git .
    ```

    > **Notes:** Replace `${repositoryownername}` with your repository owner name. The command assumes the cluster got created using the default repository name `eks-gitops-cluster`.

3. Edit the cluster specification file, the default path for the cluster file is:

    ```bash
    vi clusters/$CLUSTER_NAME/eksa-system/eksa-cluster.yaml
    ```

4. Modify the `workerNodeGroupsConfigurations[0].count` field from `1` to `2`

5. Commit the file to the git repository

    ```bash
    git add *
    git commit -m 'Scaling nodes for test'
    git push origin main
    ```

    > The flux controller will automatically make the required changes.

6. If you updated your node count you can use this command to see the current node state.

    ```bash
    kubectl get nodes
    ```

    **Output**

    {{< highlight bash "hl_lines=4" >}}
    NAME                           STATUS     ROLES                  AGE   VERSION
    gitops-lt5qd                   Ready      control-plane,master   42m   v1.21.2-eks-1-21-4
    gitops-md-0-64568d878d-c7mpq   Ready      <none>                 41m   v1.21.2-eks-1-21-4
    gitops-md-0-64568d878d-f9vcr   Ready      <none>                 5m    v1.21.2-eks-1-21-4
    {{< / highlight >}}

    > **Notes:** The `gitops-md-0-64568d878d-f9vcr` node is the new worker node added to the cluster.

## Manage Workload

Lets deploy a hello world workload the the cluster using GitOps.

1. Create a new directory for your workload.

```bash
cd clusters/$CLUSTER_NAME/eksa-system
mkdir demo-application && cd demo-application
```

2. Create a new k8s deployment file named as `deployment.yaml` using the following template:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-eks-a
  namespace: demo-application
  labels:
    app: hello-eks-a
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hello-eks-a
  template:
    metadata:
      labels:
        app: hello-eks-a
    spec:
      containers:
      - name: hello
        image: public.ecr.aws/aws-containers/hello-eks-anywhere:latest
        ports:
        - containerPort: 80
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
```

3. Create a new k8s service file named as `service.yaml` using the following template:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-eks-a
  namespace: demo-application
spec:
  type: NodePort
  selector:
    app: hello-eks-a
  ports:
    - port: 80
```

4. Create a file named as `kustomization.yaml` using the following template which maps the deployment and service files to the cluster:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: demo-application
resources:
  - deployment.yaml
  - service.yaml
```

5. Commit the files to the git repository

```bash
cd ../../../..
git add *
git commit -m 'Deploying hello world workload'
git push origin main
```

> **Notes:** The flux controller will automatically deploy the workload to the cluster.

6. In a few seconds we should see the workload deployed in the cluster. Run the following command to verify that:

```bash
kubectl get deployments | grep hello-eks-a
```

**Output**

```bash
hello-eks-a   1/1     1            1           19m
```

> **Notes:** Use `kubectl get gitrepositories -A` to verify the git repository the flux controller is connected to