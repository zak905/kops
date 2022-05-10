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

package openstackmodel

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
	"k8s.io/utils/net"
)

// ServerGroupModelBuilder configures server group objects
type ServerGroupModelBuilder struct {
	*OpenstackModelContext
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
}

var _ fi.ModelBuilder = &ServerGroupModelBuilder{}

// See https://specs.openstack.org/openstack/nova-specs/specs/newton/approved/lowercase-metadata-keys.html for details
var instanceMetadataNotAllowedCharacters = regexp.MustCompile("[^a-zA-Z0-9-_:. ]")

func (b *ServerGroupModelBuilder) buildInstances(c *fi.ModelBuilderContext, sg *openstacktasks.ServerGroup, ig *kops.InstanceGroup) error {
	sshKeyNameFull, err := b.SSHKeyName()
	if err != nil {
		return err
	}

	sshKeyName := strings.Replace(sshKeyNameFull, ":", "_", -1)

	igMeta := make(map[string]string)
	cloudTags, err := b.KopsModelContext.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return fmt.Errorf("could not get cloud tags for instance group %s: %v", ig.Name, err)
	}
	for label, labelVal := range cloudTags {
		sanitizedLabel := strings.ToLower(
			instanceMetadataNotAllowedCharacters.ReplaceAllLiteralString(label, "_"),
		)
		igMeta[sanitizedLabel] = labelVal
	}
	if ig.Spec.Role != kops.InstanceGroupRoleBastion {
		// Bastion does not belong to the cluster and will not be running protokube.

		igMeta[openstack.TagClusterName] = b.ClusterName()
	}
	igMeta["k8s"] = b.ClusterName()
	netName, err := b.GetNetworkName()
	if err != nil {
		return err
	}
	igMeta[openstack.TagKopsNetwork] = netName
	igMeta[openstack.TagKopsInstanceGroup] = ig.Name
	igMeta[openstack.TagKopsRole] = string(ig.Spec.Role)
	igMeta[openstack.INSTANCE_GROUP_GENERATION] = fmt.Sprintf("%d", ig.GetGeneration())
	igMeta[openstack.CLUSTER_GENERATION] = fmt.Sprintf("%d", b.Cluster.GetGeneration())

	if e, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.BOOT_FROM_VOLUME]; ok {
		igMeta[openstack.BOOT_FROM_VOLUME] = e
	}

	if v, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.BOOT_VOLUME_SIZE]; ok {
		igMeta[openstack.BOOT_VOLUME_SIZE] = v
	}

	startupScript, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
	if err != nil {
		return fmt.Errorf("could not create startup script for instance group %s: %v", ig.Name, err)
	}

	var securityGroups []*openstacktasks.SecurityGroup
	securityGroupName := b.SecurityGroupName(ig.Spec.Role)
	securityGroups = append(securityGroups, b.LinkToSecurityGroup(securityGroupName))

	if b.Cluster.Spec.CloudProvider.Openstack.Loadbalancer == nil && ig.Spec.Role == kops.InstanceGroupRoleMaster {
		securityGroups = append(securityGroups, b.LinkToSecurityGroup(b.Cluster.Spec.MasterPublicName))
	}

	r := strings.NewReplacer("_", "-", ".", "-")
	groupName := r.Replace(strings.ToLower(ig.Name))
	// In the future, OpenStack will use Machine API to manage groups,
	// for now create d.InstanceGroups.Spec.MinSize amount of servers
	for i := int32(0); i < *ig.Spec.MinSize; i++ {
		// FIXME: Must ensure 63 or less characters
		// replace all dots and _ with -, this is needed to get external cloudprovider working
		iName := strings.Replace(strings.ToLower(fmt.Sprintf("%s-%d.%s", ig.Name, i+1, b.ClusterName())), "_", "-", -1)
		instanceName := fi.String(strings.Replace(iName, ".", "-", -1))

		var az *string
		var subnets []*openstacktasks.Subnet
		if len(ig.Spec.Subnets) > 0 {
			subnet := ig.Spec.Subnets[int(i)%len(ig.Spec.Subnets)]
			// bastion subnet name might contain a "utility-" prefix
			if ig.Spec.Role == kops.InstanceGroupRoleBastion {
				az = fi.String(strings.Replace(subnet, "utility-", "", 1))
			} else {
				az = fi.String(subnet)
			}

			subnetName, err := b.findSubnetClusterSpec(subnet)
			if err != nil {
				return err
			}
			subnets = append(subnets, b.LinkToSubnet(s(subnetName)))
		}
		if len(ig.Spec.Zones) > 0 {
			zone := ig.Spec.Zones[int(i)%len(ig.Spec.Zones)]
			az = fi.String(zone)
		}
		// Create instance port task
		portName := fmt.Sprintf("%s-%s", "port", *instanceName)
		portTagKopsName := strings.Replace(
			strings.Replace(
				strings.ToLower(
					fmt.Sprintf("port-%s-%d", ig.Name, i+1),
				),
				"_", "-", -1,
			), ".", "-", -1,
		)
		portTask := &openstacktasks.Port{
			Name:              fi.String(portName),
			InstanceGroupName: &groupName,
			Network:           b.LinkToNetwork(),
			Tags: []string{
				fmt.Sprintf("%s=%s", openstack.TagKopsInstanceGroup, groupName),
				fmt.Sprintf("%s=%s", openstack.TagKopsName, portTagKopsName),
				fmt.Sprintf("%s=%s", openstack.TagClusterName, b.ClusterName()),
			},
			SecurityGroups:           securityGroups,
			AdditionalSecurityGroups: ig.Spec.AdditionalSecurityGroups,
			Subnets:                  subnets,
			Lifecycle:                b.Lifecycle,
		}
		c.AddTask(portTask)

		metaWithName := make(map[string]string)
		for k, v := range igMeta {
			metaWithName[k] = v
		}
		metaWithName[openstack.TagKopsName] = fi.StringValue(instanceName)
		instanceTask := &openstacktasks.Instance{
			Name:             instanceName,
			Lifecycle:        b.Lifecycle,
			GroupName:        s(groupName),
			Region:           fi.String(b.Cluster.Spec.Subnets[0].Region),
			Flavor:           fi.String(ig.Spec.MachineType),
			Image:            fi.String(ig.Spec.Image),
			SSHKey:           fi.String(sshKeyName),
			ServerGroup:      sg,
			Role:             fi.String(string(ig.Spec.Role)),
			Port:             portTask,
			UserData:         startupScript,
			Metadata:         metaWithName,
			SecurityGroups:   ig.Spec.AdditionalSecurityGroups,
			AvailabilityZone: az,
			ConfigDrive:      b.Cluster.Spec.CloudProvider.Openstack.Metadata.ConfigDrive,
		}
		c.AddTask(instanceTask)

		// Associate a floating IP to the instances if we have external network in router
		// and respective topology is "public"
		if b.Cluster.Spec.CloudProvider.Openstack.Router != nil {
			if ig.Spec.AssociatePublicIP != nil && !fi.BoolValue(ig.Spec.AssociatePublicIP) {
				continue
			}
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleBastion:
				t := &openstacktasks.FloatingIP{
					Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
					Lifecycle: b.Lifecycle,
				}
				c.AddTask(t)
				instanceTask.FloatingIP = t
			case kops.InstanceGroupRoleMaster:

				if b.Cluster.Spec.Topology == nil || b.Cluster.Spec.Topology.Masters != kops.TopologyPrivate {
					t := &openstacktasks.FloatingIP{
						Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
						Lifecycle: b.Lifecycle,
					}
					c.AddTask(t)
					b.associateFIPToKeypair(t)
					instanceTask.FloatingIP = t
				}
			default:
				if b.Cluster.Spec.Topology == nil || b.Cluster.Spec.Topology.Nodes != kops.TopologyPrivate {
					t := &openstacktasks.FloatingIP{
						Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *instanceTask.Name)),
						Lifecycle: b.Lifecycle,
					}
					c.AddTask(t)
					instanceTask.FloatingIP = t
				}
			}
		}
	}

	return nil
}

func (b *ServerGroupModelBuilder) associateFIPToKeypair(fipTask *openstacktasks.FloatingIP) {
	// Ensure the floating IP is included in the TLS certificate,
	// if we're not going to use an alias for it
	fipTask.ForAPIServer = true
}

func (b *ServerGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	clusterName := b.ClusterName()

	var masters []*openstacktasks.ServerGroup
	for _, ig := range b.InstanceGroups {
		klog.V(2).Infof("Found instance group with name %s and role %v.", ig.Name, ig.Spec.Role)
		affinityPolicies := []string{}
		if v, ok := ig.ObjectMeta.Annotations[openstack.OS_ANNOTATION+openstack.SERVER_GROUP_AFFINITY]; ok {
			affinityPolicies = append(affinityPolicies, v)
		} else {
			affinityPolicies = append(affinityPolicies, "anti-affinity")
		}
		sgTask := &openstacktasks.ServerGroup{
			Name:        s(fmt.Sprintf("%s-%s", clusterName, ig.Name)),
			ClusterName: s(clusterName),
			IGName:      s(ig.Name),
			Policies:    affinityPolicies,
			Lifecycle:   b.Lifecycle,
			MaxSize:     ig.Spec.MaxSize,
		}
		c.AddTask(sgTask)

		err := b.buildInstances(c, sgTask, ig)
		if err != nil {
			return err
		}

		if ig.Spec.Role == kops.InstanceGroupRoleMaster {
			masters = append(masters, sgTask)
		}
	}

	if b.Cluster.Spec.CloudProvider.Openstack.Loadbalancer != nil {
		var lbSubnetName string
		var err error
		for _, sp := range b.Cluster.Spec.Subnets {
			if sp.Type == kops.SubnetTypeDualStack || sp.Type == kops.SubnetTypePrivate {
				lbSubnetName, err = b.findSubnetNameByID(sp.ProviderID, sp.Name)
				if err != nil {
					return err
				}
				break
			}
		}
		if lbSubnetName == "" {
			return fmt.Errorf("could not find subnet for master loadbalancer")
		}
		lbTask := &openstacktasks.LB{
			Name:      fi.String(b.Cluster.Spec.MasterPublicName),
			Subnet:    fi.String(lbSubnetName),
			Lifecycle: b.Lifecycle,
		}

		useVIPACL := b.UseVIPACL()
		if !useVIPACL {
			lbTask.SecurityGroup = b.LinkToSecurityGroup(b.Cluster.Spec.MasterPublicName)
		}

		c.AddTask(lbTask)

		lbfipTask := &openstacktasks.FloatingIP{
			Name:      fi.String(fmt.Sprintf("%s-%s", "fip", *lbTask.Name)),
			LB:        lbTask,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(lbfipTask)

		if dns.IsGossipHostname(b.Cluster.Name) || b.UsePrivateDNS() {
			b.associateFIPToKeypair(lbfipTask)
		}

		poolTask := &openstacktasks.LBPool{
			Name:         fi.String(fmt.Sprintf("%s-https", fi.StringValue(lbTask.Name))),
			Loadbalancer: lbTask,
			Lifecycle:    b.Lifecycle,
		}
		c.AddTask(poolTask)

		listenerTask := &openstacktasks.LBListener{
			Name:      lbTask.Name,
			Lifecycle: b.Lifecycle,
			Pool:      poolTask,
		}
		if useVIPACL {
			var AllowedCIDRs []string
			// currently kOps openstack supports only ipv4 addresses
			for _, CIDR := range b.Cluster.Spec.KubernetesAPIAccess {
				if net.IsIPv4CIDRString(CIDR) {
					AllowedCIDRs = append(AllowedCIDRs, CIDR)
				}
			}
			sort.Strings(AllowedCIDRs)
			listenerTask.AllowedCIDRs = AllowedCIDRs
		}
		c.AddTask(listenerTask)

		monitorTask := &openstacktasks.PoolMonitor{
			Name:      lbTask.Name,
			Pool:      poolTask,
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(monitorTask)

		ifName, err := b.GetNetworkName()
		if err != nil {
			return err
		}
		for _, mastersg := range masters {
			associateTask := &openstacktasks.PoolAssociation{
				Name:          mastersg.Name,
				Pool:          poolTask,
				ServerGroup:   mastersg,
				InterfaceName: fi.String(ifName),
				ProtocolPort:  fi.Int(443),
				Lifecycle:     b.Lifecycle,
				Weight:        fi.Int(1),
			}

			c.AddTask(associateTask)

		}

	}

	return nil
}
