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

import (
	"k8s.io/apimachinery/pkg/util/sets"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// AWSResourceReference is a reference to a specific AWS resource by ID, ARN, or filters.
// Only one of ID, ARN or Filters may be specified. Specifying more than one will result in
// a validation error.
type AWSResourceReference struct {
	// ID of resource
	// +optional
	ID *string `json:"id,omitempty"`

	// ARN of resource
	// +optional
	ARN *string `json:"arn,omitempty"`

	// Filters is a set of key/value pairs used to identify a resource
	// They are applied according to the rules defined by the AWS API:
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html
	// +optional
	Filters []Filter `json:"filters,omitempty"`
}

// Filter is a filter used to identify an AWS resource
type Filter struct {
	// Name of the filter. Filter names are case-sensitive.
	Name string `json:"name"`

	// Values includes one or more filter values. Filter values are case-sensitive.
	Values []string `json:"values"`
}

// Instance describes an AWS instance.
type Instance struct {
	ID string `json:"id"`

	// The current state of the instance.
	State InstanceState `json:"instanceState,omitempty"`

	// The instance type.
	Type string `json:"type,omitempty"`

	// The ID of the subnet of the instance.
	SubnetID string `json:"subnetId,omitempty"`

	// The ID of the AMI used to launch the instance.
	ImageID string `json:"imageId,omitempty"`

	// The name of the SSH key pair.
	SSHKeyName *string `json:"sshKeyName,omitempty"`

	// SecurityGroupIDs are one or more security group IDs this instance belongs to.
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// UserData is the raw data script passed to the instance which is run upon bootstrap.
	// This field must not be base64 encoded and should only be used when running a new instance.
	UserData *string `json:"userData,omitempty"`

	// The name of the IAM instance profile associated with the instance, if applicable.
	IAMProfile string `json:"iamProfile,omitempty"`

	// Addresses contains the AWS instance associated addresses.
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// The private IPv4 address assigned to the instance.
	PrivateIP *string `json:"privateIp,omitempty"`

	// The public IPv4 address assigned to the instance, if applicable.
	PublicIP *string `json:"publicIp,omitempty"`

	// Specifies whether enhanced networking with ENA is enabled.
	ENASupport *bool `json:"enaSupport,omitempty"`

	// Indicates whether the instance is optimized for Amazon EBS I/O.
	EBSOptimized *bool `json:"ebsOptimized,omitempty"`

	// Configuration options for the root storage volume.
	// +optional
	RootVolume *Volume `json:"rootVolume,omitempty"`

	// Configuration options for the non root storage volumes.
	// +optional
	NonRootVolumes []*Volume `json:"nonRootVolumes,omitempty"`

	// Configuration options for the containers data volume
	// +optional
	ContainersVolume *Volume `json:"containersVolume,omitempty"`

	// Specifies ENIs attached to instance
	NetworkInterfaces []string `json:"networkInterfaces,omitempty"`

	// The tags associated with the instance.
	Tags map[string]string `json:"tags,omitempty"`

	// Availability zone of instance
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// SpotMarketOptions option for configuring instances to be run using AWS Spot instances.
	// SpotMarketOptions *SpotMarketOptions `json:"spotMarketOptions,omitempty"`

	// Tenancy indicates if instance should run on shared or single-tenant hardware.
	// +optional
	Tenancy string `json:"tenancy,omitempty"`
}

// Volume encapsulates the configuration options for the storage device
// TODO: Trim the fields that do not apply for Snow.
type Volume struct {
	// Device name
	// +optional
	DeviceName string `json:"deviceName,omitempty"`

	// Size specifies size (in Gi) of the storage device.
	// Must be greater than the image snapshot size or 8 (whichever is greater).
	// +kubebuilder:validation:Minimum=8
	Size int64 `json:"size"`

	// Type is the type of the volume (sbp1 for capacity-optimized HDD, sbg1 performance-optimized SSD, default is sbp1)
	// +optional
	// +kubebuilder:validation:Enum:=sbp1;sbg1
	Type string `json:"type,omitempty"`
}

// InstanceState describes the state of an AWS instance.
type InstanceState string

var (
	// InstanceStatePending is the string representing an instance in a pending state
	InstanceStatePending = InstanceState("pending")

	// InstanceStateRunning is the string representing an instance in a running state
	InstanceStateRunning = InstanceState("running")

	// InstanceStateShuttingDown is the string representing an instance shutting down
	InstanceStateShuttingDown = InstanceState("shutting-down")

	// InstanceStateTerminated is the string representing an instance that has been terminated
	InstanceStateTerminated = InstanceState("terminated")

	// InstanceStateStopping is the string representing an instance
	// that is in the process of being stopped and can be restarted
	InstanceStateStopping = InstanceState("stopping")

	// InstanceStateStopped is the string representing an instance
	// that has been stopped and can be restarted
	InstanceStateStopped = InstanceState("stopped")

	// InstanceRunningStates defines the set of states in which an EC2 instance is
	// running or going to be running soon
	InstanceRunningStates = sets.NewString(
		string(InstanceStatePending),
		string(InstanceStateRunning),
	)

	// InstanceOperationalStates defines the set of states in which an EC2 instance is
	// or can return to running, and supports all EC2 operations
	InstanceOperationalStates = InstanceRunningStates.Union(
		sets.NewString(
			string(InstanceStateStopping),
			string(InstanceStateStopped),
		),
	)

	// InstanceKnownStates represents all known EC2 instance states
	InstanceKnownStates = InstanceOperationalStates.Union(
		sets.NewString(
			string(InstanceStateShuttingDown),
			string(InstanceStateTerminated),
		),
	)
)

// AWSSnowMachineTemplateResource describes the data needed to create am AWSSnowMachine from a template
type AWSSnowMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec AWSSnowMachineSpec `json:"spec"`
}

type OSFamily string
