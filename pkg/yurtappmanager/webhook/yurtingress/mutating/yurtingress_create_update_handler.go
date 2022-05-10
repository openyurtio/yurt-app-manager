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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util"
	webhookutil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/webhook/util"
)

// YurtIngressCreateUpdateHandler handles YurtIngress
type YurtIngressCreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - import sigs.k8s.io/controller-runtime/pkg/client
	// - uncomment the InjectClient method at the bottom of this file.
	Client client.Client

	// Decoder decodes objects
	Decoder *admission.Decoder
}

var _ webhookutil.Handler = &YurtIngressCreateUpdateHandler{}

func (h *YurtIngressCreateUpdateHandler) SetOptions(options webhookutil.Options) {
	h.Client = options.Client
}

// Handle handles admission requests.
func (h *YurtIngressCreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	obj := &appsv1alpha1.YurtIngress{}
	err := h.Decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	appsv1alpha1.SetDefaultsYurtIngress(obj)

	marshalled, err := json.Marshal(&obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	resp := admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw,
		marshalled)
	if len(resp.Patches) > 0 {
		klog.V(5).Infof("Admit YurtIngress %s patches: %v", obj.Name, util.DumpJSON(resp.Patches))
	}
	return resp
}

var _ admission.DecoderInjector = &YurtIngressCreateUpdateHandler{}

func (h *YurtIngressCreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.Decoder = d
	return nil
}
