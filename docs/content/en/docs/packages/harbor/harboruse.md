---
title: "Harbor use cases"
linkTitle: "Harbor use cases"
weight: 20
date: 2022-04-12
description: >  
  Try some harbor use cases
---

{{% alert title="Important" color="warning" %}}

To install Harbor package, please follow the [installation guide.]({{< relref "./addharbor" >}})

{{% /alert %}}

## Proxy a public Amazon Elastic Container Registry (ECR) repository
This use case is to use Harbor to proxy and cache images from a public ECR repository, which helps limit the amount of requests made to a public ECR repository, avoiding consuming too much bandwidth or being throttled by the registry server.

1. Login

    Log in to the Harbor web portal with the default credential as shown below
    ```bash
    admin
    Harbor12345
    ```
 
    ![Harbor web portal](/images/harbor-portal.png)

1. Create a registry proxy

    Navigate to `Registries` on the left panel, and then click on `NEW ENDPOINT` button. Choose `Docker Registry` as the Provider, and enter `public-ecr` as the Name, and enter `https://public.ecr.aws/` as the Endpoint URL. Save it by clicking on OK.

    ![Harbor public ecr proxy](/images/harbor-public_ecr_proxy.png)

1. Create a proxy project

    Navigate to `Projects` on the left panel and click on the `NEW PROJECT` button. Enter `proxy-project` as the Project Name, check `Public access level`, and turn on Proxy Cache and choose `public-ecr` from the pull-down list. Save the configuration by clicking on OK.

    ![Harbor public proxy project](/images/harbor-public_proxy_project.png)

1. Pull images
   {{% alert title="Note" color="primary" %}}

   * `harbor.eksa.demo:30003` should be replaced with whatever `externalURL` is set to in the Harbor package YAML file.

   {{% /alert %}}
    ```bash
    docker pull harbor.eksa.demo:30003/proxy-project/cloudwatch-agent/cloudwatch-agent:latest
    ```

## Proxy a private Amazon Elastic Container Registry (ECR) repository
This use case is to use Harbor to proxy and cache images from a private ECR repository, which helps limit the amount of requests made to a private ECR repository, avoiding consuming too much bandwidth or being throttled by the registry server.

1. Login

    Log in to the Harbor web portal with the default credential as shown below
    ```bash
    admin
    Harbor12345
    ```
 
    ![Harbor web portal](/images/harbor-portal.png)

1. Create a registry proxy

    In order for Harbor to proxy a remote private ECR registry, an IAM credential with necessary permissions need to be created. Usually, it follows three steps: 

    1. Policy

        This is where you specify all necessary permissions. Please refer to [private repository policies](https://docs.aws.amazon.com/AmazonECR/latest/userguide/repository-policies.html), [IAM permissions for pushing an image](https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-push.html) and [ECR policy examples](https://docs.aws.amazon.com/AmazonECR/latest/userguide/security_iam_id-based-policy-examples.html) to figure out the minimal set of required permissions.

        For simplicity, the build-in policy AdministratorAccess is used here.

        ![Harbor private ecr policy](/images/harbor-private_ecr_policy.png)

    1. User group

        This is an easy way to manage a pool of users who share the same set of permissions by attaching the policy to the group.

        ![Harbor private ecr user group](/images/harbor-private_ecr_user_group.png)

    1. User

        Create a user and add it to the user group. In addition, please navigate to Security credentials to generate an access key. Access keys consists of two parts: an access key ID and a secret access key. Please save both as they are used in the next step. 

        ![Harbor private ecr user](/images/harbor-private_ecr_user.png)

    Navigate to `Registries` on the left panel, and then click on `NEW ENDPOINT` button. Choose `Aws ECR` as Provider, and enter `private-ecr` as Name, `https://[ACCOUNT NUMBER].dkr.ecr.us-west-2.amazonaws.com/` as Endpoint URL, use the access key ID part of the generated access key as Access ID, and use the secret access key part of the generated access key as Access Secret. Save it by click on OK.

    ![Harbor private ecr proxy](/images/harbor-private_ecr_proxy.png)

1. Create a proxy project

    Navigate to `Projects` on the left panel and click on `NEW PROJECT` button. Enter `proxy-private-project` as Project Name, check `Public access level`, and turn on Proxy Cache and choose `private-ecr` from the pull-down list. Save the configuration by clicking on OK.

    ![Harbor private proxy project](/images/harbor-private_proxy_project.png)

1. Pull images

    Create a repository in the target private ECR registry

    ![Harbor private ecr repository](/images/harbor-private_ecr_repository.png)

    Push an image to the created repository

    ```bash
    docker pull alpine
    docker tag alpine [ACCOUNT NUMBER].dkr.ecr.us-west-2.amazonaws.com/alpine:latest
    docker push [ACCOUNT NUMBER].dkr.ecr.us-west-2.amazonaws.com/alpine:latest
    ``` 


   {{% alert title="Note" color="primary" %}}

   * `harbor.eksa.demo:30003` should be replaced with whatever `externalURL` is set to in the Harbor package YAML file.

   {{% /alert %}}
    ```bash
    docker pull harbor.eksa.demo:30003/proxy-private-project/alpine:latest
    ```

## Repository replication from Harbor to a private Amazon Elastic Container Registry (ECR) repository
This use case is to use Harbor to replicate local images and charts to a private ECR repository in push mode. When a replication rule is set, all resources that match the defined filter patterns are replicated to the destination registry when the triggering condition is met.

1. Login

    Log in to the Harbor web portal with the default credential as shown below
    ```bash
    admin
    Harbor12345
    ```
 
    ![Harbor web portal](/images/harbor-portal.png)

1. Create a nonproxy project

   ![Harbor nonproxy project](/images/harbor-nonproxy_project.png) 

1. Create a registry proxy

    In order for Harbor to proxy a remote private ECR registry, an IAM credential with necessary permissions need to be created. Usually, it follows three steps: 

    1. Policy

        This is where you specify all necessary permissions. Please refer to [private repository policies](https://docs.aws.amazon.com/AmazonECR/latest/userguide/repository-policies.html), [IAM permissions for pushing an image](https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-push.html) and [ECR policy examples](https://docs.aws.amazon.com/AmazonECR/latest/userguide/security_iam_id-based-policy-examples.html) to figure out the minimal set of required permissions.

        For simplicity, the build-in policy AdministratorAccess is used here.

        ![Harbor private ecr policy](/images/harbor-private_ecr_policy.png)

    1. User group

        This is an easy way to manage a pool of users who share the same set of permissions by attaching the policy to the group.

        ![Harbor private ecr user group](/images/harbor-private_ecr_user_group.png)

    1. User

        Create a user and add it to the user group. In addition, please navigate to Security credentials to generate an access key. Access keys consists of two parts: an access key ID and a secret access key. Please save both as they are used in the next step. 

        ![Harbor private ecr user](/images/harbor-private_ecr_user.png)

    Navigate to `Registries` on the left panel, and then click on the `NEW ENDPOINT` button. Choose `Aws ECR` as the Provider, and enter `private-ecr` as the Name, `https://[ACCOUNT NUMBER].dkr.ecr.us-west-2.amazonaws.com/` as the Endpoint URL, use the access key ID part of the generated access key as Access ID, and use the secret access key part of the generated access key as Access Secret. Save it by clicking on OK.

    ![Harbor private ecr proxy](/images/harbor-private_ecr_proxy.png)

1. Create a replication rule

    ![Harbor replication rule](/images/harbor-replication_rule.png)

1. Prepare an image
   {{% alert title="Note" color="primary" %}}

   * `harbor.eksa.demo:30003` should be replaced with whatever `externalURL` is set to in the Harbor package YAML file.

   {{% /alert %}}
    ```bash
    docker pull alpine
    docker tag alpine:latest harbor.eksa.demo:30003/nonproxy-project/alpine:latest
    ```

1. Authenticate with Harbor with the default credential as shown below
    ```bash
    admin
    Harbor12345
    ```
   {{% alert title="Note" color="primary" %}}

   * `harbor.eksa.demo:30003` should be replaced with whatever `externalURL` is set to in the Harbor package YAML file.

   {{% /alert %}} 
    ```bash
    docker logout
    docker login harbor.eksa.demo:30003
    ```

1. Push images

    Create a repository in the target private ECR registry

    ![Harbor private ecr repository](/images/harbor-private_ecr_repository.png)


   {{% alert title="Note" color="primary" %}}

   * `harbor.eksa.demo:30003` should be replaced with whatever `externalURL` is set to in the Harbor package YAML file.

   {{% /alert %}} 
    ```bash
    docker push harbor.eksa.demo:30003/nonproxy-project/alpine:latest
    ```

    The image should appear in the target ECR repository shortly.

    ![Harbor replication result](/images/harbor-replication_result.png)

## Set up trivy image scanner in an air-gapped environment
This use case is to manually import vulnerability database to Harbor trivy when Harbor is running in an air-gapped environment. All the following commands are assuming Harbor is running in the default namespace.

1. Configure trivy

    TLS example with auto certificate generation
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
       name: my-harbor
       namespace: eksa-packages
    spec:
       packageName: harbor
       config: |-
         secretKey: "use-a-secret-key"
         externalURL: https://harbor.eksa.demo:30003
         expose:
           tls:
             certSource: auto
             auto:
               commonName: "harbor.eksa.demo"
           trivy:
             skipUpdate: true
             offlineScan: true
    ```

    Non-TLS example
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
       name: my-harbor
       namespace: eksa-packages
    spec:
       packageName: harbor
       config: |-
         secretKey: "use-a-secret-key"
         externalURL: http://harbor.eksa.demo:30002
         expose:
           tls:
             enabled: false
         trivy:
           skipUpdate: true
           offlineScan: true
    ```

    If Harbor is already running without the above trivy configurations, run the following command to update both `skipUpdate` and `offlineScan`
    ```bash
    kubectl edit statefulsets/harbor-helm-trivy
    ```

1. Download the vulnerability database to your local host

    Please follow [oras installation instruction](https://oras.land/cli/).
    ```bash
    oras pull ghcr.io/aquasecurity/trivy-db:2 -a
    ```

1. Upload database to trivy pod from your local host
    ```bash
    kubectl cp db.tar.gz harbor-helm-trivy-0:/home/scanner/.cache/trivy -c trivy
    ```

1. Set up database on Harbor trivy pod
    ```bash
    kubectl exec -it harbor-helm-trivy-0 -c trivy bash
    cd /home/scanner/.cache/trivy
    mkdir db
    mv db.tar.gz db
    cd db
    tar zxvf db.tar.gz
    ```
