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

package nodepool

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

const (
	failed  = "\u2717"
	succeed = "\u2713"
)

func TestConciliateLables(t *testing.T) {
	tests := []struct {
		name      string
		node      *corev1.Node
		oldLabels map[string]string
		newLabels map[string]string
		expect    map[string]string
	}{
		{
			"remove lable",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{"foo": "bar"},
			map[string]string{"foo": "bar"},
		},
		{
			"add labels",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-node",
					Labels: map[string]string{"foo": "bar"},
				},
			},
			map[string]string{"foo": "bar"},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
		},
		{
			"remove and add labels",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
		},
		{
			"with existing node labels",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"foo":    "bar",
						"buz":    "qux",
						"grault": "corge",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":    "bar",
				"quux":   "quuz",
				"grault": "corge",
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				conciliateLabels(st.node, st.oldLabels, st.newLabels)
				get := st.node.Labels
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestConciliateAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		node     *corev1.Node
		oldAnnos map[string]string
		newAnnos map[string]string
		expect   map[string]string
	}{
		{
			"remove an annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{"foo": "bar"},
			map[string]string{"foo": "bar"},
		},
		{
			"add an annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-node",
					Annotations: map[string]string{"foo": "bar"},
				},
			},
			map[string]string{"foo": "bar"},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
		},
		{
			"remove and add an annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						"foo": "bar",
						"buz": "qux",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
		},
		{
			"with existing node annotations",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						"foo":    "bar",
						"buz":    "qux",
						"grault": "corge",
					},
				},
			},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{
				"foo":  "bar",
				"quux": "quuz",
			},
			map[string]string{
				"foo":    "bar",
				"quux":   "quuz",
				"grault": "corge",
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				conciliateAnnotations(st.node, st.oldAnnos, st.newAnnos)
				get := st.node.Annotations
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestConciliateTaints(t *testing.T) {
	tests := []struct {
		name      string
		node      corev1.Node
		preTaints []corev1.Taint
		newTaints []corev1.Taint
		expect    []corev1.Taint
	}{
		{
			"remove the taint",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{},
			[]corev1.Taint{},
		},
		{
			"add a new taint",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
		},
		{
			"update a taint",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
		},
		{
			"with existing node taints",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
						{
							Key:    "qux",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			[]corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoExecute,
				},
				{
					Key:    "qux",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				conciliateTaints(&st.node, st.preTaints, st.newTaints)
				get := st.node.Spec.Taints
				if !taintSliceEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func taintSliceEqual(s1, s2 []corev1.Taint) bool {
	sort.Slice(s1, func(i, j int) bool {
		if s1[i].Key < s1[j].Key {
			return true
		}
		if s1[i].Key > s1[j].Key {
			return false
		}
		// if s1[i].Key == s1[j].Key, compare the Effect
		return s1[i].Effect < s1[j].Effect
	})
	sort.Slice(s2, func(i, j int) bool {
		if s2[i].Key < s2[j].Key {
			return true
		}
		if s2[i].Key > s2[j].Key {
			return false
		}
		return s2[i].Effect < s2[j].Effect
	})
	return reflect.DeepEqual(s1, s2)
}

func TestContainTaint(t *testing.T) {
	tmpTime := metav1.Now()
	tests := []struct {
		name   string
		taint  corev1.Taint
		taints []corev1.Taint
		expect bool
	}{
		{
			"containt the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoSchedule,
					TimeAdded: &tmpTime,
				},
			},
			true,
		},
		{
			"not containt the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &tmpTime,
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				_, get := containTaint(st.taint, st.taints)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestIsNodeReady(t *testing.T) {
	tests := []struct {
		name   string
		node   corev1.Node
		expect bool
	}{
		{
			"node is ready",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			true,
		},
		{
			"node is not ready",
			corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := isNodeReady(st.node)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestMergeMap(t *testing.T) {
	tests := []struct {
		name   string
		map1   map[string]string
		map2   map[string]string
		expect map[string]string
	}{
		{
			"add new key/val",
			map[string]string{"foo": "bar"},
			map[string]string{"buz": "qux"},
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
		},
		{
			"nil m1 map",
			nil,
			map[string]string{"buz": "qux"},
			map[string]string{"buz": "qux"},
		},
		{
			"replace exist key/val",
			map[string]string{
				"foo": "bar",
				"buz": "qux",
			},
			map[string]string{"buz": "quux"},
			map[string]string{
				"foo": "bar",
				"buz": "quux",
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := mergeMap(st.map1, st.map2)
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestRemoveTaint(t *testing.T) {
	tmpTime := metav1.Now()
	tests := []struct {
		name   string
		taint  corev1.Taint
		taints []corev1.Taint
		expect []corev1.Taint
	}{
		{
			"remove the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoSchedule,
					TimeAdded: &tmpTime,
				},
			},
			[]corev1.Taint{},
		},
		{
			"not contained the taint",
			corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &tmpTime,
				},
			},
			[]corev1.Taint{
				{
					Key:       "foo",
					Value:     "bar",
					Effect:    corev1.TaintEffectNoExecute,
					TimeAdded: &tmpTime,
				},
			},
		},
	}
	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := removeTaint(st.taint, st.taints)
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

// prepare variables that can be reused

var testLabel1 map[string]string = map[string]string{
	"foo":                             "bar",
	"buz":                             "qux",
	appsv1alpha1.LabelCurrentNodePool: "test",
}

var testAnnotations map[string]string = map[string]string{
	"foo": "qux",
	"buz": "qux",
}

var testTaints []corev1.Taint = []corev1.Taint{
	{
		Key:    "foo",
		Value:  "bar",
		Effect: corev1.TaintEffectNoExecute,
	},
}

var testNPRA1 NodePoolRelatedAttributes = NodePoolRelatedAttributes{
	Labels:      testLabel1,
	Annotations: testAnnotations,
	Taints:      testTaints,
}

func TestCachePrevPoolAttrs(t *testing.T) {

	npraBytes, _ := json.Marshal(testNPRA1)

	var annotationWithNP map[string]string = map[string]string{
		"foo":                            "qux",
		"buz":                            "qux",
		appsv1alpha1.AnnotationPrevAttrs: string(npraBytes),
	}

	tests := []struct {
		name   string
		node   *corev1.Node
		npra   NodePoolRelatedAttributes
		expect error
	}{
		{
			"node already cached",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      testLabel1,
					Annotations: annotationWithNP,
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			testNPRA1,
			nil,
		},
		{
			"node without cached annotation",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      testLabel1,
					Annotations: nil,
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			testNPRA1,
			nil,
		},
	}

	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := cachePrevPoolAttrs(st.node, st.npra)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
			}
		}
		t.Run(st.name, tf)
	}
}

func TestConciliateNode(t *testing.T) {

	npraBytes, _ := json.Marshal(testNPRA1)

	// test cases
	tests := []struct {
		name   string
		node   *corev1.Node
		np     appsv1alpha1.NodePool
		expect bool
	}{
		{
			"nothing to update",
			// use DeepCopy because of the side effects
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo":                             "bar",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: map[string]string{
						"foo":                            "qux",
						"buz":                            "qux",
						appsv1alpha1.AnnotationPrevAttrs: string(npraBytes),
					},
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			appsv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1alpha1.NodePoolSpec{
					Labels: map[string]string{
						"foo":                             "bar",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: testAnnotations,
					Taints:      testTaints,
				},
			},
			false,
		},
		{
			"outdated npra attrs",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo":                             "bar",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: map[string]string{
						"foo":                            "qux",
						"buz":                            "qux",
						appsv1alpha1.AnnotationPrevAttrs: string(npraBytes),
					},
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			appsv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1alpha1.NodePoolSpec{
					Labels: map[string]string{
						"foo":                             "qux",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: testAnnotations,
					Taints:      testTaints,
				},
			},
			true,
		},
		{
			"no npra attrs",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo":                             "bar",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: nil,
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			appsv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1alpha1.NodePoolSpec{
					Labels: map[string]string{
						"foo":                             "bar",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: testAnnotations,
					Taints:      testTaints,
				},
			},
			true,
		},
		{
			"nodepool name not match owner label",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo":                             "qux",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: nil,
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			appsv1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
				},
				Spec: appsv1alpha1.NodePoolSpec{
					Labels: map[string]string{
						"foo":                             "bar",
						"buz":                             "qux",
						appsv1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: testAnnotations,
					Taints:      testTaints,
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Logf("\tTestCase: %s", st.name)
			{
				get, _ := concilateNode(st.node, st.np)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
			}
		}
		t.Run(st.name, tf)
	}

}

func TestRemovePoolRelatedAttrs(t *testing.T) {

	npraBytes, _ := json.Marshal(testNPRA1)

	var annotationWithNP map[string]string = mergeMap(map[string]string{
		"foo": "qux",
		"buz": "qux",
	}, map[string]string{
		appsv1alpha1.AnnotationPrevAttrs: string(npraBytes),
	})

	// test cases
	tests := []struct {
		name   string
		node   *corev1.Node
		expect error
	}{
		{
			"node without cached annotations",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			nil,
		},
		{
			"node with cached annotations",
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      testNPRA1.Labels,
					Annotations: annotationWithNP,
				},
				Spec: corev1.NodeSpec{
					Taints: testTaints,
				},
			},
			nil,
		},
	}

	for _, st := range tests {
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := removePoolRelatedAttrs(st.node)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
			}
		}
		t.Run(st.name, tf)
	}

}

func TestConciliateNodePoolStatus(t *testing.T) {
	tests := []struct {
		name          string
		readyNodes    int32
		notReadyNodes int32
		nodes         []string
		nodePool      *appsv1alpha1.NodePool
		expect        bool
	}{
		{
			"nodepool don't need update",
			1,
			1,
			[]string{"node1"},
			&appsv1alpha1.NodePool{
				Status: appsv1alpha1.NodePoolStatus{
					ReadyNodeNum:   1,
					UnreadyNodeNum: 1,
					Nodes: []string{
						"node1",
					},
				},
			},
			false,
		}, {
			"updated ReadyNum",
			2,
			1,
			[]string{"node1"},
			&appsv1alpha1.NodePool{
				Status: appsv1alpha1.NodePoolStatus{
					ReadyNodeNum:   1,
					UnreadyNodeNum: 1,
					Nodes: []string{
						"node1",
					},
				},
			},
			true,
		},
		{
			"updated UnreadyNum",
			1,
			2,
			[]string{"node1"},
			&appsv1alpha1.NodePool{
				Status: appsv1alpha1.NodePoolStatus{
					ReadyNodeNum:   1,
					UnreadyNodeNum: 1,
					Nodes: []string{
						"node1",
					},
				},
			},
			true,
		},
		{
			"updated node list",
			1,
			1,
			[]string{"node1", "node2"},
			&appsv1alpha1.NodePool{
				Status: appsv1alpha1.NodePoolStatus{
					ReadyNodeNum:   1,
					UnreadyNodeNum: 1,
					Nodes: []string{
						"node1",
					},
				},
			},
			true,
		},
		{
			"updated node list & readynode",
			2,
			1,
			[]string{"node1", "node2"},
			&appsv1alpha1.NodePool{
				Status: appsv1alpha1.NodePoolStatus{
					ReadyNodeNum:   1,
					UnreadyNodeNum: 1,
					Nodes: []string{
						"node1",
					},
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := conciliateNodePoolStatus(st.readyNodes, st.notReadyNodes, st.nodes, st.nodePool)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
			}
		}
		t.Run(st.name, tf)
	}
}

func TestCreateNP(t *testing.T) {

	scheme := runtime.NewScheme()
	appsv1alpha1.AddToScheme(scheme)

	oldTimeSleep := timeSleep
	defer func() {
		timeSleep = oldTimeSleep
	}()

	timeSleepCnt := 0
	timeSleep = func(duration time.Duration) {
		timeSleepCnt += 1
	}

	tests := []struct {
		name     string
		c        client.Client
		npName   string
		poolType appsv1alpha1.NodePoolType
		expect   bool
	}{
		{
			"create nodepool success",
			fake.NewClientBuilder().
				WithScheme(scheme).
				Build(),
			"test",
			appsv1alpha1.Cloud,
			true,
		},
		{
			"create nodepool success",
			fake.NewClientBuilder().
				Build(),
			"test",
			appsv1alpha1.Cloud,
			false,
		},
		{
			"create nodepool fail: already exist",
			fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(&appsv1alpha1.NodePool{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: appsv1alpha1.NodePoolSpec{
						Type: appsv1alpha1.Cloud,
					},
				}).
				Build(),
			"test",
			appsv1alpha1.Cloud,
			false,
		},
	}

	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Logf("\tTestCase: %s", st.name)
			{
				get := createNodePool(st.c, st.npName, st.poolType)
				if get != st.expect {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
			}
		}
		t.Run(st.name, tf)
	}

}

func TestGetRemovedNodes(t *testing.T) {
	tests := []struct {
		name            string
		currentNodeList *corev1.NodeList
		desiredNodeList *corev1.NodeList
		expect          []corev1.Node
	}{
		{
			"no nodes should be removed",
			&corev1.NodeList{
				Items: nil,
			},
			&corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			},
			nil,
		},
		{
			"all nodes need to be removed",
			&corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			},
			&corev1.NodeList{
				Items: nil,
			},
			[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
				},
			}},
		},
		{
			"some nodes need to be kept",
			&corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test2",
						},
					},
				},
			},
			&corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			},
			[]corev1.Node{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
			}},
		},
	}

	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				get := getRemovedNodes(st.currentNodeList, st.desiredNodeList)
				if !reflect.DeepEqual(st.expect, get) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
			}
		}
		t.Run(st.name, tf)
	}
}
