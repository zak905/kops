/*
Copyright 2020 The Kubernetes Authors.

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

package model

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/distributions"
)

func TestKubectlBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/golden/minimal", "kubectl", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		nodeupModelContext.Assets = fi.NewAssetStore("")
		nodeupModelContext.Assets.AddForTest("kubectl", "/path/to/kubectl/asset", "testing kubectl content")
		// NodeUp looks for the default user and group on the machine running the tests.
		// Flatcar is unlikely to be used for such task, so tests results should be consistent.
		nodeupModelContext.Distribution = distributions.DistributionFlatcar
		builder := KubectlBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}
