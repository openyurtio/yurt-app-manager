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
	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

const (
	failed  = "\u2717"
	succeed = "\u2713"
)

func TestIsStrArrayEqual(t *testing.T) {
	tests := []struct {
		name     string
		strList1 []string
		strList2 []string
		expect   bool
	}{
		{
			"unequal len",
			[]string{
				"a",
			},
			[]string{
				"a",
				"b",
			},
			false,
		},
		{
			"lens equal 0",
			[]string{},
			[]string{},
			true,
		},
		{
			"strarray equal",
			[]string{
				"a",
				"b",
			},
			[]string{
				"a",
				"b",
			},
			true,
		},
		{
			"strarray unequal",
			[]string{
				"a",
				"b",
			},
			[]string{
				"a",
				"ab",
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
				get := isStrArrayEqual(st.strList1, st.strList2)

				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestGetPools(t *testing.T) {
	tests := []struct {
		name    string
		desired []appsv1alpha1.IngressPool
		current []appsv1alpha1.IngressPool
		expect  [3][]appsv1alpha1.IngressPool
	}{
		{
			"unchange",
			[]appsv1alpha1.IngressPool{
				{
					Name: "a",
					IngressIPs: []string{
						"192.168.0.1",
					},
				},
			},
			[]appsv1alpha1.IngressPool{
				{
					Name: "a",
					IngressIPs: []string{
						"192.168.0.1",
					},
				},
			},
			[3][]appsv1alpha1.IngressPool{
				nil,
				nil,
				{
					{
						Name: "a",
						IngressIPs: []string{
							"192.168.0.1",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			var get [3][]appsv1alpha1.IngressPool
			get[0], get[1], get[2] = getPools(tt.desired, tt.current)

			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestGetDesiredPools(t *testing.T) {
	tests := []struct {
		name        string
		yurtIngress *appsv1alpha1.YurtIngress
		expect      []appsv1alpha1.IngressPool
	}{
		{
			"normal",
			&appsv1alpha1.YurtIngress{
				ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
				Spec: appsv1alpha1.YurtIngressSpec{
					Replicas:                   3,
					IngressControllerImage:     "a",
					IngressWebhookCertGenImage: "a",
					Pools: []appsv1alpha1.IngressPool{
						{
							Name: "a",
							IngressIPs: []string{
								"192.168.0.1",
							},
						},
					},
				},
			},
			[]appsv1alpha1.IngressPool{
				{
					Name: "a",
					IngressIPs: []string{
						"192.168.0.1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := getDesiredPools(tt.yurtIngress)

			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestGetCurrentPools(t *testing.T) {
	tests := []struct {
		name        string
		yurtIngress *appsv1alpha1.YurtIngress
		expect      []appsv1alpha1.IngressPool
	}{
		{
			"normal",
			&appsv1alpha1.YurtIngress{
				ObjectMeta: metav1.ObjectMeta{Name: "a"},
				Status: appsv1alpha1.YurtIngressStatus{
					Conditions: appsv1alpha1.YurtIngressCondition{
						IngressReadyPools: []appsv1alpha1.IngressPool{
							{
								Name: "a",
								IngressIPs: []string{
									"192.168.0.1",
								},
							},
						},
					},
				},
			},
			[]appsv1alpha1.IngressPool{
				{
					Name: "a",
					IngressIPs: []string{
						"192.168.0.1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := getCurrentPools(tt.yurtIngress)

			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestGetDesiredPool(t *testing.T) {
	tests := []struct {
		name        string
		yurtIngress *appsv1alpha1.YurtIngress
		poolname    string
		expect      *appsv1alpha1.IngressPool
	}{
		{
			"true",
			&appsv1alpha1.YurtIngress{
				Spec: appsv1alpha1.YurtIngressSpec{
					Replicas:                   3,
					IngressControllerImage:     "a",
					IngressWebhookCertGenImage: "a",
					Pools: []appsv1alpha1.IngressPool{
						{
							Name: "a",
							IngressIPs: []string{
								"192.168.0.1",
							},
						},
					},
				},
			},
			"a",
			&appsv1alpha1.IngressPool{
				Name: "a",
				IngressIPs: []string{
					"192.168.0.1",
				},
			},
		},
		{
			"false",
			&appsv1alpha1.YurtIngress{
				Spec: appsv1alpha1.YurtIngressSpec{
					Replicas:                   3,
					IngressControllerImage:     "a",
					IngressWebhookCertGenImage: "a",
					Pools: []appsv1alpha1.IngressPool{
						{
							Name: "a",
							IngressIPs: []string{
								"192.168.0.1",
							},
						},
					},
				},
			},
			"b",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := getCurrentPool(tt.yurtIngress, tt.poolname)

			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestGetCurrentPool(t *testing.T) {
	tests := []struct {
		name        string
		yurtIngress *appsv1alpha1.YurtIngress
		poolname    string
		expect      *appsv1alpha1.IngressPool
	}{
		{
			"true",
			&appsv1alpha1.YurtIngress{
				Spec: appsv1alpha1.YurtIngressSpec{
					Replicas:                   3,
					IngressControllerImage:     "a",
					IngressWebhookCertGenImage: "a",
					Pools: []appsv1alpha1.IngressPool{
						{
							Name: "a",
							IngressIPs: []string{
								"192.168.0.1",
							},
						},
					},
				},
			},
			"a",
			&appsv1alpha1.IngressPool{
				Name: "a",
				IngressIPs: []string{
					"192.168.0.1",
				},
			},
		},
		{
			"false",
			&appsv1alpha1.YurtIngress{
				Spec: appsv1alpha1.YurtIngressSpec{
					Replicas:                   3,
					IngressControllerImage:     "a",
					IngressWebhookCertGenImage: "a",
					Pools: []appsv1alpha1.IngressPool{
						{
							Name: "a",
							IngressIPs: []string{
								"192.168.0.1",
							},
						},
					},
				},
			},
			"b",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := getCurrentPool(tt.yurtIngress, tt.poolname)

			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestRemovePoolfromCondition(t *testing.T) {
	tests := []struct {
		name        string
		yurtIngress *appsv1alpha1.YurtIngress
		poolname    string
		expect      bool
	}{
		{
			"true in IngressReadyPools",
			&appsv1alpha1.YurtIngress{
				Status: appsv1alpha1.YurtIngressStatus{
					ReadyNum: 1,
					Conditions: appsv1alpha1.YurtIngressCondition{
						IngressReadyPools: []appsv1alpha1.IngressPool{
							{
								Name: "a",
								IngressIPs: []string{
									"192.168.0.1",
								},
							},
						},
					},
				},
			},
			"a",
			true,
		},
		{
			"true in IngressNotReadyPool",
			&appsv1alpha1.YurtIngress{
				Status: appsv1alpha1.YurtIngressStatus{
					ReadyNum: 1,
					Conditions: appsv1alpha1.YurtIngressCondition{
						IngressNotReadyPools: []appsv1alpha1.IngressNotReadyPool{
							{
								Pool: appsv1alpha1.IngressPool{
									Name: "a",
									IngressIPs: []string{
										"192.168.0.1",
									},
								},
							},
						},
					},
				},
			},
			"a",
			true,
		},
		{
			"false",
			&appsv1alpha1.YurtIngress{
				Spec: appsv1alpha1.YurtIngressSpec{
					Replicas:                   3,
					IngressControllerImage:     "a",
					IngressWebhookCertGenImage: "a",
					Pools: []appsv1alpha1.IngressPool{
						{
							Name: "a",
							IngressIPs: []string{
								"192.168.0.1",
							},
						},
					},
				},
			},
			"b",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := removePoolfromCondition(tt.yurtIngress, tt.poolname)

			if !reflect.DeepEqual(get, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestPrepareDeploymentOwnerReferences(t *testing.T) {
	type Result struct {
		Name string
		UID  string
	}

	tests := []struct {
		name        string
		yurtIngress *appsv1alpha1.YurtIngress
		expect      Result
	}{
		{
			"normal",
			&appsv1alpha1.YurtIngress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "a",
					UID:  "a",
				},
			},
			Result{
				"a",
				"a",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := prepareDeploymentOwnerReferences(tt.yurtIngress)
			result := Result{get.Name, string(get.UID)}
			if !reflect.DeepEqual(result, tt.expect) {
				t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
			}
			t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
		})
	}
}

func TestGetUnreadyDeploymentCondition(t *testing.T) {
	type Result struct {
		conditionType appsv1alpha1.IngressNotReadyType
	}

	tests := []struct {
		name   string
		dply   *appsv1.Deployment
		expect *Result
	}{
		{
			"nil",
			&appsv1.Deployment{
				Status: appsv1.DeploymentStatus{},
			},
			nil,
		},
		{
			"fail",
			&appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type: appsv1.DeploymentReplicaFailure,
						},
					},
				},
			},
			&Result{
				conditionType: appsv1alpha1.IngressFailure,
			},
		},
		{
			"pending",
			&appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type: appsv1.DeploymentAvailable,
						},
					},
				},
			},
			&Result{
				conditionType: appsv1alpha1.IngressPending,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)

			get := getUnreadyDeploymentCondition(tt.dply)
			if get == nil {
				if !reflect.DeepEqual(&get, tt.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
			} else {
				result := Result{
					get.Type,
				}
				if !reflect.DeepEqual(&result, tt.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, result)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, result)
			}
		})
	}
}
