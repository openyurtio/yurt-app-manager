/*
Copyright 2023 The OpenYurt Authors.

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

package info

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

const (
	OTALatestManifestAnnotation = "openyurt.io/ota-latest-version"
)

// StaticPodInfo is a structure that stores some information used by static pods to upgrade.
type StaticPodInfo struct {
	// Static pod running on the node
	StaticPod *corev1.Pod
	// Indicate whether the static pod exists on the node
	// TODO: This happens when the upgrade worker pod finishes upgrading the static pod manifests.
	// TODO: But the new manifests has errors and Kubelet won't create a new static pod after deleting the old one.
	HasStaticPod bool

	// Upgrade worker pod running on the node
	WorkerPod *corev1.Pod
	// Indicate whether a work pod exists on the node
	HasWorkerPod bool

	// Indicate whether the static pod on the node needs to be upgraded.
	// If true, the static pod is not up-to-date and needs to be upgraded.
	UpgradeNeeded bool

	// Indicate whether the worker pod on the node is up-to-date.
	// If true, then the upgrade operation is in progress and does not
	// need to create a new worker pod.
	UpgradeExecuting bool

	// Indicate whether the node is ready. It's used in Auto mode.
	Ready bool
}

// ConstructStaticPodsUpgradeInfoList constructs the upgrade information for nodes which have the target static pod
func ConstructStaticPodsUpgradeInfoList(c client.Client, instance *appsv1alpha1.StaticPod, workerPodName string) (map[string]*StaticPodInfo, error) {
	spi := make(map[string]*StaticPodInfo)

	// upgrade worker pod is default in "kube-system" namespace which may be different with target static pod's namespace
	var staticPodList, workerPodList corev1.PodList
	if err := c.List(context.TODO(), &staticPodList, &client.ListOptions{Namespace: instance.Spec.Namespace}); err != nil {
		return nil, err
	}
	if err := c.List(context.TODO(), &workerPodList, &client.ListOptions{Namespace: metav1.NamespaceSystem}); err != nil {
		return nil, err
	}

	for i, pod := range staticPodList.Items {
		node := pod.Spec.NodeName
		if node == "" {
			continue
		}

		// The name format of mirror static pod is `StaticPodName-NodeName`
		if instance.Spec.StaticPodName+"-"+node == pod.Name && isStaticPod(&pod) {
			if info := spi[node]; info == nil {
				spi[node] = &StaticPodInfo{
					HasWorkerPod: false,
				}
			}
			spi[node].HasStaticPod = true
			spi[node].StaticPod = &staticPodList.Items[i]
		}
	}

	for i, pod := range workerPodList.Items {
		node := pod.Spec.NodeName
		if node == "" {
			continue
		}

		// The name format of worker pods are `WorkerPodName-NodeName-Hash` Todo: may lead to mismatch
		if strings.Contains(pod.Name, workerPodName) {
			if info := spi[node]; info == nil {
				spi[node] = &StaticPodInfo{
					HasStaticPod: false,
				}
			}
			spi[node].HasWorkerPod = true
			spi[node].WorkerPod = &workerPodList.Items[i]
		}
	}

	return spi, nil
}

// isStaticPod judges whether a pod is static by its OwnerReference
func isStaticPod(pod *corev1.Pod) bool {
	for _, ownerRef := range pod.GetOwnerReferences() {
		if ownerRef.Kind == "Node" {
			return true
		}
	}
	return false
}

// ReadyUpgradeWaitingNodes gets those nodes that satisfied
// 1. node is ready
// 2. node needs to be upgraded
// 3. no latest worker pod running on the node
// On these nodes, new worker pods need to be created for auto mode
func ReadyUpgradeWaitingNodes(staticPodInfoList map[string]*StaticPodInfo) []string {
	var nodes []string
	for node, info := range staticPodInfoList {
		if info.UpgradeNeeded && !info.UpgradeExecuting && info.Ready {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func OTAReadyUpgradeWaitingNodes(staticPodInfoList map[string]*StaticPodInfo, hash string) []string {
	var nodes []string
	for node, info := range staticPodInfoList {
		if info.StaticPod != nil && info.StaticPod.Annotations[OTALatestManifestAnnotation] == hash {
			continue
		}

		if info.UpgradeNeeded && !info.UpgradeExecuting && info.Ready {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// ReadyNodes gets nodes that are ready
func ReadyNodes(staticPodInfoList map[string]*StaticPodInfo) []string {
	var nodes []string
	for node, info := range staticPodInfoList {
		if info.Ready {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// UpgradeNeededNodes gets nodes that are not running the latest static pods
func UpgradeNeededNodes(staticPodInfoList map[string]*StaticPodInfo) []string {
	var nodes []string
	for node, info := range staticPodInfoList {
		if info.UpgradeNeeded {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// UpgradedNodes gets nodes that are running the latest static pods
func UpgradedNodes(staticPodInfoList map[string]*StaticPodInfo) []string {
	var nodes []string
	for node, info := range staticPodInfoList {
		if !info.UpgradeNeeded {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
