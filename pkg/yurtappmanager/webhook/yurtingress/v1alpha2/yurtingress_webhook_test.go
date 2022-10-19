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

package v1alpha2

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
)

var defaultYurtIngress = &v1alpha2.YurtIngress{
	ObjectMeta: metav1.ObjectMeta{
		Name: "beijing",
	},
	Spec: v1alpha2.YurtIngressSpec{
		Enabled: true,
	},
}

func TestYurtAppSetDefaulter(t *testing.T) {
	webhook := &YurtIngressHandler{}
	if err := webhook.Default(context.TODO(), defaultYurtIngress); err != nil {
		t.Fatal(err)
	}
}

func TestYurtAppSetValidator(t *testing.T) {

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	_ = v1alpha2.AddToScheme(scheme)

	bjNp := &v1beta1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "beijing",
		},
		Spec: v1beta1.NodePoolSpec{},
	}
	objs := []client.Object{bjNp}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()

	webhook := &YurtIngressHandler{Client: client}

	// set default value
	if err := webhook.Default(context.TODO(), defaultYurtIngress); err != nil {
		t.Fatal(err)
	}

	if err := webhook.ValidateCreate(context.TODO(), defaultYurtIngress); err != nil {
		t.Fatal("should create success", err)
	}

	npNotExistYurtIngress := defaultYurtIngress.DeepCopy()
	npNotExistYurtIngress.Name = "notexist"
	if err := webhook.ValidateCreate(context.TODO(), npNotExistYurtIngress); err == nil {
		t.Fatal("should create fail", err)
	}

}
