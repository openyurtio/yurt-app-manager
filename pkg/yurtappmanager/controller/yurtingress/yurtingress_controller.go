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

package yurtingress

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/gate"
	yurtapputil "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/kubernetes"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/refmanager"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	controllerName         = "yurtingress-controller"
	ingressDeploymentLabel = "yurtingress.io/nodepool"
)

const updateRetries = 5

var concurrentReconciles = 3

// YurtIngressReconciler reconciles a YurtIngress object
type YurtIngressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// Add creates a new YurtIngress Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and start it when the Manager is started.
func Add(mgr manager.Manager, ctx context.Context) error {
	if !gate.ResourceEnabled(&appsv1alpha1.YurtIngress{}) {
		return nil
	}
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
//func newReconciler(mgr manager.Manager, createSingletonPoolIngress bool) reconcile.Reconciler {
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &YurtIngressReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(controllerName),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(
		controllerName,
		mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: concurrentReconciles})
	if err != nil {
		return err
	}
	// Watch for changes to YurtIngress
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.YurtIngress{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.YurtIngress{},
	})
	if err != nil {
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=apps.openyurt.io,resources=yurtingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.openyurt.io,resources=yurtingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *YurtIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	klog.V(4).Infof("Reconcile YurtIngress: %s", req.Name)
	// Fetch the YurtIngress instance
	instance := &appsv1alpha1.YurtIngress{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Add finalizer if not exist
	if !controllerutil.ContainsFinalizer(instance, appsv1alpha1.YurtIngressFinalizer) {
		controllerutil.AddFinalizer(instance, appsv1alpha1.YurtIngressFinalizer)
		if err := r.Update(context.TODO(), instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	// Handle ingress controller resources cleanup
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.cleanupIngressResources(instance)
	}

	var desiredPools, currentPools []appsv1alpha1.IngressPool
	desiredPools = getDesiredPools(instance)
	currentPools = getCurrentPools(instance)
	isYurtIngressCRChanged := false
	if instance.Spec.IngressControllerImage == "" {
		instance.Spec.IngressControllerImage = appsv1alpha1.DefaultNginxIngressControllerImage
	}
	if instance.Spec.IngressWebhookCertGenImage == "" {
		instance.Spec.IngressWebhookCertGenImage = appsv1alpha1.DefaultNginxIngressWebhookCertGenImage
	}
	if instance.Spec.Replicas == 0 {
		klog.V(4).Infof("set default per-pool replicas to 1")
		instance.Spec.Replicas = 1
	}
	addedPools, removedPools, unchangedPools := getPools(desiredPools, currentPools)
	if addedPools != nil {
		klog.V(4).Infof("added pool list is %s", addedPools)
		isYurtIngressCRChanged = true
		ownerRef := prepareDeploymentOwnerReferences(instance)
		if currentPools == nil && !yurtapputil.IsIngressNamespaceReady(r.Client) {
			if err := yurtapputil.CreateNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
		}
		replicas := instance.Spec.Replicas
		ingress_controller_image := instance.Spec.IngressControllerImage
		ingress_webhook_certgen_image := instance.Spec.IngressWebhookCertGenImage
		for _, pool := range addedPools {
			if err := yurtapputil.CreateNginxIngressSpecificResource(r.Client, pool.Name, &pool.IngressIPs, ingress_controller_image, ingress_webhook_certgen_image, replicas, ownerRef); err != nil {
				return ctrl.Result{}, err
			}
			notReadyPool := appsv1alpha1.IngressNotReadyPool{Pool: appsv1alpha1.IngressPool{Name: pool.Name, IngressIPs: pool.IngressIPs}, Info: nil}
			instance.Status.Conditions.IngressNotReadyPools = append(instance.Status.Conditions.IngressNotReadyPools, notReadyPool)
			instance.Status.UnreadyNum += 1
		}
	}
	if removedPools != nil {
		klog.V(4).Infof("removed pool list is %s", removedPools)
		isYurtIngressCRChanged = true
		for _, pool := range removedPools {
			if desiredPools == nil {
				if err := yurtapputil.DeleteNginxIngressSpecificResource(r.Client, pool.Name, true); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				if err := yurtapputil.DeleteNginxIngressSpecificResource(r.Client, pool.Name, false); err != nil {
					return ctrl.Result{}, err
				}
			}
			if desiredPools != nil && !removePoolfromCondition(instance, pool.Name) {
				klog.V(4).Infof("Pool/%s is not found from conditions!", pool.Name)
			}
		}
		if desiredPools == nil && isOnlyYurtIngressCR(r.Client) {
			if err := yurtapputil.DeleteNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
			instance.Status.Conditions.IngressReadyPools = nil
			instance.Status.Conditions.IngressNotReadyPools = nil
			instance.Status.ReadyNum = 0
			instance.Status.UnreadyNum = 0
		}
	}
	if unchangedPools != nil {
		klog.V(4).Infof("unchanged pool list is %s", unchangedPools)
		desiredReplicas := instance.Spec.Replicas
		currentReplicas := instance.Status.Replicas
		desiredIngressControllerImage := instance.Spec.IngressControllerImage
		currentIngressControllerImage := instance.Status.IngressControllerImage
		desiredNginxWebhookCertGenImage := instance.Spec.IngressWebhookCertGenImage
		currentNginxWebhookCertGenImage := instance.Status.IngressWebhookCertGenImage
		if desiredIngressControllerImage != currentIngressControllerImage {
			klog.V(4).Infof("Ingress controller image is changed!")
			isYurtIngressCRChanged = true
			instance.Status.ReadyNum = 0
			instance.Status.UnreadyNum = int32(len(instance.Spec.Pools))
			for _, pool := range unchangedPools {
				if err := yurtapputil.UpdateNginxIngressControllerDeploymment(r.Client, pool.Name, desiredReplicas, desiredIngressControllerImage); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else if desiredReplicas != currentReplicas {
			klog.V(4).Infof("Ingress controller replicas is changed!")
			isYurtIngressCRChanged = true
			for _, pool := range unchangedPools {
				if err := yurtapputil.ScaleNginxIngressControllerDeploymment(r.Client, pool.Name, desiredReplicas); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		if desiredNginxWebhookCertGenImage != currentNginxWebhookCertGenImage {
			klog.V(4).Infof("Nginx ingress controller webhook certgen image is changed!")
			isYurtIngressCRChanged = true
			for _, pool := range unchangedPools {
				if err := yurtapputil.RecreateNginxWebhookJob(r.Client, pool.Name, desiredNginxWebhookCertGenImage); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		for _, pool := range unchangedPools {
			currentPool := getCurrentPool(instance, pool.Name)
			if currentPool != nil {
				if !isStrArrayEqual(pool.IngressIPs, currentPool.IngressIPs) {
					klog.V(4).Infof("pool %s ingressIPs is changed", pool.Name)
					if err := yurtapputil.UpdateNginxServiceExternalIPs(r.Client, pool.Name, pool.IngressIPs); err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		}
	}
	r.updateStatus(instance, isYurtIngressCRChanged)
	return ctrl.Result{}, nil
}

func isStrArrayEqual(strList1, strList2 []string) bool {
	if len(strList1) != len(strList2) {
		return false
	}
	if len(strList1) == 0 && len(strList2) == 0 {
		return true
	}
	for i, str := range strList1 {
		if str != strList2[i] {
			return false
		}
	}
	return true
}

func getPools(desired, current []appsv1alpha1.IngressPool) (added, removed, unchanged []appsv1alpha1.IngressPool) {
	swap := false
	for i := 0; i < 2; i++ {
		for _, p1 := range desired {
			found := false
			for _, p2 := range current {
				if p1.Name == p2.Name {
					found = true
					if !swap {
						unchanged = append(unchanged, p1)
					}
					break
				}
			}
			if !found {
				if !swap {
					added = append(added, p1)
				} else {
					removed = append(removed, p1)
				}
			}
		}
		if i == 0 {
			swap = true
			desired, current = current, desired
		}
	}
	return added, removed, unchanged
}

func getDesiredPools(ying *appsv1alpha1.YurtIngress) []appsv1alpha1.IngressPool {
	return ying.Spec.Pools
}

func getCurrentPools(ying *appsv1alpha1.YurtIngress) []appsv1alpha1.IngressPool {
	var currentPools []appsv1alpha1.IngressPool
	currentPools = ying.Status.Conditions.IngressReadyPools
	for _, pool := range ying.Status.Conditions.IngressNotReadyPools {
		currentPools = append(currentPools, pool.Pool)
	}
	return currentPools
}

func getDesiredPool(ying *appsv1alpha1.YurtIngress, poolname string) *appsv1alpha1.IngressPool {
	for _, pool := range ying.Spec.Pools {
		if pool.Name == poolname {
			return &pool
		}
	}
	klog.V(4).Infof("Can not find desired pool %s", poolname)
	return nil
}

func getCurrentPool(ying *appsv1alpha1.YurtIngress, poolname string) *appsv1alpha1.IngressPool {
	for _, pool := range getCurrentPools(ying) {
		if pool.Name == poolname {
			return &pool
		}
	}
	klog.V(4).Infof("Can not find current pool %s", poolname)
	return nil
}

func removePoolfromCondition(ying *appsv1alpha1.YurtIngress, poolname string) bool {
	for i, pool := range ying.Status.Conditions.IngressReadyPools {
		if pool.Name == poolname {
			length := len(ying.Status.Conditions.IngressReadyPools)
			if i == length-1 {
				ying.Status.Conditions.IngressReadyPools = ying.Status.Conditions.IngressReadyPools[:i]
			} else {
				ying.Status.Conditions.IngressReadyPools = append(ying.Status.Conditions.IngressReadyPools[:i],
					ying.Status.Conditions.IngressReadyPools[i+1:]...)
			}
			if ying.Status.ReadyNum >= 1 {
				ying.Status.ReadyNum -= 1
			}
			return true
		}
	}
	for i, pool := range ying.Status.Conditions.IngressNotReadyPools {
		if pool.Pool.Name == poolname {
			length := len(ying.Status.Conditions.IngressNotReadyPools)
			if i == length-1 {
				ying.Status.Conditions.IngressNotReadyPools = ying.Status.Conditions.IngressNotReadyPools[:i]
			} else {
				ying.Status.Conditions.IngressNotReadyPools = append(ying.Status.Conditions.IngressNotReadyPools[:i],
					ying.Status.Conditions.IngressNotReadyPools[i+1:]...)
			}
			if ying.Status.UnreadyNum >= 1 {
				ying.Status.UnreadyNum -= 1
			}
			return true
		}
	}
	return false
}

func (r *YurtIngressReconciler) updateStatus(ying *appsv1alpha1.YurtIngress, ingressCRChanged bool) error {
	ying.Status.Replicas = ying.Spec.Replicas
	ying.Status.IngressControllerImage = ying.Spec.IngressControllerImage
	ying.Status.IngressWebhookCertGenImage = ying.Spec.IngressWebhookCertGenImage
	if !ingressCRChanged {
		deployments, err := r.getAllDeployments(ying)
		if err != nil {
			klog.V(4).Infof("Fail to get all the ingress controller deployments: %v", err)
			return err
		}
		ying.Status.Conditions.IngressReadyPools = nil
		ying.Status.Conditions.IngressNotReadyPools = nil
		ying.Status.ReadyNum = 0
		for _, dply := range deployments {
			pool := dply.ObjectMeta.GetLabels()[ingressDeploymentLabel]
			if dply.Status.ReadyReplicas == ying.Spec.Replicas {
				klog.V(4).Infof("Ingress on pool %s is ready!", pool)
				ying.Status.ReadyNum += 1
				readyPool := getDesiredPool(ying, pool)
				ying.Status.Conditions.IngressReadyPools = append(ying.Status.Conditions.IngressReadyPools, *readyPool)
			} else {
				klog.V(4).Infof("Ingress on pool %s is NOT ready!", pool)
				condition := getUnreadyDeploymentCondition(dply)
				if condition == nil {
					klog.V(4).Infof("Get deployment/%s conditions nil!", dply.GetName())
				} else {
					notReadyPool := appsv1alpha1.IngressNotReadyPool{Pool: *getDesiredPool(ying, pool), Info: condition}
					ying.Status.Conditions.IngressNotReadyPools = append(ying.Status.Conditions.IngressNotReadyPools, notReadyPool)
				}
			}
		}
		ying.Status.UnreadyNum = int32(len(ying.Spec.Pools)) - ying.Status.ReadyNum
	}
	var updateErr error
	for i, obj := 0, ying; i < updateRetries; i++ {
		updateErr = r.Status().Update(context.TODO(), obj)
		if updateErr == nil {
			klog.V(4).Infof("%s status is updated!", obj.Name)
			return nil
		}
	}
	klog.Errorf("Fail to update YurtIngress %s status: %v", ying.Name, updateErr)
	return updateErr
}

func (r *YurtIngressReconciler) cleanupIngressResources(instance *appsv1alpha1.YurtIngress) (ctrl.Result, error) {
	pools := getDesiredPools(instance)
	isOnly := isOnlyYurtIngressCR(r.Client)

	if controllerutil.ContainsFinalizer(instance, appsv1alpha1.YurtIngressFinalizer) {
		controllerutil.RemoveFinalizer(instance, appsv1alpha1.YurtIngressFinalizer)
		if err := r.Update(context.TODO(), instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	if pools != nil {
		for _, pool := range pools {
			if err := yurtapputil.DeleteNginxIngressSpecificResource(r.Client, pool.Name, isOnly); err != nil {
				return ctrl.Result{}, err
			}
		}
		if isOnly {
			if err := yurtapputil.DeleteNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func prepareDeploymentOwnerReferences(instance *appsv1alpha1.YurtIngress) *metav1.OwnerReference {
	isController := true
	isBlockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		//TODO: optimize the APIVersion/Kind with instance values
		APIVersion:         "apps.openyurt.io/v1alpha1",
		Kind:               "YurtIngress",
		Name:               instance.Name,
		UID:                instance.UID,
		Controller:         &isController,
		BlockOwnerDeletion: &isBlockOwnerDeletion,
	}
	return &ownerRef
}

// getAllDeployments returns all of deployments owned by YurtIngress
func (r *YurtIngressReconciler) getAllDeployments(ying *appsv1alpha1.YurtIngress) ([]*appsv1.Deployment, error) {
	labelSelector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      ingressDeploymentLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return nil, err
	}

	dplyList := &appsv1.DeploymentList{}
	err = r.Client.List(context.TODO(), dplyList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	manager, err := refmanager.New(r.Client, &labelSelector, ying, r.Scheme)
	if err != nil {
		return nil, err
	}

	selected := make([]metav1.Object, len(dplyList.Items))
	for i, dply := range dplyList.Items {
		selected[i] = dply.DeepCopy()
	}
	claimed, err := manager.ClaimOwnedObjects(selected)
	if err != nil {
		return nil, err
	}

	claimedDplys := make([]*appsv1.Deployment, len(claimed))
	for i, dply := range claimed {
		claimedDplys[i] = dply.(*appsv1.Deployment)
	}
	return claimedDplys, nil
}

func getUnreadyDeploymentCondition(dply *appsv1.Deployment) (info *appsv1alpha1.IngressNotReadyConditionInfo) {
	len := len(dply.Status.Conditions)
	if len == 0 {
		return nil
	}
	var conditionInfo appsv1alpha1.IngressNotReadyConditionInfo
	condition := dply.Status.Conditions[len-1]
	if condition.Type == appsv1.DeploymentReplicaFailure {
		conditionInfo.Type = appsv1alpha1.IngressFailure
	} else {
		conditionInfo.Type = appsv1alpha1.IngressPending
	}
	conditionInfo.LastTransitionTime = condition.LastTransitionTime
	conditionInfo.Message = condition.Message
	conditionInfo.Reason = condition.Reason
	return &conditionInfo
}

func isOnlyYurtIngressCR(c client.Client) bool {
	ingressList := appsv1alpha1.YurtIngressList{}
	err := c.List(context.TODO(), &ingressList, &client.ListOptions{})
	if err != nil {
		klog.V(4).Infof("Get yurtingress list err: %v", err)
		return false
	}
	return len(ingressList.Items) == 1
}
