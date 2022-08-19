/*
Copyright 2021 The OpenYurt Authors.
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

package adapter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

func TestGetCurrentPartitionForStrategyOnDelete(t *testing.T) {
	currentPods := buildPodList([]int{0, 1, 2}, []string{"v1", "v2", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2}, []string{"v1", "v1", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 2 {
		t.Fatalf("expected partition 2, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2, 3}, []string{"v2", "v1", "v2", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{1, 2, 3}, []string{"v1", "v2", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 3}, []string{"v2", "v1", "v2"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 1 {
		t.Fatalf("expected partition 1, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2}, []string{"v1", "v1", "v1"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 3 {
		t.Fatalf("expected partition 3, got %d", *partition)
	}

	currentPods = buildPodList([]int{0, 1, 2, 4}, []string{"v1", "", "v2", "v3"}, t)
	if partition := getCurrentPartition(currentPods, "v2"); *partition != 3 {
		t.Fatalf("expected partition 3, got %d", *partition)
	}
}

func buildPodList(ordinals []int, revisions []string, t *testing.T) []*corev1.Pod {
	if len(ordinals) != len(revisions) {
		t.Fatalf("ordinals count should equals to revision count")
	}
	pods := []*corev1.Pod{}
	for i, ordinal := range ordinals {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      fmt.Sprintf("pod-%d", ordinal),
			},
		}
		if revisions[i] != "" {
			pod.Labels = map[string]string{
				appsv1alpha1.ControllerRevisionHashLabelKey: revisions[i],
			}
		}
		pods = append(pods, pod)
	}

	return pods
}

func TestCreateNewPatchedObject(t *testing.T) {
	cases := []struct {
		Name         string
		PatchInfo    *runtime.RawExtension
		OldObj       *appsv1.Deployment
		EqualFuntion func(new *appsv1.Deployment) bool
	}{
		{
			Name:      "replace image",
			PatchInfo: &runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"spec":{"containers":[{"image":"nginx:1.18.0","name":"nginx"}]}}}}`)},
			OldObj: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:1.19.0",
								},
							},
						},
					},
				},
			},
			EqualFuntion: func(new *appsv1.Deployment) bool {
				return new.Spec.Template.Spec.Containers[0].Image == "nginx:1.18.0"
			},
		},
		{
			Name:      "add other image",
			PatchInfo: &runtime.RawExtension{Raw: []byte(`{"spec":{"template":{"spec":{"containers":[{"image":"nginx:1.18.0","name":"nginx111"}]}}}}`)},
			OldObj: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:1.19.0",
								},
							},
						},
					},
				},
			},
			EqualFuntion: func(new *appsv1.Deployment) bool {
				if len(new.Spec.Template.Spec.Containers) != 2 {
					return false
				}
				containerMap := make(map[string]string)
				for _, container := range new.Spec.Template.Spec.Containers {
					containerMap[container.Name] = container.Image
				}
				image, ok := containerMap["nginx"]
				if !ok {
					return false
				}

				image1, ok := containerMap["nginx111"]
				if !ok {
					return false
				}
				return image == "nginx:1.19.0" && image1 == "nginx:1.18.0"
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			newObj := &appsv1.Deployment{}
			if err := CreateNewPatchedObject(c.PatchInfo, c.OldObj, newObj); err != nil {
				t.Fatalf("%s CreateNewPatchedObject error %v", c.Name, err)
			}
			if !c.EqualFuntion(newObj) {
				t.Fatalf("%s Not Expect equal function", c.Name)
			}
		})
	}

}

func Test_getPoolPrefix(t *testing.T) {

	tests := []struct {
		caseName       string
		controllerName string
		poolName       string
		prefix         string
	}{
		{caseName: "valid prefix", controllerName: "ud", poolName: "hangzhou", prefix: "ud-hangzhou-"},
		{caseName: "invalid prefix: contain uppercase characters", controllerName: "ud", poolName: "HangZhou", prefix: "ud-"},
	}
	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			if got := getPoolPrefix(tt.controllerName, tt.poolName); got != tt.prefix {
				t.Errorf("getPoolPrefix() = %v, want %v", got, tt.prefix)
			}
		})
	}
}

func Test_attachNodeAffinity(t *testing.T) {
	tests := []struct {
		name          string
		podSpec       *corev1.PodSpec
		pool          *appsv1alpha1.Pool
		expectPodSpec *corev1.PodSpec
	}{
		{
			name:    "pod Spec.Affinity is nil",
			podSpec: &corev1.PodSpec{},
			pool: &appsv1alpha1.Pool{
				NodeSelectorTerm: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "apps.openyurt.io/desired-nodepool",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"hangzhou"},
						},
					},
				},
			},

			expectPodSpec: &corev1.PodSpec{Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "apps.openyurt.io/desired-nodepool",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"hangzhou"},
									},
								},
							},
						},
					},
				},
			}},
		},
		{
			name:    "pool has pool.NodeSelector",
			podSpec: &corev1.PodSpec{},
			pool: &appsv1alpha1.Pool{
				NodeSelectorTerm: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "apps.openyurt.io/desired-nodepool",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"hangzhou"},
						},
					},
				},
			},

			expectPodSpec: &corev1.PodSpec{Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "apps.openyurt.io/desired-nodepool",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"hangzhou"},
									},
								},
							},
						},
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachNodeAffinity(tt.podSpec, tt.pool)
			assert.Equal(t, tt.expectPodSpec, tt.podSpec)
		})
	}
}

func Test_attachTolerations(t *testing.T) {
	tests := []struct {
		name          string
		podSpec       *corev1.PodSpec
		poolConfig    *appsv1alpha1.Pool
		expectPodSpec *corev1.PodSpec
	}{
		{
			name:       "poolConfig's Tolerations is nil",
			podSpec:    &corev1.PodSpec{},
			poolConfig: &appsv1alpha1.Pool{},

			expectPodSpec: &corev1.PodSpec{},
		},
		{
			name:    "poolConfig with Tolerations",
			podSpec: &corev1.PodSpec{},
			poolConfig: &appsv1alpha1.Pool{Tolerations: []corev1.Toleration{
				{
					Key:      "key",
					Operator: corev1.TolerationOpEqual,
					Value:    "value",
				},
			}},

			expectPodSpec: &corev1.PodSpec{Tolerations: []corev1.Toleration{
				{
					Key:      "key",
					Operator: corev1.TolerationOpEqual,
					Value:    "value",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachTolerations(tt.podSpec, tt.poolConfig)
			assert.Equal(t, tt.expectPodSpec, tt.podSpec)
		})
	}
}
