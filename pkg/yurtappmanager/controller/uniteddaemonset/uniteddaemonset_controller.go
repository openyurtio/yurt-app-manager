package uniteddaemonset

import (
	"context"
	"flag"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	unitv1alpha1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/util/gate"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	flag.IntVar(&concurrentReconciles, "uniteddaemonset-workers", concurrentReconciles, "Max concurrent workers for UnitedDaemonSet controller.")
}

var (
	concurrentReconciles = 3
)

const (
	controllerName              = "uniteddaemonset-controller"
	eventTypeRevisionProvision  = "RevisionProvision"
	eventTypeTemplateController = "TemplateController"
)

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
	err = c.Watch(&source.Kind{Type: &unitv1alpha1.NodePool{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &unitv1alpha1.UnitedDaemonSet{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &unitv1alpha1.UnitedDaemonSet{},
	})
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
	controls map[unitv1alpha1.TemplateType]TemplateControllor
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUnitedDaemonSet{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		recorder: mgr.GetEventRecorderFor(controllerName),
		controls: map[unitv1alpha1.TemplateType]TemplateControllor{
			//			unitv1alpha1.StatefulSetTemplateType: &StatefulSetControllor{Client: mgr.GetClient(), scheme: mgr.GetScheme()},
			unitv1alpha1.DeploymentTemplateType: &DeploymentControllor{Client: mgr.GetClient(), scheme: mgr.GetScheme()},
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
	/*
		oldStatus := instance.Status.DeepCopy()
		currentRevision, updatedRevision, collisionCount, err := yurtctlutil.ConstructUnitedRevisions(r.Client, r.scheme, &yurtctlutil.UnitedDaemonSetRevision{Object: instance})
		if err != nil {
			klog.Errorf("Fail to construct controller revision of UnitedDaemonSet %s/%s: %s", instance.Namespace, instance.Name, err)
			r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed%s", eventTypeRevisionProvision), err.Error())
			return reconcile.Result{}, err
		}

	*/

	control, templateType, err := r.getTemplateControls(instance)
	if err != nil {
		r.recorder.Event(instance.DeepCopy(), corev1.EventTypeWarning, fmt.Sprintf("Failed%s", eventTypeTemplateController), err.Error())
		return reconcile.Result{}, err
	}

	klog.V(4).Infof("Get UnitedDaemonSet %s/%s type %s all workload", request.Namespace, request.Name, templateType)

	currentNPToWorkload, err := r.getNodePoolToWorkLoad(instance, control)
	if err != nil {
		klog.Errorf("Fail to get nodePoolWorkload for UnitedDaemonSet %s/%s: %s", instance.Namespace, instance.Name, err)
		return reconcile.Result{}, nil
	}

	nodepoolSelector, err := metav1.LabelSelectorAsSelector(instance.Spec.NodePoolSelector)
	if err != nil {
		return reconcile.Result{}, err
	}

	validNodePools := unitv1alpha1.NodePoolList{}
	if err := r.Client.List(context.TODO(), &validNodePools, &client.ListOptions{LabelSelector: nodepoolSelector}); err != nil {
		klog.Errorf("Fail to get NodePoolList")
		return reconcile.Result{}, nil
	}

	nameToNodePools := make(map[string]unitv1alpha1.NodePool)
	for _, v := range validNodePools.Items {
		nameToNodePools[v.GetName()] = v
	}

	var needDeleted, needUpdate []*Workload
	for npName, load := range currentNPToWorkload {
		find := false
		for vnp, np := range nameToNodePools {
			if vnp == npName {
				find = true
				match := true
				// 判断label  和 taint
				npSelector, err := metav1.LabelSelectorAsSelector(np.Spec.Selector)
				if err != nil {
					klog.Errorf("Create nodepool[%s]  selector error, UnitedDaemonset[%s/%s]",
						npName, instance.GetNamespace(), instance.GetName())
					break
				}

				if !npSelector.Matches(labels.Set(load.Status.NodeSelector)) {
					match = false
				}

				for i, _ := range np.Spec.Taints {
					if !helper.TolerationsTolerateTaint(load.Status.Toleration, &np.Spec.Taints[i]) {
						match = false
						break
					}
				}

				if !match {
					needUpdate = append(needUpdate, load)
				}

				break
			}
		}
		if !find {
			needDeleted = append(needDeleted, load)
		}
	}

	var needCreate []unitv1alpha1.NodePool
	for vnp, vpool := range nameToNodePools {
		find := false
		for np, _ := range currentNPToWorkload {
			if vnp == np {
				find = true
				break
			}
		}
		if !find {
			needCreate = append(needCreate, vpool)
		}
	}

	// 针对于Create 的 需要创建
	// 获得UnitedDaemonset 的
	for _, c := range needCreate {
		klog.Infof("UnitedDaemonset[%s/%s] create workload %s ", instance.GetNamespace(), instance.GetName(), c.Name)
		/*
			if err := control.CreateWorkload(instance, c); err != nil {
				klog.Errorf("UnitedDaemonset[%s/%s] create workload error %v", instance.GetNamespace(), instance.GetName(),
					err)
			}
		*/
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileUnitedDaemonSet) getTemplateControls(instance *unitv1alpha1.UnitedDaemonSet) (TemplateControllor,
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

func (r *ReconcileUnitedDaemonSet) getNodePoolToWorkLoad(instance *unitv1alpha1.UnitedDaemonSet, c TemplateControllor) (map[string]*Workload, error) {
	nodePoolsToWorkloads := make(map[string]*Workload)
	workloads, err := c.GetAllWorkloads(instance)
	if err != nil {
		klog.Errorf("Get all workloads for UnitedDaemonSet[%s/%s] error %v", instance.GetNamespace(),
			instance.GetName(), err)
		return nil, err
	}

	// 获得workload 里对应的NodePool
	for i, w := range workloads {
		np, ok := w.Spec.Ref.GetAnnotations()[unitv1alpha1.AnnotationRefNodePool]
		if ok {
			nodePoolsToWorkloads[np] = workloads[i]
		}
		// TODO need consider no annotation
		// nodePoolsToWorkloads[""] = w
	}
	return nodePoolsToWorkloads, nil
}
