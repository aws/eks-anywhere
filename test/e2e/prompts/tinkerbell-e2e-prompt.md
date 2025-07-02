# Tinkerbell E2E Test Automation Prompt Plan

This document provides a structured prompt plan for automating the creation of e2e tests for new Kubernetes versions in the Tinkerbell provider using AI coding agents like Cline.

## Overview

When EKS Anywhere adds support for a new Kubernetes version, comprehensive e2e tests need to be created following established patterns. This prompt plan breaks down the task into manageable subtasks to avoid context overflow and ensure systematic implementation.

## Prerequisites

- New Kubernetes version constant should already be defined in the codebase (e.g., `v1alpha1.Kube133`)
- Base understanding of the current test structure and patterns
- **Note**: Bottlerocket is no longer supported for Tinkerbell provider
- **Note**: RedHat 8 is not supported from Kubernetes 1.32 onwards (only RedHat 9 is supported)

## Task Decomposition Strategy

Use Cline's `new_task` tool to create separate tasks for each major component to manage context effectively:

### Task 1: Build Configuration Updates
**Scope**: Update build configuration files with comprehensive image environment variables
**Files to modify**:
- `cmd/integration_test/build/buildspecs/quick-test-eks-a-cli.yml`
- `cmd/integration_test/build/buildspecs/tinkerbell-test-eks-a-cli.yml`

**Prompt**:
```
I need to add comprehensive environment variables for a new Kubernetes version {NEW_VERSION} in the Tinkerbell provider build configuration files.

For both quick-test-eks-a-cli.yml and tinkerbell-test-eks-a-cli.yml, I need to add new environment variables following the established naming convention:

**Ubuntu Images:**
1. `T_TINKERBELL_IMAGE_UBUNTU_{VERSION_NUMBER}` → "tinkerbell_ci:image_ubuntu_{version}"
2. `T_TINKERBELL_IMAGE_UBUNTU_2204_{VERSION_NUMBER}` → "tinkerbell_ci:image_ubuntu_2204_{version}"
3. `T_TINKERBELL_IMAGE_UBUNTU_2204_{VERSION_NUMBER}_RTOS` → "tinkerbell_ci:image_ubuntu_2204_{version}_rtos"
4. `T_TINKERBELL_IMAGE_UBUNTU_2204_{VERSION_NUMBER}_GENERIC` → "tinkerbell_ci:image_ubuntu_2204_{version}_generic"

**RedHat Images:**
5. `T_TINKERBELL_IMAGE_REDHAT_9_{VERSION_NUMBER}` → "tinkerbell_ci:image_redhat_9_{version}"

**Note**: Only add RedHat 9 images for Kubernetes {NEW_VERSION} if it's 1.32 or later. RedHat 8 is not supported from Kubernetes 1.32 onwards.

Where {VERSION_NUMBER} = "1_{NEW_VERSION}" (e.g., "1_33") and {version} = "1_{NEW_VERSION}" (e.g., "1_33").

Please examine the existing environment variables sections and add the new ones in the correct alphabetical/numerical order.
```

### Task 2: Quick Tests Configuration Updates
**Scope**: Update Tinkerbell quick test patterns
**Files to modify**:
- `test/e2e/QUICK_TESTS.yaml`

**Prompt**:
```
I need to update the Tinkerbell quick test configuration for Kubernetes version {NEW_VERSION}.

In test/e2e/QUICK_TESTS.yaml, I need to update the Tinkerbell test patterns to use the new version:
   - Change `^TestTinkerbellKubernetes{PREV_VERSION}UbuntuTo{NEW_VERSION}Upgrade$` to `^TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}Upgrade$`
   - Change `TestTinkerbellKubernetes{PREV_VERSION}Ubuntu2004To2204Upgrade` to `TestTinkerbellKubernetes{NEW_VERSION}Ubuntu2004To2204Upgrade`
   - Change `TestTinkerbellKubernetes{PREV_VERSION}To{NEW_VERSION}Ubuntu2204Upgrade` to `TestTinkerbellKubernetes{NEW_VERSION}To{NEXT_VERSION}Ubuntu2204Upgrade`

Please examine the existing Tinkerbell test patterns and update them to reflect the new version as the current version.
```

### Task 3: Comprehensive Test Configuration Updates
**Scope**: Update skipped tests and hardware counts with comprehensive test coverage
**Files to modify**:
- `test/e2e/SKIPPED_TESTS.yaml`
- `test/e2e/TINKERBELL_HARDWARE_COUNT.yaml`

**Prompt**:
```
I need to comprehensively update test configuration files for Kubernetes version {NEW_VERSION}, including replacing {PREV_VERSION} tests with {NEW_VERSION} tests and adding new upgrade tests.

**SKIPPED_TESTS.yaml updates:**

1. **Add new upgrade test to skipped list:**
   - `TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}Upgrade`

2. **Update comments and version references:**
   - Update version ranges in comments (e.g., "1.28 to 1.33" instead of "1.28 to 1.32")
   - Update "latest kubernetes version" references to {NEW_VERSION}

3. **Add comprehensive skipped tests for {PREV_VERSION} (to be replaced by {NEW_VERSION}):**
   - Worker node tests, scaling tests, workload cluster tests
   - OIDC, registry mirror, proxy tests
   - Simple flow tests, curated packages tests
   - In-place upgrade, OOB, management CP upgrade tests
   - All the comprehensive test patterns that were previously enabled for {PREV_VERSION}

**TINKERBELL_HARDWARE_COUNT.yaml updates:**

1. **Update primary test entries:**
   - Change `TestTinkerbellKubernetes{PREV_VERSION}AWSIamAuth: 2` to `TestTinkerbellKubernetes{NEW_VERSION}AWSIamAuth: 2`

2. **Add new upgrade test entries:**
   - `TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}Upgrade: 4`
   - `TestTinkerbellKubernetes{NEW_VERSION}To{NEXT_VERSION}Ubuntu2204Upgrade: 4`
   - `TestTinkerbellKubernetes{NEW_VERSION}Ubuntu2004To2204Upgrade: 4`
   - `TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}UpgradeCPOnly: 3`
   - `TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}UpgradeWorkerOnly: 3`

3. **Add comprehensive {NEW_VERSION} test entries (replacing {PREV_VERSION}):**
   - All simple flow tests, worker node tests, scaling tests
   - Workload cluster tests, curated packages tests
   - OIDC, registry mirror, proxy tests, OOB tests
   - Management CP upgrade, in-place upgrade tests
   - All test patterns that need hardware counts

4. **Remove {PREV_VERSION} entries that are being replaced**

Please examine the existing patterns and make these comprehensive updates.
```

### Task 4: Simple Flow Tests Addition
**Scope**: Add basic simple flow tests for the new Kubernetes version
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add simple flow test functions for Kubernetes version {NEW_VERSION} in test/e2e/tinkerbell_test.go.

Following the existing patterns, I need to create these simple flow test functions:

1. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSimpleFlow**:
   - Uses `framework.WithUbuntu{NEW_VERSION}Tinkerbell()`
   - Sets kubernetes version to `v1alpha1.Kube{NEW_VERSION}`
   - Calls `runTinkerbellSimpleFlow(test)`

2. **TestTinkerbellKubernetes{NEW_VERSION}Ubuntu2204SimpleFlow**:
   - Uses license token and provider without specific tinkerbell opt
   - Uses `provider.WithKubeVersionAndOS(v1alpha1.Kube{NEW_VERSION}, framework.Ubuntu2204, nil)`
   - Calls `runTinkerbellSimpleFlowWithoutClusterConfigGeneration(test)`

3. **TestTinkerbellKubernetes{NEW_VERSION}RedHat9SimpleFlow**:
   - Uses `framework.WithRedHat9{NEW_VERSION}Tinkerbell()`
   - Sets kubernetes version to `v1alpha1.Kube{NEW_VERSION}`
   - Calls `runTinkerbellSimpleFlow(test)`

**Note**: Do NOT add BottleRocket tests as Tinkerbell no longer supports BottleRocket.
**Note**: Do NOT add RedHat 8 tests for Kubernetes 1.32 and later as RedHat 8 is not supported from Kubernetes 1.32 onwards.

Please examine the existing simple flow test patterns and create these new functions following the same structure.
```

### Task 5: AWS IAM Auth Test Update
**Scope**: Update AWS IAM Auth test function
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to update the AWS IAM Auth test function for Kubernetes version {NEW_VERSION} in test/e2e/tinkerbell_test.go.

I need to add a new function `TestTinkerbellKubernetes{NEW_VERSION}AWSIamAuth` that:
1. Uses `framework.WithUbuntu{NEW_VERSION}Tinkerbell()`
2. Sets kubernetes version to `v1alpha1.Kube{NEW_VERSION}`
3. Uses `framework.WithAWSIam()`
4. Configures 1 CP and 1 Worker hardware
5. Calls `runTinkerbellAWSIamAuthFlow(test)`

The new function should be placed at the top of the AWS IAM Auth section, before the existing {PREV_VERSION} function.

Please examine the existing AWS IAM Auth test pattern and create the new function following the same structure.
```

### Task 6: Basic Upgrade Tests
**Scope**: Add basic upgrade test functions
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add basic upgrade test functions for Kubernetes {NEW_VERSION} in test/e2e/tinkerbell_test.go.

Following the pattern of existing upgrade tests, I need to create:

**TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}Upgrade** function that:
1. Creates a provider with `framework.WithUbuntu{NEW_VERSION}Tinkerbell()`
2. Sets up the test with `v1alpha1.Kube{NEW_VERSION}`
3. Configures hardware counts (2 CP, 2 Worker)
4. Calls `runSimpleUpgradeFlowForBareMetal` with:
   - Target version: `v1alpha1.Kube{NEXT_VERSION}`
   - Cluster upgrade: `api.WithKubernetesVersion(v1alpha1.Kube{NEXT_VERSION})`
   - Provider upgrade: `framework.Ubuntu{NEXT_VERSION}Image()`

Place this function in the upgrade tests section, maintaining the numerical order.

Please examine the existing upgrade test pattern and create the new function following the same structure.
```

### Task 7: CP/Worker Only Upgrade Tests
**Scope**: Add control plane only and worker only upgrade test functions
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add control plane only and worker only upgrade test functions for Kubernetes {NEW_VERSION} to {NEXT_VERSION} in test/e2e/tinkerbell_test.go.

Following the existing patterns, I need to create two functions:

1. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}UpgradeCPOnly**:
   - Uses `provider := framework.NewTinkerbell(t)`
   - Uses `kube{NEW_VERSION} := v1alpha1.Kube{NEW_VERSION}`
   - Sets up cluster with CP and Worker on {NEW_VERSION}
   - Upgrades only CP to {NEXT_VERSION}
   - Uses `framework.Ubuntu{NEXT_VERSION}ImageForCP()`

2. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}UpgradeWorkerOnly**:
   - Uses `provider := framework.NewTinkerbell(t)`
   - Uses both `kube{NEW_VERSION} := v1alpha1.Kube{NEW_VERSION}` and `kube{NEXT_VERSION} := v1alpha1.Kube{NEXT_VERSION}`
   - Sets up cluster with CP on {NEXT_VERSION} and Worker on {NEW_VERSION}
   - Upgrades only Worker to {NEXT_VERSION}
   - Uses `framework.Ubuntu{NEXT_VERSION}ImageForWorker()`

Please examine the existing CP-only and Worker-only upgrade test patterns and create the new functions following the same structure.
```

### Task 8: Ubuntu 2204 Upgrade Tests
**Scope**: Add Ubuntu 22.04 upgrade test functions
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add Ubuntu 22.04 upgrade test functions for Kubernetes {NEW_VERSION} in test/e2e/tinkerbell_test.go.

Following the existing patterns, I need to create two functions:

1. **TestTinkerbellKubernetes{NEW_VERSION}To{NEXT_VERSION}Ubuntu2204Upgrade**:
   - Gets license token with `framework.GetLicenseToken()`
   - Sets up cluster with Ubuntu 22.04 and {NEW_VERSION}
   - Uses `provider.WithKubeVersionAndOS(v1alpha1.Kube{NEW_VERSION}, framework.Ubuntu2204, nil)`
   - Includes license token in cluster config
   - Calls `runSimpleUpgradeFlowForBaremetalWithoutClusterConfigGeneration`
   - Uses `framework.Ubuntu2204Kubernetes{NEXT_VERSION}Image()`

2. **TestTinkerbellKubernetes{NEW_VERSION}Ubuntu2004To2204Upgrade**:
   - Gets license token with `framework.GetLicenseToken()`
   - Sets up cluster with Ubuntu 20.04 and {NEW_VERSION}
   - Uses `provider.WithKubeVersionAndOS(v1alpha1.Kube{NEW_VERSION}, framework.Ubuntu2004, nil)`
   - Upgrades OS from Ubuntu 20.04 to 22.04
   - Uses `framework.Ubuntu2204Kubernetes{NEW_VERSION}Image()`

Please examine the existing Ubuntu 2204 upgrade test patterns and create the new functions following the same structure.
```

### Task 9: Comprehensive Test Functions - Part 1
**Scope**: Add worker node, scaling, and workload cluster test functions
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add comprehensive test functions for Kubernetes {NEW_VERSION} in test/e2e/tinkerbell_test.go - Part 1: Worker node, scaling, and workload cluster tests.

Following the existing patterns, I need to create these functions:

**Worker Node Tests:**
1. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkerNodeUpgrade**
2. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkerNodeScaleUpWithAPI**
3. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuAddWorkerNodeGroupWithAPI**

**Scaling Tests:**
4. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuThreeControlPlaneReplicasSimpleFlow**
5. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuThreeWorkersSimpleFlow**
6. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuControlPlaneScaleUp**
7. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkerNodeScaleUp**
8. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkerNodeScaleDown**
9. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuControlPlaneScaleDown**

**Workload Cluster Tests:**
10. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkloadCluster**
11. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkloadClusterWithAPI**
12. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkloadClusterGitFluxWithAPI**
13. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSingleNodeWorkloadCluster**
14. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSingleNodeWorkloadClusterWithAPI**

Each function should:
- Use `framework.WithUbuntu{NEW_VERSION}Tinkerbell()` where appropriate
- Set kubernetes version to `v1alpha1.Kube{NEW_VERSION}`
- Follow the exact same pattern as the corresponding {PREV_VERSION} function
- Use appropriate hardware configurations and test flows

Please examine the existing patterns for these test types and create the new functions following the same structure.
```

### Task 10: Comprehensive Test Functions - Part 2
**Scope**: Add upgrade, OIDC, registry mirror, and other comprehensive test functions
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add comprehensive test functions for Kubernetes {NEW_VERSION} in test/e2e/tinkerbell_test.go - Part 2: Upgrade, OIDC, registry mirror, and other tests.

Following the existing patterns, I need to create these functions:

**Multicluster Upgrade Tests:**
1. **TestTinkerbellUpgrade{NEW_VERSION}MulticlusterWorkloadClusterWorkerScaleupGitFluxWithAPI**
2. **TestTinkerbellUpgrade{NEW_VERSION}MulticlusterWorkloadClusterCPScaleup**
3. **TestTinkerbellUpgradeMulticlusterWorkloadClusterK8sUpgrade{NEW_VERSION}To{NEXT_VERSION}**

**OIDC and Registry Tests:**
4. **TestTinkerbellKubernetes{NEW_VERSION}OIDC**
5. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuRegistryMirror**
6. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuInsecureSkipVerifyRegistryMirror**

**Proxy and OOB Tests:**
7. **TestTinkerbellAirgappedKubernetes{NEW_VERSION}UbuntuProxyConfigFlow**
8. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuOOB**
9. **TestTinkerbellK8sUpgrade{NEW_VERSION}to{NEXT_VERSION}WithUbuntuOOB**

**Management and In-Place Upgrade Tests:**
10. **TestTinkerbellSingleNode{NEW_VERSION}To{NEXT_VERSION}UbuntuManagementCPUpgradeAPI**
11. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}InPlaceUpgrade_1CP_1Worker**
12. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuTo{NEXT_VERSION}SingleNodeInPlaceUpgrade**

**Other Tests:**
13. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuWorkerNodeGroupsTaintsAndLabels**
14. **TestTinkerbellKubernetes{NEW_VERSION}KubeletConfigurationSimpleFlow**

Each function should:
- Use appropriate framework functions for {NEW_VERSION}
- Set kubernetes version to `v1alpha1.Kube{NEW_VERSION}`
- Follow the exact same pattern as the corresponding {PREV_VERSION} function
- Include license tokens where needed
- Use appropriate hardware configurations and test flows

Please examine the existing patterns for these test types and create the new functions following the same structure.
```

### Task 11: Curated Packages Test Functions
**Scope**: Add curated packages test functions
**Files to modify**:
- `test/e2e/tinkerbell_test.go`

**Prompt**:
```
I need to add curated packages test functions for Kubernetes {NEW_VERSION} in test/e2e/tinkerbell_test.go.

Following the existing patterns, I need to create these functions:

**Curated Packages Tests:**
1. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSingleNodeCuratedPackagesFlow**
2. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSingleNodeCuratedPackagesEmissaryFlow**
3. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSingleNodeCuratedPackagesHarborFlow**
4. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuCuratedPackagesAdotSimpleFlow**
5. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuCuratedPackagesPrometheusSimpleFlow**
6. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuCuratedPackagesClusterAutoscalerSimpleFlow**

**Single Node Test:**
7. **TestTinkerbellKubernetes{NEW_VERSION}UbuntuSingleNodeSimpleFlow**

Each function should:
- Use `framework.WithUbuntu{NEW_VERSION}Tinkerbell()` where appropriate
- Set kubernetes version to `v1alpha1.Kube{NEW_VERSION}`
- Use `packageBundleURI(v1alpha1.Kube{NEW_VERSION})` for package configurations
- Include appropriate package controller helm configurations
- Follow the exact same pattern as the corresponding {PREV_VERSION} function
- Use appropriate hardware configurations and test flows

For the cluster autoscaler test, make sure to include the minNodes/maxNodes configuration and worker hardware scaling.

Please examine the existing patterns for these test types and create the new functions following the same structure.
```

### Task 12: Framework Helper Functions
**Scope**: Add framework helper functions for the new Kubernetes version
**Files to modify**:
- `test/framework/tinkerbell.go`

**Prompt**:
```
I need to add framework helper functions for Kubernetes version {NEW_VERSION} in test/framework/tinkerbell.go.

Following the existing patterns, I need to add these functions:

**Tinkerbell Options:**
1. `WithUbuntu{NEW_VERSION}Tinkerbell() TinkerbellOpt`:
   - Returns `withKubeVersionAndOS(anywherev1.Kube{NEW_VERSION}, Ubuntu2004, "", nil)`

2. `WithRedHat9{NEW_VERSION}Tinkerbell() TinkerbellOpt`:
   - Returns `withKubeVersionAndOS(anywherev1.Kube{NEW_VERSION}, RedHat9, "", nil)`

**Note**: Do NOT add RedHat 8 framework functions for Kubernetes 1.32 and later as RedHat 8 is not supported from Kubernetes 1.32 onwards.

**Image Functions:**
4. `Ubuntu{NEW_VERSION}Image() api.TinkerbellFiller`:
   - Returns `imageForKubeVersionAndOS(anywherev1.Kube{NEW_VERSION}, Ubuntu2004, "")`

5. `Ubuntu{NEW_VERSION}ImageForCP() api.TinkerbellFiller`:
   - Returns `imageForKubeVersionAndOS(anywherev1.Kube{NEW_VERSION}, Ubuntu2004, controlPlaneIdentifier)`

6. `Ubuntu{NEW_VERSION}ImageForWorker() api.TinkerbellFiller`:
   - Returns `imageForKubeVersionAndOS(anywherev1.Kube{NEW_VERSION}, Ubuntu2004, workerIdentifier)`

7. `Ubuntu2204Kubernetes{NEW_VERSION}Image() api.TinkerbellFiller`:
   - Returns `imageForKubeVersionAndOS(anywherev1.Kube{NEW_VERSION}, Ubuntu2204, "")`

Each function should include appropriate comments following the existing pattern (e.g., "// Ubuntu{NEW_VERSION}Image represents an Ubuntu raw image corresponding to Kubernetes 1.{NEW_VERSION}.").

Please examine the existing helper function patterns and add the new functions in the correct locations, maintaining alphabetical/numerical order and including appropriate comments.
```

## Usage Instructions

1. **Preparation**: Ensure the new Kubernetes version constant is defined in the codebase
2. **Variable Substitution**: Replace placeholders in prompts:
   - `{NEW_VERSION}` → actual new version (e.g., "133")
   - `{PREV_VERSION}` → previous version (e.g., "132")
   - `{NEXT_VERSION}` → next version (e.g., "134")
   - `{VERSION_NUMBER}` → version with underscores (e.g., "1_33")

3. **Execution Order**: Execute tasks in the specified order to maintain dependencies
4. **Context Management**: Use `new_task` tool between major tasks to reset context
5. **Validation**: After each task, verify changes follow existing patterns and conventions

## Quality Assurance Checklist

- [ ] All build configuration environment variables are added
- [ ] Quick test patterns are updated correctly
- [ ] Test configuration files are updated consistently
- [ ] Simple flow tests are added for all supported OS types (Ubuntu, RedHat, RedHat9)
- [ ] AWS IAM Auth test is updated
- [ ] All upgrade test patterns are implemented
- [ ] Comprehensive test functions cover all test categories
- [ ] Curated packages tests are properly configured
- [ ] Framework helper functions are properly documented
- [ ] All references to version numbers are updated
- [ ] Code follows existing formatting and style
- [ ] No BottleRocket tests are added (no longer supported)
- [ ] License tokens are included where required
- [ ] Hardware counts match test requirements

## Important Notes

- **BottleRocket Support**: Tinkerbell no longer supports BottleRocket. Do not add any BottleRocket-related tests.
- **RedHat 8 Support**: RedHat 8 is not supported from Kubernetes 1.32 onwards. Only RedHat 9 is supported for Kubernetes 1.32 and later versions. Do not add RedHat 8 tests, environment variables, or framework functions for Kubernetes 1.32+.
- **License Tokens**: Many tests require license tokens, especially Ubuntu 22.04 tests. Ensure `framework.GetLicenseToken()` is used where needed.
- **Version Progression**: When adding {NEW_VERSION} tests, they often replace {PREV_VERSION} tests in terms of being the "current" version, while {PREV_VERSION} tests may be moved to skipped tests.
- **Comprehensive Coverage**: The updated plan covers all test categories found in the patches, including worker node tests, scaling tests, workload cluster tests, OIDC, registry mirror, proxy, OOB, in-place upgrades, management CP upgrades, curated packages, and more.
- **Context Management**: Due to the large number of changes, using `new_task` tool between major tasks is essential to avoid context overflow.

## Example Usage

For adding Kubernetes 1.34 support:
- Replace `{NEW_VERSION}` with "134"
- Replace `{PREV_VERSION}` with "133"
- Replace `{NEXT_VERSION}` with "135"
- Replace `{VERSION_NUMBER}` with "1_34"
- Execute each task in sequence using Cline's new_task tool
