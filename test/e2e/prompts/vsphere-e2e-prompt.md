# E2E Test Automation Prompt Plan for New Kubernetes Versions - vSphere Provider

This document provides a structured approach for AI coding agents (like Cline) to automate the creation of e2e tests when EKS Anywhere adds support for a new Kubernetes version for the vSphere provider.

## Overview

When adding support for a new Kubernetes version (e.g., 1.33), comprehensive e2e tests need to be created for the vSphere provider. This prompt plan breaks down the task into manageable subtasks to avoid context overflow and ensure systematic coverage.

## Prerequisites

- Reference patch file showing changes for the previous version (e.g., `vsphere-e2e-132.patch`)
- Target Kubernetes version (e.g., `1.33`)
- Target provider: `vsphere`

## Detailed Task Decomposition Strategy

Use Cline's `new_task` tool to break down the work into these granular subtasks:

### Task 1A: Quick Test Build Configuration
**Scope**: Update quick test buildspec with new template environment variables
**File**: `cmd/integration_test/build/buildspecs/quick-test-eks-a-cli.yml`
**Estimated Lines**: ~10-15 additions

**Specific Actions**:
1. Add template environment variables for new Kubernetes version
2. Follow pattern: `T_VSPHERE_TEMPLATE_{OS}_{VERSION}`

**Template Variables to Add**:
```yaml
T_VSPHERE_TEMPLATE_UBUNTU_1_33: "/SDDC-Datacenter/vm/Templates/ubuntu-kube-v1-33"
T_VSPHERE_TEMPLATE_UBUNTU_2204_1_33: "/SDDC-Datacenter/vm/Templates/ubuntu-2204-kube-v1-33"
T_VSPHERE_TEMPLATE_BR_1_33: "/SDDC-Datacenter/vm/Templates/bottlerocket-kube-v1-33"
T_VSPHERE_TEMPLATE_REDHAT_9_1_33: "/SDDC-Datacenter/vm/Templates/redhat-9-kube-v1-33"
```

**Note**: Only RedHat 9 templates are supported for Kubernetes 1.32+. RedHat 8 templates are not included.

### Task 1B: vSphere-Specific Build Configuration
**Scope**: Update vSphere-specific buildspec with new template environment variables
**File**: `cmd/integration_test/build/buildspecs/vsphere-test-eks-a-cli.yml`
**Estimated Lines**: ~10-15 additions

**Specific Actions**:
1. Add new version template variables
2. Remove deprecated older version templates (if applicable)

### Task 2: Quick Tests Configuration Update
**Scope**: Update quick test patterns for new version upgrades
**File**: `test/e2e/QUICK_TESTS.yaml`
**Estimated Lines**: ~10-15 modifications

**Specific Test Patterns to Update**:
```yaml
- ^TestVSphereKubernetes132To133RedHatUpgrade$
- TestVSphereKubernetes132To133StackedEtcdRedHatUpgrade
- ^TestVSphereKubernetes132UbuntuTo133Upgrade$
- TestVSphereKubernetes132UbuntuTo133StackedEtcdUpgrade
- TestVSphereKubernetes132To133Ubuntu2204Upgrade
- TestVSphereKubernetes132To133Ubuntu2204StackedEtcdUpgrade
- TestVSphereKubernetes133Ubuntu2004To2204Upgrade
- TestVSphereKubernetes132BottlerocketTo133Upgrade
- TestVSphereKubernetes132BottlerocketTo133StackedEtcdUpgrade
```

### Task 3A: Simple Flow Test Functions - Ubuntu Variants
**Scope**: Add Ubuntu-based simple flow test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~50-70 additions

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133Ubuntu2004SimpleFlow(t *testing.T)
func TestVSphereKubernetes133Ubuntu2204SimpleFlow(t *testing.T)
func TestVSphereKubernetes133ThreeReplicasFiveWorkersSimpleFlow(t *testing.T)
func TestVSphereKubernetes133DifferentNamespaceSimpleFlow(t *testing.T)
func TestVSphereKubernetes133StackedEtcdUbuntu(t *testing.T)
```

### Task 3B: Simple Flow Test Functions - Bottlerocket Variants
**Scope**: Add Bottlerocket-based simple flow test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~50-70 additions

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133BottleRocketSimpleFlow(t *testing.T)
func TestVSphereKubernetes133BottleRocketThreeReplicasFiveWorkersSimpleFlow(t *testing.T)
func TestVSphereKubernetes133BottleRocketDifferentNamespaceSimpleFlow(t *testing.T)
func TestVSphereKubernetes133BottleRocketWithNTP(t *testing.T)
func TestVSphereKubernetes133BottleRocketWithBottlerocketKubernetesSettings(t *testing.T)
```

### Task 3C: Simple Flow Test Functions - RedHat Variants
**Scope**: Add RedHat-based simple flow test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~20-30 additions

**Note**: RedHat 8 is not supported for Kubernetes 1.32 onwards. Only RedHat 9 is supported.

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133RedHat9SimpleFlow(t *testing.T)
func TestVSphereKubernetes133UbuntuWithNTP(t *testing.T)
```

### Task 4A: API Server Extra Args Test Functions
**Scope**: Add API Server Extra Args test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~30-40 additions

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133BottlerocketAPIServerExtraArgsSimpleFlow(t *testing.T)
func TestVSphereKubernetes133BottlerocketAPIServerExtraArgsUpgradeFlow(t *testing.T)
```

### Task 4B: Auto-import Test Functions
**Scope**: Add auto-import test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~15-20 additions

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133BottlerocketAutoimport(t *testing.T)
```

### Task 4C: AWS IAM Auth Test Functions
**Scope**: Add AWS IAM Auth test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~30-40 additions

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133BottleRocketAWSIamAuth(t *testing.T)
func TestVSphereKubernetes132To133AWSIamAuthUpgrade(t *testing.T)
```

### Task 4D: Curated Packages - Core Test Functions
**Scope**: Add core curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVSphereKubernetes133CuratedPackagesSimpleFlow(t *testing.T)
func TestVSphereKubernetes133BottleRocketCuratedPackagesSimpleFlow(t *testing.T)
```

### Task 4E: Curated Packages - Emissary Test Functions
**Scope**: Add Emissary curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133CuratedPackagesEmissarySimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketCuratedPackagesEmissarySimpleFlow(t *testing.T)
```

### Task 4F: Curated Packages - Harbor Test Functions
**Scope**: Add Harbor curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133CuratedPackagesHarborSimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketCuratedPackagesHarborSimpleFlow(t *testing.T)
```

### Task 4G: Curated Packages - ADOT Test Functions
**Scope**: Add ADOT curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133CuratedPackagesAdotUpdateFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketCuratedPackagesAdotUpdateFlow(t *testing.T)
```

### Task 4H: Curated Packages - Cluster Autoscaler Test Functions
**Scope**: Add Cluster Autoscaler curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~80-100 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketCuratedPackagesClusterAutoscalerSimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesClusterAutoscalerUpgradeFlow(t *testing.T)
```

### Task 4I: Curated Packages - Prometheus Test Functions
**Scope**: Add Prometheus curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuCuratedPackagesPrometheusSimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketCuratedPackagesPrometheusSimpleFlow(t *testing.T)
```

### Task 4J: Workload Cluster Curated Packages Test Functions
**Scope**: Add workload cluster curated packages test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~100-120 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesSimpleFlow(t *testing.T)
func TestVsphereKubernetes133UbuntuWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesEmissarySimpleFlow(t *testing.T)
func TestVsphereKubernetes133UbuntuWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T)
func TestVsphereKubernetes133BottleRocketWorkloadClusterCuratedPackagesCertManagerSimpleFlow(t *testing.T)
```

### Task 4K: GitOps - GitHub Flux Test Functions
**Scope**: Add GitHub Flux test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133GithubFlux(t *testing.T)
func TestVsphereKubernetes133BottleRocketGithubFlux(t *testing.T)
```

### Task 4L: GitOps - Git Flux Test Functions
**Scope**: Add Git Flux test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133GitFlux(t *testing.T)
func TestVsphereKubernetes133BottleRocketGitFlux(t *testing.T)
```

### Task 4M: Labels and Taints Test Functions
**Scope**: Add labels and taints test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuLabelsUpgradeFlow(t *testing.T)
func TestVsphereKubernetes133BottlerocketLabelsUpgradeFlow(t *testing.T)
func TestVsphereKubernetes133UbuntuTaintsUpgradeFlow(t *testing.T)
func TestVsphereKubernetes133BottlerocketTaintsUpgradeFlow(t *testing.T)
```

### Task 4N: Multi-cluster Test Functions
**Scope**: Add multi-cluster test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~40-60 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133MulticlusterWorkloadCluster(t *testing.T)
```

### Task 4O: OIDC Test Functions
**Scope**: Add OIDC test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~40-60 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133OIDC(t *testing.T)
func TestVsphereKubernetes132To133OIDCUpgrade(t *testing.T)
```

### Task 4P: Proxy Configuration Test Functions
**Scope**: Add proxy configuration test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuProxyConfigFlow(t *testing.T)
func TestVsphereKubernetes133BottlerocketProxyConfigFlow(t *testing.T)
```

### Task 4Q: Registry Mirror Test Functions
**Scope**: Add registry mirror test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~120-150 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuRegistryMirrorInsecureSkipVerify(t *testing.T)
func TestVsphereKubernetes133UbuntuRegistryMirrorAndCert(t *testing.T)
func TestVsphereKubernetes133BottlerocketRegistryMirrorAndCert(t *testing.T)
func TestVsphereKubernetes133UbuntuAuthenticatedRegistryMirror(t *testing.T)
func TestVsphereKubernetes133BottlerocketAuthenticatedRegistryMirror(t *testing.T)
func TestVsphereKubernetes133BottlerocketRegistryMirrorOciNamespaces(t *testing.T)
func TestVsphereKubernetes133UbuntuAuthenticatedRegistryMirrorCuratedPackagesSimpleFlow(t *testing.T)
```

### Task 4R: Clone Mode Test Functions (vSphere-specific)
**Scope**: Add clone mode test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~100-120 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133FullClone(t *testing.T)
func TestVsphereKubernetes133LinkedClone(t *testing.T)
func TestVsphereKubernetes133BottlerocketFullClone(t *testing.T)
func TestVsphereKubernetes133BottlerocketLinkedClone(t *testing.T)
```

### Task 4S: Kubelet Configuration Test Functions
**Scope**: Add kubelet configuration test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~40-60 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuKubeletConfiguration(t *testing.T)
func TestVsphereKubernetes133BottlerocketKubeletConfiguration(t *testing.T)
```

### Task 4T: Etcd Encryption Test Functions
**Scope**: Add etcd encryption test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetesUbuntu133EtcdEncryption(t *testing.T)
func TestVsphereKubernetesBottlerocket133EtcdEncryption(t *testing.T)
```

### Task 4U: Airgapped Test Functions
**Scope**: Add airgapped test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuAirgappedRegistryMirror(t *testing.T)
func TestVsphereKubernetes133UbuntuAirgappedProxy(t *testing.T)
```

### Task 5A: Simple Upgrade Test Functions - Ubuntu
**Scope**: Add Ubuntu upgrade test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~80-100 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes132UbuntuTo133Upgrade(t *testing.T)
func TestVsphereKubernetes132To133Ubuntu2204Upgrade(t *testing.T)
func TestVsphereKubernetes132To133Ubuntu2204StackedEtcdUpgrade(t *testing.T)
func TestVsphereKubernetes133Ubuntu2004To2204Upgrade(t *testing.T)
```

### Task 5B: Simple Upgrade Test Functions - Bottlerocket
**Scope**: Add Bottlerocket upgrade test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes132BottlerocketTo133Upgrade(t *testing.T)
func TestVsphereKubernetes132BottlerocketTo133StackedEtcdUpgrade(t *testing.T)
```

### Task 5C: Simple Upgrade Test Functions - RedHat9
**Scope**: Add RedHat upgrade test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes132To133RedHat9Upgrade(t *testing.T)
func TestVsphereKubernetes132To133StackedEtcdRedHat9Upgrade(t *testing.T)
```

### Task 5D: Multiple Fields Upgrade Test Functions
**Scope**: Add multiple fields upgrade test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~80-100 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes132UbuntuTo133MultipleFieldsUpgrade(t *testing.T)
func TestVsphereKubernetes132BottlerocketTo133MultipleFieldsUpgrade(t *testing.T)
```

### Task 5E: Node Scaling Upgrade Test Functions
**Scope**: Add node scaling upgrade test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~80-100 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133UbuntuControlPlaneNodeUpgrade(t *testing.T)
func TestVsphereKubernetes133UbuntuWorkerNodeUpgrade(t *testing.T)
func TestVsphereKubernetes133BottlerocketControlPlaneNodeUpgrade(t *testing.T)
func TestVsphereKubernetes133BottlerocketWorkerNodeUpgrade(t *testing.T)
```

### Task 5F: In-Place Upgrade Test Functions
**Scope**: Add in-place upgrade test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~120-150 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes132UbuntuTo133InPlaceUpgrade_1CP_1Worker(t *testing.T)
func TestVsphereKubernetes132UbuntuTo133InPlaceUpgradeWorkerOnly(t *testing.T)
func TestVsphereKubernetes133UbuntuInPlaceCPScaleUp1To3(t *testing.T)
func TestVsphereKubernetes133UbuntuInPlaceCPScaleDown3To1(t *testing.T)
func TestVsphereKubernetes133UbuntuInPlaceWorkerScaleUp1To2(t *testing.T)
func TestVsphereKubernetes133UbuntuInPlaceWorkerScaleDown2To1(t *testing.T)
```

### Task 5G: Upgrade with Features Test Functions
**Scope**: Add upgrade with features test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~60-80 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes132To133GitFluxUpgrade(t *testing.T)
func TestVsphereInPlaceUpgradeMulticlusterWorkloadClusterK8sUpgrade132To133(t *testing.T)
```

### Task 5H: Upgrade from Latest Minor Release Test Functions
**Scope**: Add upgrade from latest minor release test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~100-120 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133RedhatUpgradeFromLatestMinorRelease(t *testing.T)
func TestVsphereKubernetes133WithOIDCManagementClusterUpgradeFromLatestSideEffects(t *testing.T)
func TestVsphereKubernetes132To133UbuntuUpgradeFromLatestMinorRelease(t *testing.T)
func TestVsphereKubernetes132To133UbuntuInPlaceUpgradeFromLatestMinorRelease(t *testing.T)
func TestVsphereKubernetes132To133RedhatUpgradeFromLatestMinorRelease(t *testing.T)
func TestVsphereKubernetes133UbuntuUpgradeAndRemoveWorkerNodeGroupsAPI(t *testing.T)
func TestVsphereKubernetes132to133UpgradeFromLatestMinorReleaseBottleRocketAPI(t *testing.T)
```

### Task 5I: Etcd Scaling Test Functions
**Scope**: Add etcd scaling test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~120-150 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133BottlerocketEtcdScaleUp(t *testing.T)
func TestVsphereKubernetes133BottlerocketEtcdScaleDown(t *testing.T)
func TestVsphereKubernetes133UbuntuEtcdScaleUp(t *testing.T)
func TestVsphereKubernetes133UbuntuEtcdScaleDown(t *testing.T)
func TestVsphereKubernetes132to133UbuntuEtcdScaleUp(t *testing.T)
func TestVsphereKubernetes132to133UbuntuEtcdScaleDown(t *testing.T)
```

### Task 6A: Framework Helper Functions - OS Variants
**Scope**: Add OS variant helper functions
**File**: `test/framework/vsphere.go`
**Estimated Lines**: ~30-40 additions

**Specific Functions to Add**:
```go
func WithUbuntu133() VSphereOpt
func WithBottleRocket133() VSphereOpt
func WithRedHat9133VSphere() VSphereOpt
func (v *VSphere) WithUbuntu133() api.ClusterConfigFiller
func (v *VSphere) WithBottleRocket133() api.ClusterConfigFiller
```

**Note**: `WithRedHat133VSphere()` is not needed as RedHat 8 is not supported for Kubernetes 1.32+

### Task 6B: Framework Helper Functions - Template Functions
**Scope**: Add template helper functions
**File**: `test/framework/vsphere.go`
**Estimated Lines**: ~30-40 additions

**Specific Functions to Add**:
```go
func (v *VSphere) Ubuntu133Template() api.VSphereFiller
func (v *VSphere) Ubuntu2204Kubernetes133Template() api.VSphereFiller
func (v *VSphere) Bottlerocket133Template() api.VSphereFiller
func (v *VSphere) Redhat9133Template() api.VSphereFiller
```

**Note**: `Redhat133Template()` is not needed as RedHat 8 is not supported for Kubernetes 1.32+

### Task 7A: Workload API Test Functions
**Scope**: Add workload API test functions
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~200-250 additions

**Specific Functions to Add**:
```go
func TestVsphereKubernetes133MulticlusterWorkloadClusterAPI(t *testing.T)
func TestVsphereKubernetes133UpgradeLabelsTaintsUbuntuAPI(t *testing.T)
func TestVsphereKubernetes133UpgradeWorkerNodeGroupsUbuntuAPI(t *testing.T)
func TestVsphereKubernetes133MulticlusterWorkloadClusterGitHubFluxAPI(t *testing.T)
func TestVsphereKubernetes133CiliumUbuntuAPI(t *testing.T)
func TestVsphereKubernetes133UpgradeLabelsTaintsBottleRocketGitHubFluxAPI(t *testing.T)
func TestVsphereKubernetes133UpgradeWorkerNodeGroupsUbuntuGitHubFluxAPI(t *testing.T)
func TestVsphereUpgradeKubernetes133CiliumUbuntuGitHubFluxAPI(t *testing.T)
```

### Task 7B: Provider-Specific Helper Functions
**Scope**: Add provider-specific helper functions for labels and taints
**File**: `test/e2e/vsphere_test.go`
**Estimated Lines**: ~80-100 additions

**Specific Functions to Add**:
```go
func ubuntu133ProviderWithLabels(t *testing.T) *framework.Vsphere
func bottlerocket133ProviderWithLabels(t *testing.T) *framework.Vsphere
func ubuntu133ProviderWithTaints(t *testing.T) *framework.Vsphere
func bottlerocket133ProviderWithTaints(t *testing.T) *framework.Vsphere
```

## Implementation Guidelines

### Naming Conventions
- Test functions: `TestVSphereKubernetes{Version}{OS}{Feature}{TestType}` (Note: VSphere with capital S)
- Helper functions: `With{OS}{Version}()`, `{OS}{Version}Template()`
- Constants: Follow existing patterns in the codebase

### Code Patterns
1. **Test Structure**: Follow the established pattern of creating test framework, configuring cluster, running test flow
2. **Version References**: Use `v1alpha1.Kube{Version}` constants
3. **Template References**: Use environment variable patterns for vSphere templates
4. **Error Handling**: Follow existing error handling patterns in the test framework

### Quality Checks
1. Ensure all test functions compile
2. Verify naming consistency across all functions
3. Check that all OS variants are covered for each feature
4. Validate that upgrade paths are correctly defined
5. Ensure framework helper functions are properly integrated

## Execution Strategy

### Phase 1: Infrastructure Setup
Execute Tasks 1-2 to set up the basic infrastructure and configuration

### Phase 2: Core Test Implementation  
Execute Tasks 3-4 to implement the main test functions

### Phase 3: Advanced Features
Execute Tasks 5-7 to add upgrade tests and specialized functionality

### Phase 4: Validation
Run compilation checks and basic test validation

## Context Management

To avoid context overflow:
1. Use `new_task` tool between major task categories
2. Focus on one file at a time within each task
3. Break large files into logical sections (e.g., by feature category)
4. Preserve context about patterns and conventions between tasks

## vSphere Provider Specifics

This plan is specifically tailored for the vSphere provider and includes:
1. vSphere-specific clone mode tests (full clone vs linked clone)
2. vSphere template environment variables
3. vSphere-specific test file paths (`test/e2e/vsphere_test.go`, `test/framework/vsphere.go`)
4. vSphere datacenter configuration patterns

## Example Usage

```
Task: Add Kubernetes 1.33 e2e tests for vSphere provider

1. Create new task for build configuration updates
2. Update buildspecs with vSphere 1.33 templates
3. Create new task for test function implementation
4. Add core vSphere 1.33 test functions
5. Continue through remaining tasks...
```

This systematic approach ensures comprehensive test coverage while maintaining code quality and avoiding context limitations.
