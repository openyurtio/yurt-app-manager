package validating

import (
	webhookutil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/util"
)

// +kubebuilder:webhook:path=/validate-apps-openyurt-io-v1alpha1-uniteddaemonset,mutating=false,failurePolicy=fail,groups=apps.openyurt.io,resources=uniteddaemonsets,verbs=create;update,versions=v1alpha1,name=vuniteddaemonset.kb.io

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]webhookutil.Handler{
		"validate-apps-openyurt-io-v1alpha1-uniteddaemonset": &UnitedDaemonSetCreateUpdateHandler{},
	}
)
