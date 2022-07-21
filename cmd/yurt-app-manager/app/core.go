/*
Copyright 2020 The OpenYurt Authors.

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

package app

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/openyurtio/yurt-app-manager/cmd/yurt-app-manager/options"
	"github.com/openyurtio/yurt-app-manager/pkg/projectinfo"
	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
	extclient "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/client"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/constant"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/controller"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fieldindex"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	restConfigQPS   = flag.Int("rest-config-qps", 30, "QPS of rest config.")
	restConfigBurst = flag.Int("rest-config-burst", 50, "Burst of rest config.")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1alpha1.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)

	_ = appsv1alpha1.AddToScheme(clientgoscheme.Scheme)

	// +kubebuilder:scaffold:scheme
}

// NewCmdYurtAppManager creates a *cobra.Command object with default parameters
func NewCmdYurtAppManager(stopCh <-chan struct{}) *cobra.Command {
	yurtAppOptions := options.NewYurtAppOptions()

	cmd := &cobra.Command{
		Use:   projectinfo.GetYurtAppManagerName(),
		Short: "Launch " + projectinfo.GetYurtAppManagerName(),
		Long:  "Launch " + projectinfo.GetYurtAppManagerName(),
		Run: func(cmd *cobra.Command, args []string) {
			if yurtAppOptions.Version {
				fmt.Printf("%s: %#v\n", projectinfo.GetYurtAppManagerName(), projectinfo.Get())
				return
			}

			fmt.Printf("%s version: %#v\n", projectinfo.GetYurtAppManagerName(), projectinfo.Get())

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
			})
			if err := options.ValidateOptions(yurtAppOptions); err != nil {
				klog.Fatalf("validate options: %v", err)
			}

			Run(yurtAppOptions)
		},
	}

	yurtAppOptions.AddFlags(cmd.Flags())
	return cmd
}

func Run(opts *options.YurtAppOptions) {
	if opts.EnablePprof {
		go func() {
			if err := http.ListenAndServe(opts.PprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof")
			}
		}()
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	//ctrl.SetLogger(klogr.New())

	cfg := ctrl.GetConfigOrDie()
	setRestConfig(cfg)

	cacheDisableObjs := []client.Object{
		&appsv1alpha1.YurtIngress{},
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         opts.MetricsAddr,
		HealthProbeBindAddress:     opts.HealthProbeAddr,
		LeaderElection:             opts.EnableLeaderElection,
		LeaderElectionID:           "yurt-app-manager",
		LeaderElectionNamespace:    opts.LeaderElectionNamespace,
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock, // use lease to election
		Namespace:                  opts.Namespace,
		ClientDisableCacheFor:      cacheDisableObjs,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := registerHealthChecks(mgr); err != nil {
		setupLog.Error(err, "Unable to register ready/health checks")
		os.Exit(1)
	}

	setupLog.Info("register field index")
	if err := fieldindex.RegisterFieldIndexes(mgr.GetCache()); err != nil {
		setupLog.Error(err, "failed to register field index")
		os.Exit(1)
	}

	setupLog.Info("new clientset registry")
	err = extclient.NewRegistry(mgr)
	if err != nil {
		setupLog.Error(err, "unable to init yurtapp clientset and informer")
		os.Exit(1)
	}

	setupLog.Info("setup controllers")

	ctx := genOptCtx(opts.CreateDefaultPool)
	if err = controller.SetupWithManager(mgr, ctx); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		os.Exit(1)
	}

	setupLog.Info("setup webhook")
	if err := webhook.SetupWebhooks(mgr); err != nil {
		setupLog.Error(err, "setup webhook fail")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	stopCh := ctrl.SetupSignalHandler()
	setupLog.Info("starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}

func genOptCtx(createDefaultPool bool) context.Context {
	return context.WithValue(context.Background(),
		constant.ContextKeyCreateDefaultPool, createDefaultPool)
}

func setRestConfig(c *rest.Config) {
	if *restConfigQPS > 0 {
		c.QPS = float32(*restConfigQPS)
	}
	if *restConfigBurst > 0 {
		c.Burst = *restConfigBurst
	}
}

func registerHealthChecks(mgr ctrl.Manager) error {
	klog.Info("Create readiness/health check")
	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		return err
	}
	// TODO: change the health check to be different from readiness check
	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return err
	}
	return nil
}
