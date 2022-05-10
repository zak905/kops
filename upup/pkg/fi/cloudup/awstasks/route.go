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

package awstasks

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/cloudformation"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type Route struct {
	Name      *string
	Lifecycle fi.Lifecycle

	RouteTable *RouteTable
	Instance   *Instance
	CIDR       *string
	IPv6CIDR   *string

	// Exactly one of the below fields
	// MUST be provided.
	EgressOnlyInternetGateway *EgressOnlyInternetGateway
	InternetGateway           *InternetGateway
	NatGateway                *NatGateway
	TransitGatewayID          *string
	VPCPeeringConnectionID    *string
}

func (e *Route) Find(c *fi.Context) (*Route, error) {
	cloud := c.Cloud.(awsup.AWSCloud)

	if e.RouteTable == nil || (e.CIDR == nil && e.IPv6CIDR == nil) {
		// TODO: Move to validate?
		return nil, nil
	}

	if e.RouteTable.ID == nil {
		return nil, nil
	}

	request := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []*string{e.RouteTable.ID},
	}

	response, err := cloud.EC2().DescribeRouteTables(request)
	if err != nil {
		return nil, fmt.Errorf("error listing RouteTables: %v", err)
	}
	if response == nil || len(response.RouteTables) == 0 {
		return nil, nil
	} else {
		if len(response.RouteTables) != 1 {
			klog.Fatalf("found multiple RouteTables matching tags")
		}
		rt := response.RouteTables[0]
		for _, r := range rt.Routes {
			if (r.DestinationCidrBlock == nil || aws.StringValue(r.DestinationCidrBlock) != aws.StringValue(e.CIDR)) &&
				(r.DestinationIpv6CidrBlock == nil || aws.StringValue(r.DestinationIpv6CidrBlock) != aws.StringValue(e.IPv6CIDR)) {
				continue
			}
			actual := &Route{
				Name:       e.Name,
				RouteTable: &RouteTable{ID: rt.RouteTableId},
				CIDR:       r.DestinationCidrBlock,
				IPv6CIDR:   r.DestinationIpv6CidrBlock,
			}
			if r.EgressOnlyInternetGatewayId != nil {
				actual.EgressOnlyInternetGateway = &EgressOnlyInternetGateway{ID: r.EgressOnlyInternetGatewayId}
			}
			if r.GatewayId != nil {
				actual.InternetGateway = &InternetGateway{ID: r.GatewayId}
			}
			if r.InstanceId != nil {
				actual.Instance = &Instance{ID: r.InstanceId}
			}
			if r.NatGatewayId != nil {
				actual.NatGateway = &NatGateway{ID: r.NatGatewayId}
			}
			if r.TransitGatewayId != nil {
				actual.TransitGatewayID = r.TransitGatewayId
			}
			if r.VpcPeeringConnectionId != nil {
				actual.VPCPeeringConnectionID = r.VpcPeeringConnectionId
			}

			if aws.StringValue(r.State) == "blackhole" {
				klog.V(2).Infof("found route is a blackhole route")
				// These should be nil anyway, but just in case...
				actual.Instance = nil
				actual.InternetGateway = nil
				actual.TransitGatewayID = nil
			}

			// Prevent spurious changes
			actual.Lifecycle = e.Lifecycle

			klog.V(2).Infof("found route matching CIDR=%q IPv6CIDR=%q", aws.StringValue(e.CIDR), aws.StringValue(e.IPv6CIDR))
			return actual, nil
		}
	}

	return nil, nil
}

func (e *Route) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *Route) CheckChanges(a, e, changes *Route) error {
	if a == nil {
		// TODO: Create validate method?
		if e.RouteTable == nil {
			return fi.RequiredField("RouteTable")
		}
		if e.CIDR == nil && e.IPv6CIDR == nil {
			return fi.RequiredField("CIDR/IPv6CIDR")
		}
		if e.CIDR != nil && e.IPv6CIDR != nil {
			return fmt.Errorf("cannot set more than one CIDR or IPv6CIDR")
		}
		targetCount := 0
		if e.EgressOnlyInternetGateway != nil {
			targetCount++
			if e.CIDR != nil {
				return fmt.Errorf("cannot route IPv4 to an EgressOnlyInternetGateway")
			}
		}
		if e.InternetGateway != nil {
			targetCount++
		}
		if e.Instance != nil {
			targetCount++
		}
		if e.NatGateway != nil {
			targetCount++
		}
		if e.TransitGatewayID != nil {
			targetCount++
		}
		if e.VPCPeeringConnectionID != nil {
			targetCount++
		}
		if targetCount == 0 {
			return fmt.Errorf("EgressOnlyInternetGateway, InternetGateway, Instance, NatGateway, TransitGateway, or VpcPeeringConnection is required")
		}
		if targetCount != 1 {
			return fmt.Errorf("cannot set more than one EgressOnlyInternetGateway, InternetGateway, Instance, NatGateway, TransitGateway, or VpcPeeringConnection")
		}
	}

	if a != nil {
		if changes.RouteTable != nil {
			return fi.CannotChangeField("RouteTable")
		}
		if changes.CIDR != nil {
			return fi.CannotChangeField("CIDR")
		}
		if changes.IPv6CIDR != nil {
			return fi.CannotChangeField("IPv6CIDR")
		}
	}
	return nil
}

func (_ *Route) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *Route) error {
	if a == nil {
		request := &ec2.CreateRouteInput{}
		request.RouteTableId = checkNotNil(e.RouteTable.ID)

		if e.CIDR != nil || e.IPv6CIDR != nil {
			request.DestinationCidrBlock = e.CIDR
			request.DestinationIpv6CidrBlock = e.IPv6CIDR
		} else {
			klog.Fatal("both CIDR and IPv6CIDR were unexpectedly nil")
		}

		if e.EgressOnlyInternetGateway == nil && e.InternetGateway == nil && e.NatGateway == nil && e.TransitGatewayID == nil && e.VPCPeeringConnectionID == nil {
			return fmt.Errorf("missing target for route")
		} else if e.EgressOnlyInternetGateway != nil {
			request.EgressOnlyInternetGatewayId = checkNotNil(e.EgressOnlyInternetGateway.ID)
		} else if e.InternetGateway != nil {
			request.GatewayId = checkNotNil(e.InternetGateway.ID)
		} else if e.NatGateway != nil {
			request.NatGatewayId = checkNotNil(e.NatGateway.ID)
		} else if e.TransitGatewayID != nil {
			request.TransitGatewayId = e.TransitGatewayID
		} else if e.VPCPeeringConnectionID != nil {
			request.VpcPeeringConnectionId = e.VPCPeeringConnectionID
		}

		if e.Instance != nil {
			request.InstanceId = checkNotNil(e.Instance.ID)
		}

		klog.V(2).Infof("Creating Route with RouteTable:%q CIDR:%q IPv6CIDR:%q",
			aws.StringValue(e.RouteTable.ID), aws.StringValue(e.CIDR), aws.StringValue(e.IPv6CIDR))

		response, err := t.Cloud.EC2().CreateRoute(request)
		if err != nil {
			code := awsup.AWSErrorCode(err)
			message := awsup.AWSErrorMessage(err)
			if code == "InvalidNatGatewayID.NotFound" {
				klog.V(4).Infof("error creating Route: %s", message)
				return fi.NewTryAgainLaterError("waiting for the NAT Gateway to be created")
			}
			return fmt.Errorf("error creating Route: %s", message)
		}

		if !aws.BoolValue(response.Return) {
			return fmt.Errorf("create Route request failed: %v", response)
		}
	} else {
		request := &ec2.ReplaceRouteInput{}
		request.RouteTableId = checkNotNil(e.RouteTable.ID)

		if e.CIDR != nil || e.IPv6CIDR != nil {
			request.DestinationCidrBlock = e.CIDR
			request.DestinationIpv6CidrBlock = e.IPv6CIDR
		} else {
			klog.Fatal("both CIDR and IPv6CIDR were unexpectedly nil")
		}

		if e.InternetGateway == nil && e.NatGateway == nil && e.TransitGatewayID == nil && e.VPCPeeringConnectionID == nil {
			return fmt.Errorf("missing target for route")
		} else if e.InternetGateway != nil {
			request.GatewayId = checkNotNil(e.InternetGateway.ID)
		} else if e.NatGateway != nil {
			request.NatGatewayId = checkNotNil(e.NatGateway.ID)
		} else if e.TransitGatewayID != nil {
			request.TransitGatewayId = e.TransitGatewayID
		} else if e.VPCPeeringConnectionID != nil {
			request.VpcPeeringConnectionId = e.VPCPeeringConnectionID
		}

		if e.Instance != nil {
			request.InstanceId = checkNotNil(e.Instance.ID)
		}

		klog.V(2).Infof("Updating Route with RouteTable:%q CIDR:%q", *e.RouteTable.ID, *e.CIDR)

		if _, err := t.Cloud.EC2().ReplaceRoute(request); err != nil {
			code := awsup.AWSErrorCode(err)
			message := awsup.AWSErrorMessage(err)
			if code == "InvalidNatGatewayID.NotFound" {
				klog.V(4).Infof("error creating Route: %s", message)
				return fi.NewTryAgainLaterError("waiting for the NAT Gateway to be created")
			}
			return fmt.Errorf("error creating Route: %s", message)
		}
	}

	return nil
}

func checkNotNil(s *string) *string {
	if s == nil {
		klog.Fatal("string pointer was unexpectedly nil")
	}
	return s
}

type terraformRoute struct {
	RouteTableID                *terraformWriter.Literal `cty:"route_table_id"`
	CIDR                        *string                  `cty:"destination_cidr_block"`
	IPv6CIDR                    *string                  `cty:"destination_ipv6_cidr_block"`
	EgressOnlyInternetGatewayID *terraformWriter.Literal `cty:"egress_only_gateway_id"`
	InternetGatewayID           *terraformWriter.Literal `cty:"gateway_id"`
	NATGatewayID                *terraformWriter.Literal `cty:"nat_gateway_id"`
	TransitGatewayID            *string                  `cty:"transit_gateway_id"`
	InstanceID                  *terraformWriter.Literal `cty:"instance_id"`
	VPCPeeringConnectionID      *string                  `cty:"vpc_peering_connection_id"`
}

func (_ *Route) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Route) error {
	tf := &terraformRoute{
		RouteTableID: e.RouteTable.TerraformLink(),
		CIDR:         e.CIDR,
		IPv6CIDR:     e.IPv6CIDR,
	}

	if e.EgressOnlyInternetGateway == nil && e.InternetGateway == nil && e.NatGateway == nil && e.TransitGatewayID == nil && e.VPCPeeringConnectionID == nil {
		return fmt.Errorf("missing target for route")
	} else if e.EgressOnlyInternetGateway != nil {
		tf.EgressOnlyInternetGatewayID = e.EgressOnlyInternetGateway.TerraformLink()
	} else if e.InternetGateway != nil {
		tf.InternetGatewayID = e.InternetGateway.TerraformLink()
	} else if e.NatGateway != nil {
		tf.NATGatewayID = e.NatGateway.TerraformLink()
	} else if e.TransitGatewayID != nil {
		tf.TransitGatewayID = e.TransitGatewayID
	} else if e.VPCPeeringConnectionID != nil {
		tf.VPCPeeringConnectionID = e.VPCPeeringConnectionID
	}

	if e.Instance != nil {
		tf.InstanceID = e.Instance.TerraformLink()
	}

	// Terraform 0.12 doesn't support resource names that start with digits. See #7052
	// and https://www.terraform.io/upgrade-guides/0-12.html#pre-upgrade-checklist
	name := fmt.Sprintf("route-%v", *e.Name)
	return t.RenderResource("aws_route", name, tf)
}

type cloudformationRoute struct {
	RouteTableID           *cloudformation.Literal `json:"RouteTableId"`
	CIDR                   *string                 `json:"DestinationCidrBlock,omitempty"`
	IPv6CIDR               *string                 `json:"DestinationIpv6CidrBlock,omitempty"`
	InternetGatewayID      *cloudformation.Literal `json:"GatewayId,omitempty"`
	NATGatewayID           *cloudformation.Literal `json:"NatGatewayId,omitempty"`
	TransitGatewayID       *string                 `json:"TransitGatewayId,omitempty"`
	InstanceID             *cloudformation.Literal `json:"InstanceId,omitempty"`
	VPCPeeringConnectionID *string                 `json:"VpcPeeringConnectionId,omitempty"`
}

func (_ *Route) RenderCloudformation(t *cloudformation.CloudformationTarget, a, e, changes *Route) error {
	tf := &cloudformationRoute{
		RouteTableID: e.RouteTable.CloudformationLink(),
		CIDR:         e.CIDR,
		IPv6CIDR:     e.IPv6CIDR,
	}

	if e.InternetGateway == nil && e.NatGateway == nil && e.TransitGatewayID == nil && e.VPCPeeringConnectionID == nil {
		return fmt.Errorf("missing target for route")
	} else if e.InternetGateway != nil {
		tf.InternetGatewayID = e.InternetGateway.CloudformationLink()
	} else if e.NatGateway != nil {
		tf.NATGatewayID = e.NatGateway.CloudformationLink()
	} else if e.TransitGatewayID != nil {
		tf.TransitGatewayID = e.TransitGatewayID
	} else if e.VPCPeeringConnectionID != nil {
		tf.VPCPeeringConnectionID = e.VPCPeeringConnectionID
	}

	if e.Instance != nil {
		return fmt.Errorf("instance cloudformation routes not yet implemented")
		// tf.InstanceID = e.Instance.CloudformationLink()
	}

	return t.RenderResource("AWS::EC2::Route", *e.Name, tf)
}
