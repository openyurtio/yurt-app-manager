package webhook

import (
	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/gate"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/yurtappdaemon/mutating"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/yurtappdaemon/validating"
)

func init() {
	if !gate.ResourceEnabled(&unitv1alpha1.YurtAppDaemon{}) {
		return
	}
	addHandlers(mutating.HandlerMap)
	addHandlers(validating.HandlerMap)
}
