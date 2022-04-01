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

package kubernetes

import (
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/constant"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateNginxIngressCommonResource(client client.Client) error {
	// 1. Create Namespace
	if err := CreateNamespaceFromYaml(client, constant.NginxIngressControllerNamespace); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Create ClusterRole
	if err := CreateClusterRoleFromYaml(client, constant.NginxIngressControllerClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateClusterRoleFromYaml(client, constant.NginxIngressAdmissionWebhookClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Create ClusterRoleBinding
	if err := CreateClusterRoleBindingFromYaml(client,
		constant.NginxIngressControllerClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateClusterRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Create Role
	if err := CreateRoleFromYaml(client,
		constant.NginxIngressControllerRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateRoleFromYaml(client,
		constant.NginxIngressAdmissionWebhookRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Create RoleBinding
	if err := CreateRoleBindingFromYaml(client,
		constant.NginxIngressControllerRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 6. Create ServiceAccount
	if err := CreateServiceAccountFromYaml(client,
		constant.NginxIngressControllerServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateServiceAccountFromYaml(client,
		constant.NginxIngressAdmissionWebhookServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 7. Create Configmap
	if err := CreateConfigMapFromYaml(client,
		constant.NginxIngressControllerConfigMap); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func DeleteNginxIngressCommonResource(client client.Client) error {
	// 1. Delete Configmap
	if err := DeleteConfigMapFromYaml(client,
		constant.NginxIngressControllerConfigMap); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Delete ServiceAccount
	if err := DeleteServiceAccountFromYaml(client,
		constant.NginxIngressControllerServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteServiceAccountFromYaml(client,
		constant.NginxIngressAdmissionWebhookServiceAccount); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Delete RoleBinding
	if err := DeleteRoleBindingFromYaml(client,
		constant.NginxIngressControllerRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Delete Role
	if err := DeleteRoleFromYaml(client,
		constant.NginxIngressControllerRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteRoleFromYaml(client,
		constant.NginxIngressAdmissionWebhookRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Delete ClusterRoleBinding
	if err := DeleteClusterRoleBindingFromYaml(client,
		constant.NginxIngressControllerClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteClusterRoleBindingFromYaml(client,
		constant.NginxIngressAdmissionWebhookClusterRoleBinding); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 6. Delete ClusterRole
	if err := DeleteClusterRoleFromYaml(client, constant.NginxIngressControllerClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteClusterRoleFromYaml(client, constant.NginxIngressAdmissionWebhookClusterRole); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 7. Delete Namespace
	if err := DeleteNamespaceFromYaml(client, constant.NginxIngressControllerNamespace); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func CreateNginxIngressSpecificResource(client client.Client, poolname string, externalIPs *[]string, replicas int32, ownerRef *metav1.OwnerReference) error {
	// 1. Create Deployment
	if err := CreateDeployFromYaml(client,
		constant.NginxIngressControllerNodePoolDeployment,
		replicas,
		ownerRef,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateDeployFromYaml(client,
		constant.NginxIngressAdmissionWebhookDeployment,
		1,
		nil,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Create Service
	if err := CreateServiceFromYaml(client,
		constant.NginxIngressControllerService,
		externalIPs,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := CreateServiceFromYaml(client,
		constant.NginxIngressAdmissionWebhookService,
		nil,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Create ValidatingWebhookConfiguration
	if err := CreateValidatingWebhookConfigurationFromYaml(client,
		constant.NginxIngressValidatingWebhookConfiguration,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Create Job
	if err := CreateJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJob,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Create Job Patch
	if err := CreateJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJobPatch,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func DeleteNginxIngressSpecificResource(client client.Client, poolname string, cleanup bool) error {
	// 1. Delete Deployment
	if err := DeleteDeployFromYaml(client,
		constant.NginxIngressControllerNodePoolDeployment,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteDeployFromYaml(client,
		constant.NginxIngressAdmissionWebhookDeployment,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 2. Delete Service
	if err := DeleteServiceFromYaml(client,
		constant.NginxIngressControllerService,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	if err := DeleteServiceFromYaml(client,
		constant.NginxIngressAdmissionWebhookService,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 3. Delete ValidatingWebhookConfiguration
	if err := DeleteValidatingWebhookConfigurationFromYaml(client,
		constant.NginxIngressValidatingWebhookConfiguration,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 4. Delete Job
	if err := DeleteJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJob,
		cleanup,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	// 5. Delete Job Patch
	if err := DeleteJobFromYaml(client,
		constant.NginxIngressAdmissionWebhookJobPatch,
		cleanup,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func ScaleNginxIngressControllerDeploymment(client client.Client, poolname string, replicas int32) error {
	if err := UpdateDeployFromYaml(client,
		constant.NginxIngressControllerNodePoolDeployment,
		&replicas,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}

func UpdateNginxServiceExternalIPs(client client.Client, poolname string, externalIPs []string) error {
	if err := UpdateServiceFromYaml(client,
		constant.NginxIngressControllerService,
		&externalIPs,
		map[string]string{
			"nodepool_name": poolname}); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	return nil
}
