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

package main

import (
	"flag"
	"fmt"
	"os"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"k8s.io/kops/cmd/kops-controller/controllers"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/cmd/kops-controller/pkg/server"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/nodeidentity"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	nodeidentityazure "k8s.io/kops/pkg/nodeidentity/azure"
	nodeidentitydo "k8s.io/kops/pkg/nodeidentity/do"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	nodeidentityhetzner "k8s.io/kops/pkg/nodeidentity/hetzner"
	nodeidentityos "k8s.io/kops/pkg/nodeidentity/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm/gcetpmverifier"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/yaml"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	// +kubebuilder:scaffold:scheme
}

func main() {
	klog.InitFlags(nil)

	// Disable metrics by default (avoid port conflicts, also risky because we are host network)
	metricsAddress := ":0"
	// flag.StringVar(&metricsAddr, "metrics-addr", metricsAddress, "The address the metric endpoint binds to.")

	configPath := "/etc/kubernetes/kops-controller/config.yaml"
	flag.StringVar(&configPath, "conf", configPath, "Location of yaml configuration file")

	flag.Parse()

	if configPath == "" {
		klog.Fatalf("must specify --conf")
	}

	var opt config.Options
	opt.PopulateDefaults()

	{
		b, err := os.ReadFile(configPath)
		if err != nil {
			klog.Fatalf("failed to read configuration file %q: %v", configPath, err)
		}

		if err := yaml.Unmarshal(b, &opt); err != nil {
			klog.Fatalf("failed to parse configuration file %q: %v", configPath, err)
		}
	}

	ctrl.SetLogger(klogr.New())

	if err := buildScheme(); err != nil {
		setupLog.Error(err, "error building scheme")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddress,
		LeaderElection:     true,
		LeaderElectionID:   "kops-controller-leader",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if opt.Server != nil {
		var verifier bootstrap.Verifier
		var err error
		if opt.Server.Provider.AWS != nil {
			verifier, err = awsup.NewAWSVerifier(opt.Server.Provider.AWS)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
		} else if opt.Server.Provider.GCE != nil {
			verifier, err = gcetpmverifier.NewTPMVerifier(opt.Server.Provider.GCE)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
		} else {
			klog.Fatalf("server cloud provider config not provided")
		}

		srv, err := server.NewServer(&opt, verifier)
		if err != nil {
			setupLog.Error(err, "unable to create server")
			os.Exit(1)
		}
		mgr.Add(srv)
	}

	if opt.EnableCloudIPAM {
		setupLog.Info("enabling IPAM controller")
		if opt.Cloud != "aws" {
			klog.Error("IPAM controller only supported by aws")
			os.Exit(1)
		}
		ipamController, err := controllers.NewAWSIPAMReconciler(mgr)
		if err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "IPAMController")
			os.Exit(1)
		}
		if err := ipamController.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "IPAMController")
			os.Exit(1)
		}
	}

	if err := addNodeController(mgr, &opt); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeController")
		os.Exit(1)
	}

	if err := addGossipController(mgr, &opt); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GossipController")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func buildScheme() error {
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering corev1: %v", err)
	}
	// Needed so that the leader-election system can post events
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering coordinationv1: %v", err)
	}
	return nil
}

func addNodeController(mgr manager.Manager, opt *config.Options) error {
	var legacyIdentifier nodeidentity.LegacyIdentifier
	var identifier nodeidentity.Identifier
	var err error
	switch opt.Cloud {
	case "aws":
		identifier, err = nodeidentityaws.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "gce":
		legacyIdentifier, err = nodeidentitygce.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "openstack":
		legacyIdentifier, err = nodeidentityos.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "digitalocean":
		legacyIdentifier, err = nodeidentitydo.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "hetzner":
		identifier, err = nodeidentityhetzner.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %w", err)
		}

	case "azure":
		identifier, err = nodeidentityazure.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "":
		return fmt.Errorf("must specify cloud")

	default:
		return fmt.Errorf("identifier for cloud %q not implemented", opt.Cloud)
	}

	if identifier != nil {
		nodeController, err := controllers.NewNodeReconciler(mgr, identifier)
		if err != nil {
			return err
		}
		if err := nodeController.SetupWithManager(mgr); err != nil {
			return err
		}
	} else {
		if opt.ConfigBase == "" {
			return fmt.Errorf("must specify configBase")
		}

		nodeController, err := controllers.NewLegacyNodeReconciler(mgr, opt.ConfigBase, legacyIdentifier)
		if err != nil {
			return err
		}
		if err := nodeController.SetupWithManager(mgr); err != nil {
			return err
		}
	}

	return nil
}

func addGossipController(mgr manager.Manager, opt *config.Options) error {
	if opt.Discovery == nil || !opt.Discovery.Enabled {
		return nil
	}

	configMapID := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "coredns",
	}

	controller, err := controllers.NewHostsReconciler(mgr, configMapID)
	if err != nil {
		return err
	}

	if err := controller.SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}
