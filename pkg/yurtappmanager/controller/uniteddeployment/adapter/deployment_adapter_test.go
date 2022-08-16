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

package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

var (
	one32 int32 = 1
)

func TestDeploymentAdapter_ApplyPoolTemplate(t *testing.T) {
	scheme := clientgoscheme.Scheme
	_ = alpha1.AddToScheme(scheme)

	adapter := &DeploymentAdapter{Scheme: scheme}

	tests := []struct {
		name       string
		ud         *alpha1.UnitedDeployment
		poolName   string
		revision   string
		replicas   int32
		obj        runtime.Object
		wantDeploy *appsv1.Deployment
	}{
		{
			name: "apply pool template",
			ud: &alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: alpha1.UnitedDeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"name": "foo",
						},
					},
					WorkloadTemplate: alpha1.WorkloadTemplate{
						DeploymentTemplate: &alpha1.DeploymentTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									alpha1.AnnotationPatchKey: "annotation-v",
								},
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: &one32,
								Template: corev1.PodTemplateSpec{
									ObjectMeta: metav1.ObjectMeta{
										Labels: map[string]string{
											"name": "foo",
										},
									},
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name:  "container-a",
												Image: "nginx:1.0",
											},
										},
									},
								},
							},
						},
					},
					Topology: alpha1.Topology{
						Pools: []alpha1.Pool{
							{
								Name: "hangzhou",
								NodeSelectorTerm: corev1.NodeSelectorTerm{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      "node-name",
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{"nodeA"},
										},
									},
								},
							},
						},
					},
					RevisionHistoryLimit: &one32,
				},
			},
			poolName: "hangzhou",
			revision: "1",
			replicas: one32,
			obj:      &appsv1.Deployment{},

			wantDeploy: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Labels: map[string]string{
						"name":                                "foo",
						alpha1.ControllerRevisionHashLabelKey: "1",
						alpha1.PoolNameLabelKey:               "hangzhou",
					},
					Annotations: map[string]string{
						alpha1.AnnotationPatchKey: "",
					},
					GenerateName: "foo-hangzhou-",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"name":                  "foo",
							alpha1.PoolNameLabelKey: "hangzhou",
						},
					},
					Replicas: &one32,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"name":                                "foo",
								alpha1.ControllerRevisionHashLabelKey: "1",
								alpha1.PoolNameLabelKey:               "hangzhou",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "container-a",
									Image: "nginx:1.0",
								},
							},
							Affinity: &corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													{
														Key:      "node-name",
														Operator: corev1.NodeSelectorOpIn,
														Values:   []string{"nodeA"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					RevisionHistoryLimit: &one32,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ApplyPoolTemplate(tt.ud, tt.poolName, tt.revision, tt.replicas, tt.obj)
			if err := controllerutil.SetControllerReference(tt.ud, tt.wantDeploy, adapter.Scheme); err != nil {
				panic(err)
			}
			assert.Equal(t, nil, err)
			assert.EqualValues(t, tt.wantDeploy, tt.obj.(*appsv1.Deployment))
		})
	}
}

func TestDeploymentAdapter_GetDetails(t *testing.T) {
	adapter := DeploymentAdapter{}
	tests := []struct {
		name             string
		obj              metav1.Object
		wantReplicasInfo ReplicasInfo
	}{
		{
			name: "get deploymentAdapter details",
			obj: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: &one32,
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: one32,
				},
			},
			wantReplicasInfo: ReplicasInfo{
				Replicas:      one32,
				ReadyReplicas: one32,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := adapter.GetDetails(tt.obj)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.wantReplicasInfo, got)
		})
	}
}
