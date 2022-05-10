locals {
  bastion_autoscaling_group_ids     = [aws_autoscaling_group.bastion-privateweave-example-com.id]
  bastion_security_group_ids        = [aws_security_group.bastion-privateweave-example-com.id]
  bastions_role_arn                 = aws_iam_role.bastions-privateweave-example-com.arn
  bastions_role_name                = aws_iam_role.bastions-privateweave-example-com.name
  cluster_name                      = "privateweave.example.com"
  master_autoscaling_group_ids      = [aws_autoscaling_group.master-us-test-1a-masters-privateweave-example-com.id]
  master_security_group_ids         = [aws_security_group.masters-privateweave-example-com.id]
  masters_role_arn                  = aws_iam_role.masters-privateweave-example-com.arn
  masters_role_name                 = aws_iam_role.masters-privateweave-example-com.name
  node_autoscaling_group_ids        = [aws_autoscaling_group.nodes-privateweave-example-com.id]
  node_security_group_ids           = [aws_security_group.nodes-privateweave-example-com.id]
  node_subnet_ids                   = [aws_subnet.us-test-1a-privateweave-example-com.id]
  nodes_role_arn                    = aws_iam_role.nodes-privateweave-example-com.arn
  nodes_role_name                   = aws_iam_role.nodes-privateweave-example-com.name
  region                            = "us-test-1"
  route_table_private-us-test-1a_id = aws_route_table.private-us-test-1a-privateweave-example-com.id
  route_table_public_id             = aws_route_table.privateweave-example-com.id
  subnet_us-test-1a_id              = aws_subnet.us-test-1a-privateweave-example-com.id
  subnet_utility-us-test-1a_id      = aws_subnet.utility-us-test-1a-privateweave-example-com.id
  vpc_cidr_block                    = aws_vpc.privateweave-example-com.cidr_block
  vpc_id                            = aws_vpc.privateweave-example-com.id
}

output "bastion_autoscaling_group_ids" {
  value = [aws_autoscaling_group.bastion-privateweave-example-com.id]
}

output "bastion_security_group_ids" {
  value = [aws_security_group.bastion-privateweave-example-com.id]
}

output "bastions_role_arn" {
  value = aws_iam_role.bastions-privateweave-example-com.arn
}

output "bastions_role_name" {
  value = aws_iam_role.bastions-privateweave-example-com.name
}

output "cluster_name" {
  value = "privateweave.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-privateweave-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-privateweave-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-privateweave-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-privateweave-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-privateweave-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-privateweave-example-com.id]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-privateweave-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-privateweave-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-privateweave-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_private-us-test-1a_id" {
  value = aws_route_table.private-us-test-1a-privateweave-example-com.id
}

output "route_table_public_id" {
  value = aws_route_table.privateweave-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-privateweave-example-com.id
}

output "subnet_utility-us-test-1a_id" {
  value = aws_subnet.utility-us-test-1a-privateweave-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.privateweave-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.privateweave-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_autoscaling_group" "bastion-privateweave-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.bastion-privateweave-example-com.id
    version = aws_launch_template.bastion-privateweave-example-com.latest_version
  }
  load_balancers        = [aws_elb.bastion-privateweave-example-com.id]
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "bastion.privateweave.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privateweave.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "bastion.privateweave.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"
    propagate_at_launch = true
    value               = "node"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/role/bastion"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "bastion"
  }
  tag {
    key                 = "kubernetes.io/cluster/privateweave.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.utility-us-test-1a-privateweave-example-com.id]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-privateweave-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-privateweave-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-privateweave-example-com.latest_version
  }
  load_balancers        = [aws_elb.api-privateweave-example-com.id]
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "master-us-test-1a.masters.privateweave.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privateweave.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.privateweave.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"
    propagate_at_launch = true
    value               = "master"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/role/master"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "master-us-test-1a"
  }
  tag {
    key                 = "kubernetes.io/cluster/privateweave.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-privateweave-example-com.id]
}

resource "aws_autoscaling_group" "nodes-privateweave-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-privateweave-example-com.id
    version = aws_launch_template.nodes-privateweave-example-com.latest_version
  }
  max_size              = 2
  metrics_granularity   = "1Minute"
  min_size              = 2
  name                  = "nodes.privateweave.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privateweave.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.privateweave.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"
    propagate_at_launch = true
    value               = "node"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/role/node"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "nodes"
  }
  tag {
    key                 = "kubernetes.io/cluster/privateweave.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-privateweave-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-privateweave-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "us-test-1a.etcd-events.privateweave.example.com"
    "k8s.io/etcd/events"                             = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-privateweave-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "us-test-1a.etcd-main.privateweave.example.com"
    "k8s.io/etcd/main"                               = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_eip" "us-test-1a-privateweave-example-com" {
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "us-test-1a.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc = true
}

resource "aws_elb" "api-privateweave-example-com" {
  connection_draining         = true
  connection_draining_timeout = 300
  cross_zone_load_balancing   = false
  health_check {
    healthy_threshold   = 2
    interval            = 10
    target              = "SSL:443"
    timeout             = 5
    unhealthy_threshold = 2
  }
  idle_timeout = 300
  listener {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }
  name            = "api-privateweave-example--l94cb4"
  security_groups = [aws_security_group.api-elb-privateweave-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-privateweave-example-com.id]
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "api.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_elb" "bastion-privateweave-example-com" {
  health_check {
    healthy_threshold   = 2
    interval            = 10
    target              = "TCP:22"
    timeout             = 5
    unhealthy_threshold = 2
  }
  idle_timeout = 300
  listener {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }
  name            = "bastion-privateweave-exam-fdb6ge"
  security_groups = [aws_security_group.bastion-elb-privateweave-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-privateweave-example-com.id]
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "bastion.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "bastions-privateweave-example-com" {
  name = "bastions.privateweave.example.com"
  role = aws_iam_role.bastions-privateweave-example-com.name
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "bastions.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-privateweave-example-com" {
  name = "masters.privateweave.example.com"
  role = aws_iam_role.masters-privateweave-example-com.name
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "masters.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "nodes-privateweave-example-com" {
  name = "nodes.privateweave.example.com"
  role = aws_iam_role.nodes-privateweave-example-com.name
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "nodes.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_role" "bastions-privateweave-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_bastions.privateweave.example.com_policy")
  name               = "bastions.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "bastions.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_role" "masters-privateweave-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.privateweave.example.com_policy")
  name               = "masters.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "masters.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_role" "nodes-privateweave-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.privateweave.example.com_policy")
  name               = "nodes.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "nodes.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_iam_role_policy" "bastions-privateweave-example-com" {
  name   = "bastions.privateweave.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_bastions.privateweave.example.com_policy")
  role   = aws_iam_role.bastions-privateweave-example-com.name
}

resource "aws_iam_role_policy" "masters-privateweave-example-com" {
  name   = "masters.privateweave.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.privateweave.example.com_policy")
  role   = aws_iam_role.masters-privateweave-example-com.name
}

resource "aws_iam_role_policy" "nodes-privateweave-example-com" {
  name   = "nodes.privateweave.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.privateweave.example.com_policy")
  role   = aws_iam_role.nodes-privateweave-example-com.name
}

resource "aws_internet_gateway" "privateweave-example-com" {
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_key_pair" "kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.privateweave.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.privateweave.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_launch_template" "bastion-privateweave-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 32
      volume_type           = "gp3"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.bastions-privateweave-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.micro"
  key_name      = aws_key_pair.kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "bastion.privateweave.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.bastion-privateweave-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "privateweave.example.com"
      "Name"                                                                       = "bastion.privateweave.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/bastion"                                                        = "1"
      "kops.k8s.io/instancegroup"                                                  = "bastion"
      "kubernetes.io/cluster/privateweave.example.com"                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "privateweave.example.com"
      "Name"                                                                       = "bastion.privateweave.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/bastion"                                                        = "1"
      "kops.k8s.io/instancegroup"                                                  = "bastion"
      "kubernetes.io/cluster/privateweave.example.com"                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "privateweave.example.com"
    "Name"                                                                       = "bastion.privateweave.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/bastion"                                                        = "1"
    "kops.k8s.io/instancegroup"                                                  = "bastion"
    "kubernetes.io/cluster/privateweave.example.com"                             = "owned"
  }
}

resource "aws_launch_template" "master-us-test-1a-masters-privateweave-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 64
      volume_type           = "gp3"
    }
  }
  block_device_mappings {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.masters-privateweave-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  key_name      = aws_key_pair.kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "master-us-test-1a.masters.privateweave.example.com"
  network_interfaces {
    associate_public_ip_address = false
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.masters-privateweave-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                                                     = "privateweave.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.privateweave.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"                                      = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master"                          = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/privateweave.example.com"                                                        = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                                                     = "privateweave.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.privateweave.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"                                      = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master"                          = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/privateweave.example.com"                                                        = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                                                     = "privateweave.example.com"
    "Name"                                                                                                  = "master-us-test-1a.masters.privateweave.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"                                      = "master"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master"                          = ""
    "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
    "k8s.io/role/master"                                                                                    = "1"
    "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
    "kubernetes.io/cluster/privateweave.example.com"                                                        = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.privateweave.example.com_user_data")
}

resource "aws_launch_template" "nodes-privateweave-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 128
      volume_type           = "gp3"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-privateweave-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
  key_name      = aws_key_pair.kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "nodes.privateweave.example.com"
  network_interfaces {
    associate_public_ip_address = false
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-privateweave-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "privateweave.example.com"
      "Name"                                                                       = "nodes.privateweave.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/privateweave.example.com"                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "privateweave.example.com"
      "Name"                                                                       = "nodes.privateweave.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/privateweave.example.com"                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "privateweave.example.com"
    "Name"                                                                       = "nodes.privateweave.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/privateweave.example.com"                             = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.privateweave.example.com_user_data")
}

resource "aws_nat_gateway" "us-test-1a-privateweave-example-com" {
  allocation_id = aws_eip.us-test-1a-privateweave-example-com.id
  subnet_id     = aws_subnet.utility-us-test-1a-privateweave-example-com.id
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "us-test-1a.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.privateweave-example-com.id
  route_table_id         = aws_route_table.privateweave-example-com.id
}

resource "aws_route" "route-__--0" {
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.privateweave-example-com.id
  route_table_id              = aws_route_table.privateweave-example-com.id
}

resource "aws_route" "route-private-us-test-1a-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.us-test-1a-privateweave-example-com.id
  route_table_id         = aws_route_table.private-us-test-1a-privateweave-example-com.id
}

resource "aws_route53_record" "api-privateweave-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-privateweave-example-com.dns_name
    zone_id                = aws_elb.api-privateweave-example-com.zone_id
  }
  name    = "api.privateweave.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table" "private-us-test-1a-privateweave-example-com" {
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "private-us-test-1a.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
    "kubernetes.io/kops/role"                        = "private-us-test-1a"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_route_table" "privateweave-example-com" {
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
    "kubernetes.io/kops/role"                        = "public"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_route_table_association" "private-us-test-1a-privateweave-example-com" {
  route_table_id = aws_route_table.private-us-test-1a-privateweave-example-com.id
  subnet_id      = aws_subnet.us-test-1a-privateweave-example-com.id
}

resource "aws_route_table_association" "utility-us-test-1a-privateweave-example-com" {
  route_table_id = aws_route_table.privateweave-example-com.id
  subnet_id      = aws_subnet.utility-us-test-1a-privateweave-example-com.id
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "clusters.example.com/privateweave.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "clusters.example.com/privateweave.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "clusters.example.com/privateweave.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "clusters.example.com/privateweave.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events_content")
  key                    = "clusters.example.com/privateweave.example.com/manifests/etcd/events.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main_content")
  key                    = "clusters.example.com/privateweave.example.com/manifests/etcd/main.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "clusters.example.com/privateweave.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test-1a_content")
  key                    = "clusters.example.com/privateweave.example.com/igconfig/master/master-us-test-1a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "clusters.example.com/privateweave.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-bootstrap_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-core-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-core.addons.k8s.io_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/core.addons.k8s.io/v1.4.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-networking-weave-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-networking.weave-k8s-1.12_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/networking.weave/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateweave-example-com-addons-storage-aws-addons-k8s-io-v1-15-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateweave.example.com-addons-storage-aws.addons.k8s.io-v1.15.0_content")
  key                    = "clusters.example.com/privateweave.example.com/addons/storage-aws.addons.k8s.io/v1.15.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_security_group" "api-elb-privateweave-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "api-elb.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_security_group" "bastion-elb-privateweave-example-com" {
  description = "Security group for bastion ELB"
  name        = "bastion-elb.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "bastion-elb.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_security_group" "bastion-privateweave-example-com" {
  description = "Security group for bastion"
  name        = "bastion.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "bastion.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_security_group" "masters-privateweave-example-com" {
  description = "Security group for masters"
  name        = "masters.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "masters.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_security_group" "nodes-privateweave-example-com" {
  description = "Security group for nodes"
  name        = "nodes.privateweave.example.com"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "nodes.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-bastion-elb-privateweave-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.bastion-elb-privateweave-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-443to443-api-elb-privateweave-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-privateweave-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "from-api-elb-privateweave-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-api-elb-privateweave-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privateweave-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privateweave-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privateweave-example-com-ingress-tcp-22to22-bastion-privateweave-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.bastion-privateweave-example-com.id
  source_security_group_id = aws_security_group.bastion-elb-privateweave-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-bastion-privateweave-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-privateweave-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-privateweave-example-com-ingress-tcp-22to22-masters-privateweave-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.bastion-privateweave-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-bastion-privateweave-example-com-ingress-tcp-22to22-nodes-privateweave-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.nodes-privateweave-example-com.id
  source_security_group_id = aws_security_group.bastion-privateweave-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-privateweave-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-privateweave-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.masters-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-privateweave-example-com-ingress-all-0to0-masters-privateweave-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.masters-privateweave-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-privateweave-example-com-ingress-all-0to0-nodes-privateweave-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-privateweave-example-com.id
  source_security_group_id = aws_security_group.masters-privateweave-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-ingress-all-0to0-nodes-privateweave-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-privateweave-example-com.id
  source_security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-ingress-tcp-1to2379-masters-privateweave-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-ingress-tcp-2382to4000-masters-privateweave-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-ingress-tcp-4003to65535-masters-privateweave-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateweave-example-com-ingress-udp-1to65535-masters-privateweave-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.nodes-privateweave-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateweave-example-com.id
  source_security_group_id = aws_security_group.api-elb-privateweave-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-privateweave-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_subnet" "us-test-1a-privateweave-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.32.0/19"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "us-test-1a.privateweave.example.com"
    "SubnetType"                                     = "Private"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
    "kubernetes.io/role/internal-elb"                = "1"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_subnet" "utility-us-test-1a-privateweave-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.4.0/22"
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "utility-us-test-1a.privateweave.example.com"
    "SubnetType"                                     = "Utility"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
  vpc_id = aws_vpc.privateweave-example-com.id
}

resource "aws_vpc" "privateweave-example-com" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "172.20.0.0/16"
  enable_dns_hostnames             = true
  enable_dns_support               = true
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "privateweave-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                              = "privateweave.example.com"
    "Name"                                           = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "privateweave-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.privateweave-example-com.id
  vpc_id          = aws_vpc.privateweave-example-com.id
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 4.0.0"
    }
  }
}
