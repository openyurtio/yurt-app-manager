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

package uniteddaemonset

import (
	"context"
	"flag"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/controller/uniteddaemonset/workloadcontroller"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/gate"
)

var (
	concurrentReconciles = 3
)

const (
	controllerName            = "uniteddaemonset-controller"
	slowStartInitialBatchSize = 1

	eventTypeRevisionProvision  = "RevisionProvision"
	eventTypeTemplateController = "TemplateController"

	eventTypeWorkloadsCreated = "CreateWorkload"
	eventTypeWorkloadsUpdated = "UpdateWorkload"
	eventTypeWorkloadsDeleted = "DeleteWorkload"
)

func init() {
	flag.IntVar(&concurrentReconciles, "uniteddaemonset-workers", concurrentReconciles, "Max concurrent workers for UnitedDaemonSet controller.")
}

// Add creates a new UnitedDaemonSet Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, _ context.Context) error {
	if !gate.ResourceEnabled(&unitv1alpha1.UnitedDaemonSet{}) {
		return nil
	}
	return add(mgr, newReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r, MaxConcurrentReconciles: concurrentReconciles})
	if err != nil {
		return err
	}

	// Watch for changes to UnitedDaemonSet
	err = c.Watch(&source.Kind{Type: &unitv1alpha1.UnitedDaemonSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to NodePool
	err = c.Watch(&source.Kind{Type: &unitv1alpha1.NodePool{}}, &EnqueueUnitedDaemonsetForNodePool{client: mgr.GetClient()})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileUnitedDaemonSet{}

// ReconcileUnitedDaemonSet reconciles a UnitedDaemonSet object
type ReconcileUnitedDaemonSet struct {
	client.Client
	scheme *runtime.Scheme

	recorder record.EventRecorder
	controls map[unitv1alpha1.TemplateType]workloadcontroller.WorkloadControllor
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUnitedDaemonSet{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		recorder: mgr.GetEventRecorderFor(controllerName),
		controls: map[unitv1alpha1.TemplateType]workloadcontroller.WorkloadControllor{
			//			unitv1alpha1.StatefulSetTemplateType: &StatefulSetControllor{Client: mgr.GetClient(), scheme: mgr.GetScheme()},
			unitv1alpha1.DeploymentTemplateType: &workloadcontroller.DeploymentControllor{Client: mgr.GetClient(), Scheme: mgr.GetScheme()},
		},
	}
}

// +kubebuilder:rbac:groups=apps.openyurt.io,resources=uniteddaemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.openyurt.io,resources=uniteddaemonsets/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a UnitedDaemonSet object and makes changes based on the state read
// and what is in the UnitedDaemonSet.Spec
func (r *ReconcileUnitedDaemonSet) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	klog.V(4).Infof("Reconcile UnitedDaemonSet %s/%s", request.Namespace, request.Name)
	// Fetch the UnitedDaemonSet instance
	instance := &unitv1alpha1.UnitedDaemonSet{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	oldStatus := instance.Status.DeepCopy()

	currentRevision, updatedRevision, collisionCount, err := r.constructUnitedDaemonSetRevisions(instance)
	if err != nil {
		klog.Errorf("Fail to construct controller revision of UnitedDaemonSet %s/%s: %s", instance.Namespace, instance.Name, err)
		r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed%s", eventTypeRevisionProvision), err.Error())
		return reconcile.Result{}, err
	}

	expectedRevision := currentRevision
	if updatedRevision != nil {
		expectedRevision = updatedRevision
	}

	klog.Infof("UnitedDaemonSet [%s/%s] get expectRevision %v collisionCount %v", instance.GetNamespace(), instance.GetName(),
		expectedRevision.Name, collisionCount)

	control, templateType, err := r.getTemplateControls(instance)
	if err != nil {
		r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed%s", eventTypeTemplateController), err.Error())
		return reconcile.Result{}, err
	}

	currentNPToWorkload, err := r.getNodePoolToWorkLoad(instance, control)
	if err != nil {
		klog.Errorf("UnitedDaemonSet [%s/%s] Fail to get nodePoolWorkload, error: %s", instance.Namespace, instance.Name, err)
		return reconcile.Result{}, nil
	}

	allNameToNodePools, err := r.getNameToNodePools(instance)
	if err != nil {
		klog.Errorf("UnitedDaemonSet [%s/%s] Fail to get nameToNodePools, error: %s", instance.Namespace, instance.Name, err)
		return reconcile.Result{}, nil
	}

	newStatus, err := r.manageWorkloads(instance, currentNPToWorkload, allNameToNodePools, expectedRevision.Name, templateType)
	if err != nil {
		return reconcile.Result{}, nil
	}

	return r.updateStatus(instance, newStatus, oldStatus, currentRevision, collisionCount, templateType)
}

func (r *ReconcileUnitedDaemonSet) updateStatus(instance *unitv1alpha1.UnitedDaemonSet, newStatus, oldStatus *unitv1alpha1.UnitedDaemonSetStatus,
	currentRevision *appsv1.ControllerRevision, collisionCount int32, templateType unitv1alpha1.TemplateType) (reconcile.Result, error) {

	newStatus = r.calculateStatus(instance, newStatus, currentRevision, collisionCount, templateType)
	_, err := r.updateUnitedDaemonSet(instance, oldStatus, newStatus)

	return reconcile.Result{}, err
}

func (r *ReconcileUnitedDaemonSet) updateUnitedDaemonSet(udd *unitv1alpha1.UnitedDaemonSet, oldStatus, newStatus *unitv1alpha1.UnitedDaemonSetStatus) (*unitv1alpha1.UnitedDaemonSet, error) {
	if oldStatus.CurrentRevision == newStatus.CurrentRevision &&
		*oldStatus.CollisionCount == *newStatus.CollisionCount &&
		oldStatus.TemplateType == newStatus.TemplateType &&
		udd.Generation == newStatus.ObservedGeneration &&
		reflect.DeepEqual(oldStatus.NodePools, newStatus.NodePools) &&
		reflect.DeepEqual(oldStatus.Conditions, newStatus.Conditions) {
		klog.Infof("UnitedDaemonSet[%s/%s] oldStatus==newStatus, no need to update status", udd.GetNamespace(), udd.GetName())
		return udd, nil
	}

	newStatus.ObservedGeneration = udd.Generation

	var getErr, updateErr error
	for i, obj := 0, udd; ; i++ {
		klog.V(4).Infof(fmt.Sprintf("UnitedDaemonSet[%s/%s] The %d th time updating status for %v[%s/%s], ",
			udd.GetNamespace(), udd.GetName(), i, obj.Kind, obj.Namespace, obj.Name) +
			fmt.Sprintf("sequence No: %v->%v", obj.Status.ObservedGeneration, newStatus.ObservedGeneration))

		obj.Status = *newStatus

		updateErr = r.Client.Status().Update(context.TODO(), obj)
		if updateErr == nil {
			return obj, nil
		}
		if i >= updateRetries {
			break
		}
		tmpObj := &unitv1alpha1.UnitedDaemonSet{}
		if getErr = r.Client.Get(context.TODO(), client.ObjectKey{Namespace: obj.Namespace, Name: obj.Name}, tmpObj); getErr != nil {
			return nil, getErr
		}
		obj = tmpObj
	}

	klog.Errorf("fail to update UnitedDaemonSet %s/%s status: %s", udd.Namespace, udd.Name, updateErr)
	return nil, updateErr
}

func (r *ReconcileUnitedDaemonSet) calculateStatus(instance *unitv1alpha1.UnitedDaemonSet, newStatus *unitv1alpha1.UnitedDaemonSetStatus,
	currentRevision *appsv1.ControllerRevision, collisionCount int32, templateType unitv1alpha1.TemplateType) *unitv1alpha1.UnitedDaemonSetStatus {

	newStatus.CollisionCount = &collisionCount

	if newStatus.CurrentRevision == "" {
		// init with current revision
		newStatus.CurrentRevision = currentRevision.Name
	}

	newStatus.TemplateType = templateType

	return newStatus
}

func (r *ReconcileUnitedDaemonSet) manageWorkloads(instance *unitv1alpha1.UnitedDaemonSet, currentNodepoolToWorkload map[string]*workloadcontroller.Workload,
	allNameToNodePools map[string]unitv1alpha1.NodePool, expectedRevision string, templateType unitv1alpha1.TemplateType) (newStatus *unitv1alpha1.UnitedDaemonSetStatus, updateErr error) {

	newStatus = instance.Status.DeepCopy()

	nps := make([]string, 0, len(allNameToNodePools))
	for np, _ := range allNameToNodePools {
		nps = append(nps, np)
	}
	newStatus.NodePools = nps

	needDeleted, needUpdate, needCreate := r.classifyWorkloads(instance, currentNodepoolToWorkload, allNameToNodePools, expectedRevision)
	provision, err := r.manageWorkloadsProvision(instance, allNameToNodePools, expectedRevision, templateType, needDeleted, needCreate)
	if err != nil {
		SetUnitedDaemonSetCondition(newStatus, NewUnitedDaemonSetCondition(unitv1alpha1.WorkLoadProvisioned, corev1.ConditionFalse, "Error", err.Error()))
		return newStatus, fmt.Errorf("fail to manage workload provision: %v", err)
	}

	if provision {
		SetUnitedDaemonSetCondition(newStatus, NewUnitedDaemonSetCondition(unitv1alpha1.WorkLoadProvisioned, corev1.ConditionTrue, "", ""))
	}

	if len(needUpdate) > 0 {
		_, updateErr = util.SlowStartBatch(len(needUpdate), slowStartInitialBatchSize, func(index int) error {
			u := needUpdate[index]
			updateWorkloadErr := r.controls[templateType].UpdateWorkload(u, instance, allNameToNodePools[u.GetNodePoolName()], expectedRevision)
			if updateWorkloadErr != nil {
				r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed %s", eventTypeWorkloadsUpdated),
					fmt.Sprintf("Error updating workload type(%s) %s when updating: %s", templateType, u.Name, updateWorkloadErr))
				klog.Errorf("UnitedDaemonset[%s/%s] update workload[%s/%s/%s] error %v", instance.GetNamespace(), instance.GetName(),
					templateType, u.Namespace, u.Name, err)
			}
			return updateWorkloadErr
		})
	}

	if updateErr == nil {
		SetUnitedDaemonSetCondition(newStatus, NewUnitedDaemonSetCondition(unitv1alpha1.WorkLoadUpdated, corev1.ConditionTrue, "", ""))
	} else {
		SetUnitedDaemonSetCondition(newStatus, NewUnitedDaemonSetCondition(unitv1alpha1.WorkLoadUpdated, corev1.ConditionFalse, "Error", updateErr.Error()))
	}

	return newStatus, updateErr
}

func (r *ReconcileUnitedDaemonSet) manageWorkloadsProvision(instance *unitv1alpha1.UnitedDaemonSet,
	allNameToNodePools map[string]unitv1alpha1.NodePool, expectedRevision string, templateType unitv1alpha1.TemplateType,
	needDeleted []*workloadcontroller.Workload, needCreate []string) (bool, error) {
	// 针对于Create 的 需要创建

	var errs []error
	if len(needCreate) > 0 {
		// do not consider deletion
		var createdNum int
		var createdErr error
		createdNum, createdErr = util.SlowStartBatch(len(needCreate), slowStartInitialBatchSize, func(idx int) error {
			nodepoolName := needCreate[idx]
			err := r.controls[templateType].CreateWorkload(instance, allNameToNodePools[nodepoolName], expectedRevision)
			//err := r.poolControls[workloadType].CreatePool(ud, poolName, revision, replicas)
			if err != nil {
				klog.Errorf("UnitedDaemonset[%s/%s] templatetype %s create workload by nodepool %s error: %s",
					instance.GetNamespace(), instance.GetName(), templateType, nodepoolName, err.Error())
				if !errors.IsTimeout(err) {
					return fmt.Errorf("UnitedDaemonset[%s/%s] templatetype %s create workload by nodepool %s error: %s",
						instance.GetNamespace(), instance.GetName(), templateType, nodepoolName, err.Error())
				}
			}
			klog.Infof("UnitedDaemonset[%s/%s] templatetype %s create workload by nodepool %s success",
				instance.GetNamespace(), instance.GetName(), templateType, nodepoolName)
			return nil
		})
		if createdErr == nil {
			r.recorder.Eventf(instance.DeepCopy(), corev1.EventTypeNormal, fmt.Sprintf("Successful %s", eventTypeWorkloadsCreated), "Create %d Workload type(%s)", createdNum, templateType)
		} else {
			errs = append(errs, createdErr)
		}
	}

	// manage deleting
	if len(needDeleted) > 0 {
		var deleteErrs []error
		// need deleted
		for _, d := range needDeleted {
			if err := r.controls[templateType].DeleteWorkload(instance, d); err != nil {
				deleteErrs = append(deleteErrs, fmt.Errorf("UnitedDaemonset[%s/%s] delete workload[%s/%s/%s] error %v",
					instance.GetNamespace(), instance.GetName(), templateType, d.Namespace, d.Name, err))
			}
		}
		if len(deleteErrs) > 0 {
			errs = append(errs, deleteErrs...)
		} else {
			r.recorder.Eventf(instance.DeepCopy(), corev1.EventTypeNormal, fmt.Sprintf("Successful %s", eventTypeWorkloadsDeleted), "Delete %d Workload type(%s)", len(needDeleted), templateType)
		}
	}

	return len(needCreate) > 0 || len(needDeleted) > 0, utilerrors.NewAggregate(errs)
}

func (r *ReconcileUnitedDaemonSet) classifyWorkloads(instance *unitv1alpha1.UnitedDaemonSet, currentNodepoolToWorkload map[string]*workloadcontroller.Workload,
	allNameToNodePools map[string]unitv1alpha1.NodePool, expectedRevision string) (needDeleted, needUpdate []*workloadcontroller.Workload,
	needCreate []string) {

	for npName, load := range currentNodepoolToWorkload {
		find := false
		for vnp, np := range allNameToNodePools {
			if vnp == npName {
				find = true
				match := true
				// judge workload NodeSelector
				if !reflect.DeepEqual(load.GetNodeSelector(), workloadcontroller.CreateNodeSelectorByNodepoolName(npName)) {
					match = false
				}
				// judge workload whether toleration all taints
				match = IsTolerationsAllTaints(load.GetToleration(), np.Spec.Taints)

				// judge revision
				if load.GetRevision() != expectedRevision {
					match = false
				}

				if !match {
					klog.V(4).Infof("UnitedDaemonSet[%s/%s] need update [%s/%s/%s]", instance.GetNamespace(),
						instance.GetName(), load.GetKind(), load.Namespace, load.Name)
					needUpdate = append(needUpdate, load)
				}

				break
			}
		}
		if !find {
			needDeleted = append(needDeleted, load)
			klog.V(4).Infof("UnitedDaemonSet[%s/%s] need delete [%s/%s/%s]", instance.GetNamespace(),
				instance.GetName(), load.GetKind(), load.Namespace, load.Name)
		}
	}

	for vnp, _ := range allNameToNodePools {
		find := false
		for np, _ := range currentNodepoolToWorkload {
			if vnp == np {
				find = true
				break
			}
		}
		if !find {
			needCreate = append(needCreate, vnp)
			klog.V(4).Infof("UnitedDaemonSet[%s/%s] need create new workload by nodepool %s", instance.GetNamespace(),
				instance.GetName(), vnp)
		}
	}

	return
}

func (r *ReconcileUnitedDaemonSet) getNameToNodePools(instance *unitv1alpha1.UnitedDaemonSet) (map[string]unitv1alpha1.NodePool, error) {
	klog.V(4).Infof("UnitedDaemonSet [%s/%s] prepare to get associated nodepools",
		instance.Namespace, instance.Name)

	nodepoolSelector, err := metav1.LabelSelectorAsSelector(instance.Spec.NodePoolSelector)
	if err != nil {
		return nil, err
	}

	nodepools := unitv1alpha1.NodePoolList{}
	if err := r.Client.List(context.TODO(), &nodepools, &client.ListOptions{LabelSelector: nodepoolSelector}); err != nil {
		klog.Errorf("UnitedDaemonSet [%s/%s] Fail to get NodePoolList", instance.GetNamespace(),
			instance.GetName())
		return nil, nil
	}

	indexs := make(map[string]unitv1alpha1.NodePool)
	for i, v := range nodepools.Items {
		indexs[v.GetName()] = v
		klog.V(4).Infof("UnitedDaemonSet [%s/%s] get %d's associated nodepools %s",
			instance.Namespace, instance.Name, i, v.Name)

	}

	return indexs, nil
}

func (r *ReconcileUnitedDaemonSet) getTemplateControls(instance *unitv1alpha1.UnitedDaemonSet) (workloadcontroller.WorkloadControllor,
	unitv1alpha1.TemplateType, error) {
	switch {
	case instance.Spec.WorkloadTemplate.StatefulSetTemplate != nil:
		return r.controls[unitv1alpha1.StatefulSetTemplateType], unitv1alpha1.StatefulSetTemplateType, nil
	case instance.Spec.WorkloadTemplate.DeploymentTemplate != nil:
		return r.controls[unitv1alpha1.DeploymentTemplateType], unitv1alpha1.DeploymentTemplateType, nil
	default:
		klog.Errorf("The appropriate WorkloadTemplate was not found")
		return nil, "", fmt.Errorf("The appropriate WorkloadTemplate was not found, Now Support(%s/%s)",
			unitv1alpha1.StatefulSetTemplateType, unitv1alpha1.DeploymentTemplateType)
	}
}

func (r *ReconcileUnitedDaemonSet) getNodePoolToWorkLoad(instance *unitv1alpha1.UnitedDaemonSet, c workloadcontroller.WorkloadControllor) (map[string]*workloadcontroller.Workload, error) {
	klog.V(4).Infof("UnitedDaemonSet [%s/%s/%s] prepare to get all workload", c.GetTemplateType(), instance.Namespace, instance.Name)

	nodePoolsToWorkloads := make(map[string]*workloadcontroller.Workload)
	workloads, err := c.GetAllWorkloads(instance)
	if err != nil {
		klog.Errorf("Get all workloads for UnitedDaemonSet[%s/%s] error %v", instance.GetNamespace(),
			instance.GetName(), err)
		return nil, err
	}
	// 获得workload 里对应的NodePool
	for i, w := range workloads {
		if w.GetNodePoolName() != "" {
			nodePoolsToWorkloads[w.GetNodePoolName()] = workloads[i]
			klog.V(4).Infof("UnitedDaemonSet [%s/%s] get %d's workload[%s/%s/%s]",
				instance.Namespace, instance.Name, i, c.GetTemplateType(), w.Namespace, w.Name)
		} else {
			klog.Warningf("UnitedDaemonSet [%s/%s] %d's workload[%s/%s/%s] has no nodepool annotation",
				instance.Namespace, instance.Name, i, c.GetTemplateType(), w.Namespace, w.Name)
		}
	}
	klog.V(4).Infof("UnitedDaemonSet [%s/%s] get %d %s workloads",
		instance.Namespace, instance.Name, len(nodePoolsToWorkloads), c.GetTemplateType())
	return nodePoolsToWorkloads, nil
}

func (r *ReconcileUnitedDaemonSet) getOwnedServices(instance *unitv1alpha1.UnitedDaemonSet) (*corev1.ServiceList, error) {
	svcList := &corev1.ServiceList{}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			unitv1alpha1.LabelCurrentYurtAppDaemon: instance.GetName(),
		},
	}
	// 获得YurtAppDaemon 对应的 所有的service
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, err
	}
	// List all Service to include
	if err := r.Client.List(context.TODO(), svcList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, err
	}
	return svcList, nil
}

func (r *ReconcileUnitedDaemonSet) updateOwnedServices(instance *unitv1alpha1.UnitedDaemonSet) error {
	currentownedServices, err := r.getOwnedServices(instance)
	if err != nil {
		klog.Errorf("UnitedDaemonSet [%s/%s] Fail to get currentOwnedServices, error: %s", instance.Namespace, instance.Name, err)
		return err
	}

}
