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

package openstack

import (
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func (c *openstackCloud) CreateServerGroup(opt servergroups.CreateOptsBuilder) (*servergroups.ServerGroup, error) {
	return createServerGroup(c, opt)
}

func createServerGroup(c OpenstackCloud, opt servergroups.CreateOptsBuilder) (*servergroups.ServerGroup, error) {
	var i *servergroups.ServerGroup

	done, err := vfs.RetryWithBackoff(writeBackoff, func() (bool, error) {
		v, err := servergroups.Create(c.ComputeClient(), opt).Extract()
		if err != nil {
			return false, fmt.Errorf("error creating server group: %v", err)
		}
		i = v
		return true, nil
	})
	if err != nil {
		return i, err
	} else if done {
		return i, nil
	} else {
		return i, wait.ErrWaitTimeout
	}
}

func (c *openstackCloud) ListServerGroups(opts servergroups.ListOptsBuilder) ([]servergroups.ServerGroup, error) {
	return listServerGroups(c, opts)
}

func listServerGroups(c OpenstackCloud, opts servergroups.ListOptsBuilder) ([]servergroups.ServerGroup, error) {
	var sgs []servergroups.ServerGroup

	done, err := vfs.RetryWithBackoff(readBackoff, func() (bool, error) {
		allPages, err := servergroups.List(c.ComputeClient(), opts).AllPages()
		if err != nil {
			return false, fmt.Errorf("error listing server groups: %v", err)
		}

		r, err := servergroups.ExtractServerGroups(allPages)
		if err != nil {
			return false, fmt.Errorf("error extracting server groups from pages: %v", err)
		}
		sgs = r
		return true, nil
	})
	if err != nil {
		return sgs, err
	} else if done {
		return sgs, nil
	} else {
		return sgs, wait.ErrWaitTimeout
	}
}

// matchInstanceGroup filters a list of instancegroups for recognized cloud groups
func matchInstanceGroup(name string, clusterName string, instancegroups []*kops.InstanceGroup) (*kops.InstanceGroup, error) {
	var instancegroup *kops.InstanceGroup
	for _, g := range instancegroups {
		var groupName string

		switch g.Spec.Role {
		case kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleNode, kops.InstanceGroupRoleBastion:
			groupName = clusterName + "-" + g.ObjectMeta.Name
		default:
			klog.Warningf("Ignoring InstanceGroup of unknown role %q", g.Spec.Role)
			continue
		}

		if name == groupName {
			if instancegroup != nil {
				return nil, fmt.Errorf("found multiple instance groups matching servergrp %q", groupName)
			}
			instancegroup = g
		}
	}

	return instancegroup, nil
}

func osBuildCloudInstanceGroup(c OpenstackCloud, cluster *kops.Cluster, ig *kops.InstanceGroup, g servergroups.ServerGroup, nodeMap map[string]*v1.Node) (*cloudinstances.CloudInstanceGroup, error) {
	newLaunchConfigName := g.Name
	cg := &cloudinstances.CloudInstanceGroup{
		HumanName:     newLaunchConfigName,
		InstanceGroup: ig,
		MinSize:       int(fi.Int32Value(ig.Spec.MinSize)),
		TargetSize:    int(fi.Int32Value(ig.Spec.MinSize)), // TODO: Retrieve the target size from OpenStack?
		MaxSize:       int(fi.Int32Value(ig.Spec.MaxSize)),
		Raw:           &g,
	}
	for _, i := range g.Members {
		instanceId := i
		if instanceId == "" {
			klog.Warningf("ignoring instance with no instance id: %s", i)
			continue
		}
		server, err := servers.Get(c.ComputeClient(), instanceId).Extract()
		if err != nil {
			return nil, fmt.Errorf("Failed to get instance group member: %v", err)
		}
		igObservedGeneration := server.Metadata[INSTANCE_GROUP_GENERATION]
		clusterObservedGeneration := server.Metadata[CLUSTER_GENERATION]
		observedName := fmt.Sprintf("%s-%s", clusterObservedGeneration, igObservedGeneration)
		generationName := fmt.Sprintf("%d-%d", cluster.GetGeneration(), ig.Generation)

		status := cloudinstances.CloudInstanceStatusUpToDate
		if generationName != observedName {
			status = cloudinstances.CloudInstanceStatusNeedsUpdate
		}
		cm, err := cg.NewCloudInstance(instanceId, status, nodeMap[instanceId])
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}

		if server.Flavor["original_name"] != nil {
			cm.MachineType = server.Flavor["original_name"].(string)
		}

		ip, err := GetServerFixedIP(server, server.Metadata[TagKopsNetwork])
		if err != nil {
			return nil, fmt.Errorf("error creating cloud instance group member: %v", err)
		}

		cm.PrivateIP = ip

		cm.Roles = []string{server.Metadata["KopsRole"]}

		cm.State = cloudinstances.State(server.Status)

	}
	return cg, nil
}

func (c *openstackCloud) DeleteServerGroup(groupID string) error {
	return deleteServerGroup(c, groupID)
}

func deleteServerGroup(c OpenstackCloud, groupID string) error {
	done, err := vfs.RetryWithBackoff(deleteBackoff, func() (bool, error) {
		err := servergroups.Delete(c.ComputeClient(), groupID).ExtractErr()
		if err != nil && !isNotFound(err) {
			return false, fmt.Errorf("error deleting server group: %v", err)
		}
		if isNotFound(err) {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}
