# Patch CAPT Manager OOM

Cluster API Tinkerbell Provider (CAPT) Manager is a controller pod deployed to the cluster for bare metal provisioning. It sometimes experiences OOM due to insufficienr memory limits on the pod definition.

Patch the memory limit with increased resources.

Save following patch file contents.

```
spec:
  template:
    spec:
      containers:
      - name: manager
        resources:
          limits:
            cpu: 200m
            memory: 300Mi
```

Apply the patch.

```
$ kubectl patch -n capt-system deployment capt-controller-manager --patch-file capt-manager.patch
```