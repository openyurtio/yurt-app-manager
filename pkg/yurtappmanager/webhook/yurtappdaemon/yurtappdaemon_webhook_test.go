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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

var defaultAppDaemon = &v1alpha1.YurtAppDaemon{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "fooboo",
		Namespace: "default",
	},
	Spec: v1alpha1.YurtAppDaemonSpec{
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo"}},
		WorkloadTemplate: v1alpha1.WorkloadTemplate{
			DeploymentTemplate: &v1alpha1.DeploymentTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "demo"},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo"}},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "demo"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "demo", Image: "nginx"},
							},
						},
					},
				},
			},
		},
	},
}

func TestYurtAppSetDefaulter(t *testing.T) {
	webhook := &YurtAppDaemonHandler{}
	if err := webhook.Default(context.TODO(), defaultAppDaemon); err != nil {
		t.Fatal(err)
	}
}

func TestYurtAppSetValidator(t *testing.T) {

	webhook := &YurtAppDaemonHandler{}

	// set default value
	if err := webhook.Default(context.TODO(), defaultAppDaemon); err != nil {
		t.Fatal(err)
	}

	if err := webhook.ValidateCreate(context.TODO(), defaultAppDaemon); err != nil {
		t.Fatal("should create success", err)
	}

	updateAppSet := defaultAppDaemon.DeepCopy()
	updateAppSet.Spec.WorkloadTemplate.DeploymentTemplate.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo2"}}
	if err := webhook.ValidateUpdate(context.TODO(), defaultAppDaemon, updateAppSet); err == nil {
		t.Fatal("workload selector change should fail")
	}

}
