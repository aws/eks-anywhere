---
title: "Overview"
linkTitle: "Overview"
weight: 20
date: 2021-11-11
description: >  
---

IAM Roles for Service Account (IRSA) enables applications running in clusters to authenticate with AWS services using IAM roles. The current solution for leveraging this in EKS Anywhere involves creating your own OIDC provider for the cluster, and hosting your cluster’s public service account signing key. The public keys along with the OIDC discovery document should be hosted somewhere that AWS STS can discover it. The steps below assume the keys will be hosted on a publicly accessible S3 bucket. Refer this doc to ensure that the s3 bucket is publicly accessible.

   ![Import ova wizard](/images/irsa-arch.png) 

1. When you launch an application on kubernetes with kubectl apply -f application-job.yaml, the yaml manifest is submitted to the API server with the Amazon EKS Pod Identity webhook configured.
1. Kubernetes uses the service account set via serviceAccountName
1. Since the service account has the annotation passed "eks.amazonaws.com/role-arn" in serviceaccount.yaml the webhook injects the necessary environment variables (AWS_ROLE_ARN and AWS_WEB_IDENTITY_TOKEN) and sets up aws-iam-token projected volumes.
1. When application calls out to s3 to do any s3 operations the AWS SDK youuse in the application code base performs STS assume role with web identity performs assume role that has s3 permissions attached. It receives temporary credentials that it uses to complete the S3 operation.



