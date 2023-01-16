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
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStaticPodDefaulter(t *testing.T) {
	sp := &appsv1alpha1.StaticPod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: appsv1alpha1.StaticPodSpec{
			StaticPodName: "nginx",
			UpgradeStrategy: appsv1alpha1.StaticPodUpgradeStrategy{
				Type: appsv1alpha1.AutoStaticPodUpgradeStrategyType,
			},
		},
	}

	webhook := &StaticPodHandler{}
	if err := webhook.Default(context.TODO(), sp); err != nil {
		t.Fatal(err)
	}
	if sp.Spec.StaticPodManifest != "nginx" {
		t.Fatalf("set default StaticPodManifest failed, got %v", sp.Spec.StaticPodManifest)
	}
	if sp.Spec.UpgradeStrategy.MaxUnavailable.String() != "10%" {
		t.Fatalf("set default max-unavailable failed, got %v", sp.Spec.UpgradeStrategy.MaxUnavailable)
	}
}

func TestStaticPodValidator(t *testing.T) {
	sp := &appsv1alpha1.StaticPod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: appsv1alpha1.StaticPodSpec{
			Namespace:     "default",
			StaticPodName: "nginx",
			UpgradeStrategy: appsv1alpha1.StaticPodUpgradeStrategy{
				Type: appsv1alpha1.AutoStaticPodUpgradeStrategyType,
			},
		},
	}

	scheme := runtime.NewScheme()
	if err := appsv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal("Fail to add yurt custom resource")
	}
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal("Fail to add kubernetes clint-go custom resource")
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sp).Build()
	webhook := &StaticPodHandler{
		Client: c,
	}

	// test default
	if err := webhook.Default(context.TODO(), sp); err != nil {
		t.Fatal(err)
	}

	// test create
	if err := webhook.ValidateCreate(context.TODO(), sp); err != nil {
		t.Fatalf("Fail to validate create, %v", err)
	}

}
