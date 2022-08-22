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

package yurtappdaemon

import (
	"testing"
	"time"

	"k8s.io/client-go/util/workqueue"
)

func createQueue() workqueue.RateLimitingInterface {
	return workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(1*time.Millisecond, 1*time.Second))
}

func TestAddYurtAppDaemonToWorkQueue(t *testing.T) {
	tests := []struct {
		namespace string
		name      string
		q         workqueue.RateLimitingInterface
		added     int // the items in queue
	}{
		{
			"default",
			"test",
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
				addYurtAppDaemonToWorkQueue(st.namespace, st.name, st.q)
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
