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

package yurtappdaemon

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
func (webhook *YurtAppDaemonHandler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.YurtAppDaemon{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:path=/validate-apps-openyurt-io-v1alpha1-yurtappdaemon,mutating=false,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtappdaemons,verbs=create;update,versions=v1alpha1,name=vyurtappdaemon.kb.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-yurtappdaemon,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=yurtappdaemons,verbs=create;update,versions=v1alpha1,name=myurtappdaemon.kb.io,sideEffects=None,admissionReviewVersions=v1

// Cluster implements a validating and defaulting webhook for Cluster.
type YurtAppDaemonHandler struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &YurtAppDaemonHandler{}
var _ webhook.CustomValidator = &YurtAppDaemonHandler{}

// Default satisfies the defaulting webhook interface.
func (webhook *YurtAppDaemonHandler) Default(ctx context.Context, obj runtime.Object) error {
	daemon, ok := obj.(*v1alpha1.YurtAppDaemon)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtAppDaemon but got a %T", obj))
	}

	v1alpha1.SetDefaultsYurtAppDaemon(daemon)
	daemon.Status = v1alpha1.YurtAppDaemonStatus{}

	statefulSetTemp := daemon.Spec.WorkloadTemplate.StatefulSetTemplate
	deployTem := daemon.Spec.WorkloadTemplate.DeploymentTemplate

	if statefulSetTemp != nil {
		statefulSetTemp.Spec.Selector = daemon.Spec.Selector
	}
	if deployTem != nil {
		deployTem.Spec.Selector = daemon.Spec.Selector
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtAppDaemonHandler) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	daemon, ok := obj.(*v1alpha1.YurtAppDaemon)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtAppDaemon but got a %T", obj))
	}

	if allErrs := validateYurtAppDaemon(webhook.Client, daemon); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("YurtAppDaemon").GroupKind(), daemon.Name, allErrs)
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtAppDaemonHandler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newDaemon, ok := newObj.(*v1alpha1.YurtAppDaemon)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtAppDaemon but got a %T", newObj))
	}
	oldDaemon, ok := oldObj.(*v1alpha1.YurtAppDaemon)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a YurtAppDaemon but got a %T", oldObj))
	}

	validationErrorList := validateYurtAppDaemon(webhook.Client, newDaemon)
	updateErrorList := ValidateYurtAppDaemonUpdate(newDaemon, oldDaemon)
	if allErrs := append(validationErrorList, updateErrorList...); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("YurtAppDaemon").GroupKind(), newDaemon.Name, allErrs)
	}
	return nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *YurtAppDaemonHandler) ValidateDelete(_ context.Context, obj runtime.Object) error {
	return nil
}
