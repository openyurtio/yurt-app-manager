/*
Copyright 2022 The OpenYurt authors.
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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/controller/uniteddeployment/adapter"
)

var (
	one32 int32 = 1
	one64 int64 = 1
)

var _ = Describe("Convert to pool", func() {

	var (
		deployment = appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "deployment",
				Annotations: map[string]string{
					alpha1.AnnotationPatchKey: "deployment patch",
				},
				Labels: map[string]string{
					alpha1.PoolNameLabelKey: "hangzhou",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &one32,
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration: one64,
				ReadyReplicas:      one32,
			},
		}

		statefulset = appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "statefulset",
				Annotations: map[string]string{
					alpha1.AnnotationPatchKey: "statefulset patch",
				},
				Labels: map[string]string{
					alpha1.PoolNameLabelKey: "beijing",
				},
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &one32,
			},
			Status: appsv1.StatefulSetStatus{
				ObservedGeneration: one64,
				ReadyReplicas:      one32,
			},
		}

		wantDeployPool = Pool{
			Name:      "hangzhou",
			Namespace: "deployment",
			Spec: PoolSpec{
				PoolRef: &deployment,
			},
			Status: PoolStatus{
				ObservedGeneration: one64,
				ReplicasInfo: adapter.ReplicasInfo{
					Replicas:      one32,
					ReadyReplicas: one32,
				},
				PatchInfo: "deployment patch",
			},
		}

		wantStsPool = Pool{
			Name:      "beijing",
			Namespace: "statefulset",
			Spec: PoolSpec{
				PoolRef: &statefulset,
			},
			Status: PoolStatus{
				ObservedGeneration: one64,
				ReplicasInfo: adapter.ReplicasInfo{
					Replicas:      one32,
					ReadyReplicas: one32,
				},
				PatchInfo: "statefulset patch",
			},
		}
	)

	deployPoolCtrl := &PoolControl{adapter: &adapter.DeploymentAdapter{}}
	stsPoolCtrl := &PoolControl{adapter: &adapter.StatefulSetAdapter{}}
	It("convert deployment set to pool", func() {
		deployPool, err := deployPoolCtrl.convertToPool(&deployment)
		Expect(err).ToNot(HaveOccurred())
		Expect(*deployPool).To(Equal(wantDeployPool))
	})

	It("convert statefulset set to pool", func() {
		stsPool, err := stsPoolCtrl.convertToPool(&statefulset)
		Expect(err).ToNot(HaveOccurred())
		Expect(*stsPool).To(Equal(wantStsPool))
	})
})
