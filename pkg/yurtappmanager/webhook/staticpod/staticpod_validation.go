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
	"k8s.io/apimachinery/pkg/util/validation/field"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// validateStaticPodSpec validates the staticpod spec.
func validateStaticPodSpec(spec *appsv1alpha1.StaticPodSpec) field.ErrorList {
	var allErrs field.ErrorList

	if spec.StaticPodName == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("StaticPodName"),
			"StaticPodName is required"))
	}

	if spec.StaticPodManifest == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("StaticPodManifest"),
			"StaticPodManifest is required"))
	}

	if spec.Namespace == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("Namespace"),
			"Namespace is required"))
	}

	strategy := &spec.UpgradeStrategy
	if strategy == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("upgradeStrategy"),
			"upgrade strategy is required"))
	}

	if strategy.Type != appsv1alpha1.AutoStaticPodUpgradeStrategyType && strategy.Type != appsv1alpha1.OTAStaticPodUpgradeStrategyType {
		allErrs = append(allErrs, field.NotSupported(field.NewPath("spec").Child("upgradeStrategy"),
			strategy, []string{"auto", "ota"}))
	}

	if strategy.Type == appsv1alpha1.AutoStaticPodUpgradeStrategyType && strategy.MaxUnavailable == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("upgradeStrategy"),
			"max-unavailable is required in auto mode"))
	}

	if allErrs != nil {
		return allErrs
	}

	return nil
}

func validateStaticPodSpecUpdate(spec *appsv1alpha1.StaticPodSpec, oldSpec *appsv1alpha1.StaticPodSpec) field.ErrorList {
	return validateStaticPodSpec(spec)
}

func validateStaticPodSpecDeletion(spec *appsv1alpha1.StaticPodSpec) field.ErrorList {
	return validateStaticPodSpec(spec)
}
