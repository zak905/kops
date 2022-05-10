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
	"fmt"
	"net/url"
	"strings"

	"github.com/blang/semver/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
)

var DefaultChannelBase = "https://raw.githubusercontent.com/kubernetes/kops/master/channels/"

const (
	DefaultChannel = "stable"
)

type Channel struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChannelSpec `json:"spec,omitempty"`
}

type ChannelSpec struct {
	Images []*ChannelImageSpec `json:"images,omitempty"`

	Cluster *ClusterSpec `json:"cluster,omitempty"`

	// KopsVersions allows us to recommend/require kops versions
	KopsVersions []KopsVersionSpec `json:"kopsVersions,omitempty"`

	// KubernetesVersions allows us to recommend/requires kubernetes versions
	KubernetesVersions []KubernetesVersionSpec `json:"kubernetesVersions,omitempty"`

	// Packages specifies the package versions that correspond to this channel.
	Packages []PackageVersionSpec `json:"packages,omitempty"`
}

type KopsVersionSpec struct {
	Range string `json:"range,omitempty"`

	// RecommendedVersion is the recommended version of kops to use for this Range of kops versions
	RecommendedVersion string `json:"recommendedVersion,omitempty"`

	// RequiredVersion is the required version of kops to use for this Range of kops versions, forcing an upgrade
	RequiredVersion string `json:"requiredVersion,omitempty"`

	// KubernetesVersion is the default version of kubernetes to use with this kops version e.g. for new clusters
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
}

type KubernetesVersionSpec struct {
	Range string `json:"range,omitempty"`

	RecommendedVersion string `json:"recommendedVersion,omitempty"`
	RequiredVersion    string `json:"requiredVersion,omitempty"`
}

type ChannelImageSpec struct {
	ProviderID string `json:"providerID,omitempty"`

	ArchitectureID string `json:"architectureID,omitempty"`

	Name string `json:"name,omitempty"`

	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
}

// PackageVersionSpec specifies the version of a package
type PackageVersionSpec struct {
	// Name is the name of the package.
	Name string `json:"name"`

	// Version is the version of the package.
	Version string `json:"version"`

	// KubernetesVersion specifies that this package only applies to a semver range of kubernetes version
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// KopsVersion specifies that this package only applies to a semver range of kOps version
	KopsVersion string `json:"kopsVersion,omitempty"`
}

// ResolveChannel maps a channel to an absolute URL (possibly a VFS URL)
// If the channel is the well-known "none" value, we return (nil, nil)
func ResolveChannel(location string) (*url.URL, error) {
	if location == "none" {
		return nil, nil
	}

	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("invalid channel location: %q", location)
	}

	if !u.IsAbs() {
		base, err := url.Parse(DefaultChannelBase)
		if err != nil {
			return nil, fmt.Errorf("invalid base channel location: %q", DefaultChannelBase)
		}
		klog.V(4).Infof("resolving %q against default channel location %q", location, DefaultChannelBase)
		u = base.ResolveReference(u)
	}

	return u, nil
}

// LoadChannel loads a Channel object from the specified VFS location
func LoadChannel(location string) (*Channel, error) {
	resolvedURL, err := ResolveChannel(location)
	if err != nil {
		return nil, err
	}

	if resolvedURL == nil {
		return &Channel{}, nil
	}

	resolved := resolvedURL.String()

	klog.V(2).Infof("Loading channel from %q", resolved)
	channelBytes, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading channel %q: %v", resolved, err)
	}
	channel, err := ParseChannel(channelBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %q: %v", resolved, err)
	}
	klog.V(4).Infof("Channel contents: %s", string(channelBytes))

	return channel, nil
}

// ParseChannel parses a Channel object
func ParseChannel(channelBytes []byte) (*Channel, error) {
	channel := &Channel{}
	err := ParseRawYaml(channelBytes, channel)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %v", err)
	}

	return channel, nil
}

// FindRecommendedUpgrade returns a string with a new version, if the current version is out of date
func (v *KubernetesVersionSpec) FindRecommendedUpgrade(version semver.Version) (*semver.Version, error) {
	if v.RecommendedVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RecommendedVersion")
		return nil, nil
	}

	recommendedVersion, err := util.ParseKubernetesVersion(v.RecommendedVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing RecommendedVersion %q from channel", v.RecommendedVersion)
	}
	if recommendedVersion.GT(version) {
		klog.V(2).Infof("RecommendedVersion=%q, Have=%q.  Recommending upgrade", recommendedVersion, version)
		return recommendedVersion, nil
	}
	klog.V(4).Infof("RecommendedVersion=%q, Have=%q.  No upgrade needed.", recommendedVersion, version)
	return nil, nil
}

// FindRecommendedUpgrade returns a string with a new version, if the current version is out of date
func (v *KopsVersionSpec) FindRecommendedUpgrade(version semver.Version) (*semver.Version, error) {
	if v.RecommendedVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RecommendedVersion")
		return nil, nil
	}

	recommendedVersion, err := semver.ParseTolerant(v.RecommendedVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing RecommendedVersion %q from channel", v.RecommendedVersion)
	}
	if recommendedVersion.GT(version) {
		klog.V(2).Infof("RecommendedVersion=%q, Have=%q.  Recommending upgrade", recommendedVersion, version)
		return &recommendedVersion, nil
	}
	klog.V(4).Infof("RecommendedVersion=%q, Have=%q.  No upgrade needed.", recommendedVersion, version)
	return nil, nil
}

// IsUpgradeRequired returns true if the current version is not acceptable
func (v *KubernetesVersionSpec) IsUpgradeRequired(version semver.Version) (bool, error) {
	if v.RequiredVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RequiredVersion")
		return false, nil
	}

	requiredVersion, err := util.ParseKubernetesVersion(v.RequiredVersion)
	if err != nil {
		return false, fmt.Errorf("error parsing RequiredVersion %q from channel", v.RequiredVersion)
	}
	if requiredVersion.GT(version) {
		klog.V(2).Infof("RequiredVersion=%q, Have=%q.  Requiring upgrade", requiredVersion, version)
		return true, nil
	}
	klog.V(4).Infof("RequiredVersion=%q, Have=%q.  No upgrade needed.", requiredVersion, version)
	return false, nil
}

// IsUpgradeRequired returns true if the current version is not acceptable
func (v *KopsVersionSpec) IsUpgradeRequired(version semver.Version) (bool, error) {
	if v.RequiredVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RequiredVersion")
		return false, nil
	}

	requiredVersion, err := semver.ParseTolerant(v.RequiredVersion)
	if err != nil {
		return false, fmt.Errorf("error parsing RequiredVersion %q from channel", v.RequiredVersion)
	}
	if requiredVersion.GT(version) {
		klog.V(2).Infof("RequiredVersion=%q, Have=%q.  Requiring upgrade", requiredVersion, version)
		return true, nil
	}
	klog.V(4).Infof("RequiredVersion=%q, Have=%q.  No upgrade needed.", requiredVersion, version)
	return false, nil
}

// FindKubernetesVersionSpec returns a KubernetesVersionSpec for the current version
func FindKubernetesVersionSpec(versions []KubernetesVersionSpec, version semver.Version) *KubernetesVersionSpec {
	for i := range versions {
		v := &versions[i]
		if v.Range != "" {
			versionRange, err := semver.ParseRange(v.Range)
			if err != nil {
				klog.Warningf("unable to parse range in channel version spec: %q", v.Range)
				continue
			}
			if !versionRange(version) {
				klog.V(8).Infof("version range %q does not apply to version %q; skipping", v.Range, version)
				continue
			}
		}
		return v
	}

	return nil
}

// FindKopsVersionSpec returns a KopsVersionSpec for the current version
func FindKopsVersionSpec(versions []KopsVersionSpec, version semver.Version) *KopsVersionSpec {
	for i := range versions {
		v := &versions[i]
		if v.Range != "" {
			versionRange, err := semver.ParseRange(v.Range)
			if err != nil {
				klog.Warningf("unable to parse range in channel version spec: %q", v.Range)
				continue
			}
			if !versionRange(version) {
				klog.V(8).Infof("version range %q does not apply to version %q; skipping", v.Range, version)
				continue
			}
		}
		return v
	}

	return nil
}

type CloudProviderID string

const (
	CloudProviderAWS       CloudProviderID = "aws"
	CloudProviderDO        CloudProviderID = "digitalocean"
	CloudProviderGCE       CloudProviderID = "gce"
	CloudProviderHetzner   CloudProviderID = "hetzner"
	CloudProviderOpenstack CloudProviderID = "openstack"
	CloudProviderAzure     CloudProviderID = "azure"
)

// FindImage returns the image for the cloudprovider, or nil if none found
func (c *Channel) FindImage(provider CloudProviderID, kubernetesVersion semver.Version, architecture architectures.Architecture) *ChannelImageSpec {
	var matches []*ChannelImageSpec

	for _, image := range c.Spec.Images {
		if image.ProviderID != string(provider) {
			continue
		}
		if image.ArchitectureID != "" && image.ArchitectureID != string(architecture) {
			continue
		}
		if image.KubernetesVersion != "" {
			versionRange, err := semver.ParseRange(image.KubernetesVersion)
			if err != nil {
				klog.Warningf("cannot parse KubernetesVersion=%q", image.KubernetesVersion)
				continue
			}

			if !versionRange(kubernetesVersion) {
				klog.V(2).Infof("Kubernetes version %q does not match range: %s", kubernetesVersion, image.KubernetesVersion)
				continue
			}
		}
		matches = append(matches, image)
	}

	if len(matches) == 0 {
		klog.V(2).Infof("No matching images in channel for cloudprovider %q", provider)
		return nil
	}

	if len(matches) != 1 {
		klog.Warningf("Multiple matching images in channel for cloudprovider %q", provider)
	}
	return matches[0]
}

// RecommendedKubernetesVersion returns the recommended kubernetes version for a version of kops
// It is used by default when creating a new cluster, for example
func RecommendedKubernetesVersion(c *Channel, kopsVersionString string) *semver.Version {
	kopsVersion, err := semver.ParseTolerant(kopsVersionString)
	if err != nil {
		klog.Warningf("unable to parse kops version %q", kopsVersionString)
	} else {
		kopsVersionSpec := FindKopsVersionSpec(c.Spec.KopsVersions, kopsVersion)
		if kopsVersionSpec != nil {
			if kopsVersionSpec.KubernetesVersion != "" {
				sv, err := util.ParseKubernetesVersion(kopsVersionSpec.KubernetesVersion)
				if err != nil {
					klog.Warningf("unable to parse kubernetes version %q", kopsVersionSpec.KubernetesVersion)
				} else {
					return sv
				}
			}
		}
	}

	if c.Spec.Cluster != nil {
		sv, err := util.ParseKubernetesVersion(c.Spec.Cluster.KubernetesVersion)
		if err != nil {
			klog.Warningf("unable to parse kubernetes version %q", c.Spec.Cluster.KubernetesVersion)
		} else {
			return sv
		}
	}

	return nil
}

// Returns true if the given image name has the stable or alpha channel images prefix. Otherwise false.
func (c *Channel) HasUpstreamImagePrefix(image string) bool {
	return strings.HasPrefix(image, "kope.io/k8s-") ||
		strings.HasPrefix(image, "099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-") ||
		strings.HasPrefix(image, "cos-cloud/cos-stable-") ||
		strings.HasPrefix(image, "ubuntu-os-cloud/ubuntu-2004-focal-") ||
		strings.HasPrefix(image, "Canonical:0001-com-ubuntu-server-focal:20_04-lts-gen2:")
}

// GetPackageVersion returns the version for the package, or an error if could not be found.
func (c *Channel) GetPackageVersion(name string, kubernetesVersion *semver.Version) (*util.Version, error) {
	var matches []*PackageVersionSpec

	for i := range c.Spec.Packages {
		pkg := &c.Spec.Packages[i]
		if pkg.Name != name {
			continue
		}

		if pkg.KubernetesVersion != "" {
			versionRange, err := semver.ParseRange(pkg.KubernetesVersion)
			if err != nil {
				klog.Warningf("cannot parse KubernetesVersion=%q", pkg.KubernetesVersion)
				continue
			}

			if !versionRange(*kubernetesVersion) {
				klog.V(2).Infof("Kubernetes version %q does not match range: %s", kubernetesVersion, pkg.KubernetesVersion)
				continue
			}
		}

		if pkg.KopsVersion != "" {
			kopsVersion, err := util.ParseVersion(kopsbase.KOPS_RELEASE_VERSION)
			if err != nil {
				return nil, fmt.Errorf("parsing kops version %q: %w", kopsbase.KOPS_RELEASE_VERSION, err)
			}

			versionRange, err := semver.ParseRange(pkg.KopsVersion)
			if err != nil {
				klog.Warningf("cannot parse KopsVersion=%q", pkg.KopsVersion)
				continue
			}

			if !kopsVersion.IsInRange(versionRange) {
				klog.V(2).Infof("kOps version %q does not match range: %s", kopsVersion, pkg.KopsVersion)
				continue
			}
		}

		matches = append(matches, pkg)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("found no packages in channel for name=%q", name)
	}

	if len(matches) != 1 {
		return nil, fmt.Errorf("found multiple packages in channel for name=%q", name)
	}
	v, err := util.ParseVersion(matches[0].Version)
	if err != nil {
		return nil, fmt.Errorf("error parsing version %q for package %q", matches[0].Version, name)
	}
	return v, nil
}
