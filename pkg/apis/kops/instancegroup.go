/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kops

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelClusterName is a cluster label cloud tag
	LabelClusterName = "kops.k8s.io/cluster"
	// NodeLabelInstanceGroup is a node label set to the name of the instance group
	NodeLabelInstanceGroup = "kops.k8s.io/instancegroup"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceGroup represents a group of instances (either nodes or masters) with the same configuration
type InstanceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceGroupSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceGroupList is a list of instance groups
type InstanceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole describes the roles of the nodes in this InstanceGroup (master or nodes)
type InstanceGroupRole string

const (
	// InstanceGroupRoleMaster is a master role
	InstanceGroupRoleMaster InstanceGroupRole = "Master"
	// InstanceGroupRoleNode is a node role
	InstanceGroupRoleNode InstanceGroupRole = "Node"
	// InstanceGroupRoleBastion is a bastion role
	InstanceGroupRoleBastion InstanceGroupRole = "Bastion"
	// InstanceGroupRoleAPIServer is an API server role
	InstanceGroupRoleAPIServer InstanceGroupRole = "APIServer"
)

// AllInstanceGroupRoles is a slice of all valid InstanceGroupRole values
var AllInstanceGroupRoles = []InstanceGroupRole{
	InstanceGroupRoleMaster,
	InstanceGroupRoleAPIServer,
	InstanceGroupRoleNode,
	InstanceGroupRoleBastion,
}

const (
	// BtfsFilesystem indicates a btfs filesystem
	BtfsFilesystem = "btfs"
	// Ext4Filesystem indicates a ext3 filesystem
	Ext4Filesystem = "ext4"
	// XFSFilesystem indicates a xfs filesystem
	XFSFilesystem = "xfs"
)

// SupportedFilesystems is a list of supported filesystems to format as
var SupportedFilesystems = []string{BtfsFilesystem, Ext4Filesystem, XFSFilesystem}

type InstanceManager string

const (
	InstanceManagerCloudGroup InstanceManager = "CloudGroup"
	InstanceManagerKarpenter  InstanceManager = "Karpenter"
)

// InstanceGroupSpec is the specification for an InstanceGroup
type InstanceGroupSpec struct {
	// Manager determines what is managing the node lifecycle
	Manager InstanceManager `json:"manager,omitempty"`
	// Type determines the role of instances in this instance group: masters or nodes
	Role InstanceGroupRole `json:"role,omitempty"`
	// Image is the instance (ami etc) we should use
	Image string `json:"image,omitempty"`
	// MinSize is the minimum size of the pool
	MinSize *int32 `json:"minSize,omitempty"`
	// MaxSize is the maximum size of the pool
	MaxSize *int32 `json:"maxSize,omitempty"`
	// Autoscale determines if autoscaling will be enabled for this instance group if cluster autoscaler is enabled
	Autoscale *bool `json:"autoscale,omitempty"`
	// MachineType is the instance class
	MachineType string `json:"machineType,omitempty"`
	// RootVolumeSize is the size of the EBS root volume to use, in GB
	RootVolumeSize *int32 `json:"rootVolumeSize,omitempty"`
	// RootVolumeType is the type of the EBS root volume to use (e.g. gp2)
	RootVolumeType *string `json:"rootVolumeType,omitempty"`
	// RootVolumeIOPS is the provisioned IOPS when the volume type is io1, io2 or gp3 (AWS only).
	RootVolumeIOPS *int32 `json:"rootVolumeIOPS,omitempty"`
	// RootVolumeThroughput is the volume throughput in MBps when the volume type is gp3 (AWS only).
	RootVolumeThroughput *int32 `json:"rootVolumeThroughput,omitempty"`
	// RootVolumeOptimization enables EBS optimization for an instance
	RootVolumeOptimization *bool `json:"rootVolumeOptimization,omitempty"`
	// RootVolumeEncryption enables EBS root volume encryption for an instance
	RootVolumeEncryption *bool `json:"rootVolumeEncryption,omitempty"`
	// RootVolumeEncryptionKey provides the key identifier for root volume encryption
	RootVolumeEncryptionKey *string `json:"rootVolumeEncryptionKey,omitempty"`
	// Volumes is a collection of additional volumes to create for instances within this instance group
	Volumes []VolumeSpec `json:"volumes,omitempty"`
	// VolumeMounts a collection of volume mounts
	VolumeMounts []VolumeMountSpec `json:"volumeMounts,omitempty"`
	// Subnets is the names of the Subnets (as specified in the Cluster) where machines in this instance group should be placed
	Subnets []string `json:"subnets,omitempty"`
	// Zones is the names of the Zones where machines in this instance group should be placed
	// This is needed for regional subnets (e.g. GCE), to restrict placement to particular zones
	Zones []string `json:"zones,omitempty"`
	// Hooks is a list of hooks for this instance group, note: these can override the cluster wide ones if required
	Hooks []HookSpec `json:"hooks,omitempty"`
	// MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
	MaxPrice *string `json:"maxPrice,omitempty"`
	// SpotDurationInMinutes reserves a spot block for the period specified
	SpotDurationInMinutes *int64 `json:"spotDurationInMinutes,omitempty"`
	// CPUCredits is the credit option for CPU Usage on burstable instance types (AWS only)
	CPUCredits *string `json:"cpuCredits,omitempty"`
	// AssociatePublicIP is true if we want instances to have a public IP
	AssociatePublicIP *bool `json:"associatePublicIP,omitempty"`
	// AdditionalSecurityGroups attaches additional security groups (e.g. i-123456)
	AdditionalSecurityGroups []string `json:"additionalSecurityGroups,omitempty"`
	// CloudLabels defines additional tags or labels on cloud provider resources
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`
	// NodeLabels indicates the kubernetes labels for nodes in this instance group
	NodeLabels map[string]string `json:"nodeLabels,omitempty"`
	// FileAssets is a collection of file assets for this instance group
	FileAssets []FileAssetSpec `json:"fileAssets,omitempty"`
	// Describes the tenancy of this instance group. Can be either default or dedicated. Currently only applies to AWS.
	Tenancy string `json:"tenancy,omitempty"`
	// Kubelet overrides kubelet config from the ClusterSpec
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`
	// Taints indicates the kubernetes taints for nodes in this instance group
	Taints []string `json:"taints,omitempty"`
	// MixedInstancesPolicy defined a optional backing of an AWS ASG by a EC2 Fleet (AWS Only)
	MixedInstancesPolicy *MixedInstancesPolicySpec `json:"mixedInstancesPolicy,omitempty"`
	// AdditionalUserData is any additional user-data to be passed to the host
	AdditionalUserData []UserData `json:"additionalUserData,omitempty"`
	// SuspendProcesses disables the listed Scaling Policies
	SuspendProcesses []string `json:"suspendProcesses,omitempty"`
	// ExternalLoadBalancers define loadbalancers that should be attached to this instance group
	ExternalLoadBalancers []LoadBalancer `json:"externalLoadBalancers,omitempty"`
	// DetailedInstanceMonitoring defines if detailed-monitoring is enabled (AWS only)
	DetailedInstanceMonitoring *bool `json:"detailedInstanceMonitoring,omitempty"`
	// IAMProfileSpec defines the identity of the cloud group IAM profile (AWS only).
	IAM *IAMProfileSpec `json:"iam,omitempty"`
	// SecurityGroupOverride overrides the default security group created by Kops for this IG (AWS only).
	SecurityGroupOverride *string `json:"securityGroupOverride,omitempty"`
	// InstanceProtection makes new instances in an autoscaling group protected from scale in
	InstanceProtection *bool `json:"instanceProtection,omitempty"`
	// SysctlParameters will configure kernel parameters using sysctl(8). When
	// specified, each parameter must follow the form variable=value, the way
	// it would appear in sysctl.conf.
	SysctlParameters []string `json:"sysctlParameters,omitempty"`
	// RollingUpdate defines the rolling-update behavior
	RollingUpdate *RollingUpdate `json:"rollingUpdate,omitempty"`
	// InstanceInterruptionBehavior defines if a spot instance should be terminated, hibernated,
	// or stopped after interruption
	InstanceInterruptionBehavior *string `json:"instanceInterruptionBehavior,omitempty"`
	// CompressUserData compresses parts of the user data to save space
	CompressUserData *bool `json:"compressUserData,omitempty"`
	// InstanceMetadata defines the EC2 instance metadata service options (AWS Only)
	InstanceMetadata *InstanceMetadataOptions `json:"instanceMetadata,omitempty"`
	// UpdatePolicy determines the policy for applying upgrades automatically.
	// If specified, this value overrides a value specified in the Cluster's "spec.updatePolicy" field.
	// Valid values:
	//   'automatic' (default): apply updates automatically (apply OS security upgrades, avoiding rebooting when possible)
	//   'external': do not apply updates automatically; they are applied manually or by an external system
	UpdatePolicy *string `json:"updatePolicy,omitempty"`
	// WarmPool specifies a pool of pre-warmed instances for later use (AWS only).
	WarmPool *WarmPoolSpec `json:"warmPool,omitempty"`
	// Containerd specifies override configuration for instance group
	Containerd *ContainerdConfig `json:"containerd,omitempty"`
	// Packages specifies additional packages to be installed.
	Packages []string `json:"packages,omitempty"`
}

const (
	// SpotAllocationStrategyLowestPrices indicates a lowest-price strategy
	SpotAllocationStrategyLowestPrices = "lowest-price"
	// SpotAllocationStrategyDiversified indicates a diversified strategy
	SpotAllocationStrategyDiversified = "diversified"
	// SpotAllocationStrategyCapacityOptimized indicates a capacity optimized strategy
	SpotAllocationStrategyCapacityOptimized = "capacity-optimized"
	// SpotAllocationStrategyCapacityOptimizedPrioritized indicates a capacity optimized prioritized strategy
	SpotAllocationStrategyCapacityOptimizedPrioritized = "capacity-optimized-prioritized"
)

// SpotAllocationStrategies is a collection of supported strategies
var SpotAllocationStrategies = []string{
	SpotAllocationStrategyLowestPrices,
	SpotAllocationStrategyDiversified,
	SpotAllocationStrategyCapacityOptimized,
	SpotAllocationStrategyCapacityOptimizedPrioritized,
}

// InstanceMetadataOptions defines the EC2 instance metadata service options (AWS Only)
type InstanceMetadataOptions struct {
	// HTTPPutResponseHopLimit is the desired HTTP PUT response hop limit for instance metadata requests.
	// The larger the number, the further instance metadata requests can travel. The default value is 1.
	HTTPPutResponseHopLimit *int64 `json:"httpPutResponseHopLimit,omitempty"`
	// HTTPTokens is the state of token usage for the instance metadata requests.
	// If the parameter is not specified in the request, the default state is "required".
	HTTPTokens *string `json:"httpTokens,omitempty"`
}

// MixedInstancesPolicySpec defines the specification for an autoscaling group backed by a ec2 fleet
type MixedInstancesPolicySpec struct {
	// Instances is a list of instance types which we are willing to run in the EC2 fleet
	Instances []string `json:"instances,omitempty"`
	// InstanceRequirements is a list of requirements for any instance type we are willing to run in the EC2 fleet.
	InstanceRequirements *InstanceRequirementsSpec `json:"instanceRequirements,omitempty"`
	// OnDemandAllocationStrategy indicates how to allocate instance types to fulfill On-Demand capacity
	OnDemandAllocationStrategy *string `json:"onDemandAllocationStrategy,omitempty"`
	// OnDemandBase is the minimum amount of the Auto Scaling group's capacity that must be
	// fulfilled by On-Demand Instances. This base portion is provisioned first as your group scales.
	OnDemandBase *int64 `json:"onDemandBase,omitempty"`
	// OnDemandAboveBase controls the percentages of On-Demand Instances and Spot Instances for your
	// additional capacity beyond OnDemandBase. The range is 0–100. The default value is 100. If you
	// leave this parameter set to 100, the percentages are 100% for On-Demand Instances and 0% for
	// Spot Instances.
	OnDemandAboveBase *int64 `json:"onDemandAboveBase,omitempty"`
	// SpotAllocationStrategy diversifies your Spot capacity across multiple instance types to
	// find the best pricing. Higher Spot availability may result from a larger number of
	// instance types to choose from.
	SpotAllocationStrategy *string `json:"spotAllocationStrategy,omitempty"`
	// SpotInstancePools is the number of Spot pools to use to allocate your Spot capacity (defaults to 2)
	// pools are determined from the different instance types in the Overrides array of LaunchTemplate
	SpotInstancePools *int64 `json:"spotInstancePools,omitempty"`
}

// InstanceRequirementsSpec is a list of requirements for any instance type we are willing to run in the EC2 fleet.
type InstanceRequirementsSpec struct {
	CPU    *MinMaxSpec `json:"cpu,omitempty"`
	Memory *MinMaxSpec `json:"memory,omitempty"`
}

type MinMaxSpec struct {
	Max *resource.Quantity `json:"max,omitempty"`
	Min *resource.Quantity `json:"min,omitempty"`
}

// UserData defines a user-data section
type UserData struct {
	// Name is the name of the user-data
	Name string `json:"name,omitempty"`
	// Type is the type of user-data
	Type string `json:"type,omitempty"`
	// Content is the user-data content
	Content string `json:"content,omitempty"`
}

// VolumeSpec defined the spec for an additional volume attached to the instance group
type VolumeSpec struct {
	// DeleteOnTermination configures volume retention policy upon instance termination.
	// The volume is deleted by default. Cluster deletion does not remove retained volumes.
	DeleteOnTermination *bool `json:"deleteOnTermination,omitempty"`
	// Device is an optional device name of the block device
	Device string `json:"device,omitempty"`
	// Encrypted indicates you want to encrypt the volume
	Encrypted *bool `json:"encrypted,omitempty"`
	// IOPS is the provisioned IOPS for the volume when the volume type is io1, io2 or gp3 (AWS only).
	IOPS *int64 `json:"iops,omitempty"`
	// Throughput is the volume throughput in MBps when the volume type is gp3 (AWS only).
	Throughput *int64 `json:"throughput,omitempty"`
	// Key is the encryption key identifier for the volume
	Key *string `json:"key,omitempty"`
	// Size is the size of the volume in GB
	Size int64 `json:"size,omitempty"`
	// Type is the type of volume to create and is cloud specific
	Type string `json:"type,omitempty"`
}

// VolumeMountSpec defines the specification for mounting a device
type VolumeMountSpec struct {
	// Device is the device name to provision and mount
	Device string `json:"device,omitempty"`
	// Filesystem is the filesystem to mount
	Filesystem string `json:"filesystem,omitempty"`
	// FormatOptions is a collection of options passed when formatting the device
	FormatOptions []string `json:"formatOptions,omitempty"`
	// MountOptions is a collection of mount options - @TODO need to be added
	MountOptions []string `json:"mountOptions,omitempty"`
	// Path is the location to mount the device
	Path string `json:"path,omitempty"`
}

// IAMProfileSpec is the AWS IAM Profile to attach to instances in this instance
// group. Specify the ARN for the IAM instance profile (AWS only).
type IAMProfileSpec struct {
	// Profile is the AWS IAM Profile to attach to instances in this instance group.
	// Specify the ARN for the IAM instance profile. (AWS only)
	Profile *string `json:"profile,omitempty"`
}

// IsMaster checks if instanceGroup is a master
func (g *InstanceGroup) IsMaster() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleMaster:
		return true
	default:
		return false
	}
}

// IsAPIServerOnly checks if instanceGroup runs only the API Server
func (g *InstanceGroup) IsAPIServerOnly() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleAPIServer:
		return true
	default:
		return false
	}
}

// hasAPIServer checks if instanceGroup runs an API Server
func (g *InstanceGroup) HasAPIServer() bool {
	return g.IsMaster() || g.IsAPIServerOnly()
}

// IsBastion checks if instanceGroup is a bastion
func (g *InstanceGroup) IsBastion() bool {
	switch g.Spec.Role {
	case InstanceGroupRoleBastion:
		return true
	default:
		return false
	}
}

func (g *InstanceGroup) AddInstanceGroupNodeLabel() {
	if g.Spec.NodeLabels == nil {
		g.Spec.NodeLabels = make(map[string]string)
	}
	g.Spec.NodeLabels[NodeLabelInstanceGroup] = g.Name
}

// LoadBalancer defines a load balancer
type LoadBalancer struct {
	// LoadBalancerName to associate with this instance group (AWS ELB)
	LoadBalancerName *string `json:"loadBalancerName,omitempty"`
	// TargetGroupARN to associate with this instance group (AWS ALB/NLB)
	TargetGroupARN *string `json:"targetGroupARN,omitempty"`
}
