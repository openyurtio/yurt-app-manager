/*
Copyright 2020 The OpenYurt Authors.

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
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// dummy object for test
var enqueueNodePoolForNode EnqueueNodePoolForNode = EnqueueNodePoolForNode{}

func createQueue() workqueue.RateLimitingInterface {
	return workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(1*time.Millisecond, 1*time.Second))
}

func TestCreate(t *testing.T) {

	tests := []struct {
		name    string
		nodeObj client.Object
		q       workqueue.RateLimitingInterface
		added   int // the items in queue
	}{
		{
			"wrong runtime Object",
			&corev1.Pod{},
			createQueue(),
			0,
		},
		{
			"add node within nodepool",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
					},
				},
			},
			createQueue(),
			1,
		},
		{
			"add node not in nodepool",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			createQueue(),
			0,
		},
	}

	for _, st := range tests {
		st := st
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				enqueueNodePoolForNode.Create(event.CreateEvent{
					Object: st.nodeObj,
				}, st.q)
				get := st.q.Len()
				if get != st.added {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.added, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.added, get)

			}
		}
		t.Run(st.name, tf)
	}

}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name       string
		newNodeObj client.Object
		oldNodeObj client.Object
		q          workqueue.RateLimitingInterface
		added      int // the items in queue
	}{
		{
			"change node's nodepool to another",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test1",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelCurrentNodePool: "test2",
					},
				},
			},
			createQueue(),
			2,
		},
		{
			"remove node from old pool",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelCurrentNodePool: "test",
					},
				},
			},
			createQueue(),
			1,
		},
		{
			"add node to new pool",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelCurrentNodePool: "",
					},
				},
			},
			createQueue(),
			1,
		},
		{
			"node status change",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelCurrentNodePool: "test",
					},
				},
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{
							Type:   corev1.NodeReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			createQueue(),
			1,
		},
		{
			"nothing change ",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
						v1alpha1.LabelCurrentNodePool: "test",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
						v1alpha1.LabelCurrentNodePool: "test",
					},
				},
			},
			createQueue(),
			0,
		},
		{
			"nodepool related attrs change",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
						v1alpha1.LabelCurrentNodePool: "test",
					},
					Annotations: map[string]string{
						v1alpha1.AnnotationPrevAttrs: "test",
					},
				},
			},
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelDesiredNodePool: "test",
						v1alpha1.LabelCurrentNodePool: "test",
					},
				},
			},
			createQueue(),
			1,
		},
	}

	for _, st := range tests {
		st := st
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				enqueueNodePoolForNode.Update(event.UpdateEvent{
					ObjectNew: st.newNodeObj,
					ObjectOld: st.oldNodeObj,
				}, st.q)
				get := st.q.Len()
				if get != st.added {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.added, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.added, get)

			}
		}
		t.Run(st.name, tf)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		nodeObj client.Object
		q       workqueue.RateLimitingInterface
		added   int // the items in queue
	}{
		{
			"wrong runtime Object",
			&corev1.Pod{},
			createQueue(),
			0,
		},
		{
			"delete node within nodepool",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						v1alpha1.LabelCurrentNodePool: "test",
					},
				},
			},
			createQueue(),
			1,
		},
		{
			"delete node not in nodepool",
			&corev1.Node{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			createQueue(),
			0,
		},
	}

	for _, st := range tests {
		st := st
		tf := func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", st.name)
			{
				enqueueNodePoolForNode.Delete(event.DeleteEvent{
					Object: st.nodeObj,
				}, st.q)
				get := st.q.Len()
				if get != st.added {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, st.added, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, st.added, get)

			}
		}
		t.Run(st.name, tf)
	}

}
