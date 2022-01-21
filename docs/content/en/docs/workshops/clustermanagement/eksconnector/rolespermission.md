---
title: "Roles and permissions setup"
linkTitle: "Roles and permissions setup"
weight: 2
date: 2021-11-11
description: >  
---

Part of this article we will see how to connect and EKS Anywhere cluster to an AWS EKS console.

## Setup

### Roles and permissions

Using the Amazon EKS Connector requires the following two IAM roles, which you will have to create.

1. **Role 1** - The Amazon EKS Connector service-linked IAM role. For more information, see [Amazon EKS Connector role](https://docs.aws.amazon.com/eks/latest/userguide/using-service-linked-roles-eks-connector.html). You can create the role with this command:

    ```bash
    aws iam create-service-linked-role --aws-service-name eks-connector.amazonaws.com
    ```

    **Output:**

    ```json
    {
        "Role": {
            "Path": "/aws-service-role/eks-connector.amazonaws.com/",
            "RoleName": "AWSServiceRoleForAmazonEKSConnector",
            "RoleId": "AROATBEEL4RMPIQQP3I7A",
            "Arn": "arn:aws:iam:::role/aws-service-role/eks-connector.amazonaws.com/AWSServiceRoleForAmazonEKSConnector",
            "CreateDate": "2021-09-19T15:16:33+00:00",
            "AssumeRolePolicyDocument": {
                "Version": "2012-10-17",
                "Statement": [
                    {
                        "Action": [
                            "sts:AssumeRole"
                        ],
                        "Effect": "Allow",
                        "Principal": {
                            "Service": [
                                "eks-connector.amazonaws.com"
                            ]
                        }
                    }
                ]
            }
        }
    }
    ```

2. **Role 2** - The IAM role for the Amazon EKS Connector agent. You can create the role with the following steps:

    a. Create a file named `eks-connector-agent-trust-policy.json` that contains the following JSON to use for the IAM role.

    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "SSMAccess",
                "Effect": "Allow",
                "Principal": {
                    "Service": [
                        "ssm.amazonaws.com"
                    ]
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }
    ```

    b. Create a file named `eks-connector-agent-policy.json` that contains the following JSON to use for the IAM role.

    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "SsmControlChannel",
                "Effect": "Allow",
                "Action": [
                    "ssmmessages:CreateControlChannel"
                ],
                "Resource": "arn:aws:eks:*:*:cluster/*"
            },
            {
                "Sid": "ssmDataplaneOperations",
                "Effect": "Allow",
                "Action": [
                    "ssmmessages:CreateDataChannel",
                    "ssmmessages:OpenDataChannel",
                    "ssmmessages:OpenControlChannel"
                ],
                "Resource": "*"
            }
        ]
    }
    ```

    c. Create the Amazon EKS Connector agent role using the trust policy and policy you created in the previous steps.

    ```bash
    aws iam create-role \
        --role-name AmazonEKSConnectorAgentRole \
        --assume-role-policy-document file://eks-connector-agent-trust-policy.json
    ```

    **Output:**

    ```json
    {
        "Role": {
            "Path": "/",
            "RoleName": "AmazonEKSConnectorAgentRole",
            "RoleId": "AROATBEEL4RMJ7BJOXY4N",
            "Arn": "arn:aws:iam::1234567890:role/AmazonEKSConnectorAgentRole",
            "CreateDate": "2021-09-19T15:37:11+00:00",
            "AssumeRolePolicyDocument": {
                "Version": "2012-10-17",
                "Statement": [
                    {
                        "Sid": "SSMAccess",
                        "Effect": "Allow",
                        "Principal": {
                            "Service": [
                                "ssm.amazonaws.com"
                            ]
                        },
                        "Action": "sts:AssumeRole"
                    }
                ]
            }
        }
    }
    ```

    d. Attach the policy to your Amazon EKS Connector agent role.

    ```bash
    aws iam put-role-policy \
        --role-name AmazonEKSConnectorAgentRole \
        --policy-name AmazonEKSConnectorAgentPolicy \
        --policy-document file://eks-connector-agent-policy.json
    ```
