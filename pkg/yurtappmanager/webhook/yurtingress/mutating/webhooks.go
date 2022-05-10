/*
Copyright 2021 The Openyurt Authors.

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

package mutating

import (
	webhookutil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/util"
)

// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-yurtingress,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtingresses,verbs=create;update,versions=v1alpha1,name=myurtingress.kb.io

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]webhookutil.Handler{
		"mutate-apps-openyurt-io-v1alpha1-yurtingress": &YurtIngressCreateUpdateHandler{},
	}
)
