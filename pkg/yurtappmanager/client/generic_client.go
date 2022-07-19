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

package client

import (
	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	yurtappclientset "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/client/clientset/versioned"
)

// GenericClientset defines a generic client
type GenericClientset struct {
	KubeClient    kubeclientset.Interface
	YurtappClient yurtappclientset.Interface
}

// NewForConfig creates a new Clientset for the given config.
func newForConfig(c *rest.Config) (*GenericClientset, error) {
	kubeClient, err := kubeclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	yurtClient, err := yurtappclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &GenericClientset{
		KubeClient:    kubeClient,
		YurtappClient: yurtClient,
	}, nil
}
