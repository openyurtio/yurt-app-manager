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
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	appsv1alpha2 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha2"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1beta1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/constant"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/gate"
)

const (
	controllerName = "yurtingress-controller"
)

// YurtIngressReconciler reconciles a YurtIngress object
type YurtIngressReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	recorder  record.EventRecorder
	namespace string
}

// Add creates a new YurtIngress Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and start it when the Manager is started.
func Add(mgr manager.Manager, ctx context.Context) error {
	if !gate.ResourceEnabled(&appsv1alpha2.YurtIngress{}) {
		return nil
	}

	wn := ctx.Value(constant.ContextKeyWorkloadNamespace)
	wnStr, ok := wn.(string)
	if !ok {
		return errors.NewBadRequest("workload namespace convert fail")
	}
	if err := (&YurtIngressReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		recorder:  mgr.GetEventRecorderFor(controllerName),
		namespace: wnStr,
	}).SetupWithManager(mgr); err != nil {
		klog.Error(err, "unable to create controller", "controller", "NodeSLO")
		os.Exit(1)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *YurtIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha2.YurtIngress{}).
		Watches(&source.Kind{Type: &v1beta1.NodePool{}}, &EnqueueRequestForNodePool{}).
		Complete(r)
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
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=*
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=*
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *YurtIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Reconcile YurtIngress: %s", req.Name)

	nodePool, yurtIngress := &v1beta1.NodePool{}, &appsv1alpha2.YurtIngress{}
	nodePoolExist, yurtIngressExist := true, true
	nodePoolName, yurtIngressName := req.Name, req.Name

	if err := r.Client.Get(context.TODO(), req.NamespacedName, nodePool); err != nil {
		if !errors.IsNotFound(err) {
			klog.Errorf("failed to find nodepool %v, error: %v", nodePoolName, err)
			return ctrl.Result{Requeue: true}, err
		}
		nodePoolExist = false
	}

	if err := r.Client.Get(context.TODO(), req.NamespacedName, yurtIngress); err != nil {
		if !errors.IsNotFound(err) {
			klog.Errorf("failed to find yurtingress %v, error: %v", yurtIngressName, err)
			return ctrl.Result{Requeue: true}, err
		}
		yurtIngressExist = false
	}

	if yurtIngressExist && (!yurtIngress.GetDeletionTimestamp().IsZero() || !yurtIngress.Spec.Enabled) {
		if err := r.cleanupResources(ctx, yurtIngress); err != nil {
			return ctrl.Result{}, err
		}
		klog.Infof("delete or disable yurtingress %s", yurtIngress.Name)
		return ctrl.Result{}, nil
	}

	if !nodePoolExist && !yurtIngressExist {
		return ctrl.Result{}, nil
	} else if !nodePoolExist {
		if err := r.Client.Delete(context.TODO(), yurtIngress); err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			klog.Errorf("failed to delete yurtingress %v, error: %v", yurtIngressName, err)
			return ctrl.Result{Requeue: true}, err
		}
	} else if !yurtIngressExist {
		yurtIngress = &appsv1alpha2.YurtIngress{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodePoolName,
			},
			Spec: appsv1alpha2.YurtIngressSpec{
				Enabled: nodePool.Spec.Type == v1beta1.Edge,
			},
		}
		if err := r.Client.Create(context.TODO(), yurtIngress); err != nil {
			klog.Errorf("failed to create yurtIngress %v, error: %v", yurtIngress.Name, err)
			return ctrl.Result{Requeue: true}, err
		}
	} else {
		klog.Infof("nodepool and yurtingress all exists, %s", nodePool.Name)
	}

	if !controllerutil.ContainsFinalizer(yurtIngress, appsv1alpha1.YurtIngressFinalizer) {
		patchYi := yurtIngress.DeepCopy()
		controllerutil.AddFinalizer(patchYi, appsv1alpha1.YurtIngressFinalizer)
		if err := r.Patch(context.TODO(), patchYi, client.MergeFrom(yurtIngress)); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	if err := r.reconcileSvc(ctx, yurtIngress); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileRelease(ctx, yurtIngress, nodePool); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
