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
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var deleteOptions *client.DeleteOptions

func init() {
	policy := metav1.DeletePropagationForeground
	o := &client.DeleteOptions{PropagationPolicy: &policy}
	deleteOptions = &client.DeleteOptions{}
	o.ApplyToDelete(deleteOptions)
}

// CreateNamespaceFromYaml creates the Namespace from the yaml template.
func CreateNamespaceFromYaml(cli client.Client, crTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(crTmpl))
	if err != nil {
		return err
	}
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return fmt.Errorf("fail to assert namespace")
	}
	ns.ObjectMeta.SetOwnerReferences(ownerRefs)

	err = cli.Create(context.Background(), ns)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the namespace/%s: %v", ns.Name, err)
		}
		if apierrors.IsAlreadyExists(err) && ns.Status.Phase != corev1.NamespaceActive {
			return fmt.Errorf("namespace/%s is not active: %v", ns.Name, err)
		}
	}
	time.Sleep(time.Second)
	klog.V(4).Infof("namespace/%s is created", ns.Name)
	return nil
}

// DeleteNamespaceFromYaml deletes the Namespace from the yaml template.
func DeleteNamespaceFromYaml(client client.Client, crTmpl string) error {
	obj, err := YamlToObject([]byte(crTmpl))
	if err != nil {
		return err
	}
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return fmt.Errorf("fail to assert namespace")
	}
	err = client.Delete(context.Background(), ns)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the namespace/%s: %v", ns.Name, err)
		}
	}
	klog.V(4).Infof("namespace/%s is deleted", ns.Name)
	return nil
}

// CreateClusterRoleFromYaml creates the ClusterRole from the yaml template.
func CreateClusterRoleFromYaml(client client.Client, crTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(crTmpl))
	if err != nil {
		return err
	}
	cr, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		return fmt.Errorf("fail to assert clusterrole")
	}
	cr.ObjectMeta.SetOwnerReferences(ownerRefs)
	err = client.Create(context.Background(), cr)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the clusterrole/%s: %v", cr.Name, err)
		}
	}
	klog.V(4).Infof("clusterrole/%s is created", cr.Name)
	return nil
}

// DeleteClusterRoleFromYaml deletes the ClusterRole from the yaml template.
func DeleteClusterRoleFromYaml(client client.Client, crTmpl string) error {
	obj, err := YamlToObject([]byte(crTmpl))
	if err != nil {
		return err
	}
	cr, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		return fmt.Errorf("fail to assert clusterrole")
	}
	err = client.Delete(context.Background(), cr)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the clusterrole/%s: %v", cr.Name, err)
		}
	}
	klog.V(4).Infof("clusterrole/%s is deleted", cr.Name)
	return nil
}

// CreateClusterRoleBindingFromYaml creates the ClusterRoleBinding from the yaml template.
func CreateClusterRoleBindingFromYaml(client client.Client, crbTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(crbTmpl))
	if err != nil {
		return err
	}
	crb, ok := obj.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return fmt.Errorf("fail to assert clusterrolebinding")
	}
	crb.ObjectMeta.SetOwnerReferences(ownerRefs)
	err = client.Create(context.Background(), crb)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the clusterrolebinding/%s: %v", crb.Name, err)
		}
	}
	klog.V(4).Infof("clusterrolebinding/%s is created", crb.Name)
	return nil
}

// DeleteClusterRoleBindingFromYaml deletes the ClusterRoleBinding from the yaml template.
func DeleteClusterRoleBindingFromYaml(client client.Client, crbTmpl string) error {
	obj, err := YamlToObject([]byte(crbTmpl))
	if err != nil {
		return err
	}
	crb, ok := obj.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return fmt.Errorf("fail to assert clusterrolebinding")
	}
	err = client.Delete(context.Background(), crb)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the clusterrolebinding/%s: %v", crb.Name, err)
		}
	}
	klog.V(4).Infof("clusterrolebinding/%s is deleted", crb.Name)
	return nil
}

// CreateRoleFromYaml creates the Role from the yaml template.
func CreateRoleFromYaml(client client.Client, rTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(rTmpl))
	if err != nil {
		return err
	}
	r, ok := obj.(*rbacv1.Role)
	if !ok {
		return fmt.Errorf("fail to assert role")
	}
	r.ObjectMeta.SetOwnerReferences(ownerRefs)
	err = client.Create(context.Background(), r)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the role/%s: %v", r.Name, err)
		}
	}
	klog.V(4).Infof("role/%s is created", r.Name)
	return nil
}

// DeleteRoleFromYaml deletes the Role from the yaml template.
func DeleteRoleFromYaml(client client.Client, rTmpl string) error {
	obj, err := YamlToObject([]byte(rTmpl))
	if err != nil {
		return err
	}
	r, ok := obj.(*rbacv1.Role)
	if !ok {
		return fmt.Errorf("fail to assert role")
	}
	err = client.Delete(context.Background(), r)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the role/%s: %v", r.Name, err)
		}
	}
	klog.V(4).Infof("role/%s is deleted", r.Name)
	return nil
}

// CreateRoleBindingFromYaml creates the RoleBinding from the yaml template.
func CreateRoleBindingFromYaml(client client.Client, rbTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(rbTmpl))
	if err != nil {
		return err
	}
	rb, ok := obj.(*rbacv1.RoleBinding)
	if !ok {
		return fmt.Errorf("fail to assert rolebinding")
	}
	rb.ObjectMeta.SetOwnerReferences(ownerRefs)
	err = client.Create(context.Background(), rb)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the rolebinding/%s: %v", rb.Name, err)
		}
	}
	klog.V(4).Infof("rolebinding/%s is created", rb.Name)
	return nil
}

// DeleteRoleBindingFromYaml delete the RoleBinding from the yaml template.
func DeleteRoleBindingFromYaml(client client.Client, rbTmpl string) error {
	obj, err := YamlToObject([]byte(rbTmpl))
	if err != nil {
		return err
	}
	rb, ok := obj.(*rbacv1.RoleBinding)
	if !ok {
		return fmt.Errorf("fail to assert rolebinding")
	}
	err = client.Delete(context.Background(), rb)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the rolebinding/%s: %v", rb.Name, err)
		}
	}
	klog.V(4).Infof("rolebinding/%s is deleted", rb.Name)
	return nil
}

// CreateServiceAccountFromYaml creates the ServiceAccount from the yaml template.
func CreateServiceAccountFromYaml(client client.Client, saTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(saTmpl))
	if err != nil {
		return err
	}
	sa, ok := obj.(*corev1.ServiceAccount)
	if !ok {
		return fmt.Errorf("fail to assert serviceaccount")
	}
	sa.ObjectMeta.SetOwnerReferences(ownerRefs)
	err = client.Create(context.Background(), sa)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the serviceaccount/%s: %v", sa.Name, err)
		}
	}
	klog.V(4).Infof("serviceaccount/%s is created", sa.Name)
	return nil
}

// DeleteServiceAccountFromYaml deletes the ServiceAccount from the yaml template.
func DeleteServiceAccountFromYaml(client client.Client, saTmpl string) error {
	obj, err := YamlToObject([]byte(saTmpl))
	if err != nil {
		return err
	}
	sa, ok := obj.(*corev1.ServiceAccount)
	if !ok {
		return fmt.Errorf("fail to assert serviceaccount")
	}
	err = client.Delete(context.Background(), sa)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the serviceaccount/%s: %v", sa.Name, err)
		}
	}
	klog.V(4).Infof("serviceaccount/%s is deleted", sa.Name)
	return nil
}

// CreateConfigMapFromYaml creates the ConfigMap from the yaml template.
func CreateConfigMapFromYaml(client client.Client, cmTmpl string, ownerRefs []metav1.OwnerReference) error {
	obj, err := YamlToObject([]byte(cmTmpl))
	if err != nil {
		return err
	}
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("fail to assert configmap")
	}
	cm.ObjectMeta.SetOwnerReferences(ownerRefs)
	err = client.Create(context.Background(), cm)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the configmap/%s: %v", cm.Name, err)
		}
	}
	klog.V(4).Infof("configmap/%s is created", cm.Name)
	return nil
}

// DeleteConfigMapFromYaml deletes the ConfigMap from the yaml template.
func DeleteConfigMapFromYaml(client client.Client, cmTmpl string) error {
	obj, err := YamlToObject([]byte(cmTmpl))
	if err != nil {
		return err
	}
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return fmt.Errorf("fail to assert configmap")
	}
	err = client.Delete(context.Background(), cm)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the configmap/%s: %v", cm.Name, err)
		}
	}
	klog.V(4).Infof("configmap/%s is deleted", cm.Name)
	return nil
}

// CreateDeployFromYaml creates the Deployment from the yaml template.
func CreateDeployFromYaml(client client.Client, dplyTmpl, image string, replicas int32, ownerRef *metav1.OwnerReference, ctx interface{}) error {
	dp, err := SubsituteTemplate(dplyTmpl, ctx)
	if err != nil {
		return err
	}
	dpObj, err := YamlToObject([]byte(dp))
	if err != nil {
		return err
	}
	dply, ok := dpObj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("fail to assert deployment")
	}
	if ownerRef != nil {
		ownerRefs := dply.ObjectMeta.GetOwnerReferences()
		ownerRefs = append(ownerRefs, *ownerRef)
		dply.ObjectMeta.SetOwnerReferences(ownerRefs)
	}
	dply.Spec.Replicas = &replicas
	if image != "" {
		dply.Spec.Template.Spec.Containers[len(dply.Spec.Template.Spec.Containers)-1].Image = image
	}
	err = client.Create(context.Background(), dply)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the deployment/%s: %v", dply.Name, err)
		}
	}
	klog.V(4).Infof("deployment/%s is created", dply.Name)
	return nil
}

// DeleteDeployFromYaml delete the Deployment from the yaml template.
func DeleteDeployFromYaml(client client.Client, dplyTmpl string, ctx interface{}) error {
	dp, err := SubsituteTemplate(dplyTmpl, ctx)
	if err != nil {
		return err
	}
	dpObj, err := YamlToObject([]byte(dp))
	if err != nil {
		return err
	}
	dply, ok := dpObj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("fail to assert deployment")
	}
	err = client.Delete(context.Background(), dply)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the deployment/%s: %v", dply.Name, err)
		}
	}
	klog.V(4).Infof("deployment/%s is deleted", dply.Name)
	return nil
}

// UpdateDeployFromYaml updates the Deployment from the yaml template.
func UpdateDeployFromYaml(cli client.Client, dplyTmpl, image string, replicas *int32, ctx interface{}) error {
	dp, err := SubsituteTemplate(dplyTmpl, ctx)
	if err != nil {
		return err
	}
	dpObj, err := YamlToObject([]byte(dp))
	if err != nil {
		return err
	}
	dply, ok := dpObj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("fail to assert deployment")
	}
	if cli.Get(context.Background(), client.ObjectKey{Namespace: dply.Namespace, Name: dply.Name}, dply) != nil {
		klog.V(4).Infof("get deployment/%s failed", dply.Name)
		return nil
	}
	dply.Spec.Replicas = replicas
	if image != "" {
		dply.Spec.Template.Spec.Containers[len(dply.Spec.Template.Spec.Containers)-1].Image = image
	}
	err = cli.Update(context.Background(), dply)
	if err != nil {
		return fmt.Errorf("fail to update the deployment/%s: %v", dply.Name, err)
	}
	klog.V(4).Infof("deployment/%s is updated", dply.Name)
	return nil
}

// CreateServiceFromYaml creates the Service from the yaml template.
func CreateServiceFromYaml(client client.Client, svcTmpl string, externalIPs *[]string, ctx interface{}) error {
	sv, err := SubsituteTemplate(svcTmpl, ctx)
	if err != nil {
		return err
	}
	svcObj, err := YamlToObject([]byte(sv))
	if err != nil {
		return err
	}
	svc, ok := svcObj.(*corev1.Service)
	if !ok {
		return fmt.Errorf("fail to assert service")
	}
	if externalIPs != nil {
		svc.Spec.ExternalIPs = *externalIPs
	}
	err = client.Create(context.Background(), svc)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the service/%s: %v", svc.Name, err)
		}
	}
	klog.V(4).Infof("service/%s is created", svc.Name)
	return nil
}

// DeleteServiceFromYaml deletes the Service from the yaml template.
func DeleteServiceFromYaml(client client.Client, svcTmpl string, ctx interface{}) error {
	sv, err := SubsituteTemplate(svcTmpl, ctx)
	if err != nil {
		return err
	}
	svcObj, err := YamlToObject([]byte(sv))
	if err != nil {
		return err
	}
	svc, ok := svcObj.(*corev1.Service)
	if !ok {
		return fmt.Errorf("fail to assert service")
	}
	err = client.Delete(context.Background(), svc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the service/%s: %v", svc.Name, err)
		}
	}
	klog.V(4).Infof("service/%s is deleted", svc.Name)
	return nil
}

// UpdateServiceFromYaml updates the Service from the yaml template.
func UpdateServiceFromYaml(cli client.Client, svcTmpl string, externalIPs *[]string, ctx interface{}) error {
	sv, err := SubsituteTemplate(svcTmpl, ctx)
	if err != nil {
		return err
	}
	svcObj, err := YamlToObject([]byte(sv))
	if err != nil {
		return err
	}
	svc, ok := svcObj.(*corev1.Service)
	if !ok {
		return fmt.Errorf("fail to assert service")
	}
	if cli.Get(context.Background(), client.ObjectKey{Namespace: svc.Namespace, Name: svc.Name}, svc) != nil {
		klog.V(4).Infof("get service/%s failed", svc.Name)
		return nil
	}
	svc.Spec.ExternalIPs = *externalIPs
	err = cli.Update(context.Background(), svc)
	if err != nil {
		return fmt.Errorf("fail to update the service/%s: %v", svc.Name, err)
	}
	klog.V(4).Infof("service/%s is updated", svc.Name)
	return nil
}

// CreateValidatingWebhookConfigurationFromYaml creates the validatingwebhookconfiguration from the yaml template.
func CreateValidatingWebhookConfigurationFromYaml(client client.Client, vwcTmpl string, ownerRef *metav1.OwnerReference, ctx interface{}) error {
	vw, err := SubsituteTemplate(vwcTmpl, ctx)
	if err != nil {
		return err
	}
	vwcObj, err := YamlToObject([]byte(vw))
	if err != nil {
		return err
	}
	vwc, ok := vwcObj.(*admissionv1.ValidatingWebhookConfiguration)
	if !ok {
		return fmt.Errorf("fail to assert validatingwebhookconfiguration")
	}
	if ownerRef != nil {
		ownerRefs := vwc.ObjectMeta.GetOwnerReferences()
		ownerRefs = append(ownerRefs, *ownerRef)
		vwc.ObjectMeta.SetOwnerReferences(ownerRefs)
	}
	err = client.Create(context.Background(), vwc)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the validatingwebhookconfiguration/%s: %v", vwc.Name, err)
		}
	}
	klog.V(4).Infof("validatingwebhookconfiguration/%s is created", vwc.Name)
	return nil
}

// DeleteValidatingWebhookConfigurationFromYaml delete the validatingwebhookconfiguration from the yaml template.
func DeleteValidatingWebhookConfigurationFromYaml(client client.Client, vwcTmpl string, ctx interface{}) error {
	vw, err := SubsituteTemplate(vwcTmpl, ctx)
	if err != nil {
		return err
	}
	vwcObj, err := YamlToObject([]byte(vw))
	if err != nil {
		return err
	}
	vwc, ok := vwcObj.(*admissionv1.ValidatingWebhookConfiguration)
	if !ok {
		return fmt.Errorf("fail to assert validatingwebhookconfiguration")
	}
	err = client.Delete(context.Background(), vwc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the validatingwebhookconfiguration/%s: %s", vwc.Name, err)
		}
	}
	klog.V(4).Infof("validatingwebhookconfiguration/%s is deleted", vwc.Name)
	return nil
}

// CreateJobFromYaml creates the Job from the yaml template.
func CreateJobFromYaml(client client.Client, jobTmpl, image string, ctx interface{}) error {
	jb, err := SubsituteTemplate(jobTmpl, ctx)
	if err != nil {
		return err
	}
	jbObj, err := YamlToObject([]byte(jb))
	if err != nil {
		return err
	}
	job, ok := jbObj.(*batchv1.Job)
	if !ok {
		return fmt.Errorf("fail to assert job")
	}
	if image != "" {
		job.Spec.Template.Spec.Containers[len(job.Spec.Template.Spec.Containers)-1].Image = image
	}
	err = client.Create(context.Background(), job)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("fail to create the job/%s: %v", job.Name, err)
		}
	}
	klog.V(4).Infof("job/%s is created", job.Name)
	return nil
}

// DeleteJobFromYaml deletes the Job from the yaml template.
func DeleteJobFromYaml(client client.Client, jobTmpl string, cleanup bool, ctx interface{}) error {
	jb, err := SubsituteTemplate(jobTmpl, ctx)
	if err != nil {
		return err
	}
	jbObj, err := YamlToObject([]byte(jb))
	if err != nil {
		return err
	}
	job, ok := jbObj.(*batchv1.Job)
	if !ok {
		return fmt.Errorf("fail to assert job")
	}
	if cleanup {
		err = client.Delete(context.Background(), job)
	} else {
		err = client.Delete(context.Background(), job, deleteOptions)
	}
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("fail to delete the job/%s: %v", job.Name, err)
		}
	}
	klog.V(4).Infof("job/%s is deleted", job.Name)
	return nil
}

// YamlToObject deserializes object in yaml format to a runtime.Object
func YamlToObject(yamlContent []byte) (k8sruntime.Object, error) {
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode(yamlContent, nil, nil)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// SubsituteTemplate fills out the template based on the context
func SubsituteTemplate(tmpl string, context interface{}) (string, error) {
	t, tmplPrsErr := template.New("test").Option("missingkey=zero").Parse(tmpl)
	if tmplPrsErr != nil {
		return "", tmplPrsErr
	}
	writer := bytes.NewBuffer([]byte{})
	if err := t.Execute(writer, context); nil != err {
		return "", err
	}
	return writer.String(), nil
}
