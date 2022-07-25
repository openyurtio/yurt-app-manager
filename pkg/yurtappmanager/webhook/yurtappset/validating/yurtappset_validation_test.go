/*
Copyright 2020 The OpenYurt Authors.
Copyright 2019 The Kruise Authors.

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	y2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

func TestValidateYurtAppSet(t *testing.T) {
	deployment := v1alpha1.YurtAppSet{}

	var deployYaml []byte
	var err error
	var deployJson []byte
	// 初始化k8s客户端
	//if clientset, err = common.InitClient(); err != nil {
	//	fmt.Println(err)
	//	return
	//}

	// 读取YAML
	if deployYaml, err = ioutil.ReadFile("./tmp.yaml"); err != nil {
		fmt.Println(err)
		return
	}

	// YAML转JSON
	if deployJson, err = yaml.ToJSON(deployYaml); err != nil {
		fmt.Println(err)
		return
	}

	// JSON转struct
	if err = json.Unmarshal(deployJson, &deployment); err != nil {
		fmt.Println(err)
		return
	}

	klog.Info(deployment.Name)

	//allErrs = append(allErrs, validateYurtAppSetSpec(c, &yurtAppSet.Spec, field.NewPath("spec"))...)

	errs := validateYurtAppSetSpec(nil, &deployment.Spec, field.NewPath("spec"))

	klog.Errorln(errs)

	ss, _ := y2.Marshal(deployment.Spec)
	klog.Infoln(string(ss))
}

/*
import (
	"strconv"
	"strings"
	"testing"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

func TestValidateYurtAppSet(t *testing.T) {
	validLabels := map[string]string{"a": "b"}
	validPodTemplate := v1.PodTemplate{
		WorkloadTemplate: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
			},
		},
	}

	var val int32 = 10
	replicas1 := intstr.FromInt(1)
	replicas2 := intstr.FromString("90%")
	replicas3 := intstr.FromString("71%")
	replicas4 := intstr.FromString("29%")
	successCases := []appsv1alpha1.YurtAppSet{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name: "pool",
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
						{
							Name: "pool2",
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
						{
							Name:     "pool2",
							Replicas: &replicas2,
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas3,
						},
						{
							Name:     "pool2",
							Replicas: &replicas4,
						},
					},
				},
			},
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			setTestDefault(&successCase)
			if errs := validateYurtAppSet(&successCase); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	errorCases := map[string]appsv1alpha1.YurtAppSet{
		"no pod template label": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: v1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{},
								Spec: v1.PodSpec{
									RestartPolicy: v1.RestartPolicyAlways,
									DNSPolicy:     v1.DNSClusterFirst,
									Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
								},
							},
						},
					},
				},
			},
		},
		"no pool template": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{},
			},
		},
		"no pool name": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{},
					},
				},
			},
		},
		"invalid pool nodeSelectorTerm": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name: "pool",
							NodeSelectorTerm: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "key",
										Operator: corev1.NodeSelectorOpExists,
										Values:   []string{"unexpected"},
									},
								},
							},
						},
					},
				},
			},
		},
		"pool replicas is not enough": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
					},
				},
			},
		},
		"pool replicas is too small": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas1,
						},
						{
							Name:     "pool2",
							Replicas: &replicas3,
						},
					},
				},
			},
		},
		"pool replicas is too much": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas3,
						},
						{
							Name:     "pool2",
							Replicas: &replicas2,
						},
					},
				},
			},
		},
		"partition not exist": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
				UpdateStrategy: appsv1alpha1.YurtAppSetUpdateStrategy{
					StatefulSetUpdateStrategy: &appsv1alpha1.StatefulSetUpdateStrategy{
						Partitions: map[string]int32{
							"notExist": 1,
						},
					},
				},
				Topology: appsv1alpha1.Topology{
					Pools: []appsv1alpha1.Pool{
						{
							Name:     "pool1",
							Replicas: &replicas3,
						},
						{
							Name:     "pool2",
							Replicas: &replicas2,
						},
					},
				},
			},
		},
		"duplicated templates": {
			ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault},
			Spec: appsv1alpha1.YurtAppSetSpec{
				Replicas: &val,
				Selector: &metav1.LabelSelector{MatchLabels: validLabels},
				WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
					StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.StatefulSetSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
					DeploymentTemplate: &appsv1alpha1.DeploymentTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: validLabels,
						},
						Spec: apps.DeploymentSpec{
							WorkloadTemplate: validPodTemplate.WorkloadTemplate,
						},
					},
				},
			},
		},
	}

	for k, v := range errorCases {
		t.Run(k, func(t *testing.T) {
			setTestDefault(&v)
			errs := validateYurtAppSet(&v)
			if len(errs) == 0 {
				t.Errorf("expected failure for %s", k)
			}

			for i := range errs {
				field := errs[i].Field
				if !strings.HasPrefix(field, "spec.template") &&
					field != "spec.selector" &&
					field != "spec.topology.pools" &&
					field != "spec.topology.pools[0]" &&
					field != "spec.topology.pools[0].name" &&
					field != "spec.updateStrategy.partitions" &&
					field != "spec.topology.pools[0].nodeSelectorTerm.matchExpressions[0].values" {
					t.Errorf("%s: missing prefix for: %v", k, errs[i])
				}
			}
		})
	}
}

type UpdateCase struct {
	Old appsv1alpha1.YurtAppSet
	New appsv1alpha1.YurtAppSet
}

func TestValidateYurtAppSetUpdate(t *testing.T) {
	validLabels := map[string]string{"a": "b"}
	validPodTemplate := v1.PodTemplate{
		WorkloadTemplate: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: validLabels,
			},
			Spec: v1.PodSpec{
				RestartPolicy: v1.RestartPolicyAlways,
				DNSPolicy:     v1.DNSClusterFirst,
				Containers:    []v1.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
			},
		},
	}

	var val int32 = 10
	successCases := []UpdateCase{
		{
			Old: appsv1alpha1.YurtAppSet{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.YurtAppSetSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
						},
					},
				},
			},
			New: appsv1alpha1.YurtAppSet{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.YurtAppSetSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
							{
								Name: "pool-b",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"b"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Old: appsv1alpha1.YurtAppSet{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.YurtAppSetSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
							{
								Name: "pool-b",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"b"},
										},
									},
								},
							},
						},
					},
				},
			},
			New: appsv1alpha1.YurtAppSet{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.YurtAppSetSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, successCase := range successCases {
		t.Run("success case "+strconv.Itoa(i), func(t *testing.T) {
			setTestDefault(&successCase.Old)
			setTestDefault(&successCase.New)
			if errs := ValidateYurtAppSetUpdate(&successCase.Old, &successCase.New); len(errs) != 0 {
				t.Errorf("expected success: %v", errs)
			}
		})
	}

	errorCases := map[string]UpdateCase{
		"pool nodeSelector changed": {
			Old: appsv1alpha1.YurtAppSet{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.YurtAppSetSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a", "b"},
										},
									},
								},
							},
						},
					},
				},
			},
			New: appsv1alpha1.YurtAppSet{
				ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: metav1.NamespaceDefault, ResourceVersion: "1"},
				Spec: appsv1alpha1.YurtAppSetSpec{
					Replicas: &val,
					Selector: &metav1.LabelSelector{MatchLabels: validLabels},
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: validLabels,
							},
							Spec: apps.StatefulSetSpec{
								WorkloadTemplate: validPodTemplate.WorkloadTemplate,
							},
						},
					},
					Topology: appsv1alpha1.Topology{
						Pools: []appsv1alpha1.Pool{
							{
								Name: "pool-a",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"a"},
										},
									},
								},
							},
							{
								Name: "pool-b",
								NodeSelectorTerm: v1.NodeSelectorTerm{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Key:      "domain",
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"b"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for k, v := range errorCases {
		t.Run(k, func(t *testing.T) {
			setTestDefault(&v.Old)
			setTestDefault(&v.New)
			errs := ValidateYurtAppSetUpdate(&v.Old, &v.New)
			if len(errs) == 0 {
				t.Errorf("expected failure for %s", k)
			}

			for i := range errs {
				field := errs[i].Field
				if !strings.HasPrefix(field, "spec.template.") &&
					field != "spec.selector" &&
					field != "spec.topology.pool" &&
					field != "spec.topology.pool.name" &&
					field != "spec.updateStrategy.partitions" &&
					field != "spec.topology.pools[0].nodeSelectorTerm" {
					t.Errorf("%s: missing prefix for: %v", k, errs[i])
				}
			}
		})
	}
}

func setTestDefault(obj *appsv1alpha1.YurtAppSet) {
	if obj.Spec.Replicas == nil {
		obj.Spec.Replicas = new(int32)
		*obj.Spec.Replicas = 1
	}
	if obj.Spec.RevisionHistoryLimit == nil {
		obj.Spec.RevisionHistoryLimit = new(int32)
		*obj.Spec.RevisionHistoryLimit = 10
	}
}


*/
