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

package nodetasks

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"sync"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/hashing"
)

type Package struct {
	Name string

	Version      *string `json:"version,omitempty"`
	Source       *string `json:"source,omitempty"`
	Hash         *string `json:"hash,omitempty"`
	PreventStart *bool   `json:"preventStart,omitempty"`

	// Healthy is true if the package installation did not fail
	Healthy *bool `json:"healthy,omitempty"`

	// Additional dependencies that must be installed before this package.
	// These will actually be passed together with this package to rpm/dpkg,
	// which will then figure out the correct order in which to install them.
	// This means that Deps don't get installed unless this package needs to
	// get installed.
	Deps []*Package `json:"deps,omitempty"`
}

const (
	localPackageDir             = "/var/cache/nodeup/packages/"
	containerSelinuxPackageName = "container-selinux"
	containerdPackageName       = "containerd.io"
	dockerPackageName           = "docker-ce"
)

var _ fi.HasDependencies = &Package{}

// GetDependencies computes dependencies for the package task
func (e *Package) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	// UpdatePackages before we install any packages
	for _, v := range tasks {
		if _, ok := v.(*UpdatePackages); ok {
			deps = append(deps, v)
		}
	}

	// If this package is a bare deb, install it after OS managed packages
	if !e.isOSPackage() {
		for _, v := range tasks {
			if vp, ok := v.(*Package); ok {
				if vp.isOSPackage() {
					deps = append(deps, v)
				}
			}
		}
	}

	// containerd should wait for container-selinux to be installed
	if e.Name == containerdPackageName {
		for _, v := range tasks {
			if vp, ok := v.(*Package); ok {
				if vp.Name == containerSelinuxPackageName {
					deps = append(deps, v)
				}
			}
		}
	}

	// Docker should wait for container-selinux and containerd to be installed
	if e.Name == dockerPackageName {
		for _, v := range tasks {
			if vp, ok := v.(*Package); ok {
				if vp.Name == containerSelinuxPackageName {
					deps = append(deps, v)
				}
				if vp.Name == containerdPackageName {
					deps = append(deps, v)
				}
			}
		}
	}

	return deps
}

var _ fi.HasName = &Package{}

func (f *Package) GetName() *string {
	return &f.Name
}

// isOSPackage returns true if this is an OS provided package (as opposed to a bare .deb, for example)
func (p *Package) isOSPackage() bool {
	return fi.StringValue(p.Source) == ""
}

// String returns a string representation, implementing the Stringer interface
func (p *Package) String() string {
	return fmt.Sprintf("Package: %s", p.Name)
}

func (e *Package) Find(c *fi.Context) (*Package, error) {
	d, err := distributions.FindDistribution("/")
	if err != nil {
		return nil, fmt.Errorf("unknown or unsupported distro: %v", err)
	}

	if d.IsDebianFamily() {
		return e.findDpkg(c)
	}

	if d.IsRHELFamily() {
		return e.findYum(c)
	}

	return nil, fmt.Errorf("unsupported package system")
}

func (e *Package) findDpkg(c *fi.Context) (*Package, error) {
	args := []string{"dpkg-query", "-f", "${db:Status-Abbrev}${Version}\\n", "-W", e.Name}
	human := strings.Join(args, " ")

	klog.V(2).Infof("Listing installed packages: %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "no packages found") {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing installed packages: %v: %s", err, string(output))
	}

	installed := false
	var healthy *bool
	installedVersion := ""
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		tokens := strings.Split(line, " ")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("error parsing dpkg-query line %q", line)
		}
		state := tokens[0]
		version := tokens[1]

		switch state {
		case "ii":
			installed = true
			installedVersion = version
			healthy = fi.Bool(true)
		case "iF", "iU":
			installed = true
			installedVersion = version
			healthy = fi.Bool(false)
		case "rc":
			// removed
			installed = false
		case "un":
			// unknown
			installed = false
		case "n":
			// not installed
			installed = false
		default:
			klog.Warningf("unknown package state %q for %q in line %q", state, e.Name, line)
			return nil, fmt.Errorf("unknown package state %q for %q in line %q", state, e.Name, line)
		}
	}

	// TODO: Take InstanceGroup-level overriding of the Cluster-level update policy into account
	// here. Doing so requires that we make the current InstanceGroup available within Package's
	// methods.
	if fi.StringValue(c.Cluster.Spec.UpdatePolicy) != kops.UpdatePolicyExternal || !installed {
		return nil, nil
	}

	return &Package{
		Name:    e.Name,
		Version: fi.String(installedVersion),
		Healthy: healthy,
	}, nil
}

func (e *Package) findYum(c *fi.Context) (*Package, error) {
	args := []string{"/usr/bin/rpm", "-q", e.Name, "--queryformat", "%{NAME} %{VERSION}"}
	human := strings.Join(args, " ")

	klog.V(2).Infof("Listing installed packages: %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "is not installed") {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing installed packages: %v: %s", err, string(output))
	}

	installed := false
	var healthy *bool
	installedVersion := ""
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		tokens := strings.Split(line, " ")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("error parsing rpm line %q", line)
		}

		name := tokens[0]
		if name != e.Name {
			return nil, fmt.Errorf("error parsing rpm line %q", line)
		}
		installed = true
		installedVersion = tokens[1]
		// If we implement unhealthy; be sure to implement repair in Render
		healthy = fi.Bool(true)
	}

	// TODO: Take InstanceGroup-level overriding of the Cluster-level update policy into account
	// here. Doing so requires that we make the current InstanceGroup available within Package's
	// methods.
	if fi.StringValue(c.Cluster.Spec.UpdatePolicy) != kops.UpdatePolicyExternal || !installed {
		return nil, nil
	}

	return &Package{
		Name:    e.Name,
		Version: fi.String(installedVersion),
		Healthy: healthy,
	}, nil
}

func (e *Package) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Package) CheckChanges(a, e, changes *Package) error {
	return nil
}

// packageManagerLock is a simple lock that prevents concurrent package manager operations
// It just avoids unnecessary failures from running e.g. concurrent apt-get installs
var packageManagerLock sync.Mutex

func (_ *Package) RenderLocal(t *local.LocalTarget, a, e, changes *Package) error {
	packageManagerLock.Lock()
	defer packageManagerLock.Unlock()

	d, err := distributions.FindDistribution("/")
	if err != nil {
		return fmt.Errorf("unknown or unsupported distro: %v", err)
	}

	if a == nil || changes.Version != nil {
		klog.Infof("Installing package %q (dependencies: %v)", e.Name, e.Deps)
		var pkgs []string

		if e.Source != nil {
			// Install a deb or rpm.
			err := os.MkdirAll(localPackageDir, 0o755)
			if err != nil {
				return fmt.Errorf("error creating directories %q: %v", localPackageDir, err)
			}

			// Append file extension for local files
			var ext string
			if d.IsDebianFamily() {
				ext = ".deb"
			} else if d.IsRHELFamily() {
				ext = ".rpm"
			} else {
				return fmt.Errorf("unsupported package system")
			}

			// Download all the debs/rpms.
			pkgs = make([]string, 1+len(e.Deps))
			for i, pkg := range append([]*Package{e}, e.Deps...) {
				local := path.Join(localPackageDir, pkg.Name+ext)
				pkgs[i] = local
				var hash *hashing.Hash
				if fi.StringValue(pkg.Hash) != "" {
					parsed, err := hashing.FromString(fi.StringValue(pkg.Hash))
					if err != nil {
						return fmt.Errorf("error parsing hash: %v", err)
					}
					hash = parsed
				}
				_, err = fi.DownloadURL(fi.StringValue(pkg.Source), local, hash)
				if err != nil {
					return err
				}
			}
		} else {
			pkgs = append(pkgs, e.Name)
		}

		var args []string
		env := os.Environ()
		if d.IsDebianFamily() {
			args = []string{"apt-get", "install", "--yes", "--no-install-recommends"}
			env = append(env, "DEBIAN_FRONTEND=noninteractive")
		} else if d.IsRHELFamily() {
			if d == distributions.DistributionRhel8 || d == distributions.DistributionRocky8 {
				args = []string{"/usr/bin/dnf", "install", "-y", "--setopt=install_weak_deps=False"}
			} else {
				args = []string{"/usr/bin/yum", "install", "-y"}
			}
		} else {
			return fmt.Errorf("unsupported package system")
		}
		args = append(args, pkgs...)

		klog.Infof("running command %s", args)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error installing package %q: %v: %s", e.Name, err, string(output))
		}
	} else {
		if changes.Healthy != nil {
			if d.IsDebianFamily() {
				args := []string{"dpkg", "--configure", "-a"}
				klog.Infof("package is not healthy; running command %s", args)
				cmd := exec.Command(args[0], args[1:]...)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("error running `dpkg --configure -a`: %v: %s", err, string(output))
				}

				changes.Healthy = nil
			} else if d.IsRHELFamily() {
				// Not set on TagOSFamilyRHEL, we can't currently reach here anyway...
				return fmt.Errorf("package repair not supported on RHEL/CentOS")
			} else {
				return fmt.Errorf("unsupported package system")
			}
		}

		if !reflect.DeepEqual(changes, &Package{}) {
			klog.Warningf("cannot apply package changes for %q: %+v", e.Name, changes)
		}
	}

	return nil
}

func (_ *Package) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *Package) error {
	packageName := e.Name
	if e.Source != nil {
		localFile := path.Join(localPackageDir, packageName)
		t.AddMkdirpCommand(localPackageDir, 0o755)

		url := *e.Source
		t.AddDownloadCommand(cloudinit.Always, url, localFile)

		t.AddCommand(cloudinit.Always, "dpkg", "-i", localFile)
	} else {
		packageSpec := packageName
		if e.Version != nil {
			packageSpec += " " + *e.Version
		}
		t.Config.Packages = append(t.Config.Packages, packageSpec)
	}

	return nil
}
