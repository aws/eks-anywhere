---
title: "Configure Kubernetes Audit Policy"
linkTitle: "Configure Audit Policy"
weight: 25
aliases:
    /docs/tasks/cluster/audit-policy/
date: 2025-01-05
description: >
  Configure Kubernetes audit policy for control plane nodes to enable comprehensive logging and monitoring
---

## Kubernetes Audit Policy Support

EKS Anywhere configures a default audit policy for all clusters to provide basic logging and monitoring of API server requests. This default policy covers essential security events and resource access patterns.

{{% alert title="Note" color="primary" %}}
All EKS Anywhere clusters include audit logging with a sensible default policy. The `auditPolicyContent` field is only needed if you want to customize the audit policy beyond the default configuration.
{{% /alert %}}

## Customizing Audit Policy (Optional)

If you need to customize the audit policy beyond the default configuration, you can override it by adding the `auditPolicyContent` field to the `controlPlaneConfiguration` section of your cluster configuration:

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
spec:
  controlPlaneConfiguration:
    count: 1
    endpoint:
      host: "192.168.1.100"
    machineGroupRef:
      kind: VSphereMachineConfig
      name: my-cluster-control-plane
    auditPolicyContent: |
      apiVersion: audit.k8s.io/v1
      kind: Policy
      rules:
        - level: RequestResponse
          resources:
            - group: ""
              resources:
                - pods
                - services
                - secrets
                - configmaps
```

## Updating Audit Policy

To modify the audit policy on an existing cluster:

1. Add/Update the `auditPolicyContent` in your cluster configuration file
2. Run the cluster upgrade command:

```bash
eksctl anywhere upgrade cluster -f my-cluster.yaml
```

The upgrade process will rollout all control plane nodes with updated audit policy configuration.
