# CloudStack E2E Test Automation Prompt Plan

## Overview
This prompt plan automates the creation of e2e tests for a new Kubernetes version on the CloudStack provider, based on the reference implementation for Kubernetes 1.32. The plan is designed to be executed by AI coding agents like Cline and is broken down into manageable subtasks to avoid context overflow.

## Prerequisites
- Target Kubernetes version (e.g., 1.33, 1.34, etc.)
- Previous Kubernetes version to use as base
- CloudStack provider support already implemented in the codebase

## Important Version Compatibility Notes
- **RedHat8 Support**: RedHat8 is only supported for Kubernetes versions up to 1.31
- **RedHat9 Support**: RedHat9 is supported for all Kubernetes versions (1.28+)
- **Version-Specific Logic**: The prompt plan includes conditional logic based on target Kubernetes version
- **Key Change**: From K8s 1.32 onwards, only RedHat9 tests should be created (no new RedHat8 tests)

## Task Decomposition Strategy

### Phase 1: Configuration Files Update
**Subtask 1.1: Update QUICK_TESTS.yaml**
- File: `test/e2e/QUICK_TESTS.yaml`
- Action: Update CloudStack test entries to use new Kubernetes version
- Pattern: Replace version numbers in test names (e.g., `131To132` -> `132To133`)

**Subtask 1.2: Update SKIPPED_TESTS.yaml**
- File: `test/e2e/SKIPPED_TESTS.yaml`
- Action: Update CloudStack test entries for new Kubernetes version
- Pattern: Replace version numbers in test names and add new entries as needed

### Phase 2: Framework Helper Functions
**Subtask 2.1: Add CloudStack Framework Helpers**
- File: `test/framework/cloudstack.go`
- Actions:
  - **For Kubernetes versions 1.31 and earlier:**
    - Add `WithCloudStackRedhat{NEW_VERSION}()` function (RedHat8)
    - Add `WithCloudStackRedhat9Kubernetes{NEW_VERSION}()` function (RedHat9)
    - Add `Redhat{NEW_VERSION}Template()` function (RedHat8)
    - Add `Redhat9Kubernetes{NEW_VERSION}Template()` function (RedHat9)
    - Add `WithRedhat{NEW_VERSION}()` function (RedHat8)
    - Add `WithRedhat9Kubernetes{NEW_VERSION}()` function (RedHat9)
    - Update `WithRedhatVersion()` switch statement
  - **For Kubernetes versions 1.32 and later:**
    - Add `WithCloudStackRedhat9Kubernetes{NEW_VERSION}()` function (RedHat9)
    - Add `Redhat9Kubernetes{NEW_VERSION}Template()` function (RedHat9)
    - Add `WithRedhat9Kubernetes{NEW_VERSION}()` function (RedHat9)
    - **DO NOT** add RedHat8 functions for K8s 1.32+ (RedHat8 support discontinued)
    - **IMPORTANT**: If simple tests patch added RedHat8 functions for K8s 1.32+, remove them as they were wrongly added


### Phase 3: Test Functions - Core Infrastructure
**Subtask 3.1: API Server Extra Args Tests**
- **For Kubernetes versions 1.31 and earlier:**
  - Update `TestCloudStackKubernetes{PREV_VERSION}RedHat8APIServerExtraArgsSimpleFlow` to `TestCloudStackKubernetes{NEW_VERSION}RedHat8APIServerExtraArgsSimpleFlow`
  - Update `TestCloudStackKubernetes{PREV_VERSION}Redhat8APIServerExtraArgsUpgradeFlow` to `TestCloudStackKubernetes{NEW_VERSION}Redhat8APIServerExtraArgsUpgradeFlow`
  - Update `TestCloudStackKubernetes{PREV_VERSION}RedHat9APIServerExtraArgsSimpleFlow` to `TestCloudStackKubernetes{NEW_VERSION}RedHat9APIServerExtraArgsSimpleFlow`
  - Update `TestCloudStackKubernetes{PREV_VERSION}Redhat9APIServerExtraArgsUpgradeFlow` to `TestCloudStackKubernetes{NEW_VERSION}Redhat9APIServerExtraArgsUpgradeFlow`
- **For Kubernetes versions 1.32 and later:**
  - Update `TestCloudStackKubernetes{PREV_VERSION}RedHat9APIServerExtraArgsSimpleFlow` to `TestCloudStackKubernetes{NEW_VERSION}RedHat9APIServerExtraArgsSimpleFlow`
  - Update `TestCloudStackKubernetes{PREV_VERSION}Redhat9APIServerExtraArgsUpgradeFlow` to `TestCloudStackKubernetes{NEW_VERSION}Redhat9APIServerExtraArgsUpgradeFlow`
  - **IMPORTANT**: Do not update RedHat8 versions for K8s 1.32+ (RedHat8 support discontinued)

**Subtask 3.2: AWS IAM Auth Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}AWSIamAuth`
- Add `TestCloudStackKubernetes{PREV_VERSION}to{NEW_VERSION}AWSIamAuthUpgrade`

### Phase 4: Test Functions - Curated Packages
**Subtask 4.1: Basic Curated Packages Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesSimpleFlow`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatWorkloadClusterCuratedPackagesSimpleFlow`

**Subtask 4.2: Emissary Package Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesEmissarySimpleFlow`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatWorkloadClusterCuratedPackagesEmissarySimpleFlow`

**Subtask 4.3: Harbor Package Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesHarborSimpleFlow`

**Subtask 4.4: Cert Manager Package Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesCertManagerSimpleFlow`

**Subtask 4.5: ADOT Package Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesAdotSimpleFlow`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesAdotUpdateFlow`

**Subtask 4.6: Cluster Autoscaler Package Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedHatCuratedPackagesClusterAutoscalerSimpleFlow`

**Subtask 4.7: Prometheus Package Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatCuratedPackagesPrometheusSimpleFlow`

### Phase 5: Test Functions - GitOps and Flux
**Subtask 5.1: GitHub Flux Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}GithubFlux`
- Add `TestCloudStackKubernetes{NEW_VERSION}GitFlux`
- Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}GitFluxUpgrade`
- Add `TestCloudStackKubernetes{NEW_VERSION}InstallGitFluxDuringUpgrade`

**Subtask 5.2: Multicluster GitOps Tests**
- Add `TestCloudStackUpgradeKubernetes{NEW_VERSION}MulticlusterWorkloadClusterWithGithubFlux`

### Phase 6: Test Functions - Authentication and Security
**Subtask 6.1: OIDC Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}WithOIDCManagementClusterUpgradeFromLatestSideEffects`
- Add `TestCloudStackKubernetes{NEW_VERSION}OIDC`
- Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}OIDCUpgrade`

**Subtask 6.2: Proxy Configuration Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatProxyConfig`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatProxyConfigAPI`

### Phase 7: Test Functions - Registry and Networking
**Subtask 7.1: Registry Mirror Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatRegistryMirrorInsecureSkipVerify`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatRegistryMirrorAndCert`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatAuthenticatedRegistryMirror`

**Subtask 7.2: Cilium Policy Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}CiliumAlwaysPolicyEnforcementModeSimpleFlow`

### Phase 8: Test Functions - Basic Cluster Operations
**Subtask 8.1: Simple Flow Tests**
- **For Kubernetes versions 1.31 and earlier:**
  - Add `TestCloudStackKubernetes{NEW_VERSION}RedHat8SimpleFlow`
  - Add `TestCloudStackKubernetes{NEW_VERSION}RedHat9SimpleFlow`
- **For Kubernetes versions 1.32 and later:**
  - Add `TestCloudStackKubernetes{NEW_VERSION}RedHat9SimpleFlow` (RedHat9 continues to be supported)
  - **DO NOT** add `TestCloudStackKubernetes{NEW_VERSION}RedHat8SimpleFlow` (RedHat8 support discontinued)
- **For all versions:**
  - Add `TestCloudStackKubernetes{NEW_VERSION}ThreeReplicasFiveWorkersSimpleFlow`
  - Add `TestCloudStackKubernetes{NEW_VERSION}MultiEndpointSimpleFlow`
  - Add `TestCloudStackKubernetes{NEW_VERSION}DifferentNamespaceSimpleFlow`

**Subtask 8.2: Stacked Etcd Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}StackedEtcdRedhat`

### Phase 9: Test Functions - Node Management and Scaling
**Subtask 9.1: Labels and Node Management Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}LabelsAndNodeNameRedhat`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatLabelsUpgradeFlow`
- Add helper function `redhat{NEW_VERSION}ProviderWithLabels`

**Subtask 9.2: Taints Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatTaintsUpgradeFlow`
- Add helper function `redhat{NEW_VERSION}ProviderWithTaints`

### Phase 10: Test Functions - Multicluster Operations
**Subtask 10.1: Multicluster Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}MulticlusterWorkloadCluster`

### Phase 11: Test Functions - Upgrade Operations
**Subtask 11.1: Basic Upgrade Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatAndRemoveWorkerNodeGroups`
- **For target Kubernetes versions 1.31 and earlier:**
  - Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat8UnstackedEtcdUpgrade`
  - Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat8StackedEtcdUpgrade`
  - Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat9UnstackedEtcdUpgrade`
  - Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat9StackedEtcdUpgrade`
- **For target Kubernetes versions 1.32 and later:**
  - Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat9UnstackedEtcdUpgrade` (RedHat9 only)
  - Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat9StackedEtcdUpgrade` (RedHat9 only)
  - **DO NOT** add RedHat8 upgrade tests for K8s 1.32+

**Subtask 11.2: OS Upgrade Tests**
- **For target Kubernetes versions 1.31 and earlier:**
  - Add `TestCloudStackKubernetes{NEW_VERSION}Redhat8ToRedhat9Upgrade`
- **For target Kubernetes versions 1.32 and later:**
  - **DO NOT** add RedHat8 to RedHat9 upgrade tests (RedHat8 not supported)

**Subtask 11.3: Checkpoint Upgrade Tests**
- Add `TestCloudStackKubernetes{PREV_VERSION}RedhatTo{NEW_VERSION}UpgradeWithCheckpoint`

**Subtask 11.4: Node Scaling Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatControlPlaneNodeUpgrade`
- Add `TestCloudStackKubernetes{NEW_VERSION}RedhatWorkerNodeUpgrade`

**Subtask 11.5: Multiple Fields Upgrade Tests**
- Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}RedhatMultipleFieldsUpgrade`
- Add `TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}StackedEtcdRedhatMultipleFieldsUpgrade`

### Phase 12: Test Functions - Special Features
**Subtask 12.1: Management Components Tests**
- Update `TestCloudStackKubernetes{PREV_VERSION}UpgradeManagementComponents` to use new version

**Subtask 12.2: Download Artifacts Tests**
- Update `TestCloudStackDownloadArtifacts` to use new version
- Update `TestCloudStackRedhat9DownloadArtifacts` to use new version

**Subtask 12.3: Airgapped and Special Environment Tests**
- Update `TestCloudStackKubernetes{PREV_VERSION}RedhatAirgappedProxy` to use new version

**Subtask 12.4: Workload API Tests**
- Update API-related tests to use new version:
  - `TestCloudStackKubernetes{PREV_VERSION}MulticlusterWorkloadClusterAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}MulticlusterWorkloadClusterNewCredentialsSecretsAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}MulticlusterWorkloadClusterGitHubFluxAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}MulticlusterWorkloadClusterNewCredentialsSecretGitHubFluxAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}WorkloadClusterAWSIamAuthAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}WorkloadClusterAWSIamAuthGithubFluxAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}WorkloadClusterOIDCAuthAPI`
  - `TestCloudStackKubernetes{PREV_VERSION}WorkloadClusterOIDCAuthGithubFluxAPI`

**Subtask 12.5: Etcd and Advanced Features Tests**
- Update `TestCloudStackKubernetes{PREV_VERSION}EtcdEncryption` to use new version
- Update `TestCloudStackKubernetes{PREV_VERSION}ValidateDomainFourLevelsSimpleFlow` to use new version
- Update `TestCloudStackKubernetes{PREV_VERSION}EtcdScaleUp` to use new version
- Update `TestCloudStackKubernetes{PREV_VERSION}EtcdScaleDown` to use new version

**Subtask 12.6: Kubelet Configuration Tests**
- Add `TestCloudStackKubernetes{NEW_VERSION}KubeletConfigurationSimpleFlow`

## Implementation Guidelines

### Variable Substitution Patterns
- `{NEW_VERSION}`: Target Kubernetes version (e.g., 133, 134)
- `{PREV_VERSION}`: Previous Kubernetes version (e.g., 132, 133)
- `{NEW_K8S_VERSION}`: Kubernetes version constant (e.g., v1alpha1.Kube133)
- `{PREV_K8S_VERSION}`: Previous Kubernetes version constant (e.g., v1alpha1.Kube132)

### Code Pattern Templates

#### Test Function Template
**For Kubernetes versions 1.31 and earlier:**
```go
func TestCloudStackKubernetes{NEW_VERSION}RedHat8SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat{NEW_VERSION}()),
		framework.WithClusterFiller(api.WithKubernetesVersion({NEW_K8S_VERSION})),
	)
	runSimpleFlow(test)
}
```

**For Kubernetes versions 1.32 and later (RedHat9 continues, RedHat8 discontinued):**
```go
func TestCloudStackKubernetes{NEW_VERSION}RedHat9SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes{NEW_VERSION}()),
		framework.WithClusterFiller(api.WithKubernetesVersion({NEW_K8S_VERSION})),
	)
	runSimpleFlow(test)
}
```

#### Upgrade Test Template
```go
func TestCloudStackKubernetes{PREV_VERSION}To{NEW_VERSION}Redhat9StackedEtcdUpgrade(t *testing.T) {
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat9Kubernetes{PREV_VERSION}())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion({PREV_K8S_VERSION})),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		{NEW_K8S_VERSION},
		framework.WithClusterUpgrade(api.WithKubernetesVersion({NEW_K8S_VERSION})),
		provider.WithProviderUpgrade(provider.Redhat9Kubernetes{NEW_VERSION}Template()),
	)
}
```

#### Framework Helper Template
**For Kubernetes versions 1.31 and earlier:**
```go
// WithCloudStackRedhat{NEW_VERSION} returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.{NEW_VERSION}.
func WithCloudStackRedhat{NEW_VERSION}() CloudStackOpt {
	return withCloudStackKubeVersionAndOS({NEW_K8S_VERSION}, RedHat8, nil)
}

// Redhat{NEW_VERSION}Template returns cloudstack filler for 1.{NEW_VERSION} RedHat.
func (c *CloudStack) Redhat{NEW_VERSION}Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS({NEW_K8S_VERSION}, RedHat8, nil)
}

// WithCloudStackRedhat9Kubernetes{NEW_VERSION} returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.{NEW_VERSION} RedHat9.
func WithCloudStackRedhat9Kubernetes{NEW_VERSION}() CloudStackOpt {
	return withCloudStackKubeVersionAndOS({NEW_K8S_VERSION}, RedHat9, nil)
}

// Redhat9Kubernetes{NEW_VERSION}Template returns cloudstack filler for 1.{NEW_VERSION} RedHat9.
func (c *CloudStack) Redhat9Kubernetes{NEW_VERSION}Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS({NEW_K8S_VERSION}, RedHat9, nil)
}
```

**For Kubernetes versions 1.32 and later (RedHat9 continues, RedHat8 discontinued):**
```go
// WithCloudStackRedhat9Kubernetes{NEW_VERSION} returns a function which can be invoked to configure the Cloudstack object to be compatible with K8s 1.{NEW_VERSION} RedHat9.
func WithCloudStackRedhat9Kubernetes{NEW_VERSION}() CloudStackOpt {
	return withCloudStackKubeVersionAndOS({NEW_K8S_VERSION}, RedHat9, nil)
}

// Redhat9Kubernetes{NEW_VERSION}Template returns cloudstack filler for 1.{NEW_VERSION} RedHat9.
func (c *CloudStack) Redhat9Kubernetes{NEW_VERSION}Template() api.CloudStackFiller {
	return c.templateForKubeVersionAndOS({NEW_K8S_VERSION}, RedHat9, nil)
}
```

## Execution Strategy

### Using Cline's New Task Tool
Each phase should be executed as a separate task using Cline's `new_task` tool to avoid context overflow:

1. **Create New Task for Each Phase**: Use `new_task` with detailed context about the current phase
2. **Sequential Execution**: Complete phases in order to maintain dependencies
3. **Validation Between Phases**: Verify changes compile and tests are syntactically correct
4. **Context Preservation**: Include relevant file paths, function names, and patterns in each task context

### Task Context Template for New Task Tool
```
Current Work: Adding Kubernetes {NEW_VERSION} e2e tests for CloudStack provider - Phase {N}: {PHASE_NAME}

Key Technical Concepts:
- CloudStack provider e2e testing framework
- Kubernetes version {NEW_VERSION} support
- Test function naming patterns: TestCloudStackKubernetes{VERSION}...
- Framework helper function patterns: WithCloudStackRedhat{VERSION}()

Relevant Files:
- test/e2e/cloudstack_test.go: Main test file containing all CloudStack e2e tests
- test/framework/cloudstack.go: Framework helper functions for CloudStack provider
- test/e2e/QUICK_TESTS.yaml: Configuration for quick test suite
- test/e2e/SKIPPED_TESTS.yaml: Configuration for skipped tests

Problem Solving:
- Following established patterns from Kubernetes 1.32 implementation
- Maintaining consistency with existing test structure
- Ensuring proper version substitution throughout

Pending Tasks:
- {SPECIFIC_SUBTASKS_FOR_THIS_PHASE}

Next Steps:
- {DETAILED_NEXT_STEPS_FOR_CURRENT_PHASE}
```

## Quality Assurance

### Validation Checklist
- [ ] All version numbers correctly substituted
- [ ] Test function names follow established patterns
- [ ] Framework helper functions added with correct signatures
- [ ] Configuration files updated appropriately
- [ ] No syntax errors in generated code
- [ ] Test functions use appropriate framework methods
- [ ] Upgrade tests reference correct source and target versions

### Testing Strategy
1. **Compilation Check**: Ensure all generated code compiles successfully
2. **Pattern Verification**: Verify all functions follow established naming patterns
3. **Dependency Check**: Ensure all referenced framework functions exist
4. **Configuration Validation**: Verify YAML configuration files are valid

## Notes
- This plan is based on the CloudStack Kubernetes 1.32 implementation
- Adjust version numbers and constants according to target Kubernetes version
- Some tests may need provider-specific modifications
- Consider any breaking changes in the target Kubernetes version
- Maintain backward compatibility where possible

## Critical Version-Specific Requirements
- **RedHat8 Deprecation**: RedHat8 support ends with Kubernetes 1.31
- **RedHat9 Continuation**: RedHat9 support continues for all Kubernetes versions (was already supported before 1.32)
- **No New RedHat8 Tests**: For K8s 1.32+, do not create new RedHat8 tests or framework helpers
- **Simple Tests Patch Conflict**: The simple tests patch may have added RedHat8 tests for K8s 1.32+, but these were later removed as "wrongly added"
- **Cleanup Required**: If implementing for K8s 1.32+, ensure any RedHat8 tests from simple tests patch are not included
- **Upgrade Path Considerations**: Upgrade tests from K8s 1.31 (RedHat8) to K8s 1.32+ (RedHat9) should be OS upgrade tests
- **Test Naming**: Ensure test names reflect the correct OS version (RedHat8 vs RedHat9)
- **CloudStack Focus**: This prompt plan is specifically for CloudStack provider e2e tests only

## Version Decision Matrix
| Kubernetes Version | RedHat8 Support | RedHat9 Support | Notes |
|-------------------|----------------|----------------|-------|
| 1.28 - 1.31       | ✅ Yes         | ✅ Yes         | Both OS versions supported |
| 1.32+             | ❌ No          | ✅ Yes         | RedHat9 continues, RedHat8 discontinued |
