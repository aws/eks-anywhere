/*
Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License").
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package snow

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

const (
	// PrincipalCredentialRetrievedCondition reports on whether Principal credentials could be retrieved successfully.
	// A possible scenario, where retrieval is unsuccessful, is when SourcePrincipal is not authorized for assume role.
	PrincipalCredentialRetrievedCondition clusterv1.ConditionType = "PrincipalCredentialRetrieved"
	// PrincipalCredentialRetrievalFailedReason used when errors occur during identity credential retrieval.
	PrincipalCredentialRetrievalFailedReason = "PrincipalCredentialRetrievalFailed"
	// CredentialProviderBuildFailedReason used when errors occur during building providers before trying credential retrieval.
	CredentialProviderBuildFailedReason = "CredentialProviderBuildFailed"
	// PrincipalUsageAllowedCondition reports on whether Principal and all the nested source identities are allowed to be used in the AWSCluster namespace.
	PrincipalUsageAllowedCondition clusterv1.ConditionType = "PrincipalUsageAllowed"
	// PrincipalUsageUnauthorizedReason used when AWSCluster namespace is not in the identity's allowed namespaces list.
	PrincipalUsageUnauthorizedReason = "PrincipalUsageUnauthorized"
	// SourcePrincipalUsageUnauthorizedReason used when AWSCluster is not in the intersection of source identity allowed namespaces
	// and allowed namespaces of the identities that source identity depends to.
	SourcePrincipalUsageUnauthorizedReason = "SourcePrincipalUsageUnauthorized"
)

const (
	// VpcReadyCondition reports on the successful reconciliation of a VPC.
	VpcReadyCondition clusterv1.ConditionType = "VpcReady"
	// VpcCreationStartedReason used when attempting to create a VPC for a managed cluster.
	// Will not be applied to unmanaged clusters.
	VpcCreationStartedReason = "VpcCreationStarted"
	// VpcReconciliationFailedReason used when errors occur during VPC reconciliation.
	VpcReconciliationFailedReason = "VpcReconciliationFailed"
)

const (
	// SubnetsReadyCondition reports on the successful reconciliation of subnets.
	SubnetsReadyCondition clusterv1.ConditionType = "SubnetsReady"
	// SubnetsReconciliationFailedReason used to report failures while reconciling subnets.
	SubnetsReconciliationFailedReason = "SubnetsReconciliationFailed"
)

const (
	// InternetGatewayReadyCondition reports on the successful reconciliation of internet gateways.
	// Only applicable to managed clusters.
	InternetGatewayReadyCondition clusterv1.ConditionType = "InternetGatewayReady"
	// InternetGatewayFailedReason used when errors occur during internet gateway reconciliation.
	InternetGatewayFailedReason = "InternetGatewayFailed"
)

const (
	// NatGatewaysReadyCondition reports successful reconciliation of NAT gateways.
	// Only applicable to managed clusters.
	NatGatewaysReadyCondition clusterv1.ConditionType = "NatGatewaysReady"
	// NatGatewaysCreationStartedReason set once when creating new NAT gateways.
	NatGatewaysCreationStartedReason = "NatGatewaysCreationStarted"
	// NatGatewaysReconciliationFailedReason used when any errors occur during reconciliation of NAT gateways.
	NatGatewaysReconciliationFailedReason = "NatGatewaysReconciliationFailed"
)

const (
	// RouteTablesReadyCondition reports successful reconciliation of route tables.
	// Only applicable to managed clusters.
	RouteTablesReadyCondition clusterv1.ConditionType = "RouteTablesReady"
	// RouteTableReconciliationFailedReason used when any errors occur during reconciliation of route tables.
	RouteTableReconciliationFailedReason = "RouteTableReconciliationFailed"
)

const (
	// SecondaryCidrsReadyCondition reports successful reconciliation of secondary CIDR blocks.
	// Only applicable to managed clusters.
	SecondaryCidrsReadyCondition clusterv1.ConditionType = "SecondaryCidrsReady"
	// SecondaryCidrReconciliationFailedReason used when any errors occur during reconciliation of secondary CIDR blocks.
	SecondaryCidrReconciliationFailedReason = "SecondaryCidrReconciliationFailed"
)

const (
	// ClusterSecurityGroupsReadyCondition reports successful reconciliation of security groups.
	ClusterSecurityGroupsReadyCondition clusterv1.ConditionType = "ClusterSecurityGroupsReady"
	// ClusterSecurityGroupReconciliationFailedReason used when any errors occur during reconciliation of security groups.
	ClusterSecurityGroupReconciliationFailedReason = "SecurityGroupReconciliationFailed"
)

const (
	// BastionHostReadyCondition reports whether a bastion host is ready. Depending on the configuration, a cluster
	// may not require a bastion host and this condition will be skipped.
	BastionHostReadyCondition clusterv1.ConditionType = "BastionHostReady"
	// BastionCreationStartedReason used when creating a new bastion host.
	BastionCreationStartedReason = "BastionCreationStarted"
	// BastionHostFailedReason used when an error occurs during the creation of a bastion host.
	BastionHostFailedReason = "BastionHostFailed"
)

const (
	// LoadBalancerReadyCondition reports on whether a control plane load balancer was successfully reconciled.
	LoadBalancerReadyCondition clusterv1.ConditionType = "LoadBalancerReady"
	// WaitForDNSNameReason used while waiting for a DNS name for the API server to be populated.
	WaitForDNSNameReason = "WaitForDNSName"
	// WaitForDNSNameResolveReason used while waiting for DNS name to resolve.
	WaitForDNSNameResolveReason = "WaitForDNSNameResolve"
	// LoadBalancerFailedReason used when an error occurs during load balancer reconciliation.
	LoadBalancerFailedReason = "LoadBalancerFailed"
)

const (
	// InstanceReadyCondition reports on current status of the EC2 instance. Ready indicates the instance is in a Running state.
	InstanceReadyCondition clusterv1.ConditionType = "InstanceReady"

	// InstanceNotFoundReason used when the instance couldn't be retrieved.
	InstanceNotFoundReason = "InstanceNotFound"
	// InstanceTerminatedReason instance is in a terminated state.
	InstanceTerminatedReason = "InstanceTerminated"
	// InstanceStoppedReason instance is in a stopped state.
	InstanceStoppedReason = "InstanceStopped"
	// InstanceNotReadyReason used when the instance is in a pending state.
	InstanceNotReadyReason = "InstanceNotReady"
	// InstanceProvisionStartedReason set when the provisioning of an instance started.
	InstanceProvisionStartedReason = "InstanceProvisionStarted"
	// InstanceProvisionFailedReason used for failures during instance provisioning.
	InstanceProvisionFailedReason = "InstanceProvisionFailed"
	// WaitingForClusterInfrastructureReason used when machine is waiting for cluster infrastructure to be ready before proceeding.
	WaitingForClusterInfrastructureReason = "WaitingForClusterInfrastructure"
	// WaitingForBootstrapDataReason used when machine is waiting for bootstrap data to be ready before proceeding.
	WaitingForBootstrapDataReason = "WaitingForBootstrapData"
)

// TODO add Snowball Device conditions, e.g., direct-network-interface ready condition

const (
	// SecurityGroupsReadyCondition indicates the security groups are up to date on the AWSMachine.
	SecurityGroupsReadyCondition clusterv1.ConditionType = "SecurityGroupsReady"

	// SecurityGroupsFailedReason used when the security groups could not be synced.
	SecurityGroupsFailedReason = "SecurityGroupsSyncFailed"
)

const (
	// ELBAttachedCondition will report true when a control plane is successfully registered with an ELB.
	// When set to false, severity can be an Error if the subnet is not found or unavailable in the instance's AZ.
	// Note this is only applicable to control plane machines.
	// Only applicable to control plane machines.
	ELBAttachedCondition clusterv1.ConditionType = "ELBAttached"

	// ELBAttachFailedReason used when a control plane node fails to attach to the ELB.
	ELBAttachFailedReason = "ELBAttachFailed"
	// ELBDetachFailedReason used when a control plane node fails to detach from an ELB.
	ELBDetachFailedReason = "ELBDetachFailed"
)

const (
	Bottlerocket OSFamily = "bottlerocket"
	Ubuntu       OSFamily = "ubuntu"
)
