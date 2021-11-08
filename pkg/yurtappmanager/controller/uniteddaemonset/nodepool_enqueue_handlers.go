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

package uniteddaemonset

import (
	"context"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EnqueueUnitedDaemonsetForNodePool struct {
	client client.Client
}

func (e *EnqueueUnitedDaemonsetForNodePool) Create(event event.CreateEvent, limitingInterface workqueue.RateLimitingInterface) {
	e.addAllUnitedDaemonsetToWorkQueue(limitingInterface)
}

func (e *EnqueueUnitedDaemonsetForNodePool) Update(event event.UpdateEvent, limitingInterface workqueue.RateLimitingInterface) {
	e.addAllUnitedDaemonsetToWorkQueue(limitingInterface)
}

func (e *EnqueueUnitedDaemonsetForNodePool) Delete(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	e.addAllUnitedDaemonsetToWorkQueue(limitingInterface)
}

func (e *EnqueueUnitedDaemonsetForNodePool) Generic(event event.GenericEvent, limitingInterface workqueue.RateLimitingInterface) {
	return
}

func (e *EnqueueUnitedDaemonsetForNodePool) addAllUnitedDaemonsetToWorkQueue(limitingInterface workqueue.RateLimitingInterface) {
	udds := &v1alpha1.YurtAppDaemonList{}
	if err := e.client.List(context.TODO(), udds); err != nil {
		return
	}

	for _, ud := range udds.Items {
		addUnitedDaemonsetToWorkQueue(ud.GetNamespace(), ud.GetName(), limitingInterface)
	}
}

var _ handler.EventHandler = &EnqueueUnitedDaemonsetForNodePool{}

// addUnitedDaemonsetToWorkQueue adds the unitedDaemonset the reconciler's workqueue
func addUnitedDaemonsetToWorkQueue(namespace, name string,
	q workqueue.RateLimitingInterface) {
	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: namespace},
	})
}
