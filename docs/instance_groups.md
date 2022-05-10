# The `InstanceGroup` resource

The `InstanceGroup` resource represents a group of similar machines typically provisioned in the same availability zone. On AWS, instance groups map directly to an autoscaling group.

The complete list of keys can be found at the [InstanceGroup](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#InstanceGroupSpec) reference page. 

You can also find concrete use cases for the configurations on the [Instance Group operations page](instance_groups.md)

On this page, we will expand on the more important configuration keys.

## cloudLabels

If you need to add tags on auto scaling groups or instances (propagate ASG tags), you can add it in the instance group specs with `cloudLabels`. Cloud Labels defined at the cluster spec level will also be inherited.

```YAML
spec:
  cloudLabels:
    billing: infra
    environment: dev
```

## suspendProcess

Autoscaling groups automatically include multiple [scaling processes](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-suspend-resume-processes.html#process-types)
that keep our ASGs healthy.  In some cases, you may want to disable certain scaling activities.

An example of this is if you are running multiple AZs in an ASG while using a Kubernetes Autoscaler.
The autoscaler will remove specific instances that are not being used. In some cases, the `AZRebalance` process
will rescale the ASG without warning.

```YAML
spec:
  suspendProcesses:
  - AZRebalance
```


## instanceProtection

Autoscaling groups may scale up or down automatically to balance types of instances, regions, etc.
[Instance protection](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-instance-termination.html#instance-protection) prevents the ASG from being scaled in.

```YAML
spec:
  instanceProtection: true
```

## instanceMetadata

By default IMDSv2 are enabled as of kOps 1.22 on new clusters using Kubernetes 1.22. The default hop limit is 3 on control plane nodes, and 1 on other roles.

On other versions, you can enable IMDSv2 like this:

```YAML
spec:
  instanceMetadata:
    httpPutResponseHopLimit: 1
    httpTokens: required
```

## externalLoadBalancers

Instance groups can be linked to up to 10 load balancers. When attached, any instance launched will
automatically register itself to the load balancer. For example, if you can create an instance group
dedicated to running an ingress controller exposed on a
[NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport), you can
manually create a load balancer and link it to the instance group. Traffic to the load balancer will now
automatically go to one of the nodes.

You can specify either `loadBalancerName` to link the instance group to an AWS Classic ELB or you can
specify `targetGroupArn` to link the instance group to a target group, which are used by Application
load balancers and Network load balancers.

```YAML
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: ingress
spec:
  machineType: m4.large
  maxSize: 2
  minSize: 2
  role: Node
  externalLoadBalancers:
  - targetGroupArn: arn:aws:elasticloadbalancing:eu-west-1:123456789012:targetgroup/my-ingress-target-group/0123456789abcdef
  - loadBalancerName: my-elb-classic-load-balancer
```

## detailedInstanceMonitoring

Detailed monitoring will cause the monitoring data to be available every 1 minute instead of every 5 minutes. [Enabling Detailed Monitoring](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch-new.html). In production environments you may want to consider to enable detailed monitoring for quicker troubleshooting.

**Note: that enabling detailed monitoring is a subject for [charge](https://aws.amazon.com/cloudwatch)**

```YAML
spec:
  detailedInstanceMonitoring: true
```

## additionalUserData

kOps utilizes cloud-init to initialize and setup a host at boot time. However in certain cases you may already be leveraging certain features of cloud-init in your infrastructure and would like to continue doing so. More information on cloud-init can be found [here](http://cloudinit.readthedocs.io/en/latest/). 

Additional user-data can be passed to the host provisioning by setting the `additionalUserData` field. A list of valid user-data content-types can be found [here](http://cloudinit.readthedocs.io/en/latest/topics/format.html#mime-multi-part-archive).

Scripts will be run in alphabetical order as documented [here](https://cloudinit.readthedocs.io/en/latest/topics/modules.html#scripts-per-instance).

Note: Passing additionalUserData in Flatcar-OS is not supported, it results in node not coming up.

Example:

```YAML
spec:
  additionalUserData:
  - name: myscript.sh
    type: text/x-shellscript
    content: |
      #!/bin/sh
      echo "Hello World.  The time is now $(date -R)!" | tee /root/output.txt
  - name: local_repo.txt
    type: text/cloud-config
    content: |
      #cloud-config
      apt:
        primary:
          - arches: [default]
            uri: http://local-mirror.mydomain
            search:
              - http://local-mirror.mydomain
              - http://archive.ubuntu.com
```

## compressUserData
{{ kops_feature_table(kops_added_default='1.19') }}

Compresses parts of the user-data to save space and help with the size limit 
in certain clouds. Currently only the Specs in nodeup.sh will be compressed.

```YAML
spec:
  compressUserData: true
```

## packages
{{ kops_feature_table(kops_added_default='1.24') }}

To install additional packages to hosts in the instance group, specify the `packages` field as an array of strings.

Package names are distro specific and are not validated in any way. Specifying incorrect package names may prevent nodes from starting.

For example:

```YAML
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: nodes
spec:
  packages:
  - nfs-common
```

## sysctlParameters
{{ kops_feature_table(kops_added_default='1.17') }}

To add custom kernel runtime parameters to your instance group, specify the
`sysctlParameters` field as an array of strings. Each string must take the form
of `variable=value` the way it would appear in sysctl.conf (see also
`sysctl(8)` manpage).

Unlike a simple file asset, specifying kernel runtime parameters in this manner
would correctly invoke `sysctl --system` automatically for you to apply said
parameters.

For example:

```YAML
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: nodes
spec:
  sysctlParameters:
  - fs.pipe-user-pages-soft=524288
  - net.ipv4.tcp_keepalive_time=200
```

which would end up in a drop-in file on nodes of the instance group in question.

## mixedInstancesPolicy (AWS Only)

A Mixed Instances Policy utilizing EC2 Spot and the `capacity-optimized` allocation strategy allows an EC2 Autoscaling Group to select the instance types with the highest capacity. This reduces the chance of a spot interruption on your instance group. 

Instance groups with a mixedInstancesPolicy can be generated with the `kops toolbox instance-selector` command. 
The instance-selector accepts user supplied resource parameters like vcpus, memory, and much more to dynamically select instance types that match your criteria. 

```bash
kops toolbox instance-selector --vcpus 4 --flexible --usage-class spot --instance-group-name spotgroup
```

```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: spot.k8s.local
  name: spotgroup
spec:
  image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20200528
  machineType: c3.xlarge
  maxSize: 15
  minSize: 2
  mixedInstancesPolicy:
    instances:
    - c3.xlarge
    - c4.xlarge
    - c5.xlarge
    - c5a.xlarge
    onDemandAboveBase: 0
    onDemandBase: 0
    spotAllocationStrategy: capacity-optimized
  nodeLabels:
    kops.k8s.io/instancegroup: spotgroup
  role: Node
  subnets:
  - us-east-1a
  - us-east-1b
  - us-east-1c
```

### Instances

Instances is a list of instance types which we are willing to run in the EC2 Auto Scaling group.

### onDemandAllocationStrategy

Indicates how to allocate instance types to fulfill On-Demand capacity

### onDemandBase

OnDemandBase is the minimum amount of the Auto Scaling group's capacity that must be
fulfilled by On-Demand Instances. This base portion is provisioned first as your group scales.

### onDemandAboveBase

OnDemandAboveBase controls the percentages of On-Demand Instances and Spot Instances for your
additional capacity beyond OnDemandBase. The range is 0–100. The default value is 100. If you
leave this parameter set to 100, the percentages are 100% for On-Demand Instances and 0% for
Spot Instances.

### spotAllocationStrategy
SpotAllocationStrategy Indicates how to allocate instances across Spot Instance pools.

If the allocation strategy is lowest-price, the Auto Scaling group launches instances using the Spot pools with the lowest price, and evenly allocates your instances across the number of Spot pools that you specify in spotInstancePools. If the allocation strategy is [capacity-optimized](https://aws.amazon.com/blogs/compute/introducing-the-capacity-optimized-allocation-strategy-for-amazon-ec2-spot-instances/), the Auto Scaling group launches instances using Spot pools that are optimally chosen based on the available Spot capacity.
https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_InstancesDistribution.html

### spotInstancePools
Used only when the Spot allocation strategy is lowest-price.
The number of Spot Instance pools across which to allocate your Spot Instances. The Spot pools are determined from the different instance types in the Overrides array of LaunchTemplate. Default if not set is 2.

### instanceRequirements

{{ kops_feature_table(kops_added_default='1.24') }}

Instead of configuring specific machine types, the InstanceGroup can be configured to use all machine types that satisfy a given set of requirements.

```
spec:
  mixedInstancesPolicy:
    instanceRquirements:
      cpu:
        min: "2"
        max: "16"
      memory:
        min: "2G"
```

Note that burstable instances are always included in the set of eligible instances.

## warmPool (AWS Only)

{{ kops_feature_table(kops_added_default='1.21') }}

A Warm Pool contains pre-initialized EC2 instances that can join the cluster significantly faster than regular instances. These instances run the kOps configuration process, pull known container images, and then shut down. When the ASG needs to scale out it will pull instances from the warm pool if any are available.

You can enable the warm pool by adding the following:

```yaml
spec:
  warmPool: {}
```

This will use the AWS default settings. You can change the pool size like this:

```yaml
spec:
  warmPool:
    minSize: 3
    maxSize: 10
```

You can also specify defaults for all instance groups of type Node or APIServer by setting the `warmPool` field in the cluster spec.
If warm pools are enabled at the cluster spec level, you can disable them at the instance group level by setting `maxSize: 0`.

### Lifecycle hook

By default AWS does not guarantee that the kOps configuration will run to completion. Nor that the instance will timely shut down after completion if the instance is allowed to run that long. In order to guarantee this, a lifecycle hook is needed.

**You have to ensure your metadata API is protected if you enable this. If not, any Pod in the cluster will be able to complete the lifecycle hook with the `ABANDONED` result, preventing any instance from ever joining the cluster.**

The following config will enable the lifecycle hook as well as protect the metadata API from abuse:

```yaml
spec:
  warmPool:
    enableLifecycleHook: true
  instanceMetadata:
    httpPutResponseHopLimit: 1
    httpTokens: required
```
