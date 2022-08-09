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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

var defaultNodePool = &v1alpha1.NodePool{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "fooboo",
		Namespace: "default",
	},
	Spec: v1alpha1.NodePoolSpec{
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "demo"}},
		Type:     v1alpha1.Cloud,
	},
}

func TestNodePoolDefaulter(t *testing.T) {
	webhook := &NodePoolHandler{}
	if err := webhook.Default(context.TODO(), defaultNodePool); err != nil {
		t.Fatal(err)
	}
}

func TestNodePoolValidator(t *testing.T) {

	// prepare fake client
	scheme := runtime.NewScheme()
	v1alpha1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(defaultNodePool).Build()
	webhook := &NodePoolHandler{
		Client: cl,
	}

	// test default
	if err := webhook.Default(context.TODO(), defaultNodePool); err != nil {
		t.Fatal(err)
	}

	// test create
	if err := webhook.ValidateCreate(context.TODO(), defaultNodePool); err != nil {
		t.Fatal("should create success", err)
	}

	// test update
	updatedNp := defaultNodePool.DeepCopy()
	updatedNp.Spec.Type = v1alpha1.Edge
	if err := webhook.ValidateUpdate(context.TODO(), defaultNodePool, updatedNp); err == nil {
		t.Fatal("workload selector change should fail")
	}

	// test delete
	if err := webhook.ValidateDelete(context.TODO(), defaultNodePool); err != nil {
		t.Fatal(err)
	}

}
