/*
Copyright 2022 The OpenYurt Authors.

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

package webhook

import (
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/nodepool"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/nodepool/v1beta1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/uniteddeployment"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/yurtappdaemon"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/yurtappset"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/yurtingress"
)

func SetupWebhooks(mgr ctrl.Manager) error {
	// Our existing call to SetupWebhookWithManager registers our conversion webhooks with the manager, too.
	if err := (&nodepool.NodePoolHandler{Client: mgr.GetClient()}).SetupWebhookWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create webhook for NodePool")
	}

	if err := (&v1beta1.NodePoolHandler{}).SetupWebhookWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create webhook for v1beta1 NodePool")
	}

	if err := (&uniteddeployment.UnitedDeploymentHandler{Client: mgr.GetClient()}).SetupWebhookWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create webhook for UnitedDeployment")
	}

	if err := (&yurtappdaemon.YurtAppDaemonHandler{Client: mgr.GetClient()}).SetupWebhookWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create webhook for YurtAppDaemon")
	}

	if err := (&yurtappset.YurtAppSetHandler{Client: mgr.GetClient()}).SetupWebhookWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create webhook for YurtAppSet")
	}

	if err := (&yurtingress.YurtIngressHandler{Client: mgr.GetClient()}).SetupWebhookWithManager(mgr); err != nil {
		return errors.Wrapf(err, "unable to create webhook for YurtIngress")
	}

	return nil
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
