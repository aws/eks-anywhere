---
toc_hide: true
hide_summary: true
---

### Test GitOps controller

After your cluster is created you can test the GitOps controller by modifying the cluster specification.

1. Clone your git repo and modify the cluster specification.
   The default path for the cluster file is:

    ```
    clusters/$CLUSTER_NAME/eksa-system/eksa-cluster.yaml
    ```

1. Modify the `workerNodeGroupsConfigurations[0].count` field with your desired changes.

1. Commit the file to your git repository

    ```bash
    git add eksa-cluster.yaml
    git commit -m 'Scaling nodes for test'
    git push origin main
    ```

1. The flux controller will automatically make the required changes.

   If you updated your node count you can use this command to see the current node state.
    ```bash
    kubectl get nodes 
    ```