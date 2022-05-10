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

package spotinsttasks

import (
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// +kops:fitask
type Elastigroup struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Region                   *string
	MinSize                  *int64
	MaxSize                  *int64
	SpotPercentage           *float64
	UtilizeReservedInstances *bool
	UtilizeCommitments       *bool
	FallbackToOnDemand       *bool
	DrainingTimeout          *int64
	HealthCheckType          *string
	Product                  *string
	Orientation              *string
	Tags                     map[string]string
	UserData                 fi.Resource
	ImageID                  *string
	OnDemandInstanceType     *string
	SpotInstanceTypes        []string
	IAMInstanceProfile       *awstasks.IAMInstanceProfile
	LoadBalancers            []*awstasks.ClassicLoadBalancer
	TargetGroups             []*awstasks.TargetGroup
	SSHKey                   *awstasks.SSHKey
	Subnets                  []*awstasks.Subnet
	SecurityGroups           []*awstasks.SecurityGroup
	Monitoring               *bool
	AssociatePublicIPAddress *bool
	Tenancy                  *string
	RootVolumeOpts           *RootVolumeOpts
	AutoScalerOpts           *AutoScalerOpts
}

type RootVolumeOpts struct {
	Type         *string
	Size         *int64
	IOPS         *int64
	Throughput   *int64
	Optimization *bool
	Encryption   *bool
}

type AutoScalerOpts struct {
	Enabled                *bool
	AutoConfig             *bool
	AutoHeadroomPercentage *int
	ClusterID              *string
	Cooldown               *int
	Labels                 map[string]string
	Taints                 []*corev1.Taint
	Headroom               *AutoScalerHeadroomOpts
	Down                   *AutoScalerDownOpts
	ResourceLimits         *AutoScalerResourceLimitsOpts
}

type AutoScalerHeadroomOpts struct {
	CPUPerUnit *int
	GPUPerUnit *int
	MemPerUnit *int
	NumOfUnits *int
}

type AutoScalerDownOpts struct {
	MaxPercentage     *float64
	EvaluationPeriods *int
}

type AutoScalerResourceLimitsOpts struct {
	MaxVCPU   *int
	MaxMemory *int
}

var (
	_ fi.Task            = &Elastigroup{}
	_ fi.CompareWithID   = &Elastigroup{}
	_ fi.HasDependencies = &Elastigroup{}
)

func (e *Elastigroup) CompareWithID() *string {
	return e.Name
}

func (e *Elastigroup) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	if e.IAMInstanceProfile != nil {
		deps = append(deps, e.IAMInstanceProfile)
	}

	if e.LoadBalancers != nil {
		for _, lb := range e.LoadBalancers {
			deps = append(deps, lb)
		}
	}

	if e.TargetGroups != nil {
		for _, tg := range e.TargetGroups {
			deps = append(deps, tg)
		}
	}

	if e.SSHKey != nil {
		deps = append(deps, e.SSHKey)
	}

	if e.Subnets != nil {
		for _, subnet := range e.Subnets {
			deps = append(deps, subnet)
		}
	}

	if e.SecurityGroups != nil {
		for _, sg := range e.SecurityGroups {
			deps = append(deps, sg)
		}
	}

	if e.UserData != nil {
		deps = append(deps, fi.FindDependencies(tasks, e.UserData)...)
	}

	return deps
}

func (e *Elastigroup) find(svc spotinst.InstanceGroupService) (*aws.Group, error) {
	klog.V(4).Infof("Attempting to find Elastigroup: %q", fi.StringValue(e.Name))

	groups, err := svc.List(context.Background())
	if err != nil {
		return nil, fmt.Errorf("spotinst: failed to find elastigroup %s: %v", fi.StringValue(e.Name), err)
	}

	var out *aws.Group
	for _, group := range groups {
		if group.Name() == fi.StringValue(e.Name) {
			out = group.Obj().(*aws.Group)
			break
		}
	}
	if out == nil {
		return nil, fmt.Errorf("spotinst: failed to find elastigroup %q", fi.StringValue(e.Name))
	}

	klog.V(4).Infof("Elastigroup/%s: %s", fi.StringValue(e.Name), stringutil.Stringify(out))
	return out, nil
}

var _ fi.HasCheckExisting = &Elastigroup{}

func (e *Elastigroup) Find(c *fi.Context) (*Elastigroup, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	group, err := e.find(cloud.Spotinst().Elastigroup())
	if err != nil {
		return nil, err
	}

	actual := &Elastigroup{}
	actual.Name = group.Name
	actual.Region = group.Region

	// Capacity.
	{
		actual.MinSize = fi.Int64(int64(fi.IntValue(group.Capacity.Minimum)))
		actual.MaxSize = fi.Int64(int64(fi.IntValue(group.Capacity.Maximum)))
	}

	// Strategy.
	{
		actual.SpotPercentage = group.Strategy.Risk
		actual.Orientation = group.Strategy.AvailabilityVsCost
		actual.FallbackToOnDemand = group.Strategy.FallbackToOnDemand
		actual.UtilizeReservedInstances = group.Strategy.UtilizeReservedInstances
		actual.UtilizeCommitments = group.Strategy.UtilizeCommitments

		if group.Strategy.DrainingTimeout != nil {
			actual.DrainingTimeout = fi.Int64(int64(fi.IntValue(group.Strategy.DrainingTimeout)))
		}
	}

	// Compute.
	{
		compute := group.Compute
		actual.Product = compute.Product

		// Instance types.
		{
			actual.OnDemandInstanceType = compute.InstanceTypes.OnDemand
			actual.SpotInstanceTypes = compute.InstanceTypes.Spot
		}

		// Subnets.
		{
			for _, subnetID := range compute.SubnetIDs {
				actual.Subnets = append(actual.Subnets,
					&awstasks.Subnet{ID: fi.String(subnetID)})
			}
			if subnetSlicesEqualIgnoreOrder(actual.Subnets, e.Subnets) {
				actual.Subnets = e.Subnets
			}
		}
	}

	// Launch specification.
	{
		lc := group.Compute.LaunchSpecification

		// Image.
		{
			actual.ImageID = lc.ImageID

			if e.ImageID != nil && actual.ImageID != nil &&
				fi.StringValue(actual.ImageID) != fi.StringValue(e.ImageID) {
				image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
				if err != nil {
					return nil, err
				}
				if fi.StringValue(image.ImageId) == fi.StringValue(lc.ImageID) {
					actual.ImageID = e.ImageID
				}
			}
		}

		// Tags.
		{
			if lc.Tags != nil && len(lc.Tags) > 0 {
				actual.Tags = make(map[string]string)
				for _, tag := range lc.Tags {
					actual.Tags[fi.StringValue(tag.Key)] = fi.StringValue(tag.Value)
				}
			}
		}

		// Security groups.
		{
			if lc.SecurityGroupIDs != nil {
				for _, sgID := range lc.SecurityGroupIDs {
					actual.SecurityGroups = append(actual.SecurityGroups,
						&awstasks.SecurityGroup{ID: fi.String(sgID)})
				}
			}
		}

		// Root volume options.
		{
			// Block device mappings.
			{
				if lc.BlockDeviceMappings != nil {
					for _, b := range lc.BlockDeviceMappings {
						if b.EBS == nil || b.EBS.SnapshotID != nil {
							continue // not the root
						}
						if actual.RootVolumeOpts == nil {
							actual.RootVolumeOpts = new(RootVolumeOpts)
						}
						if b.EBS.VolumeType != nil {
							actual.RootVolumeOpts.Type = fi.String(strings.ToLower(fi.StringValue(b.EBS.VolumeType)))
						}
						if b.EBS.VolumeSize != nil {
							actual.RootVolumeOpts.Size = fi.Int64(int64(fi.IntValue(b.EBS.VolumeSize)))
						}
						if b.EBS.IOPS != nil {
							actual.RootVolumeOpts.IOPS = fi.Int64(int64(fi.IntValue(b.EBS.IOPS)))
						}
						if b.EBS.Throughput != nil {
							actual.RootVolumeOpts.Throughput = fi.Int64(int64(fi.IntValue(b.EBS.Throughput)))
						}
						if b.EBS.Encrypted != nil {
							actual.RootVolumeOpts.Encryption = b.EBS.Encrypted
						}
					}
				}
			}

			// EBS optimization.
			{
				if fi.BoolValue(lc.EBSOptimized) {
					if actual.RootVolumeOpts == nil {
						actual.RootVolumeOpts = new(RootVolumeOpts)
					}

					actual.RootVolumeOpts.Optimization = lc.EBSOptimized
				}
			}
		}

		// User data.
		{
			var userData []byte

			if lc.UserData != nil {
				userData, err = base64.StdEncoding.DecodeString(fi.StringValue(lc.UserData))
				if err != nil {
					return nil, err
				}
			}

			actual.UserData = fi.NewStringResource(string(userData))
		}

		// Network interfaces.
		{
			associatePublicIP := false

			if lc.NetworkInterfaces != nil && len(lc.NetworkInterfaces) > 0 {
				for _, iface := range lc.NetworkInterfaces {
					if fi.BoolValue(iface.AssociatePublicIPAddress) {
						associatePublicIP = true
						break
					}
				}
			}

			actual.AssociatePublicIPAddress = fi.Bool(associatePublicIP)
		}

		// Load balancers.
		{
			if conf := lc.LoadBalancersConfig; conf != nil && len(conf.LoadBalancers) > 0 {
				for _, lb := range conf.LoadBalancers {
					switch fi.StringValue(lb.Type) {
					case "CLASSIC":
						actual.LoadBalancers = append(actual.LoadBalancers,
							&awstasks.ClassicLoadBalancer{
								Name: lb.Name,
							})
					case "TARGET_GROUP":
						actual.TargetGroups = append(actual.TargetGroups,
							&awstasks.TargetGroup{
								ARN: lb.Arn,
							})
					}
				}

				var apiLBTask *awstasks.ClassicLoadBalancer
				for _, elb := range e.LoadBalancers {
					if !fi.BoolValue(elb.Shared) {
						apiLBTask = elb
					}
				}
				if apiLBTask != nil && len(actual.LoadBalancers) > 0 {
					apiLBDesc, err := cloud.FindELBByNameTag(fi.StringValue(apiLBTask.Name))
					if err != nil {
						return nil, err
					}
					if apiLBDesc != nil {
						for i := 0; i < len(actual.LoadBalancers); i++ {
							lb := actual.LoadBalancers[i]
							if fi.StringValue(apiLBDesc.LoadBalancerName) == fi.StringValue(lb.Name) {
								actual.LoadBalancers[i] = apiLBTask
							}
						}
					}
				}
			}
		}

		// IAM instance profile.
		if lc.IAMInstanceProfile != nil {
			actual.IAMInstanceProfile = &awstasks.IAMInstanceProfile{Name: lc.IAMInstanceProfile.Name}
		}

		// SSH key.
		if lc.KeyPair != nil {
			actual.SSHKey = &awstasks.SSHKey{Name: lc.KeyPair}
		}

		// Tenancy.
		if lc.Tenancy != nil {
			actual.Tenancy = lc.Tenancy
		}

		// Monitoring.
		if lc.Monitoring != nil {
			actual.Monitoring = lc.Monitoring
		}

		// Health check.
		if lc.HealthCheckType != nil {
			actual.HealthCheckType = lc.HealthCheckType
		}
	}

	// Auto Scaler.
	{
		if group.Integration != nil && group.Integration.Kubernetes != nil {
			integration := group.Integration.Kubernetes

			actual.AutoScalerOpts = new(AutoScalerOpts)
			actual.AutoScalerOpts.ClusterID = integration.ClusterIdentifier

			if integration.AutoScale != nil {
				actual.AutoScalerOpts.Enabled = integration.AutoScale.IsEnabled
				actual.AutoScalerOpts.Cooldown = integration.AutoScale.Cooldown

				// Headroom.
				if headroom := integration.AutoScale.Headroom; headroom != nil {
					actual.AutoScalerOpts.Headroom = new(AutoScalerHeadroomOpts)

					if v := fi.IntValue(headroom.CPUPerUnit); v > 0 {
						actual.AutoScalerOpts.Headroom.CPUPerUnit = headroom.CPUPerUnit
					}
					if v := fi.IntValue(headroom.GPUPerUnit); v > 0 {
						actual.AutoScalerOpts.Headroom.GPUPerUnit = headroom.GPUPerUnit
					}
					if v := fi.IntValue(headroom.MemoryPerUnit); v > 0 {
						actual.AutoScalerOpts.Headroom.MemPerUnit = headroom.MemoryPerUnit
					}
					if v := fi.IntValue(headroom.NumOfUnits); v > 0 {
						actual.AutoScalerOpts.Headroom.NumOfUnits = headroom.NumOfUnits
					}
				}

				// Scale down.
				if down := integration.AutoScale.Down; down != nil {
					actual.AutoScalerOpts.Down = &AutoScalerDownOpts{
						MaxPercentage:     down.MaxScaleDownPercentage,
						EvaluationPeriods: down.EvaluationPeriods,
					}
				}

				// Labels.
				if labels := integration.AutoScale.Labels; labels != nil {
					actual.AutoScalerOpts.Labels = make(map[string]string)

					for _, label := range labels {
						actual.AutoScalerOpts.Labels[fi.StringValue(label.Key)] = fi.StringValue(label.Value)
					}
				}
			}
		}
	}

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *Elastigroup) CheckExisting(c *fi.Context) bool {
	cloud := c.Cloud.(awsup.AWSCloud)
	group, err := e.find(cloud.Spotinst().Elastigroup())
	return err == nil && group != nil
}

func (e *Elastigroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Elastigroup) CheckChanges(a, e, changes *Elastigroup) error {
	if e.Name == nil {
		return fi.RequiredField("Name")
	}
	return nil
}

func (eg *Elastigroup) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Elastigroup) error {
	return eg.createOrUpdate(t.Cloud, a, e, changes)
}

func (eg *Elastigroup) createOrUpdate(cloud awsup.AWSCloud, a, e, changes *Elastigroup) error {
	if a == nil {
		return eg.create(cloud, a, e, changes)
	} else {
		return eg.update(cloud, a, e, changes)
	}
}

func (_ *Elastigroup) create(cloud awsup.AWSCloud, a, e, changes *Elastigroup) error {
	klog.V(2).Infof("Creating Elastigroup %q", *e.Name)
	e.applyDefaults()

	group := &aws.Group{
		Capacity: new(aws.Capacity),
		Strategy: new(aws.Strategy),
		Compute: &aws.Compute{
			LaunchSpecification: new(aws.LaunchSpecification),
			InstanceTypes:       new(aws.InstanceTypes),
		},
	}

	// General.
	{
		group.SetName(e.Name)
		group.SetDescription(e.Name)
		group.SetRegion(e.Region)
	}

	// Capacity.
	{
		group.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
		group.Capacity.SetMinimum(fi.Int(int(*e.MinSize)))
		group.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
	}

	// Strategy.
	{
		group.Strategy.SetRisk(e.SpotPercentage)
		group.Strategy.SetAvailabilityVsCost(fi.String(string(normalizeOrientation(e.Orientation))))
		group.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
		group.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
		group.Strategy.SetUtilizeCommitments(e.UtilizeCommitments)

		if e.DrainingTimeout != nil {
			group.Strategy.SetDrainingTimeout(fi.Int(int(*e.DrainingTimeout)))
		}
	}

	// Compute.
	{
		group.Compute.SetProduct(e.Product)

		// Instance types.
		{
			group.Compute.InstanceTypes.SetOnDemand(e.OnDemandInstanceType)
			group.Compute.InstanceTypes.SetSpot(e.SpotInstanceTypes)
		}

		// Subnets.
		{
			subnets := make([]string, len(e.Subnets))
			for i, subnet := range e.Subnets {
				subnets[i] = fi.StringValue(subnet.ID)
			}
			group.Compute.SetSubnetIDs(subnets)
		}

		// Launch Specification.
		{
			group.Compute.LaunchSpecification.SetMonitoring(e.Monitoring)
			group.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)

			if e.Tenancy != nil {
				group.Compute.LaunchSpecification.SetTenancy(e.Tenancy)
			}

			// Block device mappings.
			{
				rootDevice, err := buildRootDevice(cloud, e.RootVolumeOpts, e.ImageID)
				if err != nil {
					return err
				}

				mappings := []*aws.BlockDeviceMapping{
					e.convertBlockDeviceMapping(rootDevice),
				}

				ephemeralDevices, err := buildEphemeralDevices(cloud, e.OnDemandInstanceType)
				if err != nil {
					return err
				}

				for _, bdm := range ephemeralDevices {
					mappings = append(mappings, e.convertBlockDeviceMapping(bdm))
				}

				group.Compute.LaunchSpecification.SetBlockDeviceMappings(mappings)
			}

			// Image.
			{
				image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
				if err != nil {
					return err
				}
				group.Compute.LaunchSpecification.SetImageId(image.ImageId)
			}

			// User data.
			{
				if e.UserData != nil {
					userData, err := fi.ResourceAsString(e.UserData)
					if err != nil {
						return err
					}

					if len(userData) > 0 {
						encoded := base64.StdEncoding.EncodeToString([]byte(userData))
						group.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
					}
				}
			}

			// IAM instance profile.
			{
				if e.IAMInstanceProfile != nil {
					iprof := new(aws.IAMInstanceProfile)
					iprof.SetName(e.IAMInstanceProfile.GetName())
					group.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
				}
			}

			// Security groups.
			{
				if e.SecurityGroups != nil {
					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}
					group.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
				}
			}

			// Public IP.
			{
				if e.AssociatePublicIPAddress != nil {
					iface := &aws.NetworkInterface{
						Description:              fi.String("eth0"),
						DeviceIndex:              fi.Int(0),
						DeleteOnTermination:      fi.Bool(true),
						AssociatePublicIPAddress: e.AssociatePublicIPAddress,
					}

					group.Compute.LaunchSpecification.SetNetworkInterfaces([]*aws.NetworkInterface{iface})
				}
			}

			// Load balancers.
			{
				if len(e.LoadBalancers) > 0 {
					lbs, err := e.buildLoadBalancers(cloud)
					if err != nil {
						return err
					}

					if group.Compute.LaunchSpecification.LoadBalancersConfig == nil {
						group.Compute.LaunchSpecification.LoadBalancersConfig = new(aws.LoadBalancersConfig)
					}

					group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers = append(group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers, lbs...)
				}
			}

			// Target groups.
			{
				if len(e.TargetGroups) > 0 {
					tgs := e.buildTargetGroups()

					if group.Compute.LaunchSpecification.LoadBalancersConfig == nil {
						group.Compute.LaunchSpecification.LoadBalancersConfig = new(aws.LoadBalancersConfig)
					}

					group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers = append(group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers, tgs...)
				}
			}

			// Tags.
			{
				if e.Tags != nil {
					group.Compute.LaunchSpecification.SetTags(e.buildTags())
				}
			}

			// Health check.
			{
				if e.HealthCheckType != nil {
					group.Compute.LaunchSpecification.SetHealthCheckType(e.HealthCheckType)
				}
			}
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			k8s := new(aws.KubernetesIntegration)
			k8s.SetIntegrationMode(fi.String("pod"))
			k8s.SetClusterIdentifier(opts.ClusterID)

			if opts.Enabled != nil {
				autoScaler := new(aws.AutoScaleKubernetes)
				autoScaler.IsEnabled = opts.Enabled
				autoScaler.IsAutoConfig = fi.Bool(true)
				autoScaler.Cooldown = opts.Cooldown

				// Headroom.
				if headroom := opts.Headroom; headroom != nil {
					autoScaler.IsAutoConfig = fi.Bool(false)
					autoScaler.Headroom = &aws.AutoScaleHeadroom{
						CPUPerUnit:    headroom.CPUPerUnit,
						GPUPerUnit:    headroom.GPUPerUnit,
						MemoryPerUnit: headroom.MemPerUnit,
						NumOfUnits:    headroom.NumOfUnits,
					}
				}

				// Scale down.
				if down := opts.Down; down != nil {
					autoScaler.Down = &aws.AutoScaleDown{
						MaxScaleDownPercentage: down.MaxPercentage,
						EvaluationPeriods:      down.EvaluationPeriods,
					}
				}

				// Labels.
				if labels := opts.Labels; labels != nil {
					autoScaler.Labels = e.buildAutoScaleLabels(labels)
				}

				k8s.SetAutoScale(autoScaler)
			}

			integration := new(aws.Integration)
			integration.SetKubernetes(k8s)

			group.SetIntegration(integration)
		}
	}

	attempt := 0
	maxAttempts := 10

readyLoop:
	for {
		attempt++
		klog.V(2).Infof("(%d/%d) Attempting to create Elastigroup: %q, config: %s",
			attempt, maxAttempts, *e.Name, stringutil.Stringify(group))

		// Wait for IAM instance profile to be ready.
		time.Sleep(10 * time.Second)

		// Wrap the raw object as an Elastigroup.
		eg, err := spotinst.NewElastigroup(cloud.ProviderID(), group)
		if err != nil {
			return err
		}

		// Create the Elastigroup.
		_, err = cloud.Spotinst().Elastigroup().Create(context.Background(), eg)
		if err == nil {
			break
		}

		if errs, ok := err.(client.Errors); ok {
			for _, err := range errs {
				if strings.Contains(err.Message, "Invalid IAM Instance Profile name") {
					if attempt > maxAttempts {
						return fmt.Errorf("IAM instance profile not yet created/propagated (original error: %v)", err)
					}

					klog.V(4).Infof("Got an error indicating that the IAM instance profile %q is not ready %q", fi.StringValue(e.IAMInstanceProfile.Name), err)
					klog.Infof("Waiting for IAM instance profile %q to be ready", fi.StringValue(e.IAMInstanceProfile.Name))
					goto readyLoop
				}
			}

			return fmt.Errorf("spotinst: failed to create elastigroup: %v", err)
		}
	}

	return nil
}

func (_ *Elastigroup) update(cloud awsup.AWSCloud, a, e, changes *Elastigroup) error {
	klog.V(2).Infof("Updating Elastigroup %q", *e.Name)

	actual, err := e.find(cloud.Spotinst().Elastigroup())
	if err != nil {
		klog.Errorf("Unable to resolve Elastigroup %q, error: %v", *e.Name, err)
		return err
	}

	var changed bool
	group := new(aws.Group)
	group.SetId(actual.ID)

	// Region.
	if changes.Region != nil {
		group.SetRegion(e.Region)
		changes.Region = nil
		changed = true
	}

	// Strategy.
	{
		// Spot percentage.
		if changes.SpotPercentage != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetRisk(e.SpotPercentage)
			changes.SpotPercentage = nil
			changed = true
		}

		// Orientation.
		if changes.Orientation != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetAvailabilityVsCost(fi.String(string(normalizeOrientation(e.Orientation))))
			changes.Orientation = nil
			changed = true
		}

		// Fallback to on-demand.
		if changes.FallbackToOnDemand != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetFallbackToOnDemand(e.FallbackToOnDemand)
			changes.FallbackToOnDemand = nil
			changed = true
		}

		// Utilize reserved instances.
		if changes.UtilizeReservedInstances != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetUtilizeReservedInstances(e.UtilizeReservedInstances)
			changes.UtilizeReservedInstances = nil
			changed = true
		}

		// Utilize commitments.
		if changes.UtilizeCommitments != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetUtilizeCommitments(e.UtilizeCommitments)
			changes.UtilizeCommitments = nil
			changed = true
		}

		// Draining timeout.
		if changes.DrainingTimeout != nil {
			if group.Strategy == nil {
				group.Strategy = new(aws.Strategy)
			}

			group.Strategy.SetDrainingTimeout(fi.Int(int(*e.DrainingTimeout)))
			changes.DrainingTimeout = nil
			changed = true
		}
	}

	// Compute.
	{
		// Product.
		if changes.Product != nil {
			if group.Compute == nil {
				group.Compute = new(aws.Compute)
			}

			group.Compute.SetProduct(e.Product)
			changes.Product = nil
			changed = true
		}

		// On-demand instance type.
		{
			if changes.OnDemandInstanceType != nil {
				if group.Compute == nil {
					group.Compute = new(aws.Compute)
				}
				if group.Compute.InstanceTypes == nil {
					group.Compute.InstanceTypes = new(aws.InstanceTypes)
				}

				group.Compute.InstanceTypes.SetOnDemand(e.OnDemandInstanceType)
				changes.OnDemandInstanceType = nil
				changed = true
			}
		}

		// Spot instance types.
		{
			if changes.SpotInstanceTypes != nil {
				if group.Compute == nil {
					group.Compute = new(aws.Compute)
				}
				if group.Compute.InstanceTypes == nil {
					group.Compute.InstanceTypes = new(aws.InstanceTypes)
				}

				types := make([]string, len(e.SpotInstanceTypes))
				copy(types, e.SpotInstanceTypes)

				group.Compute.InstanceTypes.SetSpot(types)
				changes.SpotInstanceTypes = nil
				changed = true
			}
		}

		// Subnets.
		{
			if changes.Subnets != nil {
				if group.Compute == nil {
					group.Compute = new(aws.Compute)
				}

				subnets := make([]string, len(e.Subnets))
				for i, subnet := range e.Subnets {
					subnets[i] = fi.StringValue(subnet.ID)
				}

				group.Compute.SetSubnetIDs(subnets)
				changes.Subnets = nil
				changed = true
			}
		}

		// Launch specification.
		{
			// Security groups.
			{
				if changes.SecurityGroups != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					securityGroupIDs := make([]string, len(e.SecurityGroups))
					for i, sg := range e.SecurityGroups {
						securityGroupIDs[i] = *sg.ID
					}

					group.Compute.LaunchSpecification.SetSecurityGroupIDs(securityGroupIDs)
					changes.SecurityGroups = nil
					changed = true
				}
			}

			// User data.
			{
				if changes.UserData != nil {
					userData, err := fi.ResourceAsString(e.UserData)
					if err != nil {
						return err
					}

					if len(userData) > 0 {
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						encoded := base64.StdEncoding.EncodeToString([]byte(userData))
						group.Compute.LaunchSpecification.SetUserData(fi.String(encoded))
						changed = true
					}

					changes.UserData = nil
				}
			}

			// Network interfaces.
			{
				if changes.AssociatePublicIPAddress != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					iface := &aws.NetworkInterface{
						Description:              fi.String("eth0"),
						DeviceIndex:              fi.Int(0),
						DeleteOnTermination:      fi.Bool(true),
						AssociatePublicIPAddress: changes.AssociatePublicIPAddress,
					}

					group.Compute.LaunchSpecification.SetNetworkInterfaces([]*aws.NetworkInterface{iface})
					changes.AssociatePublicIPAddress = nil
					changed = true
				}
			}

			// Root volume options.
			{
				if opts := changes.RootVolumeOpts; opts != nil {
					// Block device mappings.
					{
						if opts.Type != nil || opts.Size != nil || opts.IOPS != nil {
							rootDevice, err := buildRootDevice(cloud, opts, e.ImageID)
							if err != nil {
								return err
							}

							mappings := []*aws.BlockDeviceMapping{
								e.convertBlockDeviceMapping(rootDevice),
							}

							ephemeralDevices, err := buildEphemeralDevices(cloud, e.OnDemandInstanceType)
							if err != nil {
								return err
							}

							for _, bdm := range ephemeralDevices {
								mappings = append(mappings, e.convertBlockDeviceMapping(bdm))
							}

							if group.Compute == nil {
								group.Compute = new(aws.Compute)
							}
							if group.Compute.LaunchSpecification == nil {
								group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							group.Compute.LaunchSpecification.SetBlockDeviceMappings(mappings)
							changed = true
						}
					}

					// EBS optimization.
					{
						if opts.Optimization != nil {
							if group.Compute == nil {
								group.Compute = new(aws.Compute)
							}
							if group.Compute.LaunchSpecification == nil {
								group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
							}

							group.Compute.LaunchSpecification.SetEBSOptimized(e.RootVolumeOpts.Optimization)
							changed = true
						}
					}

					changes.RootVolumeOpts = nil
				}
			}

			// Image.
			{
				if changes.ImageID != nil {
					image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
					if err != nil {
						return err
					}

					if *actual.Compute.LaunchSpecification.ImageID != *image.ImageId {
						if group.Compute == nil {
							group.Compute = new(aws.Compute)
						}
						if group.Compute.LaunchSpecification == nil {
							group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
						}

						group.Compute.LaunchSpecification.SetImageId(image.ImageId)
						changed = true
					}

					changes.ImageID = nil
				}
			}

			// Tags.
			{
				if changes.Tags != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetTags(e.buildTags())
					changes.Tags = nil
					changed = true
				}
			}

			// IAM instance profile.
			{
				if changes.IAMInstanceProfile != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					iprof := new(aws.IAMInstanceProfile)
					iprof.SetName(e.IAMInstanceProfile.GetName())

					group.Compute.LaunchSpecification.SetIAMInstanceProfile(iprof)
					changes.IAMInstanceProfile = nil
					changed = true
				}
			}

			// Monitoring.
			{
				if changes.Monitoring != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetMonitoring(e.Monitoring)
					changes.Monitoring = nil
					changed = true
				}
			}

			// SSH key.
			{
				if changes.SSHKey != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetKeyPair(e.SSHKey.Name)
					changes.SSHKey = nil
					changed = true
				}
			}

			// Load balancers.
			{
				if len(changes.LoadBalancers) > 0 {
					lbs, err := e.buildLoadBalancers(cloud)
					if err != nil {
						return err
					}

					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}
					if group.Compute.LaunchSpecification.LoadBalancersConfig == nil {
						group.Compute.LaunchSpecification.LoadBalancersConfig = new(aws.LoadBalancersConfig)
					}

					group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers = append(group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers, lbs...)

					changes.LoadBalancers = nil
					changed = true
				}
			}

			// Target groups.
			{
				if len(changes.TargetGroups) > 0 {
					tgs := e.buildTargetGroups()

					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}
					if group.Compute.LaunchSpecification.LoadBalancersConfig == nil {
						group.Compute.LaunchSpecification.LoadBalancersConfig = new(aws.LoadBalancersConfig)
					}

					group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers = append(group.Compute.LaunchSpecification.LoadBalancersConfig.LoadBalancers, tgs...)

					changes.TargetGroups = nil
					changed = true
				}
			}

			// Tenancy.
			{
				if changes.Tenancy != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetTenancy(e.Tenancy)
					changes.Tenancy = nil
					changed = true
				}
			}

			// Health check.
			{
				if changes.HealthCheckType != nil {
					if group.Compute == nil {
						group.Compute = new(aws.Compute)
					}
					if group.Compute.LaunchSpecification == nil {
						group.Compute.LaunchSpecification = new(aws.LaunchSpecification)
					}

					group.Compute.LaunchSpecification.SetHealthCheckType(e.HealthCheckType)
					changes.HealthCheckType = nil
					changed = true
				}
			}
		}
	}

	// Capacity.
	{
		if changes.MinSize != nil {
			if group.Capacity == nil {
				group.Capacity = new(aws.Capacity)
			}

			group.Capacity.SetMinimum(fi.Int(int(*e.MinSize)))
			changes.MinSize = nil
			changed = true

			// Scale up the target capacity, if needed.
			if int64(*actual.Capacity.Target) < *e.MinSize {
				group.Capacity.SetTarget(fi.Int(int(*e.MinSize)))
			}
		}
		if changes.MaxSize != nil {
			if group.Capacity == nil {
				group.Capacity = new(aws.Capacity)
			}

			group.Capacity.SetMaximum(fi.Int(int(*e.MaxSize)))
			changes.MaxSize = nil
			changed = true
		}
	}

	// Auto Scaler.
	{
		if opts := changes.AutoScalerOpts; opts != nil {
			if opts.Enabled != nil {
				autoScaler := new(aws.AutoScaleKubernetes)
				autoScaler.IsEnabled = e.AutoScalerOpts.Enabled
				autoScaler.Cooldown = e.AutoScalerOpts.Cooldown

				// Headroom.
				if headroom := opts.Headroom; headroom != nil {
					autoScaler.IsAutoConfig = fi.Bool(false)
					autoScaler.Headroom = &aws.AutoScaleHeadroom{
						CPUPerUnit:    e.AutoScalerOpts.Headroom.CPUPerUnit,
						GPUPerUnit:    e.AutoScalerOpts.Headroom.GPUPerUnit,
						MemoryPerUnit: e.AutoScalerOpts.Headroom.MemPerUnit,
						NumOfUnits:    e.AutoScalerOpts.Headroom.NumOfUnits,
					}
				} else if a.AutoScalerOpts != nil && a.AutoScalerOpts.Headroom != nil {
					autoScaler.IsAutoConfig = fi.Bool(true)
					autoScaler.SetHeadroom(nil)
				}

				// Scale down.
				if down := opts.Down; down != nil {
					autoScaler.Down = &aws.AutoScaleDown{
						MaxScaleDownPercentage: down.MaxPercentage,
						EvaluationPeriods:      down.EvaluationPeriods,
					}
				} else if a.AutoScalerOpts.Down != nil {
					autoScaler.SetDown(nil)
				}

				// Labels.
				if labels := opts.Labels; labels != nil {
					autoScaler.Labels = e.buildAutoScaleLabels(e.AutoScalerOpts.Labels)
				} else if a.AutoScalerOpts.Labels != nil {
					autoScaler.SetLabels(nil)
				}

				k8s := new(aws.KubernetesIntegration)
				k8s.SetAutoScale(autoScaler)

				integration := new(aws.Integration)
				integration.SetKubernetes(k8s)

				group.SetIntegration(integration)
				changed = true
			}

			changes.AutoScalerOpts = nil
		}
	}

	empty := &Elastigroup{}
	if !reflect.DeepEqual(empty, changes) {
		klog.Warningf("Not all changes applied to Elastigroup %q: %v", *e.Name, changes)
	}

	if !changed {
		klog.V(2).Infof("No changes detected in Elastigroup %q", *e.Name)
		return nil
	}

	klog.V(2).Infof("Updating Elastigroup %q (config: %s)", *e.Name, stringutil.Stringify(group))

	// Wrap the raw object as an Elastigroup.
	eg, err := spotinst.NewElastigroup(cloud.ProviderID(), group)
	if err != nil {
		return err
	}

	// Update the Elastigroup.
	if err := cloud.Spotinst().Elastigroup().Update(context.Background(), eg); err != nil {
		return fmt.Errorf("spotinst: failed to update elastigroup: %v", err)
	}

	return nil
}

type terraformElastigroup struct {
	Name                 *string                                 `cty:"name"`
	Description          *string                                 `cty:"description"`
	Product              *string                                 `cty:"product"`
	Region               *string                                 `cty:"region"`
	SubnetIDs            []*terraformWriter.Literal              `cty:"subnet_ids"`
	LoadBalancers        []*terraformWriter.Literal              `cty:"elastic_load_balancers"`
	TargetGroups         []*terraformWriter.Literal              `cty:"target_group_arns"`
	NetworkInterfaces    []*terraformElastigroupNetworkInterface `cty:"network_interface"`
	RootBlockDevice      *terraformElastigroupBlockDevice        `cty:"ebs_block_device"`
	EphemeralBlockDevice []*terraformElastigroupBlockDevice      `cty:"ephemeral_block_device"`
	Integration          *terraformElastigroupIntegration        `cty:"integration_kubernetes"`
	Tags                 []*terraformKV                          `cty:"tags"`

	MinSize         *int64  `cty:"min_size"`
	MaxSize         *int64  `cty:"max_size"`
	DesiredCapacity *int64  `cty:"desired_capacity"`
	CapacityUnit    *string `cty:"capacity_unit"`

	SpotPercentage           *float64 `cty:"spot_percentage"`
	Orientation              *string  `cty:"orientation"`
	FallbackToOnDemand       *bool    `cty:"fallback_to_ondemand"`
	UtilizeReservedInstances *bool    `cty:"utilize_reserved_instances"`
	UtilizeCommitments       *bool    `cty:"utilize_commitments"`
	DrainingTimeout          *int64   `cty:"draining_timeout"`

	OnDemand *string  `cty:"instance_types_ondemand"`
	Spot     []string `cty:"instance_types_spot"`

	Monitoring         *bool                      `cty:"enable_monitoring"`
	EBSOptimized       *bool                      `cty:"ebs_optimized"`
	ImageID            *string                    `cty:"image_id"`
	HealthCheckType    *string                    `cty:"health_check_type"`
	SecurityGroups     []*terraformWriter.Literal `cty:"security_groups"`
	UserData           *terraformWriter.Literal   `cty:"user_data"`
	IAMInstanceProfile *terraformWriter.Literal   `cty:"iam_instance_profile"`
	KeyName            *terraformWriter.Literal   `cty:"key_name"`
}

type terraformElastigroupBlockDevice struct {
	DeviceName          *string `cty:"device_name"`
	VirtualName         *string `cty:"virtual_name"`
	VolumeType          *string `cty:"volume_type"`
	VolumeSize          *int64  `cty:"volume_size"`
	VolumeIOPS          *int64  `cty:"iops"`
	VolumeThroughput    *int64  `cty:"throughput"`
	Encrypted           *bool   `cty:"encrypted"`
	DeleteOnTermination *bool   `cty:"delete_on_termination"`
}

type terraformElastigroupNetworkInterface struct {
	Description              *string `cty:"description"`
	DeviceIndex              *int    `cty:"device_index"`
	AssociatePublicIPAddress *bool   `cty:"associate_public_ip_address"`
	DeleteOnTermination      *bool   `cty:"delete_on_termination"`
}

type terraformElastigroupIntegration struct {
	IntegrationMode   *string `cty:"integration_mode"`
	ClusterIdentifier *string `cty:"cluster_identifier"`

	Enabled    *bool                        `cty:"autoscale_is_enabled"`
	AutoConfig *bool                        `cty:"autoscale_is_auto_config"`
	Cooldown   *int                         `cty:"autoscale_cooldown"`
	Headroom   *terraformAutoScalerHeadroom `cty:"autoscale_headroom"`
	Down       *terraformAutoScalerDown     `cty:"autoscale_down"`
	Labels     []*terraformKV               `cty:"autoscale_labels"`
}

type terraformAutoScaler struct {
	Enabled                *bool                              `cty:"autoscale_is_enabled"`
	AutoConfig             *bool                              `cty:"autoscale_is_auto_config"`
	AutoHeadroomPercentage *int                               `cty:"auto_headroom_percentage"`
	Cooldown               *int                               `cty:"autoscale_cooldown"`
	Headroom               *terraformAutoScalerHeadroom       `cty:"autoscale_headroom"`
	Down                   *terraformAutoScalerDown           `cty:"autoscale_down"`
	ResourceLimits         *terraformAutoScalerResourceLimits `cty:"resource_limits"`
	Labels                 []*terraformKV                     `cty:"autoscale_labels"`
}

type terraformAutoScalerHeadroom struct {
	CPUPerUnit *int `cty:"cpu_per_unit"`
	GPUPerUnit *int `cty:"gpu_per_unit"`
	MemPerUnit *int `cty:"memory_per_unit"`
	NumOfUnits *int `cty:"num_of_units"`
}

type terraformAutoScalerDown struct {
	MaxPercentage     *float64 `cty:"max_scale_down_percentage"`
	EvaluationPeriods *int     `cty:"evaluation_periods"`
}

type terraformAutoScalerResourceLimits struct {
	MaxVCPU   *int `cty:"max_vcpu"`
	MaxMemory *int `cty:"max_memory_gib"`
}

type terraformKV struct {
	Key   *string `cty:"key"`
	Value *string `cty:"value"`
}

type terraformTaint struct {
	Key    *string `cty:"key"`
	Value  *string `cty:"value"`
	Effect *string `cty:"effect"`
}

func (_ *Elastigroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Elastigroup) error {
	cloud := t.Cloud.(awsup.AWSCloud)
	e.applyDefaults()

	tf := &terraformElastigroup{
		Name:        e.Name,
		Description: e.Name,
		Product:     e.Product,
		Region:      e.Region,

		DesiredCapacity: e.MinSize,
		MinSize:         e.MinSize,
		MaxSize:         e.MaxSize,
		CapacityUnit:    fi.String("instance"),

		SpotPercentage:           e.SpotPercentage,
		Orientation:              fi.String(string(normalizeOrientation(e.Orientation))),
		FallbackToOnDemand:       e.FallbackToOnDemand,
		UtilizeReservedInstances: e.UtilizeReservedInstances,
		UtilizeCommitments:       e.UtilizeCommitments,
		DrainingTimeout:          e.DrainingTimeout,

		OnDemand: e.OnDemandInstanceType,
		Spot:     e.SpotInstanceTypes,
	}

	// Image.
	if e.ImageID != nil {
		image, err := resolveImage(cloud, fi.StringValue(e.ImageID))
		if err != nil {
			return err
		}
		tf.ImageID = image.ImageId
	}

	var role string
	for key := range e.Tags {
		if strings.HasPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix) {
			suffix := strings.TrimPrefix(key, awstasks.CloudTagInstanceGroupRolePrefix)
			if role != "" && role != suffix {
				return fmt.Errorf("spotinst: found multiple role tags %q vs %q", role, suffix)
			}
			role = suffix
		}
	}

	// Security groups.
	if e.SecurityGroups != nil {
		for _, sg := range e.SecurityGroups {
			tf.SecurityGroups = append(tf.SecurityGroups, sg.TerraformLink())
			if role != "" {
				if err := t.AddOutputVariableArray(role+"_security_groups", sg.TerraformLink()); err != nil {
					return err
				}
			}
		}
	}

	// User data.
	if e.UserData != nil {
		var err error
		tf.UserData, err = t.AddFileResource("spotinst_elastigroup_aws", *e.Name, "user_data", e.UserData, false)
		if err != nil {
			return err
		}
	}

	// IAM instance profile.
	if e.IAMInstanceProfile != nil {
		tf.IAMInstanceProfile = e.IAMInstanceProfile.TerraformLink()
	}

	// Monitoring.
	if e.Monitoring != nil {
		tf.Monitoring = e.Monitoring
	}

	// Health check.
	if e.HealthCheckType != nil {
		tf.HealthCheckType = e.HealthCheckType
	}

	// SSH key.
	if e.SSHKey != nil {
		tf.KeyName = e.SSHKey.TerraformLink()
	}

	// Subnets.
	if e.Subnets != nil {
		for _, subnet := range e.Subnets {
			tf.SubnetIDs = append(tf.SubnetIDs, subnet.TerraformLink())
			if role != "" {
				if err := t.AddOutputVariableArray(role+"_subnet_ids", subnet.TerraformLink()); err != nil {
					return err
				}
			}
		}
	}

	// Load balancers.
	if e.LoadBalancers != nil {
		for _, lb := range e.LoadBalancers {
			tf.LoadBalancers = append(tf.LoadBalancers, lb.TerraformLink())
		}
	}

	// Target groups.
	if e.TargetGroups != nil {
		for _, tg := range e.TargetGroups {
			tf.TargetGroups = append(tf.TargetGroups, tg.TerraformLink())
		}
	}

	// Public IP.
	if e.AssociatePublicIPAddress != nil {
		tf.NetworkInterfaces = append(tf.NetworkInterfaces, &terraformElastigroupNetworkInterface{
			Description:              fi.String("eth0"),
			DeviceIndex:              fi.Int(0),
			DeleteOnTermination:      fi.Bool(true),
			AssociatePublicIPAddress: e.AssociatePublicIPAddress,
		})
	}

	// Root volume options.
	{
		if opts := e.RootVolumeOpts; opts != nil {
			// Block device mappings.
			{
				rootDevice, err := buildRootDevice(t.Cloud.(awsup.AWSCloud), e.RootVolumeOpts, e.ImageID)
				if err != nil {
					return err
				}

				tf.RootBlockDevice = &terraformElastigroupBlockDevice{
					DeviceName:          rootDevice.DeviceName,
					VolumeType:          rootDevice.EbsVolumeType,
					VolumeSize:          rootDevice.EbsVolumeSize,
					VolumeIOPS:          rootDevice.EbsVolumeIops,
					VolumeThroughput:    rootDevice.EbsVolumeThroughput,
					Encrypted:           rootDevice.EbsEncrypted,
					DeleteOnTermination: fi.Bool(true),
				}

				ephemeralDevices, err := buildEphemeralDevices(cloud, e.OnDemandInstanceType)
				if err != nil {
					return err
				}

				if len(ephemeralDevices) != 0 {
					tf.EphemeralBlockDevice = make([]*terraformElastigroupBlockDevice, len(ephemeralDevices))
					for i, bdm := range ephemeralDevices {
						tf.EphemeralBlockDevice[i] = &terraformElastigroupBlockDevice{
							DeviceName:  bdm.DeviceName,
							VirtualName: bdm.VirtualName,
						}
					}
				}
			}

			// EBS optimization.
			{
				if opts.Optimization != nil {
					tf.EBSOptimized = opts.Optimization
				}
			}
		}
	}

	// Auto Scaler.
	{
		if opts := e.AutoScalerOpts; opts != nil {
			tf.Integration = &terraformElastigroupIntegration{
				IntegrationMode:   fi.String("pod"),
				ClusterIdentifier: opts.ClusterID,
			}

			if opts.Enabled != nil {
				tf.Integration.Enabled = opts.Enabled
				tf.Integration.AutoConfig = fi.Bool(true)
				tf.Integration.Cooldown = opts.Cooldown

				// Headroom.
				if headroom := opts.Headroom; headroom != nil {
					tf.Integration.AutoConfig = fi.Bool(false)
					tf.Integration.Headroom = &terraformAutoScalerHeadroom{
						CPUPerUnit: headroom.CPUPerUnit,
						GPUPerUnit: headroom.GPUPerUnit,
						MemPerUnit: headroom.MemPerUnit,
						NumOfUnits: headroom.NumOfUnits,
					}
				}

				// Scale down.
				if down := opts.Down; down != nil {
					tf.Integration.Down = &terraformAutoScalerDown{
						MaxPercentage:     down.MaxPercentage,
						EvaluationPeriods: down.EvaluationPeriods,
					}
				}

				// Labels.
				if labels := opts.Labels; labels != nil {
					tf.Integration.Labels = make([]*terraformKV, 0, len(labels))
					for k, v := range labels {
						tf.Integration.Labels = append(tf.Integration.Labels, &terraformKV{
							Key:   fi.String(k),
							Value: fi.String(v),
						})
					}
				}
			}
		}
	}

	// Tags.
	{
		if e.Tags != nil {
			tags := e.buildTags()
			for _, tag := range tags {
				tf.Tags = append(tf.Tags, &terraformKV{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
		}
	}

	return t.RenderResource("spotinst_elastigroup_aws", *e.Name, tf)
}

func (e *Elastigroup) TerraformLink() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("spotinst_elastigroup_aws", *e.Name, "id")
}

func (e *Elastigroup) buildTags() []*aws.Tag {
	tags := make([]*aws.Tag, 0, len(e.Tags))

	for key, value := range e.Tags {
		tags = append(tags, &aws.Tag{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return tags
}

func (e *Elastigroup) buildAutoScaleLabels(labelsMap map[string]string) []*aws.AutoScaleLabel {
	labels := make([]*aws.AutoScaleLabel, 0, len(labelsMap))
	for key, value := range labelsMap {
		labels = append(labels, &aws.AutoScaleLabel{
			Key:   fi.String(key),
			Value: fi.String(value),
		})
	}

	return labels
}

func (e *Elastigroup) buildLoadBalancers(cloud awsup.AWSCloud) ([]*aws.LoadBalancer, error) {
	lbs := make([]*aws.LoadBalancer, len(e.LoadBalancers))
	for i, lb := range e.LoadBalancers {
		if lb.LoadBalancerName == nil {
			lbName := fi.StringValue(lb.GetName())
			lbDesc, err := cloud.FindELBByNameTag(lbName)
			if err != nil {
				return nil, err
			}
			if lbDesc == nil {
				return nil, fmt.Errorf("spotinst: could not find "+
					"load balancer to attach: %s", lbName)
			}
			lbs[i] = &aws.LoadBalancer{
				Type: fi.String("CLASSIC"),
				Name: lbDesc.LoadBalancerName,
			}
		} else {
			lbs[i] = &aws.LoadBalancer{
				Type: fi.String("CLASSIC"),
				Name: lb.LoadBalancerName,
			}
		}
	}
	return lbs, nil
}

func (e *Elastigroup) buildTargetGroups() []*aws.LoadBalancer {
	tgs := make([]*aws.LoadBalancer, len(e.TargetGroups))
	for i, tg := range e.TargetGroups {
		tgs[i] = &aws.LoadBalancer{
			Type: fi.String("TARGET_GROUP"),
			Arn:  tg.ARN,
		}
	}
	return tgs
}

func buildEphemeralDevices(cloud awsup.AWSCloud, machineType *string) ([]*awstasks.BlockDeviceMapping, error) {
	info, err := awsup.GetMachineTypeInfo(cloud, fi.StringValue(machineType))
	if err != nil {
		return nil, err
	}

	bdms := make([]*awstasks.BlockDeviceMapping, len(info.EphemeralDevices()))
	for i, ed := range info.EphemeralDevices() {
		bdms[i] = &awstasks.BlockDeviceMapping{
			DeviceName:  fi.String(ed.DeviceName),
			VirtualName: fi.String(ed.VirtualName),
		}
	}

	return bdms, nil
}

func buildRootDevice(cloud awsup.AWSCloud, volumeOpts *RootVolumeOpts,
	imageID *string) (*awstasks.BlockDeviceMapping, error,
) {
	img, err := resolveImage(cloud, fi.StringValue(imageID))
	if err != nil {
		return nil, err
	}

	bdm := &awstasks.BlockDeviceMapping{
		DeviceName:             img.RootDeviceName,
		EbsVolumeSize:          volumeOpts.Size,
		EbsVolumeType:          volumeOpts.Type,
		EbsEncrypted:           volumeOpts.Encryption,
		EbsDeleteOnTermination: fi.Bool(true),
	}

	// IOPS is not supported for gp2 volumes.
	if volumeOpts.IOPS != nil && fi.StringValue(volumeOpts.Type) != "gp2" {
		bdm.EbsVolumeIops = volumeOpts.IOPS
	}

	// Throughput is only supported for gp3 volumes.
	if volumeOpts.Throughput != nil && fi.StringValue(volumeOpts.Type) == "gp3" {
		bdm.EbsVolumeThroughput = volumeOpts.Throughput
	}

	return bdm, nil
}

func (e *Elastigroup) convertBlockDeviceMapping(in *awstasks.BlockDeviceMapping) *aws.BlockDeviceMapping {
	out := &aws.BlockDeviceMapping{
		DeviceName:  in.DeviceName,
		VirtualName: in.VirtualName,
	}

	if in.EbsDeleteOnTermination != nil || in.EbsVolumeSize != nil || in.EbsVolumeType != nil {
		out.EBS = &aws.EBS{
			VolumeType:          in.EbsVolumeType,
			VolumeSize:          fi.Int(int(fi.Int64Value(in.EbsVolumeSize))),
			Encrypted:           in.EbsEncrypted,
			DeleteOnTermination: in.EbsDeleteOnTermination,
		}

		// IOPS is not valid for gp2 volumes.
		if in.EbsVolumeIops != nil && fi.StringValue(in.EbsVolumeType) != "gp2" {
			out.EBS.IOPS = fi.Int(int(fi.Int64Value(in.EbsVolumeIops)))
		}

		// Throughput is only valid for gp3 volumes.
		if in.EbsVolumeThroughput != nil && fi.StringValue(in.EbsVolumeType) == "gp3" {
			out.EBS.Throughput = fi.Int(int(fi.Int64Value(in.EbsVolumeThroughput)))
		}
	}

	return out
}

func (e *Elastigroup) applyDefaults() {
	if e.FallbackToOnDemand == nil {
		e.FallbackToOnDemand = fi.Bool(true)
	}

	if e.UtilizeReservedInstances == nil {
		e.UtilizeReservedInstances = fi.Bool(true)
	}

	if e.Product == nil || (e.Product != nil && fi.StringValue(e.Product) == "") {
		e.Product = fi.String("Linux/UNIX")
	}

	if e.Orientation == nil || (e.Orientation != nil && fi.StringValue(e.Orientation) == "") {
		e.Orientation = fi.String("balanced")
	}

	if e.Monitoring == nil {
		e.Monitoring = fi.Bool(false)
	}

	if e.HealthCheckType == nil {
		e.HealthCheckType = fi.String("K8S_NODE")
	}
}

func resolveImage(cloud awsup.AWSCloud, name string) (*ec2.Image, error) {
	image, err := cloud.ResolveImage(name)
	if err != nil {
		return nil, fmt.Errorf("spotinst: unable to resolve image %q: %v", name, err)
	} else if image == nil {
		return nil, fmt.Errorf("spotinst: unable to resolve image %q: not found", name)
	}

	return image, nil
}

func subnetSlicesEqualIgnoreOrder(l, r []*awstasks.Subnet) bool {
	var lIDs []string
	for _, s := range l {
		lIDs = append(lIDs, *s.ID)
	}

	var rIDs []string
	for _, s := range r {
		if s.ID == nil {
			klog.V(4).Infof("Subnet ID not set; returning not-equal: %v", s)
			return false
		}
		rIDs = append(rIDs, *s.ID)
	}

	return utils.StringSlicesEqualIgnoreOrder(lIDs, rIDs)
}

type Orientation string

const (
	OrientationBalanced              Orientation = "balanced"
	OrientationCost                  Orientation = "costOriented"
	OrientationAvailability          Orientation = "availabilityOriented"
	OrientationEqualZoneDistribution Orientation = "equalAzDistribution"
)

func normalizeOrientation(orientation *string) Orientation {
	out := OrientationBalanced

	// Fast path.
	if orientation == nil {
		return out
	}

	switch *orientation {
	case "cost":
		out = OrientationCost
	case "availability":
		out = OrientationAvailability
	case "equal-distribution":
		out = OrientationEqualZoneDistribution
	}

	return out
}
