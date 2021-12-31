/*
Copyright 2021 The OpenYurt Authors.

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
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util"
	webhookutil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/util"
)

// YurtIngressCreateUpdateHandler handles YurtIngress
type YurtIngressCreateUpdateHandler struct {
	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ webhookutil.Handler = &YurtIngressCreateUpdateHandler{}

func (h *YurtIngressCreateUpdateHandler) SetOptions(options webhookutil.Options) {
	//return
}

// Handle handles admission requests.
func (h *YurtIngressCreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	np_ing := appsv1alpha1.YurtIngress{}
	err := h.Decoder.Decode(req, &np_ing)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	marshalled, err := json.Marshal(&np_ing)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	resp := admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw,
		marshalled)
	if len(resp.Patches) > 0 {
		klog.V(5).Infof("Admit YurtIngress %s patches: %v", np_ing.Name, util.DumpJSON(resp.Patches))
	}
	return resp
}

var _ admission.DecoderInjector = &YurtIngressCreateUpdateHandler{}

func (h *YurtIngressCreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}
