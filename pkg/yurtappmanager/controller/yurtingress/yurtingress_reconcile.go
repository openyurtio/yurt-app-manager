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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"time"

	v2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha2 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd/events"
)

const (
	IngressNginxChartPath     = "/tmp/ingress-nginx/"
	IngressControllerLabelKey = "openyurt.io/yurtingress"
	IngressControllerLabelVal = "ingress-nginx"
	IngressControllerService  = "yurtingress-ingresscontroller"
)

func (r *YurtIngressReconciler) reconcileYurtIngress(ctx context.Context, yurtingress *appsv1alpha2.YurtIngress, np *v1beta1.NodePool) (ctrl.Result, error) {

	chart, err := loader.LoadDir(IngressNginxChartPath)
	if err != nil {
		return ctrl.Result{}, err
	}

	releaseName := fmt.Sprintf("yurtingress-%s", yurtingress.Name)
	values, err := composeChartValues(yurtingress, np)
	if err != nil {
		return ctrl.Result{}, err
	}

	klog.Infof("chart: %s, release: %s, values: %v", IngressNginxChartPath, releaseName, values.AsMap())
	// for reuse helm-controller's implementation, copy YurtIngress's attribute to HelmRelease
	hr := v2.HelmRelease{
		ObjectMeta: yurtingress.ObjectMeta,
		Spec: v2.HelmReleaseSpec{
			TargetNamespace: r.namespace,
			ReleaseName:     releaseName,
			Timeout:         &metav1.Duration{Duration: 1 * time.Minute},
			Install:         &v2.Install{Remediation: &v2.InstallRemediation{Retries: 10}},
			Upgrade:         &v2.Upgrade{Remediation: &v2.UpgradeRemediation{Retries: 10}},
		},
		Status: yurtingress.Status,
	}

	reconciledHr, reconcileErr := r.reconcileRelease(ctx, *hr.DeepCopy(), chart, values)

	if reconcileErr != nil {
		r.event(ctx, hr, chart.Metadata.Version, events.EventSeverityError,
			fmt.Sprintf("reconciliation failed: %s", reconcileErr.Error()))
	}

	if updateStatusErr := r.patchStatus(ctx, &reconciledHr); updateStatusErr != nil {
		klog.Error(updateStatusErr, "unable to update status after reconciliation")
		return ctrl.Result{}, updateStatusErr
	}

	// whether reconcile success or fail, retry reconcile
	return ctrl.Result{RequeueAfter: yurtingress.Spec.Interval.Duration}, nil

}

func (r *YurtIngressReconciler) cleanupResources(ctx context.Context, yi *appsv1alpha2.YurtIngress) error {
	klog.Infof("cleanup yurtingress %s resources", yi.Name)
	run, err := fluxcd.NewDefaultRunner(context.TODO(), r.namespace)
	if err != nil {
		return err
	}

	releaseName := fmt.Sprintf("yurtingress-%s", yi.Name)

	hr := v2.HelmRelease{
		ObjectMeta: yi.ObjectMeta,
		Spec: v2.HelmReleaseSpec{
			TargetNamespace: r.namespace,
			ReleaseName:     releaseName,
		},
		Status: yi.Status,
	}

	if err := run.Uninstall(hr); err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		newCondition := metav1.Condition{
			Type:    "Released",
			Status:  metav1.ConditionFalse,
			Reason:  "Uninstall",
			Message: "Helm uninstall failed",
		}
		apimeta.SetStatusCondition(hr.GetStatusConditions(), newCondition)
		if patchErr := r.patchStatus(ctx, &hr); patchErr != nil {
			klog.Errorf("uninstall patch condition error %s", yi.Name)
			return patchErr
		}
		return err
	}

	newCondition := metav1.Condition{
		Type:    "Released",
		Status:  metav1.ConditionTrue,
		Reason:  "Uninstall",
		Message: "Helm uninstall succeeded",
	}
	apimeta.SetStatusCondition(hr.GetStatusConditions(), newCondition)
	if patchErr := r.patchStatus(ctx, &hr); patchErr != nil {
		klog.Errorf("uninstall patch condition error %s", yi.Name)
		return patchErr
	}
	klog.Infof("yurtingress %s uninstall succeeded", yi.Name)

	if !yi.GetDeletionTimestamp().IsZero() {
		if controllerutil.ContainsFinalizer(yi, appsv1alpha2.YurtIngressFinalizer) {
			newIns := yi.DeepCopy()
			controllerutil.RemoveFinalizer(newIns, appsv1alpha2.YurtIngressFinalizer)
			if err := r.Patch(context.TODO(), newIns, client.MergeFrom(yi)); err != nil {
				return err
			}
		}
		klog.Infof("yurtingress %s remove finalizer succeeded", yi.Name)
	}

	klog.Infof("yurtingress %s cleanup succeeded", yi.Name)

	return nil
}

func (r *YurtIngressReconciler) reconcileTopologyService(ctx context.Context, yurtingress *appsv1alpha2.YurtIngress) error {

	ns := r.namespace
	ingressService := &v1.Service{}
	svcExist := true
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: ns, Name: IngressControllerService}, ingressService); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		svcExist = false
	}

	if !svcExist {
		if err := r.Client.Create(ctx, generateTopologyService(ns, ingressService)); err != nil {
			klog.Errorf("create yurtingress serivce fail %v", err)
			return err
		}

	} else {
		newSvc := generateTopologyService(ns, ingressService)
		if err := r.Client.Patch(ctx, newSvc, client.MergeFrom(ingressService)); err != nil {
			return err
		} else {
			r.EventRecorder.Event(yurtingress, v1.EventTypeNormal, "ServiceReconcile", "patch svc success")
		}
	}

	return nil
}

// overwrite helm-controller's patchStatus
// reconcileRelease return HelmRelease result, convert it to YurtIngress, and update YurtIngress status
func (r *YurtIngressReconciler) patchStatus(ctx context.Context, hr *v2.HelmRelease) error {
	key := client.ObjectKeyFromObject(hr)

	latest := &appsv1alpha2.YurtIngress{}
	if err := r.Client.Get(ctx, key, latest); err != nil {
		return err
	}

	yi := &appsv1alpha2.YurtIngress{
		TypeMeta:   latest.TypeMeta,
		ObjectMeta: hr.ObjectMeta,
		Status:     hr.Status,
	}

	if err := r.Client.Status().Patch(ctx, yi, client.MergeFrom(latest)); err != nil {
		return err
	}

	newLatest := &appsv1alpha2.YurtIngress{}
	if err := r.Client.Get(ctx, key, newLatest); err != nil {
		return err
	}

	hr.ObjectMeta = newLatest.ObjectMeta
	hr.Status = newLatest.Status

	return nil
}

func generateTopologyService(ns string, svc *v1.Service) *v1.Service {
	if svc == nil {
		svc = &v1.Service{}
	}

	newSvc := svc.DeepCopy()

	newSvc.Namespace = ns
	newSvc.Name = IngressControllerService

	if newSvc.Annotations == nil {
		newSvc.Annotations = map[string]string{}
	}
	newSvc.Annotations["openyurt.io/topologyKeys"] = "openyurt.io/nodepool"

	newSvc.Spec = v1.ServiceSpec{
		Type:     v1.ServiceTypeNodePort,
		Selector: map[string]string{IngressControllerLabelKey: IngressControllerLabelVal},
		Ports:    []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromString("http")}, {Name: "https", Port: 443, TargetPort: intstr.FromString("https")}},
	}

	return newSvc
}

// generate default helm values by nodepool settings
func composeChartValues(in *appsv1alpha2.YurtIngress, np *v1beta1.NodePool) (chartutil.Values, error) {

	strVals := `
defaultBackend:
  enabled: true
controller:
  labels:
    {{.controllerDefaultLabelKey}}: {{.controllerDefaultLabelVal}}
  service:
    type: NodePort
  admissionWebhooks:
    enabled: false
  ingressClassByName: true
  ingressClassResource:
    name: {{.controllerIngressClassResourceName}}
    controllerValue: {{.controllerIngressClassResourceControllerValue}}
`

	nodePoolTolerations := make([]Toleration, 0)
	for _, tmp := range np.Spec.Taints {
		npTolera := Toleration{
			Key: tmp.Key,
			//Value:    tmp.Value,
			Operator: v1.TolerationOpExists,
			Effect:   tmp.Effect,
		}
		nodePoolTolerations = append(nodePoolTolerations, npTolera)
	}

	strVal := map[string]string{
		"controllerIngressClassResourceName":            np.Name,
		"controllerIngressClassResourceControllerValue": fmt.Sprintf("openyurt.io/%s", np.Name),
		"controllerDefaultLabelKey":                     IngressControllerLabelKey,
		"controllerDefaultLabelVal":                     IngressControllerLabelVal,
	}

	var tpl bytes.Buffer

	templ, err := template.New("helm").Parse(strVals)
	if err != nil {
		return nil, err
	}

	if err := templ.Execute(&tpl, strVal); err != nil {
		return nil, err
	}

	defaultValues, err := chartutil.ReadValues(tpl.Bytes())
	if err != nil {
		return nil, err
	}

	// key defaultBackend already exist
	backendVal := defaultValues["defaultBackend"].(map[string]interface{})
	backendVal["tolerations"] = nodePoolTolerations
	backendVal["nodeSelector"] = np.Spec.Selector.MatchLabels

	// key controller already exist
	controllerVal := defaultValues["controller"].(map[string]interface{})
	controllerVal["tolerations"] = nodePoolTolerations
	controllerVal["nodeSelector"] = np.Spec.Selector.MatchLabels

	var values map[string]interface{}
	if in.Spec.Values != nil {
		if err := json.Unmarshal(in.Spec.Values.Raw, &values); err != nil {
			return nil, err
		}
	}

	return fluxcd.MergeMaps(defaultValues, values), nil
}

type Toleration struct {
	Key string `json:"key,omitempty" protobuf:"bytes,1,opt,name=key"`

	Operator v1.TolerationOperator `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator,casttype=TolerationOperator"`

	Value string `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`

	Effect v1.TaintEffect `json:"effect,omitempty" protobuf:"bytes,4,opt,name=effect,casttype=TaintEffect"`
}
