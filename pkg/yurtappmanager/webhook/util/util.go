/*
Copyright 2020 The OpenYurt Authors.
Copyright 2020 The Kruise Authors.

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

package util

import (
	"os"
	"strconv"

	"k8s.io/klog"
)

func GetHost() string {
	return os.Getenv("WEBHOOK_HOST")
}

func GetNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); len(ns) > 0 {
		return ns
	}
	return "kube-system"
}

func GetSecretName() string {
	if name := os.Getenv("SECRET_NAME"); len(name) > 0 {
		return name
	}
	return "yurt-app-webhook-certs"
}

func GetServiceName() string {
	if name := os.Getenv("SERVICE_NAME"); len(name) > 0 {
		return name
	}
	return "yurt-app-webhook-service"
}

func GetPort() int {
	port := 9876
	if p := os.Getenv("WEBHOOK_PORT"); len(p) > 0 {
		if p, err := strconv.ParseInt(p, 10, 32); err == nil {
			port = int(p)
		} else {
			klog.Fatalf("failed to convert WEBHOOK_PORT=%v in env: %v", p, err)
		}
	}
	return port
}

func GetCertDir() string {
	if p := os.Getenv("WEBHOOK_CERT_DIR"); len(p) > 0 {
		return p
	}
	return "/tmp/yurt-app-webhook-certs"
}

func GetMutatingWebhookConfigurationName() string {
	if p := os.Getenv("MUTATING_WEBHOOK_CONFIGURATION_NAME"); len(p) > 0 {
		return p
	}
	return "yurt-app-mutating-webhook-configuration"
}

func GetValidatingWebhookConfigurationName() string {
	if p := os.Getenv("VALIDATING_WEBHOOK_CONFIGURATION_NAME"); len(p) > 0 {
		return p
	}
	return "yurt-app-validating-webhook-configuration"
}
