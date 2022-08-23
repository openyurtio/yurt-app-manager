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

package yurtappdaemon

import (
	"reflect"
	"testing"

	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

//var cfg *rest.Config
//var testenv *envtest.Environment = &envtest.Environment{}
//var option manager.Options
//var rl resourcelock.Interface
//
//// clientTransport is used to force-close keep-alives in tests that check for leaks.
//var clientTransport *http.Transport
//
//func init() {
//	testenv = &envtest.Environment{}
//	cfg, _ = testenv.Start()
//	option = manager.Options{
//		NewCache: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
//			return nil, fmt.Errorf("expected error")
//		},
//	}
//}

//func TestAdd(t *testing.T) {
//	tests := []struct {
//		name   string
//		mng    manager.Manager
//		cxt    context.Context
//		expect map[string]string
//	}{
//		{
//			"add new key/val",
//			manager.New(cfg, option),
//			context.TODO(),
//			map[string]string{
//				"foo": "bar",
//				"buz": "qux",
//			},
//		},
//	}
//	for _, tt := range tests {
//		st := tt
//		tf := func(t *testing.T) {
//			t.Parallel()
//			t.Logf("\tTestCase: %s", st.name)
//			{
//				get := Add()
//				if !reflect.DeepEqual(get, st.expect) {
//					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
//				}
//				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
//
//			}
//		}
//		t.Run(st.name, tf)
//	}
//}

//func TestUpdateStatus(t *testing.T) {
//	tests := []struct {
//		name            string
//		instance        *unitv1alpha1.YurtAppDaemon
//		newStatus       *unitv1alpha1.YurtAppDaemonStatus
//		oldStatus       *unitv1alpha1.YurtAppDaemonStatus
//		currentRevision *appsv1.ControllerRevision
//		collisionCount  int32
//		templateType    unitv1alpha1.TemplateType
//		expect          reconcile.Result
//	}{
//		{
//			"equal",
//			&unitv1alpha1.YurtAppDaemon{},
//			&unitv1alpha1.YurtAppDaemonStatus{
//
//			},
//			&unitv1alpha1.YurtAppDaemonStatus{},
//			&appsv1.ControllerRevision{},
//			1,
//			"StatefulSet",
//			reconcile.Result{},
//		},
//	}
//
//	for _, tt := range tests {
//		st := tt
//		tf := func(t *testing.T) {
//			t.Parallel()
//			t.Logf("\tTestCase: %s", st.name)
//			{
//				rc := &ReconcileYurtAppDaemon{}
//				get, _ := rc.updateStatus(
//					st.instance, st.newStatus, st.oldStatus, st.currentRevision, st.collisionCount, st.templateType)
//				if !reflect.DeepEqual(get, st.expect) {
//					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
//				}
//				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)
//
//			}
//		}
//		t.Run(st.name, tf)
//	}
//}

func TestUpdateYurtAppDaemon(t *testing.T) {
	var int1 int32 = 11
	var yad *unitv1alpha1.YurtAppDaemon
	yad = &unitv1alpha1.YurtAppDaemon{}
	yad.Generation = 1

	tests := []struct {
		name      string
		instance  *unitv1alpha1.YurtAppDaemon
		newStatus *unitv1alpha1.YurtAppDaemonStatus
		oldStatus *unitv1alpha1.YurtAppDaemonStatus
		expect    *unitv1alpha1.YurtAppDaemon
	}{
		{
			"equal",
			yad,
			&unitv1alpha1.YurtAppDaemonStatus{
				CurrentRevision:    controllerName,
				CollisionCount:     &int1,
				TemplateType:       "StatefulSet",
				ObservedGeneration: 1,
				NodePools: []string{
					"192.168.1.1",
				},
				Conditions: []unitv1alpha1.YurtAppDaemonCondition{
					{
						Type: unitv1alpha1.WorkLoadProvisioned,
					},
				},
			},
			&unitv1alpha1.YurtAppDaemonStatus{
				CurrentRevision:    controllerName,
				CollisionCount:     &int1,
				TemplateType:       "StatefulSet",
				ObservedGeneration: 1,
				NodePools: []string{
					"192.168.1.1",
				},
				Conditions: []unitv1alpha1.YurtAppDaemonCondition{
					{
						Type: unitv1alpha1.WorkLoadProvisioned,
					},
				},
			},
			yad,
		},
	}

	for _, tt := range tests {
		st := tt
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				rc := &ReconcileYurtAppDaemon{}
				get, _ := rc.updateYurtAppDaemon(
					st.instance, st.newStatus, st.oldStatus)
				if !reflect.DeepEqual(get, st.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.expect, get)

			}
		}
		t.Run(st.name, tf)
	}
}
