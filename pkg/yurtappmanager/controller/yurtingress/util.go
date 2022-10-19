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
	"context"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/runner"
)

func NewDefaultRunner(ctx context.Context, releaseNamespace string) (*runner.Runner, error) {
	getter, err := buildRESTClientGetter(releaseNamespace)
	if err != nil {
		return nil, err
	}

	log := ctrl.LoggerFrom(ctx)
	run, err := runner.NewRunner(getter, releaseNamespace, log)
	if err != nil {
		return nil, err
	}
	return run, nil
}

func buildRESTClientGetter(namespace string) (genericclioptions.RESTClientGetter, error) {
	cfg, err := ctrl.GetConfig()
	flags := genericclioptions.NewConfigFlags(false)
	flags.APIServer = pointer.String(cfg.Host)
	flags.BearerToken = pointer.String(cfg.BearerToken)
	flags.CAFile = pointer.String(cfg.CAFile)
	flags.Namespace = pointer.String(namespace)
	if sa := cfg.Impersonate.UserName; sa != "" {
		flags.Impersonate = pointer.String(sa)
	}
	flags.CacheDir = nil

	return flags, err
}

// MergeMaps merges map b into given map a and returns the result.
// It allows overwrites of map values with flat values, and vice versa.
//
// Originally copied over from https://github.com/helm/helm/blob/v3.3.0/pkg/cli/values/options.go#L88.
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
