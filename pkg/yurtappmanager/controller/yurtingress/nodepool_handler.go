/*
Copyright 2022 The OpenYurt Authors.

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
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
)

var _ handler.EventHandler = &EnqueueRequestForNodePool{}

type EnqueueRequestForNodePool struct {
	client.Client
}

func (n *EnqueueRequestForNodePool) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	if np, ok := e.Object.(*v1beta1.NodePool); !ok {
		return
	} else {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: np.Name,
			},
		})
	}
}

func (n *EnqueueRequestForNodePool) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	newNodePool, oldNodePool := e.ObjectNew.(*v1beta1.NodePool), e.ObjectOld.(*v1beta1.NodePool)
	if !isNodePoolUpdated(newNodePool, oldNodePool) {
		return
	}
	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: newNodePool.Name,
		},
	})
}

func (n *EnqueueRequestForNodePool) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if np, ok := e.Object.(*v1beta1.NodePool); !ok {
		return
	} else {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: np.Name,
			},
		})
	}
}

func (n *EnqueueRequestForNodePool) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
}

func isNodePoolUpdated(newNodePool *v1beta1.NodePool, oldNodePool *v1beta1.NodePool) bool {
	if newNodePool == nil || oldNodePool == nil {
		return false
	}
	return !reflect.DeepEqual(oldNodePool.Spec, newNodePool.Spec)
}
