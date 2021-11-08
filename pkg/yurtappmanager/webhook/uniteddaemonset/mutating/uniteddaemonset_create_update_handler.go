package mutating

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util"
	webhookutil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/util"
)

// UnitedDaemonsetCreateUpdateHandler handles UnitedDeployment
type UnitedDaemonsetCreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	Client client.Client

	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ webhookutil.Handler = &UnitedDaemonsetCreateUpdateHandler{}

func (h *UnitedDaemonsetCreateUpdateHandler) SetOptions(options webhookutil.Options) {
	h.Client = options.Client
	return
}

// Handle handles admission requests.
func (h *UnitedDaemonsetCreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &unitv1alpha1.UnitedDaemonSet{}

	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	unitv1alpha1.SetDefaultsUnitedDaemonSet(obj)
	obj.Status = unitv1alpha1.UnitedDaemonSetStatus{}

	statefulSetTemp := obj.Spec.WorkloadTemplate.StatefulSetTemplate
	deployTem := obj.Spec.WorkloadTemplate.DeploymentTemplate
	svcTem := obj.Spec.ServiceTemplate

	if statefulSetTemp != nil {
		statefulSetTemp.Spec.Selector = obj.Spec.Selector
	}
	if deployTem != nil {
		deployTem.Spec.Selector = obj.Spec.Selector
	}
	if svcTem != nil {
		svcTem.Labels[unitv1alpha1.LabelCurrentYurtAppDaemon] = obj.GetName()
		svcTem.Spec.Selector = nil
	}

	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	resp := admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, marshalled)
	if len(resp.Patches) > 0 {
		klog.V(5).Infof("Admit UnitedDaemonSet %s/%s patches: %v", obj.Namespace, obj.Name, util.DumpJSON(resp.Patches))
	}
	return resp
}

var _ admission.DecoderInjector = &UnitedDaemonsetCreateUpdateHandler{}

// InjectDecoder injects the decoder into the UnitedDeploymentCreateUpdateHandler
func (h *UnitedDaemonsetCreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}
