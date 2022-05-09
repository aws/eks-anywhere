---
title: "Test deploying Pod with AWS Access"
linkTitle: "Test deploying Pod with AWS Access"
weight: 70
date: 2021-11-11
description: >  
---

1.	You get the arn of the IAM role `Sample-app-role` you created above.
   ![Import ova wizard](/images/sample-app-iam-role.png) 


1. Then, yougo to the workload cluster to create a service account in the namespace where youare going to run the pod that needs access to AWS and annotate it with arn of the IAM role
    ```bash
    export KUBECONFIG=${PWD}/${WORKLOAD_CLUSTER_NAME}/${WORKLOAD_CLUSTER_NAME}-eks-a-cluster.kubeconfig

    mkdir test & cd test
    cat <<EOF > test-sa.yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    name: test
    namespace: default
    annotations:
        eks.amazonaws.com/role-arn: arn:aws:iam::[AWS_ACCOUNT]:role/Sample-app-role
    EOF

    kubectl apply -f test-sa.yaml
    ```

1.	Then, go to AWS IAM console to check the trust policy of the IAM role youcreated if it has a reference for our service account. If it’s not there add it.
 

1.	Then, let’s create a job with the service account associated with it.

    ```bash
    cat <<EoF> job-s3.yaml
    apiVersion: batch/v1
    kind: Job
    metadata:
    name: eks-iam-test-s3
    spec:
    template:
        metadata:
        labels:
            app: eks-iam-test-s3
        spec:
        serviceAccountName: test
        containers:
        - name: eks-iam-test
            image: amazon/aws-cli:latest
            args: ["s3", "ls"]
        restartPolicy: Never
    EoF

    kubectl apply -f job-s3.yaml
    ```

1. Make sure your job is completed.

    ```bash
    kubectl get job -l app=eks-iam-test-s3
    ```
    Output:
    ```bash
    NAME              COMPLETIONS   DURATION   AGE
    eks-iam-test-s3   1/1           2s         21m
    ```
1.	Let’s check the logs to verify that the command listed the s3 buckets and ran successfully.
    ```bash
    kubectl logs -l app=eks-iam-test-s3
    ```
