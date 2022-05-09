---
title: "IAM Roles for Service Accounts"
linkTitle: "IAM Roles for Service Accounts"
weight: 20
date: 2021-11-11
description: >  
---

In Kubernetes version 1.12, support was added for a new `ProjectedServiceAccountToken` feature, which is an OIDC JSON web token that also contains the service account identity, and supports a configurable audience.

By enabling IAM for pods comonents on an EKS Anywherecluster, you enable your EKS-A cluster to host a public OIDC discovery endpoint containing the signing keys for the ProjectedServiceAccountToken JSON web tokens so external systems, like IAM, can validate and accept the Kubernetes-issued OIDC tokens.

OIDC federation access allows you to assume IAM roles via the Secure Token Service (STS), enabling authentication with an OIDC provider, receiving a JSON Web Token (JWT), which in turn can be used to assume an IAM role. Kubernetes, on the other hand, can issue so-called projected service account tokens, which happen to be valid OIDC JWTs for pods. Our setup equips each pod with a cryptographically-signed token that can be verified by STS against the OIDC provider of your choice to establish the pod’s identity.

new credential provider `sts:AssumeRoleWithWebIdentity`