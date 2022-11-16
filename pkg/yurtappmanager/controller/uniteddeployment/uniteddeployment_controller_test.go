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

package uniteddeployment

import (
	"context"
	"strconv"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	fakeclint "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	adpt "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/controller/uniteddeployment/adapter"
)

var (
	ten     int32 = 10
	forteen int64 = 14
	fifteen int64 = 15
)

func TestReconcileUnitedDeployment_Reconcile(t *testing.T) {
	instance := &appsv1alpha1.UnitedDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "foo-ns",
		},
		Spec: appsv1alpha1.UnitedDeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "foo",
				},
			},
			WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
				DeploymentTemplate: &appsv1alpha1.DeploymentTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
						Labels: map[string]string{
							"name": "foo",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &two,
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"name": "foo",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "container",
										Image: "nginx:1.0",
									},
								},
							},
						},
					},
				},
			},
			Topology: appsv1alpha1.Topology{
				Pools: []appsv1alpha1.Pool{
					{
						Name:     "foo-0",
						Replicas: &one,
						NodeSelectorTerm: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "app.openyurt.io/nodepool",
									Operator: corev1.NodeSelectorOpIn,
									Values: []string{
										"foo-0",
									},
								},
							},
						},
					},
					{
						Name:     "foo-1",
						Replicas: &two,
						NodeSelectorTerm: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "app.openyurt.io/nodepool",
									Operator: corev1.NodeSelectorOpIn,
									Values: []string{
										"foo-1",
									},
								},
							},
						},
					},
				},
			},
			RevisionHistoryLimit: &two,
		},
		Status: appsv1alpha1.UnitedDeploymentStatus{
			ObservedGeneration: fifteen,
			CollisionCount:     &two,
			CurrentRevision:    "v0.1.0",
			Replicas:           2,
			ReadyReplicas:      2,
			TemplateType:       appsv1alpha1.DeploymentTemplateType,
		},
	}

	scheme := runtime.NewScheme()
	if err := appsv1alpha1.AddToScheme(scheme); err != nil {
		t.Logf("failed to add yurt custom resource")
		return
	}
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Logf("failed to add kubernetes clint-go custom resource")
		return
	}
	fc := fakeclint.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(instance).Build()

	var req = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "foo-ns"}}
	rud := ReconcileUnitedDeployment{
		Client: fc,
		scheme: scheme,
		poolControls: map[appsv1alpha1.TemplateType]ControlInterface{
			appsv1alpha1.StatefulSetTemplateType: &PoolControl{
				Client:  fc,
				scheme:  scheme,
				adapter: &adpt.StatefulSetAdapter{Client: fc, Scheme: scheme},
			},
			appsv1alpha1.DeploymentTemplateType: &PoolControl{
				Client:  fc,
				scheme:  scheme,
				adapter: &adpt.DeploymentAdapter{Client: fc, Scheme: scheme},
			},
		},
	}

	for i := 0; i < 2; i++ {
		tf := rud.poolControls[appsv1alpha1.DeploymentTemplateType].CreatePool(instance, "foo-"+strconv.FormatInt(int64(i), 10), "v0.1.0", two)
		if tf != nil {
			t.Logf("failed create node pool resource")
		}
	}

	_, err := rud.Reconcile(context.TODO(), req)
	if err != nil {
		t.Logf("failed to control uniteddeployment controller")
	}

}

func TestGetPoolTemplateType(t *testing.T) {
	instances := []struct {
		ud   *appsv1alpha1.UnitedDeployment
		want appsv1alpha1.TemplateType
	}{
		{
			ud: &appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "foo-ns",
				},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						DeploymentTemplate: &appsv1alpha1.DeploymentTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "foo",
								Namespace: "foo-ns",
							},
						},
					},
				},
			},
			want: "Deployment",
		},
		{
			ud: &appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "foo",
				},
				Spec: appsv1alpha1.UnitedDeploymentSpec{
					WorkloadTemplate: appsv1alpha1.WorkloadTemplate{
						StatefulSetTemplate: &appsv1alpha1.StatefulSetTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "foo",
								Namespace: "foo-ns",
							},
						},
					},
				},
			},
			want: "StatefulSet",
		},
		{
			ud: &appsv1alpha1.UnitedDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-uniteddeployment",
					Namespace: "foo",
				},
			},
			want: "",
		},
	}

	for _, v := range instances {
		tt := getPoolTemplateType(v.ud)
		if tt != v.want {
			t.Logf("failed to get pool template type")
		}
	}
}

func TestReconcileUnitedDeployment_UpdateUnitedDeployment(t *testing.T) {
	instance := struct {
		ud        *appsv1alpha1.UnitedDeployment
		oldStatus *appsv1alpha1.UnitedDeploymentStatus
		newStatus *appsv1alpha1.UnitedDeploymentStatus
	}{
		ud: &appsv1alpha1.UnitedDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "foo",
				Namespace:  "foo-ns",
				Generation: forteen,
			},
			Spec: appsv1alpha1.UnitedDeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"name": "foo",
					},
				},
			},
			Status: appsv1alpha1.UnitedDeploymentStatus{},
		},
		oldStatus: &appsv1alpha1.UnitedDeploymentStatus{
			CurrentRevision:    "v0.1.0",
			CollisionCount:     &ten,
			Replicas:           two,
			ReadyReplicas:      one,
			ObservedGeneration: fifteen,
			PoolReplicas: map[string]int32{
				"foo-pool": ten,
			},
			Conditions: []appsv1alpha1.UnitedDeploymentCondition{},
		},
		newStatus: &appsv1alpha1.UnitedDeploymentStatus{
			CurrentRevision:    "v0.2.0",
			CollisionCount:     &ten,
			Replicas:           two,
			ReadyReplicas:      one,
			ObservedGeneration: fifteen,
			PoolReplicas: map[string]int32{
				"foo-pool": ten,
			},
			Conditions: []appsv1alpha1.UnitedDeploymentCondition{},
		},
	}

	scheme := runtime.NewScheme()
	if err := appsv1alpha1.AddToScheme(scheme); err != nil {
		t.Logf("failed to add yurt custom resource")
		return
	}
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Logf("failed to add kubernetes clint-go custom resource")
		return
	}
	fc := fakeclint.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(instance.ud).Build()
	r := ReconcileUnitedDeployment{Client: fc, scheme: scheme}
	o, err := r.updateUnitedDeployment(instance.ud, instance.oldStatus, instance.newStatus)
	if err != nil || o.Status.CurrentRevision != instance.newStatus.CurrentRevision {
		t.Logf("failed to update uniteddeployment")
	}
}
