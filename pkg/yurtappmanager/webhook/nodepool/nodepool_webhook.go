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

package nodepool

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// SetupWebhookWithManager sets up Cluster webhooks.
func (webhook *NodePoolHandler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.NodePool{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-apps-openyurt-io-v1alpha1-nodepool,mutating=false,failurePolicy=fail,groups=apps.openyurt.io,resources=nodepools,versions=v1alpha1,name=vnodepool.kb.io,sideEffects=None,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/mutate-apps-openyurt-io-v1alpha1-nodepool,mutating=true,failurePolicy=fail,groups=apps.openyurt.io,resources=nodepools,verbs=create;update,versions=v1alpha1,name=mnodepool.kb.io,sideEffects=None,admissionReviewVersions=v1

// Cluster implements a validating and defaulting webhook for Cluster.
type NodePoolHandler struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &NodePoolHandler{}
var _ webhook.CustomValidator = &NodePoolHandler{}

// Default satisfies the defaulting webhook interface.
func (webhook *NodePoolHandler) Default(ctx context.Context, obj runtime.Object) error {
	np, ok := obj.(*v1alpha1.NodePool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NodePool but got a %T", obj))
	}

	np.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{v1alpha1.LabelCurrentNodePool: np.Name},
	}

	// add NodePool.Spec.Type to NodePool labels
	if np.Labels == nil {
		np.Labels = make(map[string]string)
	}
	np.Labels[v1alpha1.NodePoolTypeLabelKey] = string(np.Spec.Type)

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *NodePoolHandler) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	np, ok := obj.(*v1alpha1.NodePool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NodePool but got a %T", obj))
	}

	if allErrs := validateNodePoolSpec(&np.Spec); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("NodePool").GroupKind(), np.Name, allErrs)
	}

	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *NodePoolHandler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newNp, ok := newObj.(*v1alpha1.NodePool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NodePool but got a %T", newObj))
	}
	oldNp, ok := oldObj.(*v1alpha1.NodePool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NodePool but got a %T", oldObj))
	}

	if allErrs := validateNodePoolSpecUpdate(&newNp.Spec, &oldNp.Spec); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("NodePool").GroupKind(), newNp.Name, allErrs)
	}

	return nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type.
func (webhook *NodePoolHandler) ValidateDelete(_ context.Context, obj runtime.Object) error {
	np, ok := obj.(*v1alpha1.NodePool)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a NodePool but got a %T", obj))
	}
	if allErrs := validateNodePoolDeletion(webhook.Client, np); len(allErrs) > 0 {
		return apierrors.NewInvalid(v1alpha1.GroupVersion.WithKind("NodePool").GroupKind(), np.Name, allErrs)
	}

	return nil
}
