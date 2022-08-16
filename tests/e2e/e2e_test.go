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

package e2e

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	nodeutil "k8s.io/kubernetes/pkg/controller/util/node"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	_ "github.com/openyurtio/yurt-app-manager/tests/e2e/nodepool"
	_ "github.com/openyurtio/yurt-app-manager/tests/e2e/yurtappset"
	"github.com/openyurtio/yurt-app-manager/tests/e2e/yurtconfig"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appsv1alpha1.AddToScheme(scheme))
}

var ReportDir = flag.String("report-dir", "", "Path to the directory where the JUnit XML reports should be saved. Default is empty, which doesn't generate these reports.")

func handleFlags() {
	flag.Parse()
}

func CheckYurtArg() bool {
	return true
}

func PreCheckOk() (bool, error) {
	c := yurtconfig.YurtE2eCfg.KubeClient
	nodes := &apiv1.NodeList{}
	err := c.List(context.Background(), nodes)
	if err != nil {
		klog.Errorf("pre_check_get_nodes failed errmsg:%v", err)
		return false, client.IgnoreNotFound(err)
	}

	for _, node := range nodes.Items {
		_, readyCondition := nodeutil.GetNodeCondition(&node.Status, apiv1.NodeReady)
		if readyCondition == nil || readyCondition.Status != apiv1.ConditionTrue {
			klog.Warningf("node not ready, name: %s, status: %v, reason: %s", node.Name, readyCondition.Status, readyCondition.Reason)
			return false, nil
		}
	}

	pods := &apiv1.PodList{}
	if err := c.List(context.TODO(), pods, client.InNamespace("kube-system")); err != nil {
		return false, client.IgnoreNotFound(err)
	}

	for _, tmpPod := range pods.Items {
		if tmpPod.Status.Phase != apiv1.PodRunning {
			klog.Warningf("pod not ready, name: %s, status: %s, reason: %s", tmpPod.Name, tmpPod.Status.Phase, tmpPod.Status.Reason)
			return false, nil
		}
	}

	return true, nil
}

func SetYurtE2eCfg() error {
	k8sClient, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	yurtconfig.YurtE2eCfg.KubeClient = k8sClient
	yurtconfig.YurtE2eCfg.ReportDir = *ReportDir
	return nil
}

func TestMain(m *testing.M) {
	handleFlags()

	if !CheckYurtArg() {
		os.Exit(-1)
	}

	if err := SetYurtE2eCfg(); err != nil {
		os.Exit(-1)
	}

	if err := wait.Poll(1*time.Second, 300*time.Second,
		func() (bool, error) {
			if paas, err := PreCheckOk(); !paas {
				return false, err
			}
			return true, nil
		}); err != nil {
		os.Exit(-1)
	}

	//framework.AfterReadingAllFlags(&framework.TestContext)
	rand.Seed(time.Now().UnixNano())

	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}
