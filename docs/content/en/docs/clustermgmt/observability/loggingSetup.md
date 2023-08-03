---
title: "Install Fluentbit and Configuring Kubernetes Service Account to assume an IAM Role"
linkTitle: "Fluentbit and Configure Kubernetes Service Account to assume IAM Role"
weight: 4
#aliases:
#    /docs/clustermgmt/oservability/
date: 2023-07-28
description: >  
---

Fluentbit is an open source and multi-platform Log Processor and Forwarder which allows you to collect data/logs from different sources, unify and send them to multiple destinations. It’s fully compatible with Docker and Kubernetes environments. Due to its lightweight nature, using Fluent Bit as the default log forwarder for EKS Anywhere nodes will allow you to stream application logs into CloudWatch Logs efficiently and reliably.

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
4.  Click on Create Policy as shown:

     ![Observability Policy Name](/images/observability_policy_name.png)

After performing these steps, lets create IAM Role.

### Create IAM Role

1. Click on [IAM Role](https://us-east-1.console.aws.amazon.com/iamv2/home?region=us-west-2#/roles/create?step=selectEntities)

2. Please follow the steps as shown below:

     ![Observability Role Creation](/images/observability_role_creation.png)

     In **Identity Provider**, mention your OIDC provider which you have created as a part of IRSA configuration.

     In **Audience**, select sts.amazonaws.com. Click on Next.

3. Select permission name which we have created in this [Section](#Create IAM Policy)

     ![Observability Select Permission](/images/observability_select_permission.png)

4. Provide Role name as shown below and click Next

     ![Observability Review Role](/images/observability_review_role.png)

5. Copy the ARN as shown below as save it locally for next step

     ![Observability Copy ARN](/images/observability_arn_copy.png)

### Install Fluentbit

1. Create the `amazon-cloudwatch` namespace using the below command

    ```bash 
    kubectl create namespace amazon-cloudwatch
    ```

2. Create the Service Account for `cloudwatch-agent` and `fluent-bit` under namespace `amazon-cloudwatch`. In this section, we will use Role ARN which we saved [earlier](#create-iam-role). Please replace `$RoleARN` with the actual value.

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

3. Now deploy fluentbit in EKS Anywhere cluster to scrape metrics and send it to CloudWatch:

     ```bash
     kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/fluentbit.yaml"
     ```

     You should see below output:

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

4. You can verify all the `DaemonSets` have been deployed by running the following command:

     ```bash
     kubectl -n amazon-cloudwatch get daemonsets
     ```

- If you are running an [EKS connector](https://anywhere.eks.amazonaws.com/docs/clustermgmt/observability/cluster-connect/)
, you can verify the status of `DaemonSets` by logging into AWS console and navigate to Amazon EKS -> Cluster -> Resources -> DaemonSets

     ![Observability Verify DaemonSet](/images/observability_cluster_verify_Daemonset.png)

### Deploy a test application

We’ve created a simple [test application](https://anywhere.eks.amazonaws.com/docs/workloadmgmt/test-app/) for you to verify your cluster is working properly.

## View Cluster Logs and Metrics

In the previous section, you have completed the basic setup for Observability in EKS Anywhere. Now lets move to the fun part of seeing the logs and metrics for our deployed test application.

### Cloudwatch Logs
1. Open the [CloudWatch console](https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#logsV2:log-groups). The link opens the console and displays your current available log groups.

2. Choose the EKS Anywhere clustername that you want to view logs for. The log group name format is /aws/containerinsights/`my-EKS-Anywhere-cluster`/cluster.

     ![Observability Container Insights](/images/observability_logGroups.png)

    Log group name `/aws/containerinsights/my-EKS-Anywhere-cluster/application` have log source from /var/log/containers.

    Log group name `/aws/containerinsights/my-EKS-Anywhere-cluster/dataplane` have log source for `kubelet.service`, `kubeproxy.service`, and `docker.service`

3. To view the deployed [test application](#deploy-a-test-application) logs, click on the application LogGroup, and click on Search All

     ![Observability Container Insights](/images/observability_search_logstream.png)

4. Now, type `HTTP 1.1 200` in the search box and enter, you should see below logs

     ![Observability Container Insights](/images/observability_logGroups_filterlog.png)

### Cloudwatch Container Insights

1. Open the [CloudWatch console](https://us-west-2.console.aws.amazon.com/cloudwatch/home?region=us-west-2#container-insights:performance). The link opens the Container Insights performance Monitoring console and displays dropdown to select your `EKS Cluster`.

     ![Observability Container Insights](/images/observability_container_insights.png)

For more details on CloudWatch logs please refer [here](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/WhatIsCloudWatchLogs.html)