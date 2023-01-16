/*
Copyright 2023 The OpenYurt authors.
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

package staticpod

import (
	"context"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// SetupWebhookWithManager sets up Cluster webhooks.
func (webhook *StaticPodHandler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1alpha1.StaticPod{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// Cluster implements a validating and defaulting webhook for Cluster.
type StaticPodHandler struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &StaticPodHandler{}
var _ webhook.CustomValidator = &StaticPodHandler{}

//+kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-staticpod,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.openyurt.io,resources=staticpods,verbs=create;update,versions=v1alpha1,name=mstaticpod.kb.io,admissionReviewVersions=v1

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (webhook *StaticPodHandler) Default(ctx context.Context, obj runtime.Object) error {
	sp, ok := obj.(*appsv1alpha1.StaticPod)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected StaticPod but got %T", obj))
	}

	// 1. Set default max-unavailable to "10%" in auto mode if it's not set
	strategy := sp.Spec.UpgradeStrategy.DeepCopy()
	if strategy != nil && strategy.Type == appsv1alpha1.AutoStaticPodUpgradeStrategyType && strategy.MaxUnavailable == nil {
		v := intstr.FromString("10%")
		sp.Spec.UpgradeStrategy.MaxUnavailable = &v
	}

	// 2. Set StaticPodManifest to the same as StaticPodName if it's not set
	if sp.Spec.StaticPodManifest == "" {
		sp.Spec.StaticPodManifest = sp.Spec.StaticPodName
	}

	return nil
}

//+kubebuilder:webhook:path=/validate-apps-openyurt-io-v1alpha1-staticpod,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.openyurt.io,resources=staticpods,verbs=create;update,versions=v1alpha1,name=vstaticpod.kb.io,admissionReviewVersions=v1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (webhook *StaticPodHandler) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	sp, ok := obj.(*appsv1alpha1.StaticPod)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected StaticPod but got %T", obj))
	}

	if allErrs := validateStaticPodSpec(&sp.Spec); len(allErrs) > 0 {
		return apierrors.NewInvalid(appsv1alpha1.GroupVersion.WithKind("StaticPod").GroupKind(), sp.Name, allErrs)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (webhook *StaticPodHandler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newSp, ok := newObj.(*appsv1alpha1.StaticPod)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected StaticPod but got %T", newObj))
	}
	oldNSp, ok := oldObj.(*appsv1alpha1.StaticPod)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected StaticPod but got %T", oldObj))
	}

	if allErrs := validateStaticPodSpecUpdate(&newSp.Spec, &oldNSp.Spec); len(allErrs) > 0 {
		return apierrors.NewInvalid(appsv1alpha1.GroupVersion.WithKind("StaticPod").GroupKind(), newSp.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (webhook *StaticPodHandler) ValidateDelete(_ context.Context, obj runtime.Object) error {

	// TODO: do not allowed to delete StaticPod when exist upgrade worker pods in cluster
	return nil
}
