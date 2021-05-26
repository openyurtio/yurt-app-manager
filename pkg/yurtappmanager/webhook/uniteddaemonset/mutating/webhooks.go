package mutating

import (
	webhookutil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/util"
)

// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-uniteddaemonset,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=uniteddaemonsets,verbs=create;update,versions=v1alpha1,name=muniteddaemonset.kb.io

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]webhookutil.Handler{
		"mutate-apps-openyurt-io-v1alpha1-uniteddaemonset": &UnitedDaemonsetCreateUpdateHandler{},
	}
)
