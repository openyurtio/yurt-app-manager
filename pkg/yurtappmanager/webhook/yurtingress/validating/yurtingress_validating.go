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

package validating

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// validateYurtIngressSpec validates the yurt ingress spec.
func validateYurtIngressSpec(c client.Client, ingressName string, spec *appsv1alpha1.YurtIngressSpec, isdelete bool) field.ErrorList {
	if len(spec.Pools) > 0 {
		var err error
		var errmsg string

		ingressList := appsv1alpha1.YurtIngressList{}
		if err = c.List(context.TODO(), &ingressList, &client.ListOptions{}); err != nil {
			errmsg = "List YurtIngressList error!"
			klog.Errorf(errmsg)
			return field.ErrorList([]*field.Error{
				field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
		}

		//get all the nodepools with ingress enabled
		var npIngressMap map[string]string = make(map[string]string)
		if !isdelete && len(ingressList.Items) > 0 {
			for _, ingress := range ingressList.Items { //go through all the yurtingress
				for _, np := range ingress.Spec.Pools { //get all the nodepools with ingress enabled
					npIngressMap[np.Name] = ingress.ObjectMeta.Name
				}
			}
		}

		nps := appsv1alpha1.NodePoolList{}
		if err = c.List(context.TODO(), &nps, &client.ListOptions{}); err != nil {
			errmsg = "List nodepool list error!"
			klog.Errorf(errmsg)
			return field.ErrorList([]*field.Error{
				field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
		}

		// validate whether the nodepool exist
		if len(nps.Items) > 0 {
			var found = false
			for _, snp := range spec.Pools { //go through the nodepools setting in yaml
				for _, cnp := range nps.Items { //go through the nodepools in cluster
					if snp.Name == cnp.ObjectMeta.Name {
						// check if ingress is already enabled in certain nodepool
						for inp := range npIngressMap {
							if snp.Name == inp && ingressName != npIngressMap[inp] {
								errmsg = "Nodepool \"" + snp.Name + "\" has been enabled in \"" + npIngressMap[inp] + "\" already!"
								klog.Errorf(errmsg)
								return field.ErrorList([]*field.Error{
									field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
							}
						}
						found = true
						break
					}
				}
				if !found {
					errmsg = snp.Name + " does not exist in the cluster!"
					klog.Errorf(errmsg)
					return field.ErrorList([]*field.Error{
						field.Forbidden(field.NewPath("spec").Child("pools"), errmsg)})
				}
				found = false
			}

		}
	}
	return nil
}

func validateYurtIngressSpecUpdate(c client.Client, ingressName string, spec *appsv1alpha1.YurtIngressSpec, oldSpec *appsv1alpha1.YurtIngressSpec) field.ErrorList {
	return validateYurtIngressSpec(c, ingressName, spec, false)
}

func validateYurtIngressSpecDeletion(c client.Client, ingressName string, spec *appsv1alpha1.YurtIngressSpec) field.ErrorList {
	return validateYurtIngressSpec(c, ingressName, spec, true)
}
