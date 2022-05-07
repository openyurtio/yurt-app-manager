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

package uniteddeployment

import (
	"context"
	"fmt"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/test/e2e/util"
	ycfg "github.com/openyurtio/yurt-app-manager/test/e2e/yurtconfig"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("uniteddeployment test", func() {
	ctx := context.Background()
	k8sClient := ycfg.YurtE2eCfg.KubeClient
	var namespaceName string
	//var app v1alpha1.UnitedDeployment

	bjNpName := "beijing"
	hzNpName := "hangzhou"
	poolToNodesMap := make(map[string]sets.String)
	poolToNodesMap[bjNpName] = sets.NewString("kind-worker")
	poolToNodesMap[hzNpName] = sets.NewString("kind-worker2")
	nodeToPoolMap := make(map[string]string)
	for k, v := range poolToNodesMap {
		for _, n := range v.List() {
			nodeToPoolMap[n] = k
		}
	}

	createNamespace := func() {
		ns := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		}
		Eventually(
			func() error {
				return k8sClient.Delete(ctx, &ns, client.PropagationPolicy(metav1.DeletePropagationForeground))
			},
			time.Second*120, time.Millisecond*500).Should(SatisfyAny(BeNil(), &util.NotFoundMatcher{}))
		By("make sure all the resources are removed")

		res := &corev1.Namespace{}
		Eventually(
			func() error {
				return k8sClient.Get(ctx, client.ObjectKey{
					Name: namespaceName,
				}, res)
			},
			time.Second*120, time.Millisecond*500).Should(&util.NotFoundMatcher{})
		Eventually(
			func() error {
				return k8sClient.Create(ctx, &ns)
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))
	}

	removeNodeLabel := func() {
		nodes := &corev1.NodeList{}
		Eventually(
			func() error {
				return k8sClient.List(ctx, nodes)
			},
			time.Second*3, time.Millisecond*300).Should(BeNil())

		for _, originNode := range nodes.Items {

			newNode := originNode.DeepCopy()
			if newNode.Labels != nil {
				for k, _ := range newNode.Labels {
					if k == "openyurt.io/is-edge-worker" {
						delete(newNode.Labels, "openyurt.io/is-edge-worke")
					}
					if k == "apps.openyurt.io/desired-nodepool" {
						delete(newNode.Labels, "apps.openyurt.io/desired-nodepool")
					}
				}
			}
			Eventually(func() error {
				return k8sClient.Patch(context.TODO(), newNode, client.MergeFrom(&originNode))
			},
				time.Second*3, time.Millisecond*300).Should(BeNil())
		}

		Eventually(
			func() error {
				np := &v1alpha1.NodePool{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Name: bjNpName}, np); err != nil {
					return err
				}
				if len(np.Status.Nodes) != 0 {
					return fmt.Errorf("nodepool not reconciled")
				}
				return nil
			},
			time.Second*60, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.NotFoundMatcher{}))

		Eventually(
			func() error {
				np := &v1alpha1.NodePool{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Name: hzNpName}, np); err != nil {
					return err
				}
				if len(np.Status.Nodes) != 0 {
					return fmt.Errorf("nodepool not reconciled")
				}
				return nil
			},
			time.Second*60, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.NotFoundMatcher{}))
	}

	labelNodes := func() {

		nodes := &corev1.NodeList{}
		Eventually(
			func() error {
				return k8sClient.List(ctx, nodes)
			},
			time.Second*3, time.Millisecond*300).Should(BeNil())

		for _, originNode := range nodes.Items {
			nodeLabels := originNode.Labels
			if nodeLabels == nil {
				nodeLabels = map[string]string{}
			}
			switch originNode.Name {
			case "kind-control-plane":
				nodeLabels["openyurt.io/is-edge-worker"] = "false"
			case "kind-worker":
				nodeLabels["openyurt.io/is-edge-worker"] = "true"
			case "kind-worker2":
				nodeLabels["openyurt.io/is-edge-worker"] = "true"
			}

			newNode := originNode.DeepCopy()
			newNode.Labels = nodeLabels
			Eventually(
				func() error {
					return k8sClient.Patch(context.TODO(), newNode, client.MergeFrom(&originNode))
				},
				time.Second*3, time.Millisecond*300).Should(BeNil())
		}
	}

	addNodeToNodePool := func() {

		Eventually(
			func() error {
				return k8sClient.Delete(ctx, &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: bjNpName}})
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.NotFoundMatcher{}))

		Eventually(
			func() error {
				return k8sClient.Delete(ctx, &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "hangzhou"}})
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.NotFoundMatcher{}))

		bjNp := &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: bjNpName,
			},
			Spec: v1alpha1.NodePoolSpec{
				Type: v1alpha1.Cloud,
			},
		}

		Eventually(
			func() error {
				return k8sClient.Create(ctx, bjNp)
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))

		hzNp := &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: hzNpName,
			},
			Spec: v1alpha1.NodePoolSpec{
				Type:        v1alpha1.Edge,
				Annotations: map[string]string{"apps.openyurt.io/example": "test-hangzhou"},
				Labels:      map[string]string{"apps.openyurt.io/example": "test-hangzhou"},
				Taints: []corev1.Taint{corev1.Taint{
					Key:    "apps.openyurt.io/example",
					Value:  "test-hangzhou",
					Effect: "NoSchedule",
				}},
			},
		}

		Eventually(
			func() error {
				return k8sClient.Create(ctx, hzNp)
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))

		nodes := &corev1.NodeList{}

		Eventually(
			func() error {
				return k8sClient.List(ctx, nodes)
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil()))

		for _, originNode := range nodes.Items {
			newNode := originNode.DeepCopy()
			nodeLabels := newNode.Labels
			if nodeLabels == nil {
				nodeLabels = map[string]string{}
			}
			nodeLabels["apps.openyurt.io/desired-nodepool"] = nodeToPoolMap[originNode.Name]
			newNode.Labels = nodeLabels
			Eventually(
				func() error {
					return k8sClient.Patch(ctx, newNode, client.MergeFrom(&originNode))
				},
				time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil()))
		}

		Eventually(
			func() error {
				np := &v1alpha1.NodePool{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Name: bjNpName}, np); err != nil {
					return err
				}
				npNodes := sets.NewString(np.Status.Nodes...)
				if !npNodes.Equal(poolToNodesMap[bjNpName]) {
					return fmt.Errorf("node controller not reconciled")
				}
				return nil
			},
			time.Second*60, time.Millisecond*300).Should(BeNil())
		Eventually(
			func() error {
				np := &v1alpha1.NodePool{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Name: hzNpName}, np); err != nil {
					return err
				}
				npNodes := sets.NewString(np.Status.Nodes...)
				if !npNodes.Equal(poolToNodesMap[hzNpName]) {
					return fmt.Errorf("node controller not reconciled")
				}
				return nil
			},
			time.Second*60, time.Millisecond*300).Should(BeNil())

	}

	topologyTest := func() {
		appName := "test-np"
		Eventually(
			func() error {
				return k8sClient.Delete(ctx, &v1alpha1.UnitedDeployment{ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: namespaceName}})
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.NotFoundMatcher{}))

		testLabel := map[string]string{"app": appName}
		var bjPatchTemplate = func() *runtime.RawExtension {
			yamlStr := `
        spec:
          template:
            spec:
              containers:
                - name: nginx
                  image: nginx:1.19.0
`
			b, _ := yaml.YAMLToJSON([]byte(yamlStr))
			return &runtime.RawExtension{Raw: b}
		}

		testNp := &v1alpha1.UnitedDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespaceName,
				Name:      appName,
			},
			Spec: v1alpha1.UnitedDeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: testLabel},
				WorkloadTemplate: v1alpha1.WorkloadTemplate{
					DeploymentTemplate: &v1alpha1.DeploymentTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: testLabel},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: testLabel,
								},
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{corev1.Container{
										Name:  "nginx",
										Image: "nginx:1.19.3",
									}},
								},
							},
						},
					},
				},
				Topology: v1alpha1.Topology{
					Pools: []v1alpha1.Pool{
						{Name: "beijing", NodeSelectorTerm: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "apps.openyurt.io/nodepool",
									Operator: "In",
									Values:   []string{"beijing"},
								},
							},
						},
							Replicas: pointer.Int32Ptr(1),
							Patch:    bjPatchTemplate(),
						},
						{Name: "hangzhou", NodeSelectorTerm: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "apps.openyurt.io/nodepool",
									Operator: "In",
									Values:   []string{"hangzhou"},
								},
							},
						},
							Replicas: pointer.Int32Ptr(2),
							Tolerations: []corev1.Toleration{
								{
									Effect:   "NoSchedule",
									Key:      "apps.openyurt.io/example",
									Operator: "Exists",
								},
							},
						},
					},
				},
			},
		}

		Eventually(
			func() error {
				return k8sClient.Create(ctx, testNp)
			},
			time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))

		Eventually(func() error {
			testPods := &corev1.PodList{}
			if err := k8sClient.List(ctx, testPods, client.InNamespace(namespaceName), client.MatchingLabels{"apps.openyurt.io/pool-name": bjNpName}); err != nil {
				return err
			}
			if len(testPods.Items) != 1 {
				return fmt.Errorf("not reconcile")
			}
			return nil
		}, time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil()))

		Eventually(func() error {
			testPods := &corev1.PodList{}
			if err := k8sClient.List(ctx, testPods, client.InNamespace(namespaceName), client.MatchingLabels{"apps.openyurt.io/pool-name": hzNpName}); err != nil {
				return err
			}
			if len(testPods.Items) != 2 {
				return fmt.Errorf("not reconcile")
			}
			return nil
		}, time.Second*3, time.Millisecond*300).Should(SatisfyAny(BeNil()))
	}

	BeforeEach(func() {
		By("Start to run a test, clean up previous resources")
		namespaceName = "uniteddeployment-e2e-test" + "-" + rand.String(4)
		removeNodeLabel()
		createNamespace()
		labelNodes()
		addNodeToNodePool()
	})

	AfterEach(func() {
		By("Clean up resources after a test")
		By(fmt.Sprintf("Delete the entire namespaceName %s", namespaceName))
		Expect(k8sClient.Delete(ctx,&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}, client.PropagationPolicy(metav1.DeletePropagationBackground))).Should(BeNil())
		removeNodeLabel()
	})

	It("Test unitedDeployment topology", func() {
		By("Run unitedDeployment topology test")
		topologyTest()
	})

})
