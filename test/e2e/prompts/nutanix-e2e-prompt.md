# Nutanix E2E Test Automation Prompt Plan

This document provides a structured prompt plan for automating the creation of e2e tests for a new Kubernetes version in the Nutanix provider using AI coding agents like Cline.

## Overview

When EKS Anywhere adds support for a new Kubernetes version, comprehensive e2e tests need to be added across multiple files. This process involves updating build configurations, test definitions, framework helpers, and quick test configurations.

## Task Decomposition Strategy

Due to the large number of changes required across multiple files, this task should be decomposed into subtasks using Cline's `new_task` tool to avoid context overflow. Each subtask focuses on a specific file or category of changes.

## Subtask 1: Update Build Configuration Files

**Objective**: Update buildspec YAML files to include environment variables for the new Kubernetes version templates.

**Files to modify**:
- `cmd/integration_test/build/buildspecs/nutanix-test-eks-a-cli.yml`
- `cmd/integration_test/build/buildspecs/quick-test-eks-a-cli.yml`

**Changes required**:
1. Add new environment variables for the new Kubernetes version:
   - `T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_XX: "nutanix_ci:nutanix_template_ubuntu_1_XX"`
   - `T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_XX: "nutanix_ci:nutanix_template_rhel_9_1_XX"`

**Critical Note**: 
- Only add new environment variables. Do not remove existing old version variables.
- **⚠️ IMPORTANT: RedHat 8 is only supported up to Kubernetes 1.31**. For Kubernetes 1.32 and later, **DO NOT** add any RedHat 8 environment variables. Only add RedHat 9 and Ubuntu environment variables.
- If RedHat 8 environment variables were previously added for K8s 1.32+, they must be removed as they are incorrect.

**Pattern to follow**:
```yaml
# For Kubernetes 1.32 and later (RedHat 8 NOT supported - only Ubuntu and RedHat 9)
T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_32: "nutanix_ci:nutanix_template_ubuntu_1_32"
T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_32: "nutanix_ci:nutanix_template_rhel_9_1_32"

# For Kubernetes 1.31 and earlier (RedHat 8 still supported)
T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_31: "nutanix_ci:nutanix_template_ubuntu_1_31"
T_NUTANIX_TEMPLATE_NAME_REDHAT_1_31: "nutanix_ci:nutanix_template_rhel_8_1_31"
T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_31: "nutanix_ci:nutanix_template_rhel_9_1_31"

# ❌ WRONG - Do NOT add these for K8s 1.32+:
# T_NUTANIX_TEMPLATE_NAME_REDHAT_1_32: "nutanix_ci:nutanix_template_rhel_8_1_32"
# T_NUTANIX_TEMPLATE_NAME_REDHAT_1_33: "nutanix_ci:nutanix_template_rhel_8_1_33"
```

## Subtask 2: Update Quick Tests Configuration

**Objective**: Update the quick tests configuration to include upgrade tests for the new Kubernetes version.

**Files to modify**:
- `test/e2e/QUICK_TESTS.yaml`

**Changes required**:
1. Update Nutanix upgrade test names to reflect the new version range
2. Change from previous version upgrade (e.g., `130to131`) to new version upgrade (e.g., `131to132`)

**Pattern to follow**:
```yaml
# Nutanix - For Kubernetes 1.32 and later (NO RedHat 8 support - only Ubuntu and RedHat 9)
- TestNutanixKubernetes131to132RedHat9Upgrade
- TestNutanixKubernetes131to132StackedEtcdRedHat9Upgrade
- TestNutanixKubernetes131To132UbuntuUpgrade
- TestNutanixKubernetes131To132StackedEtcdUbuntuUpgrade

# For upgrades to Kubernetes 1.31 and earlier (RedHat 8 still supported)
- TestNutanixKubernetes130to131RedHat9Upgrade
- TestNutanixKubernetes130to131StackedEtcdRedHat9Upgrade
- TestNutanixKubernetes130to131RedHat8Upgrade
- TestNutanixKubernetes130to131StackedEtcdRedHat8Upgrade
- TestNutanixKubernetes130To131UbuntuUpgrade
- TestNutanixKubernetes130To131StackedEtcdUbuntuUpgrade

# ❌ WRONG - Do NOT add these for upgrades to K8s 1.32+:
# - TestNutanixKubernetes132to133RedHat8Upgrade
# - TestNutanixKubernetes132to133StackedEtcdRedHat8Upgrade
```

## Subtask 3: Add New Test Functions (Part 1 - Simple Flow Tests)

**Objective**: Add new test functions for basic functionality with the new Kubernetes version.

**Files to modify**:
- `test/e2e/nutanix_test.go`

**Changes required**:
Add new test functions for:
1. Simple flow tests with different OS combinations:
   - `TestNutanixKubernetes{VERSION}UbuntuSimpleFlowWithName`
   - `TestNutanixKubernetes{VERSION}RedHat9SimpleFlowWithName` (RedHat 8 only supported up to K8s 1.31)
   - `TestNutanixKubernetes{VERSION}UbuntuSimpleFlowWithUUID`
   - `TestNutanixKubernetes{VERSION}RedHat9SimpleFlowWithUUID` (RedHat 8 only supported up to K8s 1.31)

**Critical Note**: 
- **⚠️ For Kubernetes 1.32 and later, DO NOT add any RedHat 8 test functions**. Only add Ubuntu and RedHat 9 tests.
- If RedHat 8 test functions were previously added for K8s 1.32+, they must be removed as they are incorrect.
- RedHat 8 support ends at Kubernetes 1.31.

**Pattern to follow**:
```go
func TestNutanixKubernetes132UbuntuSimpleFlowWithName(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
	)
	runSimpleFlow(test)
}
```

## Subtask 4: Add New Test Functions (Part 2 - Curated Packages Tests)

**Objective**: Add new test functions for curated packages with the new Kubernetes version.

**Files to modify**:
- `test/e2e/nutanix_test.go`

**Changes required**:
Add new test functions for curated packages:
1. Basic curated packages: `TestNutanixKubernetes{VERSION}UbuntuCuratedPackagesSimpleFlow`
2. Emissary: `TestNutanixKubernetes{VERSION}UbuntuCuratedPackagesEmissarySimpleFlow`
3. ADOT: `TestNutanixKubernetes{VERSION}UbuntuCuratedPackagesAdotSimpleFlow`
4. Prometheus: `TestNutanixKubernetes{VERSION}UbuntuCuratedPackagesPrometheusSimpleFlow`
5. Cluster Autoscaler: `TestNutanixKubernetes{VERSION}UbuntuCuratedPackagesClusterAutoscalerSimpleFlow`
6. Harbor: `TestNutanixKubernetes{VERSION}UbuntuCuratedPackagesHarborSimpleFlow`

**Pattern to follow**:
```go
func TestNutanixKubernetes132UbuntuCuratedPackagesSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu132Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube132)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube132),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues, nil),
	)
	runCuratedPackageInstallSimpleFlow(test)
}
```

## Subtask 5: Add New Test Functions (Part 3 - Upgrade Tests)

**Objective**: Add new upgrade test functions for the new Kubernetes version.

**Files to modify**:
- `test/e2e/nutanix_test.go`

**Changes required**:
Add new upgrade test functions:
1. Ubuntu upgrades:
   - `TestNutanixKubernetes{PREV}To{NEW}StackedEtcdUbuntuUpgrade`
   - `TestNutanixKubernetes{PREV}To{NEW}UbuntuUpgrade`
2. RedHat 9 upgrades:
   - `TestNutanixKubernetes{PREV}to{NEW}StackedEtcdRedHat9Upgrade`
   - `TestNutanixKubernetes{PREV}to{NEW}RedHat9Upgrade`

**Critical Note**: 
- **⚠️ For upgrades to Kubernetes 1.32 and later, DO NOT add any RedHat 8 upgrade tests**. RedHat 8 support ends at Kubernetes 1.31.
- If RedHat 8 upgrade tests were previously added for upgrades to K8s 1.32+, they must be removed as they are incorrect.
- Only add Ubuntu and RedHat 9 upgrade tests for K8s 1.32+.

**Pattern to follow**:
```go
func TestNutanixKubernetes131To132StackedEtcdUbuntuUpgrade(t *testing.T) {
	provider := framework.NewNutanix(t, framework.WithUbuntu131Nutanix())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube131)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube132,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube132)),
		provider.WithProviderUpgrade(provider.Ubuntu132Template()),
	)
}
```

## Subtask 6: Add New Test Functions (Part 4 - Scaling and Other Tests)

**Objective**: Add new test functions for scaling, OIDC, AWS IAM Auth, and other features.

**Files to modify**:
- `test/e2e/nutanix_test.go`

**Changes required**:
Add new test functions for:
1. Worker node scaling: `TestNutanixKubernetes{VERSION}UbuntuWorkerNodeScaleUp1To3`, `TestNutanixKubernetes{VERSION}UbuntuWorkerNodeScaleDown3To1`
2. Control plane scaling: `TestNutanixKubernetes{VERSION}UbuntuControlPlaneNodeScaleUp1To3`, `TestNutanixKubernetes{VERSION}UbuntuControlPlaneNodeScaleDown3To1`
3. OIDC: `TestNutanixKubernetes{VERSION}OIDC`
4. AWS IAM Auth: `TestNutanixKubernetes{VERSION}AWSIamAuth`
5. Management CP upgrade: `TestNutanixKubernetes{VERSION}UbuntuManagementCPUpgradeAPI`
6. Kubelet configuration: `TestNutanixKubernetes{VERSION}KubeletConfigurationSimpleFlow`

## Subtask 7: Update Framework Helper Functions (Part 1 - Constants and Variables)

**Objective**: Add new constants and environment variables for the new Kubernetes version.

**Files to modify**:
- `test/framework/nutanix.go`

**Changes required**:
1. Add new template name constants:
   - `nutanixTemplateNameUbuntu{VERSION}Var`
   - `nutanixTemplateNameRedHat9{VERSION}Var`

**Critical Note**: 
- **⚠️ For Kubernetes 1.32 and later, DO NOT add `nutanixTemplateNameRedHat{VERSION}Var` constants** as RedHat 8 is not supported.
- If RedHat 8 constants were previously added for K8s 1.32+, they must be removed from both the constants section and the `requiredNutanixEnvVars` slice.
- Only add Ubuntu and RedHat 9 constants for K8s 1.32+.

2. Update `requiredNutanixEnvVars` slice to include new variables

**Note**: Only add new constants and variables. Do not remove existing old version constants.

**Pattern to follow**:
```go
// For Kubernetes 1.32 and later (NO RedHat 8 support)
nutanixTemplateNameUbuntu132Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_32"
nutanixTemplateNameRedHat9132Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_32"

// For Kubernetes 1.31 and earlier (RedHat 8 still supported)
nutanixTemplateNameUbuntu131Var  = "T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_31"
nutanixTemplateNameRedHat131Var  = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_31"
nutanixTemplateNameRedHat9131Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_9_1_31"

// ❌ WRONG - Do NOT add these for K8s 1.32+:
// nutanixTemplateNameRedHat132Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_32"
// nutanixTemplateNameRedHat133Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_33"
```

## Subtask 8: Update Framework Helper Functions (Part 2 - With Functions)

**Objective**: Add new "With" functions for the new Kubernetes version.

**Files to modify**:
- `test/framework/nutanix.go`

**Changes required**:
Add new functions:
1. `WithUbuntu{VERSION}Nutanix()`
2. `WithRedHat9Kubernetes{VERSION}Nutanix()`
3. `WithUbuntu{VERSION}NutanixUUID()`
4. `WithRedHat9Kubernetes{VERSION}NutanixUUID()`

**Critical Note**: 
- **⚠️ For Kubernetes 1.32 and later, DO NOT add `WithRedHat{VERSION}Nutanix()` or `WithRedHat{VERSION}NutanixUUID()` functions** as RedHat 8 is not supported.
- If RedHat 8 "With" functions were previously added for K8s 1.32+, they must be removed as they are incorrect.
- Only add Ubuntu and RedHat 9 "With" functions for K8s 1.32+.

**Pattern to follow**:
```go
// WithUbuntu132Nutanix returns a NutanixOpt that adds API fillers to use a Ubuntu Nutanix template for k8s 1.32
// and the "ubuntu" osFamily in all machine configs.
func WithUbuntu132Nutanix() NutanixOpt {
	return withNutanixKubeVersionAndOS(anywherev1.Kube132, Ubuntu2004, nil)
}
```

## Subtask 9: Update Framework Helper Functions (Part 3 - Template Functions)

**Objective**: Add new template functions for the new Kubernetes version.

**Files to modify**:
- `test/framework/nutanix.go`

**Changes required**:
Add new template functions:
1. `Ubuntu{VERSION}Template()`
2. `RedHat9Kubernetes{VERSION}Template()`

**Critical Note**: 
- **⚠️ For Kubernetes 1.32 and later, DO NOT add `RedHat{VERSION}Template()` functions** as RedHat 8 is not supported.
- If RedHat 8 template functions were previously added for K8s 1.32+, they must be removed as they are incorrect.
- Only add Ubuntu and RedHat 9 template functions for K8s 1.32+.

**Pattern to follow**:
```go
// Ubuntu132Template returns NutanixFiller by reading the env var and setting machine config's
// image name parameter in the spec.
func (n *Nutanix) Ubuntu132Template() api.NutanixFiller {
	return n.templateForKubeVersionAndOS(anywherev1.Kube132, Ubuntu2004, nil)
}
```

## Input Parameters for Automation

When using this prompt plan, provide the following parameters:

1. **NEW_K8S_VERSION**: The new Kubernetes version (e.g., "132", "133")
2. **PREV_K8S_VERSION**: The previous Kubernetes version for upgrade tests (e.g., "131", "132")
3. **VERSION_CONSTANT**: The version constant used in code (e.g., "v1alpha1.Kube132")

## Execution Strategy

1. **Use Cline's new_task tool** to create separate tasks for each subtask
2. **Execute subtasks sequentially** to avoid context overflow
3. **Validate changes** after each subtask completion
4. **Run tests** to ensure new test functions work correctly

## Validation Steps

After completing all subtasks:

1. Verify all new test functions compile without errors
2. Check that environment variables are properly referenced
3. Ensure upgrade test version ranges are correct
4. Validate that framework helper functions return correct values
5. Run a subset of new tests to verify functionality
6. **⚠️ Critical RedHat 8 Validation**: Verify NO RedHat 8 support exists for K8s 1.32+:
   - Search for `REDHAT_1_32`, `REDHAT_1_33`, etc. in buildspec files
   - Search for `RedHat8.*132`, `RedHat8.*133`, etc. in test files
   - Search for `RedHat132`, `RedHat133`, etc. in framework files
   - If found, these must be removed as they are incorrect


## Example Usage

```
Please add e2e tests for Kubernetes version 1.33 for the Nutanix provider. 
Use the following parameters:
- NEW_K8S_VERSION: 133
- PREV_K8S_VERSION: 132  
- VERSION_CONSTANT: v1alpha1.Kube133

Follow the subtask decomposition strategy outlined in nutanix-e2e-prompt.md to avoid context overflow.
```

## Notes

- This prompt plan is based on the reference commit for Kubernetes 1.32 Nutanix e2e tests
- The pattern should be adaptable to other providers with similar structure
- Always verify that the new Kubernetes version constant exists in the codebase before proceeding
- Consider running existing tests to ensure no regressions are introduced
- **Important**: This plan only adds support for new Kubernetes versions. It does not remove support for old versions, as old version tests should be preserved for backward compatibility and testing purposes.

## ⚠️ Critical RedHat 8 Support Limitation

**RedHat 8 is ONLY supported up to Kubernetes 1.31**. This is a hard limitation that must be strictly enforced:

- **For Kubernetes 1.32 and later**: Only Ubuntu and RedHat 9 are supported
- **Do NOT create any RedHat 8 tests, constants, functions, or environment variables for K8s 1.32+**
- **If RedHat 8 support was previously added for K8s 1.32+, it must be removed** (as shown in the reference patch)

### Files that commonly have incorrect RedHat 8 entries for K8s 1.32+:
1. `cmd/integration_test/build/buildspecs/*.yml` - Environment variables
2. `test/e2e/QUICK_TESTS.yaml` - Quick test entries
3. `test/e2e/nutanix_test.go` - Test functions
4. `test/framework/nutanix.go` - Constants, variables, and helper functions

### Common incorrect patterns to avoid:
```yaml
# ❌ WRONG - Do not add for K8s 1.32+
T_NUTANIX_TEMPLATE_NAME_REDHAT_1_32: "nutanix_ci:nutanix_template_rhel_8_1_32"
- TestNutanixKubernetes132to133RedHat8Upgrade
```

```go
// ❌ WRONG - Do not add for K8s 1.32+
nutanixTemplateNameRedHat132Var = "T_NUTANIX_TEMPLATE_NAME_REDHAT_1_32"
func WithRedHat132Nutanix() NutanixOpt { ... }
func (n *Nutanix) RedHat132Template() api.NutanixFiller { ... }
```

**Always double-check that no RedHat 8 support is being added for Kubernetes 1.32 and later versions.**
