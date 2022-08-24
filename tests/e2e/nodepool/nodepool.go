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

package nodepool

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/tests/e2e/util"
	ycfg "github.com/openyurtio/yurt-app-manager/tests/e2e/yurtconfig"
)

var _ = Describe("nodepool test", func() {
	ctx := context.Background()
	k8sClient := ycfg.YurtE2eCfg.KubeClient
	poolToNodesMap := make(map[string]sets.String)

	testNodePoolStatus := func() error {
		nps := &v1alpha1.NodePoolList{}
		if err := k8sClient.List(ctx, nps); err != nil {
			return err
		}
		for _, tmp := range nps.Items {
			if int(tmp.Status.ReadyNodeNum) != poolToNodesMap[tmp.Name].Len() {
				return errors.New("node size not match")
			}
		}
		return nil
	}

	BeforeEach(func() {
		By("Start to run nodepool test, clean up previous resources")
		util.CleanupNodeLabel(ctx, k8sClient)
		util.CleanupNodePool(ctx, k8sClient)
	})

	AfterEach(func() {
		By("Clean up resources after a test")
		util.CleanupNodeLabel(ctx, k8sClient)
		util.CleanupNodePool(ctx, k8sClient)
	})

	It("Test NodePool empty", func() {
		By("Run noolpool empty")
		poolToNodesMap = map[string]sets.String{}
		Eventually(
			func() error {
				return util.InitNodeAndNodePool(ctx, k8sClient, poolToNodesMap)
			},
			time.Second*5, time.Millisecond*500).Should(SatisfyAny(BeNil()))

		Eventually(
			func() error {
				return testNodePoolStatus()
			},
			time.Second*5, time.Millisecond*500).Should(SatisfyAny(BeNil()))
	})

	It("Test NodePool Size One", func() {
		By("Run nodepool size one")

		bjNpName := "beijing"
		poolToNodesMap[bjNpName] = sets.NewString("kind-control-plane")
		Eventually(
			func() error {
				return util.InitNodeAndNodePool(ctx, k8sClient, poolToNodesMap)
			},
			time.Second*5, time.Millisecond*500).Should(SatisfyAny(BeNil()))

		Eventually(
			func() error {
				return testNodePoolStatus()
			},
			time.Second*5, time.Millisecond*500).Should(SatisfyAny(BeNil()))
	})

	It("Test NodePool With Multi Nodes", func() {
		poolToNodesMap["beijing"] = sets.NewString("kind-control-plane")
		poolToNodesMap["hangzhou"] = sets.NewString("kind-worker", "kind-worker2")

		Eventually(
			func() error {
				return util.InitNodeAndNodePool(ctx, k8sClient, poolToNodesMap)
			},
			time.Second*5, time.Millisecond*500).Should(SatisfyAny(BeNil()))

		Eventually(
			func() error {
				return testNodePoolStatus()
			},
			time.Second*5, time.Millisecond*500).Should(SatisfyAny(BeNil()))
	})

})
