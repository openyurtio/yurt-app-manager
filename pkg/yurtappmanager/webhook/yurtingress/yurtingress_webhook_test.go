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

package yurtingress

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

var defaultYurtIngress = &v1alpha1.YurtIngress{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "fooboo",
		Namespace: "default",
	},
	Spec: v1alpha1.YurtIngressSpec{
		Replicas:                   1,
		IngressControllerImage:     "registry.k8s.io/ingress-nginx/controller:v0.49.0",
		IngressWebhookCertGenImage: "registry.k8s.io/ingress-nginx/kube-webhook-certgen:v0.49.0",
		Pools:                      []v1alpha1.IngressPool{{Name: "beijing"}},
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
	_ = v1alpha1.AddToScheme(scheme)

	bjNp := &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "beijing",
			Namespace: "default",
		},
		Spec: v1alpha1.NodePoolSpec{},
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

	npNotExist := defaultYurtIngress.DeepCopy()
	npNotExist.Spec.Pools = []v1alpha1.IngressPool{{Name: "noneexist"}}
	if err := webhook.ValidateCreate(context.TODO(), npNotExist); err == nil {
		t.Fatal("should create fail", err)
	}

}
