---
title: "Admission Webhook Protection"
linkTitle: "Admission Webhook Protection"
weight: 10
description: >
  Preventing custom admission webhooks from interfering with cluster operations
---

Custom admission webhooks installed in Kubernetes clusters can interfere with cluster upgrade operations and normal cluster functioning. EKS Anywhere provides an admission webhook protection mechanism to prevent webhook interference with system operations.

## Overview

The admission webhook protection feature uses the upstream EKS-Distro admission webhook exclusion mechanism that AWS EKS uses to protect system operations during cluster upgrades and normal operations. When enabled, custom admission webhooks are prevented from intercepting operations on specific system resources.

Faulty or misconfigured admission webhooks can block cluster upgrades, disrupt networking components such as Cilium, or prevent cluster self-healing by targeting control plane resources. This feature protects critical operations without modifying existing webhook configurations.

{{% alert title="Note" color="primary" %}}
This feature is only available for Kubernetes version 1.31 and later. It requires kubeadm v1beta4 API support for configuring API server environment variables.
{{% /alert %}}

## Configuration

Enable admission webhook protection by adding the `skipAdmissionForSystemResources` field to your cluster specification:

```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
spec:
  controlPlaneConfiguration:
    skipAdmissionForSystemResources: true
    count: 3
    endpoint:
      host: "198.51.100.10"
    machineGroupRef:
      kind: VSphereMachineConfig
      name: my-cluster-cp
  kubernetesVersion: "1.33"
```

### Default Behavior

- **Current default**: `nil` or `false` (opt-in - feature disabled by default)
- **Recommendation**: Enable this feature for all production clusters to protect critical system operations

### Requirements

- **Kubernetes version**: 1.31 or later (requires kubeadm v1beta4 API for `extraEnvs` support)
- **Cluster lifecycle**: Can be enabled during cluster creation or via cluster upgrade
- **Supported providers**: All EKS Anywhere providers

## How It Works

When enabled, the feature configures the Kubernetes API server with exclusion rules that bypass custom webhook processing for specific critical operations.

### Technical Implementation

1. **Exclusion Rules File**: Deploys `/etc/kubernetes/admission-plugin-exclusion-rules.json` to control plane nodes containing targeted exclusion rules

2. **API Server Configuration**: Sets environment variable `EKS_PATCH_EXCLUSION_RULES_FILE=/etc/kubernetes/admission-plugin-exclusion-rules.json`

3. **Upstream Patch**: The EKS-Distro Kubernetes distribution includes an admission webhook exclusion patch that reads these rules and bypasses custom webhook processing for matching operations

## Protected Resources

The exclusion rules protect specific critical operations based on resource type, namespace, name patterns, and requesting user identity:

### Control Plane Operations
- **Leader election leases**: kube-controller-manager, kube-scheduler, cilium-operator, cert-manager in kube-system namespace
- **ServiceAccount operations**: Token requests and ServiceAccount management by kube-controller-manager
- **Node leases**: All node heartbeat leases in kube-node-lease namespace

### RBAC and API Registration
- **Cluster-level RBAC**: ClusterRoles and ClusterRoleBindings modified by system:apiserver
- **Namespace-level RBAC**: Roles and RoleBindings in kube-system and kube-public by system:apiserver
- **APIService registrations**: Core Kubernetes API group registrations by system:apiserver

### Critical Networking Components
- **Cilium**: DaemonSet and operator deployments 
- **CoreDNS**: Deployment operations 
- **kube-proxy**: DaemonSet operations

### Flow Control Resources
- **FlowSchema**: Flow control schema by system:apiserver
- **PriorityLevelConfiguration**: Priority level configuration by system:apiserver

### System Pods
- **DaemonSet Pods**: Pods in kube-system
- **ReplicaSet Pods**: Pods in kube-system

For the complete list of exclusion rules, see the [admission-plugin-exclusion-rules.json](https://github.com/aws/eks-anywhere/blob/main/pkg/providers/common/config/admission-plugin-exclusion-rules.json) file in the repository.

## Enabling the Feature

### New Cluster Creation

Add the configuration when creating a new cluster:

```bash
eksctl anywhere create cluster -f cluster.yaml
```

Where `cluster.yaml` includes:

```yaml
spec:
  controlPlaneConfiguration:
    skipAdmissionForSystemResources: true
```

### Existing Cluster Upgrade

Enable the feature by upgrading an existing cluster:

1. Edit your cluster specification to add the field:

```yaml
spec:
  controlPlaneConfiguration:
    skipAdmissionForSystemResources: true
```

2. Apply the upgrade:

```bash
eksctl anywhere upgrade cluster -f cluster.yaml
```

The feature will be activated during the control plane upgrade process.

## Disabling the Feature

To disable the feature, set the field to `false` and perform a cluster upgrade:

```yaml
spec:
  controlPlaneConfiguration:
    skipAdmissionForSystemResources: false
```

The API server environment variable and exclusion rules file configuration will be removed during the next cluster reconciliation.

## Impact on Custom Webhooks

### What Changes

- Custom webhooks will **not** be invoked for operations on protected system resources
- Webhook processing continues normally for all other resources
- No modifications are made to webhook configurations themselves

### What Doesn't Change

- Custom webhooks continue to process application workloads
- Webhook configurations remain unchanged
- Custom business logic and validation continues for non-system resources

### Example Scenario

Consider a ValidatingWebhookConfiguration that targets all DaemonSets with `failurePolicy: Fail`:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: custom-webhook
webhooks:
- name: validate.example.com
  failurePolicy: Fail
  rules:
  - apiGroups: ["apps"]
    apiVersions: ["v1"]
    operations: ["UPDATE"]
    resources: ["daemonsets"]
```

**With admission webhook protection disabled:**
- Updates to Cilium DaemonSet are blocked if webhook service is unavailable or rejects the request
- Cluster upgrades may fail if Cilium cannot be updated

**With admission webhook protection enabled:**
- Updates to Cilium DaemonSet bypass the webhook (protected system resource)
- Cluster upgrades proceed successfully
- Updates to your application DaemonSets still invoke the webhook normally

## Troubleshooting

### Verifying Feature Activation

Check if the feature is active on control plane nodes:

```bash
kubectl get kubeadmcontrolplane -n eksa-system <cluster-name> -o yaml
```

Look for these configurations:

```yaml
spec:
  kubeadmConfigSpec:
    clusterConfiguration:
      apiServer:
        extraEnvs:
        - name: EKS_PATCH_EXCLUSION_RULES_FILE
          value: /etc/kubernetes/admission-plugin-exclusion-rules.json
        extraVolumes:
        - name: admission-exclusion-rules
          hostPath: /etc/kubernetes/admission-plugin-exclusion-rules.json
          mountPath: /etc/kubernetes/admission-plugin-exclusion-rules.json
          pathType: File
          readOnly: true
    files:
    - path: /etc/kubernetes/admission-plugin-exclusion-rules.json
      owner: root:root
```

### Common Issues

**Issue**: Feature not taking effect after upgrade
- **Solution**: Verify the Kubernetes version is 1.31+. Check kubeadm config for extraEnvs support.

**Issue**: Webhook still blocking system operations
- **Solution**: Verify the resource is in the protected list. Check API server logs for exclusion rule processing.

**Issue**: Application webhooks not working
- **Solution**: Verify webhooks are correctly configured for non-system resources. Protected resources are limited to specific system components.

## Best Practices

1. **Enable for production clusters**: We recommend enabling this feature for all production clusters to prevent upgrade failures

2. **Test in non-production first**: When enabling on existing clusters, test the upgrade in a non-production environment first

3. **Review webhook configurations**: Audit existing webhooks to ensure they don't unnecessarily target system resources

4. **Use appropriate failurePolicy**: Set webhooks to `Ignore` for non-critical validations to avoid blocking operations

5. **Scope webhooks appropriately**: Use namespace and object selectors to limit webhook scope to intended resources

## Related Resources

- [EKS Best Practices: Admission Webhooks](https://docs.aws.amazon.com/eks/latest/best-practices/control-plane.html#reliability_cpadmission_webhooks)
- [Security Best Practices]({{< relref "../security/best-practices.md" >}})
