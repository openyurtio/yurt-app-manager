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

package yurtingress

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// SetupWebhookWithManager sets up Cluster webhooks.
func (webhook *YurtIngressHandler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.YurtIngress{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-yurtingress,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtingresses,verbs=create;update,versions=v1alpha1,name=myurtingress.kb.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/validate-apps-openyurt-io-v1alpha1-yurtingress,mutating=false,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtingresses,verbs=create;update;delete,versions=v1alpha1,name=vyurtingress.kb.io,sideEffects=None,admissionReviewVersions=v1

// Cluster implements a validating and defaulting webhook for Cluster.
type YurtIngressHandler struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &YurtIngressHandler{}
var _ webhook.CustomValidator = &YurtIngressHandler{}

// Default satisfies the defaulting webhook interface.
func (webhook *YurtIngressHandler) Default(ctx context.Context, obj runtime.Object) error {
	ing, ok := obj.(*v1alpha1.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", obj))
	}

	v1alpha1.SetDefaultsYurtIngress(ing)

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtIngressHandler) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	ingress, ok := obj.(*v1alpha1.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", obj))
	}

	if allErrs := validateYurtIngressSpec(webhook.Client, ingress.ObjectMeta.Name, &ingress.Spec, false); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("YurtIngress").GroupKind(), ingress.Name, allErrs)
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtIngressHandler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newIngress, ok := newObj.(*v1alpha1.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", newObj))
	}
	oldIngress, ok := oldObj.(*v1alpha1.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", oldObj))
	}

	if allErrs := validateYurtIngressSpecUpdate(webhook.Client, newIngress.ObjectMeta.Name, &newIngress.Spec, &oldIngress.Spec); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("YurtIngress").GroupKind(), newIngress.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtIngressHandler) ValidateDelete(_ context.Context, obj runtime.Object) error {
	ingress, ok := obj.(*v1alpha1.YurtIngress)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtIngress but got a %T", obj))
	}

	if allErrs := validateYurtIngressSpecDeletion(webhook.Client, ingress.ObjectMeta.Name, &ingress.Spec); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("YurtIngress").GroupKind(), ingress.Name, allErrs)
	}
	return nil
}
