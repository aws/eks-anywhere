---
toc_hide: true
hide_summary: true
---

### Create a GitOps enabled cluster

Generate your cluster configuration and add the GitOps configuration.
For a full spec reference see the [Cluster Spec reference]({{< relref "../../../reference/clusterspec/gitops" >}}).

>**_NOTE:_** After your cluster is created the cluster configuration will automatically be commited to your git repo.

1. Create an EKS Anywhere cluster with GitOps enabled.

    ```bash
    CLUSTER_NAME=gitops
    eksctl anywhere create cluster -f ${CLUSTER_NAME}.yaml
    ```

2. Add the following GitOps configuration to the cluster config file:

    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
        name: mynewgitopscluster
    spec:
    ... # collapsed cluster spec fields
    # Below added for gitops support
        gitOpsRef:
            kind: GitOpsConfig
            name: my-cluster-name
    ---
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: GitOpsConfig
    metadata:
        name: my-cluster-name
    spec:
        flux:
            github:
                personal: true
                repository: mygithubrepository
                owner: mygithubusername
    ```

    > **Note:** Replace `mygithubusername` with your GitHub username.

3. Create an EKS Anywhere cluster with GitOps enabled.

    ```bash
    eksctl anywhere create cluster -f ${CLUSTER_NAME}.yaml
    ```

    **Output**

    ```bash
    Checking Github Access Token permissions
    ✅ Github personal access token has the required repo permissions
    Performing setup and validations
    Warning: The docker infrastructure provider is meant for local development and testing only
    ✅ Docker Provider setup is valid
    ✅ Flux path
    Creating new bootstrap cluster
    Installing cluster-api providers on bootstrap cluster
    Provider specific setup
    Creating new workload cluster
    Installing networking on workload cluster
    Installing storage class on workload cluster
    Installing cluster-api providers on workload cluster
    Moving cluster management from bootstrap to workload cluster
    Installing EKS-A custom components (CRD and controller) on workload cluster
    Creating EKS-A CRDs instances on workload cluster
    Installing AddonManager and GitOps Toolkit on workload cluster
    Adding cluster configuration files to Git
    Finalized commit and committed to local repository	{"hash": "1cc5a6016118b1bae8744dd8433f42189ba90a21"}
    Finalized commit and committed to local repository	{"hash": "3811581b14fc8e585a67318986298854d4d4787c"}
    Writing cluster config file
    Deleting bootstrap cluster
    🎉 Cluster created!
    ```

    > **NOTE:** After your cluster is created the cluster configuration will automatically be commited to your git repo.

----

### Verification

#### Cluster verification

1. Connect to cluster - Once the cluster is created you can use it with the generated `KUBECONFIG` file in your local directory

    ```bash
    export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
    kubectl get ns
    ```

    **Output**

    <pre><code>
    NAME                                STATUS   AGE
    capd-system                         Active   66m
    capi-kubeadm-bootstrap-system       Active   67m
    capi-kubeadm-control-plane-system   Active   67m
    capi-system                         Active   68m
    capi-webhook-system                 Active   68m
    cert-manager                        Active   70m
    default                             Active   71m
    eksa-system                         Active   65m
    etcdadm-bootstrap-provider-system   Active   67m
    etcdadm-controller-system           Active   67m
    <strong>flux-system                         Active   64m</strong>
    kube-node-lease                     Active   71m
    kube-public                         Active   71m
    kube-system                         Active   71m
    </code></pre>

    > **Note:** You can the **flux controller** deployed which takes care of syncing github repository to the cluster.

2. To verify that a cluster control plane is up and running, use the `kubectl` command to show that the control plane pods are all running.

    ```bash
    kubectl get po -A -l control-plane=controller-manager
    ```

    **Output**

    ```bash
    NAMESPACE                           NAME                                                             READY   STATUS    RESTARTS   AGE
    capd-system                         capd-controller-manager-659dd5f8bc-9w2ml                         2/2     Running   0          69m
    capi-kubeadm-bootstrap-system       capi-kubeadm-bootstrap-controller-manager-69889cb844-k6w9m       2/2     Running   0          70m
    capi-kubeadm-control-plane-system   capi-kubeadm-control-plane-controller-manager-6ddc66fb75-6hpvs   2/2     Running   0          69m
    capi-system                         capi-controller-manager-db59f5789-nx8zb                          2/2     Running   0          70m
    capi-webhook-system                 capi-controller-manager-64b8c548db-hcq4m                         2/2     Running   0          70m
    capi-webhook-system                 capi-kubeadm-bootstrap-controller-manager-68b8cc9759-v5hct       2/2     Running   0          70m
    capi-webhook-system                 capi-kubeadm-control-plane-controller-manager-7dc88f767d-pcfrj   2/2     Running   0          69m
    etcdadm-bootstrap-provider-system   etcdadm-bootstrap-provider-controller-manager-54476b7bf9-l44q9   2/2     Running   0          69m
    etcdadm-controller-system           etcdadm-controller-controller-manager-d5795556-bk6hz             2/2     Running   0          69m
    ```

3. To verify that the expected number of cluster worker nodes are up and running, use the `kubectl` command to show that nodes are Ready. This will confirm that the expected number of worker nodes, named following the format `$CLUSTERNAME-md-0`, are present.

    ```bash
    kubectl get nodes
    ```

    **Output**

    ```bash
    NAME                           STATUS   ROLES                  AGE   VERSION
    gitops-cqrdp                   Ready    control-plane,master   73m   v1.21.2-eks-1-21-4
    gitops-md-0-746dd56cf7-f78cx   Ready    <none>                 73m   v1.21.2-eks-1-21-4
    ```

### Github verification

You should see a github repository created by EKS Anywhere cluster with the same name specified in the cluster config file, like below:

![repo](../images/repo.png)