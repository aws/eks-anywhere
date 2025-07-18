# E2E Test Automation Prompt Plan for New Kubernetes Versions

This document provides a structured prompt plan for automating the creation of e2e tests when adding support for a new Kubernetes version to a provider in EKS Anywhere.

## Overview

When EKS Anywhere adds support for a new Kubernetes version, comprehensive e2e tests need to be created following established patterns. This automation plan breaks down the task into manageable subtasks to avoid context overflow and ensure systematic implementation.

## Task Decomposition Strategy

Use Cline's `new_task` tool to create separate tasks for each major component to manage context effectively:

### Task 1: Update Test Configuration Files
**Scope**: Update configuration files that control test execution
**Files**: 
- `test/e2e/QUICK_TESTS.yaml`
- `test/e2e/constants.go`

### Task 2: Update Core Test Functions (Part 1)
**Scope**: Update existing test functions to use the new Kubernetes version
**Files**: 
- `test/e2e/docker_test.go` (first 500 lines)

### Task 3: Update Core Test Functions (Part 2) 
**Scope**: Continue updating existing test functions
**Files**: 
- `test/e2e/docker_test.go` (lines 500-1000)

### Task 4: Add New Test Functions (Part 1)
**Scope**: Add new test functions for the new Kubernetes version
**Files**: 
- `test/e2e/docker_test.go` (curated packages tests)

### Task 5: Add New Test Functions (Part 2)
**Scope**: Add remaining new test functions
**Files**: 
- `test/e2e/docker_test.go` (auth, registry, simple flow tests)

### Task 6: Update Upgrade Test Functions
**Scope**: Update upgrade test functions to use new version transitions
**Files**: 
- `test/e2e/docker_test.go` (upgrade test functions)

## Detailed Implementation Prompts

### Task 1 Prompt: Update Test Configuration Files

```
I need to update the e2e test configuration files to support Kubernetes version {NEW_VERSION} for the Docker provider.

Based on the reference pattern, please:

1. Update `test/e2e/QUICK_TESTS.yaml`:
   - Change the test pattern from `TestDocker.*{PREVIOUS_VERSION}` to `TestDocker.*{NEW_VERSION}`

2. Update `test/e2e/constants.go`:
   - Add `v1alpha1.Kube{NEW_VERSION}` to the `KubeVersions` slice
   - Maintain the existing order (ascending version numbers)

Variables to replace:
- {NEW_VERSION}: The new Kubernetes version number (e.g., 134)
- {PREVIOUS_VERSION}: The previous Kubernetes version number (e.g., 133)
```

### Task 2 Prompt: Update Core Test Functions (Part 1)

```
I need to update existing e2e test functions in the first part of `test/e2e/docker_test.go` to use Kubernetes version {NEW_VERSION}.

Please update the following types of functions (approximately first 500 lines):

1. **Label Tests**: Update functions like `TestDockerKubernetesLabels` to use `v1alpha1.Kube{NEW_VERSION}`

2. **Flux Tests**: Update functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}GithubFlux` → `TestDockerKubernetes{NEW_VERSION}GithubFlux`
   - `TestDockerKubernetes{PREVIOUS_VERSION}GitFlux` → `TestDockerKubernetes{NEW_VERSION}GitFlux`
   - Update the `api.WithKubernetesVersion()` calls within these functions

3. **Flux Upgrade Tests**: Update functions like:
   - `TestDockerInstallGitFluxDuringUpgrade`
   - `TestDockerInstallGithubFluxDuringUpgrade`
   - Update both function names and internal version references

Pattern to follow:
- Function names: Change version numbers in function names
- Internal calls: Update `api.WithKubernetesVersion(v1alpha1.Kube{OLD})` to `api.WithKubernetesVersion(v1alpha1.Kube{NEW})`
- Upgrade flows: Update version parameters in upgrade function calls

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 3 Prompt: Update Core Test Functions (Part 2)

```
Continue updating existing e2e test functions in `test/e2e/docker_test.go` (lines 500-1000) to use Kubernetes version {NEW_VERSION}.

Focus on updating:

1. **Workload Cluster Tests**: Update functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}UpgradeWorkloadClusterWithGithubFlux`
   - Update both the function name and internal version references
   - Update upgrade target versions appropriately

2. **Taints Tests**: Update functions like:
   - `TestDockerKubernetes{PREVIOUS_VERSION}Taints` → `TestDockerKubernetes{NEW_VERSION}Taints`
   - `TestDockerKubernetes{PREVIOUS_VERSION}WorkloadClusterTaints` → `TestDockerKubernetes{NEW_VERSION}WorkloadClusterTaints`

3. **Simple Flow Tests**: Update functions like:
   - Update any remaining simple flow test functions to use the new version

Follow the same pattern as Task 2:
- Update function names with version numbers
- Update `api.WithKubernetesVersion()` calls
- Update upgrade flow version parameters

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 4 Prompt: Add New Test Functions (Part 1)

```
I need to add new e2e test functions for Kubernetes version {NEW_VERSION} in `test/e2e/docker_test.go`.

Please add the following new test functions by copying and modifying existing patterns:

1. **Curated Packages Tests**: Add new functions for:
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesEmissarySimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesHarborSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesAdotSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesPrometheusSimpleFlow`
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesDisabled`

2. **MetalLB Test**: Add:
   - `TestDockerKubernetes{NEW_VERSION}CuratedPackagesMetalLB`

Pattern to follow:
- Copy the corresponding {PREVIOUS_VERSION} function
- Update function name to use {NEW_VERSION}
- Update `api.WithKubernetesVersion(v1alpha1.Kube{PREVIOUS_VERSION})` to `api.WithKubernetesVersion(v1alpha1.Kube{NEW_VERSION})`
- Update `packageBundleURI(v1alpha1.Kube{PREVIOUS_VERSION})` to `packageBundleURI(v1alpha1.Kube{NEW_VERSION})`
- Keep all other parameters and function calls identical

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 5 Prompt: Add New Test Functions (Part 2)

```
Continue adding new e2e test functions for Kubernetes version {NEW_VERSION} in `test/e2e/docker_test.go`.

Please add the following new test functions:

1. **Authentication Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}AWSIamAuth`
   - `TestDockerKubernetes{NEW_VERSION}OIDC`

2. **Registry Mirror Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}RegistryMirrorInsecureSkipVerify`

3. **Simple Flow Test**: Add:
   - `TestDockerKubernetes{NEW_VERSION}SimpleFlow`

4. **Kubelet Configuration Test**: Add:
   - `TestDockerKubernetes{NEW_VERSION}KubeletConfigurationSimpleFlow`

5. **Etcd Scale Tests**: Add:
   - `TestDockerKubernetes{NEW_VERSION}EtcdScaleUp`
   - `TestDockerKubernetes{NEW_VERSION}EtcdScaleDown`

Follow the same pattern as Task 4:
- Copy the corresponding {PREVIOUS_VERSION} function
- Update function name to use {NEW_VERSION}
- Update all `api.WithKubernetesVersion()` calls
- Keep all other parameters identical

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
```

### Task 6 Prompt: Update Upgrade Test Functions

```
I need to update upgrade test functions in `test/e2e/docker_test.go` to support upgrading to Kubernetes version {NEW_VERSION}.

Please update the following types of upgrade tests:

1. **Version-to-Version Upgrade Tests**: Update functions like:
   - `TestDockerKubernetes{PREV_PREV}To{PREVIOUS_VERSION}StackedEtcdUpgrade` → `TestDockerKubernetes{PREVIOUS_VERSION}To{NEW_VERSION}StackedEtcdUpgrade`
   - `TestDockerKubernetes{PREV_PREV}To{PREVIOUS_VERSION}ExternalEtcdUpgrade` → `TestDockerKubernetes{PREVIOUS_VERSION}To{NEW_VERSION}ExternalEtcdUpgrade`

2. **Upgrade from Latest Release Tests**: Update functions like:
   - `TestDockerKubernetes{PREV_PREV}to{PREVIOUS_VERSION}UpgradeFromLatestMinorRelease` → `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}UpgradeFromLatestMinorRelease`
   - `TestDockerKubernetes{PREV_PREV}to{PREVIOUS_VERSION}GithubFluxEnabledUpgradeFromLatestMinorRelease` → `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}GithubFluxEnabledUpgradeFromLatestMinorRelease`

3. **Workload Cluster Upgrade Tests**: Update functions like:
   - `TestDockerUpgradeKubernetes{PREV_PREV}to{PREVIOUS_VERSION}WorkloadClusterScaleupAPI` → `TestDockerUpgradeKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}WorkloadClusterScaleupAPI`
   - Update similar workload cluster upgrade functions

4. **Management Cluster Tests**: Update:
   - `TestDockerKubernetes{PREVIOUS_VERSION}WithOIDCManagementClusterUpgradeFromLatestSideEffects` → `TestDockerKubernetes{NEW_VERSION}WithOIDCManagementClusterUpgradeFromLatestSideEffects`

5. **Etcd Scale with Upgrade Tests**: Update:
   - `TestDockerKubernetes{PREV_PREV}to{PREVIOUS_VERSION}EtcdScaleUp` → `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}EtcdScaleUp`
   - `TestDockerKubernetes{PREV_PREV}to{PREVIOUS_VERSION}EtcdScaleDown` → `TestDockerKubernetes{PREVIOUS_VERSION}to{NEW_VERSION}EtcdScaleDown`

Pattern for upgrade tests:
- Update function names to reflect new version transitions
- Update initial cluster version: `api.WithKubernetesVersion(v1alpha1.Kube{PREVIOUS_VERSION})`
- Update target upgrade version: `api.WithKubernetesVersion(v1alpha1.Kube{NEW_VERSION})`
- Update upgrade flow parameters: `v1alpha1.Kube{NEW_VERSION}`

Variables:
- {NEW_VERSION}: {NEW_VERSION}
- {PREVIOUS_VERSION}: {PREVIOUS_VERSION}
- {PREV_PREV}: {PREV_PREV_VERSION} (version before previous)
```

## Usage Instructions

1. **Preparation**: 
   - Identify the new Kubernetes version number (e.g., 134)
   - Identify the previous version number (e.g., 133)
   - Identify the provider name (Docker)

2. **Variable Substitution**:
   Replace the following variables in all prompts:
   - `{NEW_VERSION}`: New Kubernetes version (e.g., 134)
   - `{PREVIOUS_VERSION}`: Previous Kubernetes version (e.g., 133)
   - `{PREV_PREV_VERSION}`: Version before previous (e.g., 132)

3. **Execution**:
   - Use Cline's `new_task` tool to create separate tasks for each section
   - Execute tasks in order (1-6)
   - Verify each task completion before proceeding to the next

4. **Validation**:
   - After all tasks are complete, run the tests to ensure they compile and execute correctly
   - Check that all version references are updated consistently
   - Verify that upgrade paths are logical (previous → new version)

## Notes

- This plan is specifically for the Docker provider
- Kubernetes 1.33 support has already been added to the Docker provider (as of the referenced patch)
- Always verify that the new Kubernetes version constant exists in the codebase before starting
- Consider running a subset of tests first to validate the changes before full test suite execution
- When adding support for Kubernetes 1.34, use 1.33 as the {PREVIOUS_VERSION} and 1.32 as the {PREV_PREV_VERSION}
