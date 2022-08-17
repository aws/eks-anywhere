## Prerequisites
Before installing any curated packages for EKS Anywhere, do the following:

* Check that the version of `eksctl anywhere` is `v0.9.0` or above with the following commands:
    ```bash
    eksctl anywhere version
    ```
    Example command output
    ```
    NAME                                       READY   STATUS     RESTARTS   AGE
    eks-anywhere-packages-57778bc88f-587tq     2/2     Running    0          16h
    ```
* Make sure cert-manager is up and running in the cluster. Note cert-manager is not installed on workload clusters by default. If cert-manager is not installed, you can manually install cert-manager and follow the instructions below to finish the package controller installation.

* Check the existence of package controller:

    ```bash
    kubectl get pods -n eksa-packages | grep "eks-anywhere-packages"
    ```
    If the returned result is empty, you need to install the package controller.

* If the packages controller is not yet installed, follow these steps:

    1. Install the package controller
        ```bash
        export CURATED_PACKAGES_SUPPORT=true
        eksctl anywhere install packagecontroller -f $CLUSTER_NAME.yaml
        ```

    1. Check the package controller
        ```bash
        kubectl get pods -n eksa-packages
        ```
