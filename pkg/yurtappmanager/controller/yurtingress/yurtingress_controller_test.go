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

package yurtingress

import (
	"testing"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	yaml2 "sigs.k8s.io/yaml"

	appsv1alpha2 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
)

func TestIngressReconcile(t *testing.T) {

	yiCustomValue := `
imagePullSecrets:
- name: xxx
controller:
  service:
    type: ClusterIP
  image:
    registry: docker.io
  tolerations:
  - key: statiskey
    operator: Exists
    value: xyz
    effect: NoSchedule
`
	yiBytes, err := yaml2.YAMLToJSON([]byte(yiCustomValue))
	if err != nil {
		t.Fatal(err)
	}
	yi := &appsv1alpha2.YurtIngress{
		ObjectMeta: v1.ObjectMeta{
			Name: "xx",
		},
		Spec: appsv1alpha2.YurtIngressSpec{
			Values: &apiextensionsv1.JSON{Raw: yiBytes},
		},
	}
	now := v1.Now()
	npBj := &v1beta1.NodePool{
		ObjectMeta: v1.ObjectMeta{
			Name: "bj",
		},
		Spec: v1beta1.NodePoolSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"nodepool": "foo",
				},
			},
			Taints: []corev1.Taint{
				{
					Effect:    corev1.TaintEffectNoSchedule,
					Key:       "apps.openyurt.io/nodepool",
					Value:     "bj",
					TimeAdded: &now,
				},
			},
		},
	}

	cv, err := genChartValues(yi, npBj)
	if err != nil {
		t.Fatal(err)
	}

	str, _ := yaml.Marshal(cv)
	klog.Infof("val: \n%v", string(str))

	val, err := cv.PathValue("controller.service.type")
	if err != nil {
		t.Fatal(err)
	}

	if val.(string) != "ClusterIP" {
		t.Fatal("val not right")
	}
}
