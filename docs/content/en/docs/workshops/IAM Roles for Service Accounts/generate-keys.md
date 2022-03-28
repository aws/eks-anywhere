---
title: "Generating keys.json and Make it Publicly Accessible"
linkTitle: "Generating keys.json and Make it Publicly Accessible"
weight: 50
date: 2021-11-11
description: >  
---

1. The cluster provisioning workflow generates a pair of service account signing keys. Retrieve the public signing key generated and used by the workload cluster w01-cluster, but located on the management cluster mgmt.-cluster, and create a keys.json document containing the public signing key.

    ```bash
    kubectl get secret ${WORKLOAD_CLUSTER_NAME}-sa --ignore-not-found \
    -n eksa-system --kubeconfig \
    ${MGMT_CLUSTER_NAME}/${MGMT_CLUSTER_NAME}-eks-a-cluster.kubeconfig \
    -o jsonpath={.data.tls\\.crt} | base64 \
    --decode > ${WORKLOAD_CLUSTER_NAME}-sa.pub

    git clone https://github.com/aws/amazon-eks-pod-identity-webhook.git
    cd amazon-eks-pod-identity-webhook/
    git checkout a65cc3d9c61cf6fc43f0f985818c474e0867d786

    wget https://raw.githubusercontent.com/aws/amazon-eks-pod-identity-webhook/master/hack/self-hosted/main.go -O keygenerator.go

    sudo go run keygenerator.go -key ../${WORKLOAD_CLUSTER_NAME}-sa.pub \
    | jq '.keys += [.keys[0]] | .keys[1].kid = ""' > ../keys.json
    ```
    
1. Upload the keys.json document to the s3 bucket.
    ```bash
    aws s3 cp --acl public-read ../keys.json s3://$S3_BUCKET/keys.json
    ```
