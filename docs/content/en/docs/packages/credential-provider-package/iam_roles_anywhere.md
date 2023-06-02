---
title: "Credential Provider Package with IAM Roles Anywhere"
linkTitle: "Credential Provider Package with IAM Roles Anywhere"
weight: 2
aliases:
    /docs/workshops/packages/credential-provider-package/iam_roles_anywhere/
date: 2023-03-31
description: >
---

This tutorial demonstrates how to configure the credential provider package to authenticate using [IAM Roles Anywhere](https://docs.aws.amazon.com/rolesanywhere/latest/userguide/introduction.html) to pull from a private AWS Elastic Container Registry (ECR).  
IAM Roles Anywhere enables workloads outside of AWS to access AWS resources by using X.509 digital certificates to obtain temporary AWS credentials. A trust anchor is used to reference a certificate authority with IAM Roles Anywhere. For this use case, the **Kubernetes Cluster CA** can be registered and each kubelet client's x509 cert can be used to authenticate to get temporary AWS credentials.

## Prerequisites 
1. For setting up the certificate authority later, you will need to obtain your cluster's CA. This can be obtain by:
    ```bash
    # Assuming CLUSTER_NAME and KUBECONFIG are set:
    kubectl get secret -n eksa-system ${CLUSTER_NAME}-ca -o yaml | yq '.data."tls.crt"' | base64 -d
    ```
1. A role should be created to allow read access for curated packages. This role can be extended to include private registries that you would also like to pull from into your cluster. A sample policy for curated packages would be.
   ```
   {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "ECRRead",
            "Effect": "Allow",
            "Action": [
                "ecr:DescribeImageScanFindings",
                "ecr:GetDownloadUrlForLayer",
                "ecr:DescribeRegistry",
                "ecr:DescribePullThroughCacheRules",
                "ecr:DescribeImageReplicationStatus",
                "ecr:ListTagsForResource",
                "ecr:ListImages",
                "ecr:BatchGetImage",
                "ecr:DescribeImages",
                "ecr:DescribeRepositories",
                "ecr:BatchCheckLayerAvailability"
            ],
            "Resource": "arn:aws:ecr:*:783794618700:repository/*"
        },
        {
            "Sid": "ECRLogin",
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken"
            ],
            "Resource": "*"
        }
     ]
    }
   ```

1. Next create a trust anchor and profile. The trust anchor will be a reference to the CA certificate from step 1 and the profile should point to the role created in step 2. See [here](https://docs.aws.amazon.com/rolesanywhere/latest/userguide/getting-started.html) for instructions on creating the trust anchor and profile.

1. Create a secret that will be referenced by the credential-provider-package to authenticate the kubelet with ECR. 
    ```bash
    # Set PROFILE_ARN, ROLE_ARN, and TRUST_ANCHOR_ARN obtained in previous step
    # Set AWS_REGION to region to pull images from
    # This will create a file credfile which will then be turned into a secret
    cat << EOF >> credfile
    [default]
    region = $AWS_REGION
    credential_process = aws_signing_helper credential-process --certificate /var/lib/kubelet/pki/kubelet-client-current.pem --private-key /var/lib/kubelet/pki/kubelet-client-current.pem --profile-arn $PROFILE_ARN --role-arn $ROLE_ARN --trust-anchor-arn $TRUST_ANCHOR_ARN
    EOF
    
    # Create secret, for this example the secret name aws-config is used and the package will be installed in eksa-packages
    kubectl create secret generic aws-config --from-file=config=credfile -n eksa-packages
    ```

1. Either edit the existing package or delete and create a new credential-provider-package that points towards the new secret. For more information on specific configuration option refer to [installation guide]({{< relref "../credential-provider-package" >}}) for details]  
   The example below changes the default secret name from aws-secret to newly created aws-config. It also changes the match images to pull from multiple regions as well as across multiple accounts. Make sure to change cluster-name to match your CLUSTER_NAME
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: my-credential-provider-package
      namespace: eksa-packages-<clusterName>
      annotations:
        "helm.sh/resource-policy": keep
        "anywhere.eks.aws.com/internal": "true"
    spec:
      packageName: credential-provider-package
      targetNamespace: eksa-packages
      config: |-
        tolerations:
          - key: "node-role.kubernetes.io/master"
            operator: "Exists"
            effect: "NoSchedule"
          - key: "node-role.kubernetes.io/control-plane"
            operator: "Exists"
            effect: "NoSchedule"
        sourceRegistry: public.ecr.aws/eks-anywhere
        credential:
          - matchImages:
            - *.dkr.ecr.*.amazonaws.com
            profile: "default"
            secretName: aws-config
            defaultCacheDuration: "5h"
    ```
