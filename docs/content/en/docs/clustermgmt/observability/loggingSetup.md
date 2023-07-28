---
title: "Install Fluentbit and Configuring Kubernetes Service Account to assume an IAM Role"
linkTitle: "Fluentbit and Configure Kubernetes Service Account to assume IAM Role"
weight: 3
#aliases:
#    /docs/clustermgmt/oservability/
date: 2023-07-28
description: >  
---

Fluentbit is an open source and multi-platform Log Processor and Forwarder which allows you to collect data/logs from different sources, unify and send them to multiple destinations. It’s fully compatible with Docker and Kubernetes environments. Due to its lightweight nature, using Fluent Bit as the default log forwarder for EKS Anywhere nodes will allow you to stream application logs into CloudWatch logs efficiently and reliably.

Before setting up fluentbit, first create right IAM Policy and Role to send logs to CloudWatch.

### Create IAM Policy

1. Click on [IAM Policy](https://us-east-1.console.aws.amazon.com/iamv2/home?region=us-west-2#/policies/create?step=addPermissions);
1. Click on JSON as shown below:

     ![Observability Create Policy](/images/observability_create_policy.png)
     
1. Create policy on the IAM Console as shown below:

     ![Observability Policy JSON](/images/observability_policy_json.png)

```
        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Sid": "EKSAnywhereLogging",
                    "Effect": "Allow",
                    "Action": "cloudwatch:*",
                    "Resource": "*"
                }
            ]
        }
```
1. Click on Create Policy as shown:

     ![Observability Policy Name](/images/observability_policy_name.png)

After performing these steps, lets create IAM Role.

### Create IAM Role

1. Click on [IAM Role](https://us-east-1.console.aws.amazon.com/iamv2/home?region=us-west-2#/roles/create?step=selectEntities)

1. Please follow the steps as shown below:

     ![Observability Role Creation](/images/observability_role_creation.png)

In identity provider, please mention your OIDC provider which you have created as a part of IRSA configuration.

In Audience, please select sts.amazonaws.com. Click on Next.

1. Select permission name which we have created in this [Section](#Create IAM Policy)

     ![Observability Select Permission](/images/observability_select_permission.png)

1. Provide Role name as shown below and click Next

     ![Observability Review Role](/images/observability_review_role.png)

1. Copy the ARN as shown below as save it locally for next step

     ![Observability Copy ARN](/images/observability_arn_copy.png)

### Install Fluentbit

Now deploy fluentbit in EKS Anywhere cluster to scrap metrics and send it to CloudWatch:

```bash
kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/fluentbit.yaml"
```

You should see below output:

```
namespace/amazon-cloudwatch changed
clusterrole.rbac.authorization.k8s.io/cloudwatch-agent-role changed
clusterrolebinding.rbac.authorization.k8s.io/cloudwatch-agent-role-binding changed
configmap/cwagentconfig changed
daemonset.apps/cloudwatch-agent changed
configmap/fluent-bit-cluster-info changed
clusterrole.rbac.authorization.k8s.io/fluent-bit-role changed
clusterrolebinding.rbac.authorization.k8s.io/fluent-bit-role-binding changed
configmap/fluent-bit-config changed
daemonset.apps/fluent-bit changed
```

Fluentbit.yaml has created all the required resources except service account. In next steps, we will create Service Account for `cloudwatch-agent` and `fluent-bit` under namespace `amazon-cloudwatch`. In this section, we will use Role ARN which we saved [earlier](#create-iam-role). Please replace `$RoleARN` with the actual value.

```
cat << EOF | kubectl apply -f -
# create cwagent service account and role binding
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloudwatch-agent
  namespace: amazon-cloudwatch
  annotations:
    # set this with value of OIDC_IAM_ROLE
    eks.amazonaws.com/role-arn: "$RoleARN"
    # optional: Defaults to "sts.amazonaws.com" if not set
    eks.amazonaws.com/audience: "sts.amazonaws.com"
    # optional: When set to "true", adds AWS_STS_REGIONAL_ENDPOINTS env var
    #   to containers
    eks.amazonaws.com/sts-regional-endpoints: "true"
    # optional: Defaults to 86400 for expirationSeconds if not set
    #   Note: This value can be overwritten if specified in the pod
    #         annotation as shown in the next step.
    eks.amazonaws.com/token-expiration: "86400"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluent-bit
  namespace: amazon-cloudwatch
  annotations:
    # set this with value of OIDC_IAM_ROLE
    eks.amazonaws.com/role-arn: "$RoleARN"
    # optional: Defaults to "sts.amazonaws.com" if not set
    eks.amazonaws.com/audience: "sts.amazonaws.com"
    # optional: When set to "true", adds AWS_STS_REGIONAL_ENDPOINTS env var
    #   to containers
    eks.amazonaws.com/sts-regional-endpoints: "true"
    # optional: Defaults to 86400 for expirationSeconds if not set
    #   Note: This value can be overwritten if specified in the pod
    #         annotation as shown in the next step.
    eks.amazonaws.com/token-expiration: "86400"
EOF
```
The above command will create two service account with below output:

```
serviceaccount/cloudwatch-agent created
serviceaccount/fluent-bit created
```

