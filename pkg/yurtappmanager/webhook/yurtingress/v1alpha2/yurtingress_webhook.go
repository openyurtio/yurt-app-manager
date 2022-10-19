/*
Copyright 2022 The OpenYurt authors.
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
// +kubebuilder:docs-gen:collapse=Apache License

package v1alpha2

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
)

func (webhook *YurtIngressHandler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha2.YurtIngress{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha2-yurtingress,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtingresses,verbs=create;update,versions=v1alpha2,name=myurtingress.kb.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/validate-apps-openyurt-io-v1alpha2-yurtingress,mutating=false,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtingresses,verbs=create;update;delete,versions=v1alpha2,name=vyurtingress.kb.io,sideEffects=None,admissionReviewVersions=v1

// Cluster implements a validating and defaulting webhook for Cluster.
type YurtIngressHandler struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &YurtIngressHandler{}
var _ webhook.CustomValidator = &YurtIngressHandler{}

// Default satisfies the defaulting webhook interface.
func (webhook *YurtIngressHandler) Default(ctx context.Context, obj runtime.Object) error {
	yi, ok := obj.(*v1alpha2.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", obj))
	}

	if yi.Spec.Interval == nil {
		yi.Spec.Interval = &metav1.Duration{Duration: 1 * time.Minute}
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtIngressHandler) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	ingress, ok := obj.(*v1alpha2.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", obj))
	}

	np := &v1beta1.NodePool{}
	if err := webhook.Client.Get(ctx, types.NamespacedName{Name: ingress.Name}, np); err != nil {
		if apierrors.IsNotFound(err) {
			return apierrors.NewBadRequest(fmt.Sprintf("nodepool %s not exist", ingress.Name))
		}
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtIngressHandler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	return nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtIngressHandler) ValidateDelete(_ context.Context, obj runtime.Object) error {
	return nil
}
