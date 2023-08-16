---
title: "Configure Fluent Bit for CloudWatch"
linkTitle: "Fluent Bit for CloudWatch"
weight: 90
date: 2023-08-11
description: >
  Using Fluent Bit for logging with EKS Anywhere clusters and CloudWatch
---

[Fluent Bit](https://docs.fluentbit.io/manual) is an open source, multi-platform log processor and forwarder which allows you to collect data/logs from different sources, then unify and send them to multiple destinations. It’s fully compatible with Docker and Kubernetes environments. Due to its lightweight nature, using Fluent Bit as the log forwarder for EKS Anywhere clusters enables you to stream application logs into Amazon CloudWatch Logs efficiently and reliably.

You can additionally use [CloudWatch Container Insights](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/ContainerInsights.html) to collect, aggregate, and summarize metrics and logs from your containerized applications and microservices running on EKS Anywhere clusters. CloudWatch automatically collects metrics for many resources, such as CPU, memory, disk, and network. Container Insights also provides diagnostic information, such as container restart failures, to help you isolate issues and resolve them quickly. You can also set CloudWatch alarms on metrics that Container Insights collects.

On this page, we show how to set up Fluent Bit and Container Insights to send logs and metrics from your EKS Anywhere clusters to CloudWatch.

#### Prerequisites

- An AWS Account (see [AWS documentation](https://docs.aws.amazon.com/accounts/latest/reference/welcome-first-time-user.html) to get started)
- An EKS Anywhere cluster with IAM Roles for Service Account (IRSA) enabled: With IRSA, an IAM role can be associated with a Kubernetes service account. This service account can provide AWS permissions to the containers in any Pod that use the service account, which enables the containers to securely communicate with AWS services. This removes the need to hardcode AWS security credentials as environment variables on your nodes. See the [IRSA configuration page]({{< relref "../../getting-started/optional/irsa/" >}}) for details.

{{% alert title="Note" color="primary" %}}
- The example uses `eksapoc` as the EKS Anywhere cluster name. You must adjust the configuration in the examples below if you use a different cluster name. Specifically, make sure to adjust the `fluentbit.yaml` manifest accordingly.
- The example uses the `us-west-2` AWS Region. You must adjust the configuration in the examples below if you are using a different region.
{{% /alert %}}

Before setting up Fluent Bit, first create an IAM Policy and Role to send logs to CloudWatch.

## Step 1: Create IAM Policy

1. Go to [IAM Policy](https://us-east-1.console.aws.amazon.com/iamv2/home?region=us-west-2#/policies/create?step=addPermissions) in the AWS console.
1. Click on JSON as shown below:

     ![Observability Create Policy](/images/observability_create_policy.png)
     
1. Create below policy on the IAM Console. Click on Create Policy as shown:

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

## Step 2: Create IAM Role

1. Go to [IAM Role](https://us-east-1.console.aws.amazon.com/iamv2/home?region=us-west-2#/roles/create?step=selectEntities) in the AWS console.

2. Follow the steps as shown below:

     ![Observability Role Creation](/images/observability_role_creation.png)

     In **Identity Provider**, enter the OIDC provider you created as a part of IRSA configuration.

     In **Audience**, select sts.amazonaws.com. Click on Next.

3. Select permission name which we have created in [Create IAM Policy](#step-1-create-iam-policy)

     ![Observability Select Permission](/images/observability_select_permission.png)

4. Provide a Role name `EKSAnywhereLogging` and click Next.

5. Copy the ARN as shown below and save it locally for the next step.

     ![Observability Copy ARN](/images/observability_arn_copy.png)

## Step 3: Install Fluent Bit

1. Create the `amazon-cloudwatch` namespace using this command:

    ```bash 
    kubectl create namespace amazon-cloudwatch
    ```

2. Create the Service Account for `cloudwatch-agent` and `fluent-bit` under the `amazon-cloudwatch` namespace. In this section, we will use Role ARN which we saved [earlier](#step-2-create-iam-role). Replace `$RoleARN` with your actual value.

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

     The above command creates two Service Accounts:

     ```
     serviceaccount/cloudwatch-agent created
     serviceaccount/fluent-bit created
     ```

3. Now deploy Fluent Bit in your EKS Anywhere cluster to scrape and send logs to CloudWatch:

     ```bash
     kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/fluentbit.yaml"
     ```

     You should see the following output:

     ```
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

4. You can verify the `DaemonSets` have been deployed with the following command:

     ```bash
     kubectl -n amazon-cloudwatch get daemonsets
     ```

- If you are running the [EKS connector]({{< relref "./cluster-connect" >}})
, you can verify the status of `DaemonSets` by logging into AWS console and navigate to Amazon EKS -> Cluster -> Resources -> DaemonSets

     ![Observability Verify DaemonSet](/images/observability_cluster_verify_Daemonset.png)

## Step 4: Deploy a test application

Deploy a simple [test application]({{< relref "../../workloadmgmt/test-app/" >}}) to verify your setup is working properly.

## Step 5: View cluster logs and metrics

### Cloudwatch Logs
1. Open the [CloudWatch console](https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#logsV2:log-groups). The link opens the console and displays your current available log groups.

2. Choose the EKS Anywhere clustername that you want to view logs for. The log group name format is /aws/containerinsights/`my-EKS-Anywhere-cluster`/cluster.

     ![Observability Container Insights](/images/observability_logGroups.png)

    Log group name `/aws/containerinsights/my-EKS-Anywhere-cluster/application` has log source from /var/log/containers.

    Log group name `/aws/containerinsights/my-EKS-Anywhere-cluster/dataplane` has log source for `kubelet.service`, `kubeproxy.service`, and `docker.service`

3. To view the deployed [test application](#step-4-deploy-a-test-application) logs, click on the application LogGroup, and click on Search All

     ![Observability Container Insights](/images/observability_search_logstream.png)

4. Type `HTTP 1.1 200` in the search box and press enter. You should see logs as shown below:

     ![Observability Container Insights](/images/observability_logGroups_filterlog.png)

### Cloudwatch Container Insights

1. Open the [CloudWatch console](https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#container-insights:performance). The link opens the Container Insights performance Monitoring console and displays a dropdown to select your `EKS Clusters`.

     ![Observability Container Insights](/images/observability_container_insights.png)

For more details on CloudWatch logs, please refer [What is Amazon CloudWatch Logs?](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/WhatIsCloudWatchLogs.html)
