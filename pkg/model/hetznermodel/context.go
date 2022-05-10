/*
Copyright 2022 The Kubernetes Authors.

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

package hetznermodel

import (
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

type HetznerModelContext struct {
	*model.KopsModelContext
}

func (b *HetznerModelContext) LinkToNetwork() *hetznertasks.Network {
	name := b.ClusterName()
	return &hetznertasks.Network{Name: &name}
}

func (b *HetznerModelContext) LinkToSSHKey() *hetznertasks.SSHKey {
	name := b.ClusterName()
	return &hetznertasks.SSHKey{Name: &name}
}
