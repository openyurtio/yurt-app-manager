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

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1alpha2 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
	v2 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/fluxcd/v2beta1"
)

const (
	IngressNginxChartPath     = "/tmp/ingress-nginx/"
	IngressControllerLabelKey = "openyurt.io/yurtingress"
	IngressControllerLabelVal = "ingress-nginx"
	IngressControllerService  = "yurtingress-ingresscontroller"
)

func (r *YurtIngressReconciler) reconcileSvc(ctx context.Context, yurtingress *appsv1alpha2.YurtIngress) error {

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
		if err := r.Client.Create(ctx, genNginxSvc(ns, ingressService)); err != nil {
			klog.Errorf("create yurtingress serivce fail %v", err)
			return err
		}

	} else {
		newSvc := genNginxSvc(ns, ingressService)
		if err := r.Client.Patch(ctx, newSvc, client.MergeFrom(ingressService)); err != nil {
			return err
		} else {
			r.recorder.Event(yurtingress, v1.EventTypeNormal, "ServiceReconcile", "patch svc success")
		}
	}

	return nil
}

func (r *YurtIngressReconciler) reconcileRelease(ctx context.Context, yurtingress *appsv1alpha2.YurtIngress, np *v1beta1.NodePool) error {

	chart, err := loader.LoadDir(IngressNginxChartPath)
	klog.Infof("load chart info %s", IngressNginxChartPath)
	if err != nil {
		return err
	}

	releaseName := fmt.Sprintf("yurtingress-%s", yurtingress.Name)
	values, err := genChartValues(yurtingress, np)
	if err != nil {
		return err
	}

	klog.Infof("prepare process chart %s %v", releaseName, values.AsMap())
	hr := v2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseName,
			Namespace: r.namespace,
		},
		Spec: v2.HelmReleaseSpec{
			Upgrade: &v2.Upgrade{
				Force: true,
			},
		},
	}

	run, err := NewDefaultRunner(ctx, r.namespace)
	if err != nil {
		return err
	}

	rel, observeLastReleaseErr := run.ObserveLastRelease(releaseName)
	if observeLastReleaseErr != nil {
		err = fmt.Errorf("failed to get last release revision: %w", observeLastReleaseErr)
		return err
	}

	if rel != nil {
		fmt.Printf("name: %v,status: %v\n", rel.Name, rel.Info.Status)
	}
	//relJson,_ := yaml.Marshal(rel)
	//fmt.Println(string(relJson))
	//klog.Infof("HelmRelease info: %v",string(relJson))

	if rel != nil {
		oldValues := chartutil.Values{}
		if rel.Config != nil {
			oldValues = chartutil.Values(rel.Config)
		}

		oldValue, oldErr := oldValues.YAML()
		newValue, newErr := values.YAML()
		if oldErr != nil {
			return oldErr
		}
		if newErr != nil {
			return newErr
		}

		if oldValue == newValue && yurtingress.Status.Status == release.StatusDeployed.String() {
			r.recorder.Event(yurtingress, v1.EventTypeNormal, "HelmReconcile", "helm value not change, just return")
			return nil
		} else {
			klog.Infof("release value change, release: %s %s, old: %s, newZ: %s", r.namespace, releaseName, oldValue, newValue)
		}
	} else {
		klog.Infof("release value not found, release: %s %s", r.namespace, releaseName)
	}

	if rel != nil {
		klog.Infof(" release exist %s %s %v", hr.Namespace, hr.Name, rel.Version)
		_, err = run.Upgrade(hr, chart, values)

		klog.Infof("check error %v %v", errors.Is(err, driver.ErrNoDeployedReleases), errors.Is(err, driver.ErrReleaseNotFound))
		//https://phoenixnap.com/kb/helm-has-no-deployed-releases
		if errors.Is(err, driver.ErrNoDeployedReleases) || errors.Is(err, driver.ErrReleaseNotFound) {
			relSecret := &v1.Secret{}
			relSecretName := fmt.Sprintf("sh.helm.release.v1.%s.v%d", hr.Name, rel.Version-1)
			if err := r.Get(ctx, types.NamespacedName{Namespace: r.namespace, Name: relSecretName}, relSecret); err != nil {
				return err
			}

			newReleaseSecret := relSecret.DeepCopy()
			if _, ok := newReleaseSecret.Labels["status"]; ok {
				newReleaseSecret.Labels["status"] = "deployed"
			}
			if err := r.Patch(ctx, newReleaseSecret, client.MergeFrom(relSecret)); err != nil {
				return err
			}
			klog.Infof("update secret status %s %s %s", hr.Namespace, hr.Name, relSecretName)
		}
	} else {
		klog.Infof(" release not exist %s", releaseName)
		rel, err = run.Install(hr, chart, values)
	}

	if err != nil {
		newCondition := metav1.Condition{
			Type:    "Released",
			Status:  metav1.ConditionFalse,
			Reason:  "ReconciliationFailed",
			Message: err.Error(),
		}
		newYurtIngress := yurtingress.DeepCopy()

		apimeta.SetStatusCondition(newYurtIngress.GetStatusConditions(), newCondition)
		newYurtIngress.Status.Status = release.StatusFailed.String()
		newYurtIngress.Status.LastAppliedRevision = hr.Status.LastAttemptedRevision
		if patchErr := r.Client.Status().Patch(ctx, newYurtIngress, client.MergeFrom(yurtingress)); patchErr != nil {
			klog.Errorf("helm action fail, patch yurtingress fail, err: %v", patchErr)
			return patchErr
		}
		klog.Errorf("helm action fail %v", err)
		return err

	} else {
		newCondition := metav1.Condition{
			Type:    "Released",
			Status:  metav1.ConditionTrue,
			Reason:  "ReconciliationSucceeded",
			Message: "Release reconciliation succeeded",
		}
		newYurtIngress := yurtingress.DeepCopy()

		apimeta.SetStatusCondition(newYurtIngress.GetStatusConditions(), newCondition)
		newYurtIngress.Status.Status = release.StatusDeployed.String()
		newYurtIngress.Status.LastAppliedRevision = hr.Status.LastAttemptedRevision
		if err := r.Client.Status().Patch(ctx, newYurtIngress, client.MergeFrom(yurtingress)); err != nil {
			return err
		}

	}

	return err
}

func (r *YurtIngressReconciler) cleanupResources(ctx context.Context, instance *appsv1alpha2.YurtIngress) error {
	klog.Infof("cleanup yurtingress %s", instance.Name)
	run, err := NewDefaultRunner(context.TODO(), r.namespace)
	if err != nil {
		return err
	}

	hr := v2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("yurtingress-%s", instance.Name),
			Namespace: r.namespace,
		},
		Spec: v2.HelmReleaseSpec{},
	}

	newYurtIngress := instance.DeepCopy()

	if err := run.Uninstall(hr); err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		newCondition := metav1.Condition{
			Type:    "Released",
			Status:  metav1.ConditionFalse,
			Reason:  "Uninstall",
			Message: "Release uninstall failed",
		}
		apimeta.SetStatusCondition(newYurtIngress.GetStatusConditions(), newCondition)
		if err := r.Client.Status().Patch(ctx, newYurtIngress, client.MergeFrom(instance)); err != nil {
			return err
		}
		return err
	}

	newCondition := metav1.Condition{
		Type:    "Released",
		Status:  metav1.ConditionTrue,
		Reason:  "Uninstall",
		Message: "Release uninstall succeeded",
	}

	apimeta.SetStatusCondition(newYurtIngress.GetStatusConditions(), newCondition)
	newYurtIngress.Status.Status = release.StatusUninstalled.String()
	if err := r.Client.Status().Patch(ctx, newYurtIngress, client.MergeFrom(instance)); err != nil {
		return err
	}

	if controllerutil.ContainsFinalizer(instance, appsv1alpha2.YurtIngressFinalizer) {
		newIns := instance.DeepCopy()
		controllerutil.RemoveFinalizer(newIns, appsv1alpha2.YurtIngressFinalizer)
		if err := r.Patch(context.TODO(), newIns, client.MergeFrom(instance)); err != nil {
			return err
		}
	}
	klog.Infof("cleanup yurtingress %s succ", instance.Name)

	return nil
}

func genNginxSvc(ns string, svc *v1.Service) *v1.Service {
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

func genChartValues(in *appsv1alpha2.YurtIngress, np *v1beta1.NodePool) (chartutil.Values, error) {

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

	return MergeMaps(defaultValues, values), nil
}

type Toleration struct {
	Key string `json:"key,omitempty" protobuf:"bytes,1,opt,name=key"`

	Operator v1.TolerationOperator `json:"operator,omitempty" protobuf:"bytes,2,opt,name=operator,casttype=TolerationOperator"`

	Value string `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`

	Effect v1.TaintEffect `json:"effect,omitempty" protobuf:"bytes,4,opt,name=effect,casttype=TaintEffect"`
}
