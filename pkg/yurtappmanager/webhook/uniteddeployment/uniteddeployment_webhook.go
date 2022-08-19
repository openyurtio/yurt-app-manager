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

package uniteddeployment

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
func (webhook *UnitedDeploymentHandler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.UnitedDeployment{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:path=/validate-apps-openyurt-io-v1alpha1-uniteddeployment,mutating=false,failurePolicy=fail,groups=apps.openyurt.io,resources=uniteddeployments,verbs=create;update,versions=v1alpha1,name=vuniteddeployment.kb.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-uniteddeployment,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=uniteddeployments,verbs=create;update,versions=v1alpha1,name=muniteddeployment.kb.io,sideEffects=None,admissionReviewVersions=v1

// Cluster implements a validating and defaulting webhook for Cluster.
type UnitedDeploymentHandler struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &UnitedDeploymentHandler{}
var _ webhook.CustomValidator = &UnitedDeploymentHandler{}

// Default satisfies the defaulting webhook interface.
func (webhook *UnitedDeploymentHandler) Default(ctx context.Context, obj runtime.Object) error {
	appset, ok := obj.(*v1alpha1.UnitedDeployment)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a UnitedDeployment but got a %T", obj))
	}

	v1alpha1.SetDefaultsUnitedDeployment(appset)
	appset.Status = v1alpha1.UnitedDeploymentStatus{}

	statefulSetTemp := appset.Spec.WorkloadTemplate.StatefulSetTemplate
	deployTem := appset.Spec.WorkloadTemplate.DeploymentTemplate

	if statefulSetTemp != nil {
		statefulSetTemp.Spec.Selector = appset.Spec.Selector
	}
	if deployTem != nil {
		deployTem.Spec.Selector = appset.Spec.Selector
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *UnitedDeploymentHandler) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	appset, ok := obj.(*v1alpha1.UnitedDeployment)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a UnitedDeployment but got a %T", obj))
	}

	if allErrs := validateUnitedDeployment(webhook.Client, appset); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("UnitedDeployment").GroupKind(), appset.Name, allErrs)
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *UnitedDeploymentHandler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newAppSet, ok := newObj.(*v1alpha1.UnitedDeployment)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a UnitedDeployment but got a %T", newObj))
	}
	oldAppSet, ok := oldObj.(*v1alpha1.UnitedDeployment)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a UnitedDeployment but got a %T", oldObj))
	}

	validationErrorList := validateUnitedDeployment(webhook.Client, newAppSet)
	updateErrorList := ValidateUnitedDeploymentUpdate(newAppSet, oldAppSet)

	if allErrs := append(validationErrorList, updateErrorList...); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("UnitedDeployment").GroupKind(), newAppSet.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *UnitedDeploymentHandler) ValidateDelete(_ context.Context, obj runtime.Object) error {
	return nil
}
