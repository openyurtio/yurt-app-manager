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
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
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
	if req.Name != appsv1alpha1.SingletonYurtIngressInstanceName {
		return ctrl.Result{}, nil
	}
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
	// Set the default version at current stage
	instance.Status.Version = appsv1alpha1.NginxIngressControllerVersion

	var desiredPoolNames, currentPoolNames []string
	desiredPoolNames = getDesiredPoolNames(instance)
	currentPoolNames = getCurrentPoolNames(instance)
	isIngressCRChanged := false
	addedPools, removedPools, unchangedPools := getPools(desiredPoolNames, currentPoolNames)
	if addedPools != nil {
		klog.V(4).Infof("added pool list is %s", addedPools)
		isIngressCRChanged = true
		ownerRef := prepareDeploymentOwnerReferences(instance)
		if currentPoolNames == nil {
			if err := yurtapputil.CreateNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
		}
		for _, pool := range addedPools {
			replicas := instance.Spec.Replicas
			if err := yurtapputil.CreateNginxIngressSpecificResource(r.Client, pool, replicas, ownerRef); err != nil {
				return ctrl.Result{}, err
			}
			notReadyPool := appsv1alpha1.IngressNotReadyPool{Name: pool, Info: nil}
			instance.Status.Conditions.IngressNotReadyPools = append(instance.Status.Conditions.IngressNotReadyPools, notReadyPool)
		}
	}
	if removedPools != nil {
		klog.V(4).Infof("removed pool list is %s", removedPools)
		isIngressCRChanged = true
		for _, pool := range removedPools {
			if desiredPoolNames == nil {
				if err := yurtapputil.DeleteNginxIngressSpecificResource(r.Client, pool, true); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				if err := yurtapputil.DeleteNginxIngressSpecificResource(r.Client, pool, false); err != nil {
					return ctrl.Result{}, err
				}
			}
			if !removePoolfromCondition(instance, pool) {
				klog.V(4).Infof("Pool/%s is not found from conditions!", pool)
			}
		}
		if desiredPoolNames == nil {
			if err := yurtapputil.DeleteNginxIngressCommonResource(r.Client); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	if unchangedPools != nil {
		klog.V(4).Infof("unchanged pool list is %s", unchangedPools)
		desiredReplicas := instance.Spec.Replicas
		currentReplicas := instance.Status.Replicas
		if desiredReplicas != currentReplicas {
			klog.V(4).Infof("Per-Pool ingress controller replicas is changed!")
			isIngressCRChanged = true
			for _, pool := range unchangedPools {
				if err := yurtapputil.ScaleNginxIngressControllerDeploymment(r.Client, pool, desiredReplicas); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}
	r.updateStatus(instance, isIngressCRChanged)
	return ctrl.Result{}, nil
}

func getPools(desired, current []string) (added, removed, unchanged []string) {
	swap := false
	for i := 0; i < 2; i++ {
		for _, s1 := range desired {
			found := false
			for _, s2 := range current {
				if s1 == s2 {
					found = true
					if !swap {
						unchanged = append(unchanged, s1)
					}
					break
				}
			}
			if !found {
				if !swap {
					added = append(added, s1)
				} else {
					removed = append(removed, s1)
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

func getDesiredPoolNames(ying *appsv1alpha1.YurtIngress) []string {
	var desiredPoolNames []string
	for _, pool := range ying.Spec.Pools {
		desiredPoolNames = append(desiredPoolNames, pool.Name)
	}
	return desiredPoolNames
}

func getCurrentPoolNames(ying *appsv1alpha1.YurtIngress) []string {
	var currentPoolNames []string
	currentPoolNames = ying.Status.Conditions.IngressReadyPools
	for _, pool := range ying.Status.Conditions.IngressNotReadyPools {
		currentPoolNames = append(currentPoolNames, pool.Name)
	}
	return currentPoolNames
}

func removePoolfromCondition(ying *appsv1alpha1.YurtIngress, poolname string) bool {
	for i, pool := range ying.Status.Conditions.IngressReadyPools {
		if pool == poolname {
			length := len(ying.Status.Conditions.IngressReadyPools)
			if i == length-1 {
				if length == 1 {
					ying.Status.Conditions.IngressReadyPools = nil
				} else {
					ying.Status.Conditions.IngressReadyPools = ying.Status.Conditions.IngressReadyPools[:i-1]
				}
			} else {
				ying.Status.Conditions.IngressReadyPools[i] = ying.Status.Conditions.IngressReadyPools[i+1]
			}
			if ying.Status.ReadyNum >= 1 {
				ying.Status.ReadyNum -= 1
			}
			return true
		}
	}
	for i, pool := range ying.Status.Conditions.IngressNotReadyPools {
		if pool.Name == poolname {
			length := len(ying.Status.Conditions.IngressNotReadyPools)
			if i == length-1 {
				if length == 1 {
					ying.Status.Conditions.IngressNotReadyPools = nil
				} else {
					ying.Status.Conditions.IngressNotReadyPools = ying.Status.Conditions.IngressNotReadyPools[:i-1]
				}
			} else {
				ying.Status.Conditions.IngressNotReadyPools[i] = ying.Status.Conditions.IngressNotReadyPools[i+1]
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
	if !ingressCRChanged {
		deployments, err := r.getAllDeployments(ying)
		if err != nil {
			klog.V(4).Infof("Get all the ingress controller deployments err: %v", err)
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
				ying.Status.Conditions.IngressReadyPools = append(ying.Status.Conditions.IngressReadyPools, pool)
			} else {
				klog.V(4).Infof("Ingress on pool %s is NOT ready!", pool)
				condition := getUnreadyDeploymentCondition(dply)
				if condition == nil {
					klog.V(4).Infof("Get deployment/%s conditions nil!", dply.GetName())
				} else {
					notReadyPool := appsv1alpha1.IngressNotReadyPool{Name: pool, Info: condition}
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
	pools := getDesiredPoolNames(instance)
	if pools != nil {
		for _, pool := range pools {
			if err := yurtapputil.DeleteNginxIngressSpecificResource(r.Client, pool, true); err != nil {
				return ctrl.Result{}, err
			}
		}
		if err := yurtapputil.DeleteNginxIngressCommonResource(r.Client); err != nil {
			return ctrl.Result{}, err
		}
	}
	if controllerutil.ContainsFinalizer(instance, appsv1alpha1.YurtIngressFinalizer) {
		controllerutil.RemoveFinalizer(instance, appsv1alpha1.YurtIngressFinalizer)
		if err := r.Update(context.TODO(), instance); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func prepareDeploymentOwnerReferences(instance *appsv1alpha1.YurtIngress) *metav1.OwnerReference {
	isController := true
	isBlockOwnerDeletion := true
	ownerRef := metav1.OwnerReference{
		//TODO: optimze the APIVersion/Kind with instance values
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
